-- ============================================================
-- Denormalize User Profile Counts
-- Adds cached count columns to users table with triggers
-- Also adds functional index for movie year search
-- ============================================================

-- ============================================================
-- 1. Add count columns to users table
-- ============================================================

ALTER TABLE users
    ADD COLUMN follower_count INT NOT NULL DEFAULT 0,
    ADD COLUMN following_count INT NOT NULL DEFAULT 0,
    ADD COLUMN rating_count INT NOT NULL DEFAULT 0,
    ADD COLUMN comment_count INT NOT NULL DEFAULT 0;

-- ============================================================
-- 2. Backfill existing counts
-- ============================================================

UPDATE users u SET
    follower_count = (SELECT COUNT(*) FROM follows WHERE following_id = u.id),
    following_count = (SELECT COUNT(*) FROM follows WHERE follower_id = u.id),
    rating_count = (SELECT COUNT(*) FROM ratings WHERE user_id = u.id),
    comment_count = (SELECT COUNT(*) FROM comments WHERE user_id = u.id AND deleted_at IS NULL);

-- ============================================================
-- 3. Create triggers for follows (follower_count, following_count)
-- ============================================================

CREATE OR REPLACE FUNCTION update_user_follow_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment follower_count for the user being followed
        UPDATE users SET follower_count = follower_count + 1
        WHERE id = NEW.following_id;
        -- Increment following_count for the user who follows
        UPDATE users SET following_count = following_count + 1
        WHERE id = NEW.follower_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement follower_count for the user being unfollowed
        UPDATE users SET follower_count = GREATEST(0, follower_count - 1)
        WHERE id = OLD.following_id;
        -- Decrement following_count for the user who unfollows
        UPDATE users SET following_count = GREATEST(0, following_count - 1)
        WHERE id = OLD.follower_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER follows_user_counts_insert
    AFTER INSERT ON follows
    FOR EACH ROW EXECUTE FUNCTION update_user_follow_counts();

CREATE TRIGGER follows_user_counts_delete
    AFTER DELETE ON follows
    FOR EACH ROW EXECUTE FUNCTION update_user_follow_counts();

-- ============================================================
-- 4. Create triggers for ratings (rating_count)
-- ============================================================

CREATE OR REPLACE FUNCTION update_user_rating_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET rating_count = rating_count + 1
        WHERE id = NEW.user_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET rating_count = GREATEST(0, rating_count - 1)
        WHERE id = OLD.user_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ratings_user_count_insert
    AFTER INSERT ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_user_rating_count();

CREATE TRIGGER ratings_user_count_delete
    AFTER DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_user_rating_count();

-- ============================================================
-- 5. Create triggers for comments (comment_count)
-- Handles both hard delete and soft delete (deleted_at)
-- ============================================================

CREATE OR REPLACE FUNCTION update_user_comment_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Only count if not soft-deleted
        IF NEW.deleted_at IS NULL THEN
            UPDATE users SET comment_count = comment_count + 1
            WHERE id = NEW.user_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle soft delete transition
        IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
            -- Comment was soft-deleted
            UPDATE users SET comment_count = GREATEST(0, comment_count - 1)
            WHERE id = NEW.user_id;
        ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
            -- Comment was restored
            UPDATE users SET comment_count = comment_count + 1
            WHERE id = NEW.user_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Only decrement if wasn't already soft-deleted
        IF OLD.deleted_at IS NULL THEN
            UPDATE users SET comment_count = GREATEST(0, comment_count - 1)
            WHERE id = OLD.user_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comments_user_count_insert
    AFTER INSERT ON comments
    FOR EACH ROW EXECUTE FUNCTION update_user_comment_count();

CREATE TRIGGER comments_user_count_update
    AFTER UPDATE OF deleted_at ON comments
    FOR EACH ROW EXECUTE FUNCTION update_user_comment_count();

CREATE TRIGGER comments_user_count_delete
    AFTER DELETE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_user_comment_count();

-- ============================================================
-- 6. Add functional index for movie year search
-- ============================================================

CREATE INDEX movies_release_year_idx ON movies (EXTRACT(YEAR FROM release_date));

-- ============================================================
-- 7. Add index hints for common user profile queries
-- ============================================================

CREATE INDEX users_id_status_idx ON users (id) WHERE status = 'ACTIVE';
