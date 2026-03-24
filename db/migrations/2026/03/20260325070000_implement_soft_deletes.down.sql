-- ============================================================
-- Rollback: Remove Soft Delete Infrastructure
--
-- WARNING: This migration will create a placeholder user for
-- orphaned comments/ratings if any exist.
-- ============================================================

-- ============================================================
-- PHASE 1: Handle orphaned content (comments/ratings with NULL user_id)
-- ============================================================

DO $$
DECLARE
    placeholder_user_id INTEGER;
    orphaned_comments_exist BOOLEAN;
    orphaned_ratings_exist BOOLEAN;
BEGIN
    -- Check if there are any orphaned comments or ratings
    SELECT EXISTS(SELECT 1 FROM comments WHERE user_id IS NULL) INTO orphaned_comments_exist;
    SELECT EXISTS(SELECT 1 FROM ratings WHERE user_id IS NULL) INTO orphaned_ratings_exist;

    -- If orphaned content exists, create a placeholder user
    IF orphaned_comments_exist OR orphaned_ratings_exist THEN
        -- Create placeholder user account for orphaned content
        INSERT INTO users (
            email,
            username,
            display_name,
            status,
            is_verified,
            created_at,
            updated_at
        ) VALUES (
            'system_deleted_user@filmophilia.local',
            'deleted_user_system',
            'Deleted User',
            'BANNED',
            FALSE,
            NOW(),
            NOW()
        ) RETURNING id INTO placeholder_user_id;

        -- Reassign orphaned comments to placeholder user
        UPDATE comments
        SET user_id = placeholder_user_id
        WHERE user_id IS NULL;

        -- Reassign orphaned ratings to placeholder user
        UPDATE ratings
        SET user_id = placeholder_user_id
        WHERE user_id IS NULL;

        RAISE NOTICE 'Created placeholder user (id: %) for orphaned content', placeholder_user_id;
    END IF;
END $$;

-- ============================================================
-- PHASE 2: Restore ratings constraints
-- ============================================================

-- Drop partial unique index
DROP INDEX IF EXISTS ratings_user_movie_unique_idx;

-- Drop SET NULL foreign key
ALTER TABLE ratings DROP CONSTRAINT ratings_user_id_fkey;

-- Restore user_id as NOT NULL
ALTER TABLE ratings ALTER COLUMN user_id SET NOT NULL;

-- Recreate original unique constraint
ALTER TABLE ratings ADD CONSTRAINT ratings_user_id_movie_id_key
    UNIQUE (user_id, movie_id);

-- Recreate original CASCADE foreign key
ALTER TABLE ratings ADD CONSTRAINT ratings_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================================
-- PHASE 3: Restore comments constraints
-- ============================================================

-- Drop active user index
DROP INDEX IF EXISTS comments_user_id_active_idx;

-- Drop SET NULL foreign key
ALTER TABLE comments DROP CONSTRAINT comments_user_id_fkey;

-- Restore user_id as NOT NULL
ALTER TABLE comments ALTER COLUMN user_id SET NOT NULL;

-- Recreate original CASCADE foreign key
ALTER TABLE comments ADD CONSTRAINT comments_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ============================================================
-- PHASE 4: Remove soft delete infrastructure
-- ============================================================

-- Drop indexes
DROP INDEX IF EXISTS users_active_idx;
DROP INDEX IF EXISTS users_deleted_at_idx;

-- Drop constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_anonymization_after_deletion_check;

-- Drop columns
ALTER TABLE users DROP COLUMN IF EXISTS anonymized_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;

-- Note: Cannot easily remove enum value 'DELETED' from user_status
-- It will remain in the enum but unused (this is safe)

-- ============================================================
-- ROLLBACK COMPLETE
-- ============================================================
