-- ============================================================
-- Rollback: Remove denormalized user counts
-- ============================================================

-- Drop indexes
DROP INDEX IF EXISTS users_id_status_idx;
DROP INDEX IF EXISTS movies_release_year_idx;

-- Drop comment triggers
DROP TRIGGER IF EXISTS comments_user_count_delete ON comments;
DROP TRIGGER IF EXISTS comments_user_count_update ON comments;
DROP TRIGGER IF EXISTS comments_user_count_insert ON comments;
DROP FUNCTION IF EXISTS update_user_comment_count();

-- Drop rating triggers
DROP TRIGGER IF EXISTS ratings_user_count_delete ON ratings;
DROP TRIGGER IF EXISTS ratings_user_count_insert ON ratings;
DROP FUNCTION IF EXISTS update_user_rating_count();

-- Drop follow triggers
DROP TRIGGER IF EXISTS follows_user_counts_delete ON follows;
DROP TRIGGER IF EXISTS follows_user_counts_insert ON follows;
DROP FUNCTION IF EXISTS update_user_follow_counts();

-- Remove columns from users
ALTER TABLE users
    DROP COLUMN IF EXISTS follower_count,
    DROP COLUMN IF EXISTS following_count,
    DROP COLUMN IF EXISTS rating_count,
    DROP COLUMN IF EXISTS comment_count;
