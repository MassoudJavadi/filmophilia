-- ============================================================
-- DROP REVIEWS AND CLEANUP DEPENDENCIES
-- ============================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS reviews_updated_at ON reviews;

-- Drop indexes on reactions that reference reviews
DROP INDEX IF EXISTS reactions_user_review_unique;
DROP INDEX IF EXISTS reactions_review_id_idx;

-- Remove the CHECK constraint and review_id column from reactions
ALTER TABLE reactions DROP CONSTRAINT IF EXISTS reactions_check;
ALTER TABLE reactions DROP COLUMN IF EXISTS review_id;

-- Add new constraint (comment_id must be NOT NULL now)
ALTER TABLE reactions ALTER COLUMN comment_id SET NOT NULL;

-- Drop the reviews table
DROP TABLE IF EXISTS reviews;

-- Update entity_type enum to remove REVIEW
ALTER TYPE entity_type RENAME TO entity_type_old;
CREATE TYPE entity_type AS ENUM ('MOVIE', 'COMMENT', 'USER');
ALTER TABLE activities ALTER COLUMN entity_type TYPE entity_type USING entity_type::text::entity_type;
DROP TYPE entity_type_old;

-- Update notification_type enum to remove NEW_REVIEW
ALTER TYPE notification_type RENAME TO notification_type_old;
CREATE TYPE notification_type AS ENUM (
    'NEW_FOLLOWER', 'NEW_LIKE', 'NEW_COMMENT',
    'ACCOUNT_ACTIVATED', 'ACCOUNT_SUSPENDED', 'ACCOUNT_BANNED', 'SYSTEM_ALERT'
);
ALTER TABLE notifications ALTER COLUMN type TYPE notification_type USING type::text::notification_type;
DROP TYPE notification_type_old;
