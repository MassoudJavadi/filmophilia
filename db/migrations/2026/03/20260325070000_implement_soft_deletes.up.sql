-- ============================================================
-- Implement Soft Deletes for User Accounts
--
-- Strategy:
-- 1. Add soft delete infrastructure (deleted_at, anonymized_at)
-- 2. Preserve public content (comments, ratings) when users delete
-- 3. Cascade delete private data (sessions, notifications, activities)
-- ============================================================

-- ============================================================
-- PHASE 1: Add soft delete infrastructure
-- ============================================================

-- Add soft delete timestamp columns
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN anonymized_at TIMESTAMPTZ;

-- Add DELETED status to user_status enum
ALTER TYPE user_status ADD VALUE IF NOT EXISTS 'DELETED';

-- Create index for efficient filtering of active users
CREATE INDEX users_deleted_at_idx ON users (deleted_at) WHERE deleted_at IS NOT NULL;

-- Create index for active users (most common query)
CREATE INDEX users_active_idx ON users (id) WHERE deleted_at IS NULL;

-- ============================================================
-- PHASE 2: Make comments.user_id nullable (preserve content)
-- ============================================================

-- Allow comments to exist without a user (for deleted accounts)
ALTER TABLE comments ALTER COLUMN user_id DROP NOT NULL;

-- Drop existing foreign key constraint
ALTER TABLE comments DROP CONSTRAINT comments_user_id_fkey;

-- Recreate with SET NULL to preserve comments when user deletes
ALTER TABLE comments ADD CONSTRAINT comments_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Add index for querying comments by existing users only
CREATE INDEX comments_user_id_active_idx ON comments (user_id)
    WHERE user_id IS NOT NULL;

-- ============================================================
-- PHASE 3: Make ratings.user_id nullable (preserve ratings)
-- ============================================================

-- Allow ratings to exist without a user (for deleted accounts)
ALTER TABLE ratings ALTER COLUMN user_id DROP NOT NULL;

-- Drop existing foreign key constraint
ALTER TABLE ratings DROP CONSTRAINT ratings_user_id_fkey;

-- Recreate with SET NULL to preserve ratings when user deletes
ALTER TABLE ratings ADD CONSTRAINT ratings_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Handle unique constraint: one rating per user per movie
-- Drop the old unique constraint
ALTER TABLE ratings DROP CONSTRAINT ratings_user_id_movie_id_key;

-- Create partial unique index that only applies when user_id is NOT NULL
-- This allows multiple NULL user_id entries (from deleted users)
-- but maintains uniqueness for active users
CREATE UNIQUE INDEX ratings_user_movie_unique_idx
    ON ratings (user_id, movie_id)
    WHERE user_id IS NOT NULL;

-- ============================================================
-- PHASE 4: Add helpful constraints
-- ============================================================

-- Ensure anonymized_at is only set after deleted_at
ALTER TABLE users ADD CONSTRAINT users_anonymization_after_deletion_check
    CHECK (anonymized_at IS NULL OR deleted_at IS NOT NULL);

-- When user is marked DELETED, deleted_at should be set
-- Note: This is enforced at application level, not as DB constraint
-- since the status might be set before deleted_at in a transaction

-- ============================================================
-- NOTES FOR APPLICATION LAYER
-- ============================================================
--
-- 1. User Deletion Flow:
--    UPDATE users SET
--        status = 'DELETED',
--        deleted_at = NOW()
--    WHERE id = ?;
--
-- 2. Query Active Users:
--    WHERE deleted_at IS NULL
--    -- or: WHERE status != 'DELETED'
--
-- 3. Display Deleted User Content:
--    Comments with NULL user_id → show as "Deleted User"
--    Ratings with NULL user_id → count in analytics but no attribution
--
-- 4. Anonymization (after 30 days):
--    UPDATE users SET
--        email = 'deleted_' || id || '@filmophilia.local',
--        username = 'deleted_user_' || id,
--        display_name = NULL,
--        avatar_url = NULL,
--        bio = NULL,
--        password_hash = NULL,
--        is_verified = FALSE,
--        anonymized_at = NOW()
--    WHERE deleted_at < NOW() - INTERVAL '30 days'
--      AND anonymized_at IS NULL;
--
--    -- Then cascade delete auth/private data:
--    DELETE FROM sessions WHERE user_id IN (...);
--    DELETE FROM accounts WHERE user_id IN (...);
--
-- 5. Hard Delete (GDPR compliance only):
--    DELETE FROM users WHERE id = ?;
--    -- This will SET NULL on comments/ratings (preserved)
--    -- And CASCADE delete remaining private data
-- ============================================================
