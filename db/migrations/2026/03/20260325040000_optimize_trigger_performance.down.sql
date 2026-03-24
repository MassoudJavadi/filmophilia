-- Drop optimized triggers
DROP TRIGGER IF EXISTS reactions_like_count_delete ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_insert ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_update ON reactions;
DROP TRIGGER IF EXISTS ratings_stats_delete ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_insert ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_update ON ratings;

-- Restore original update_like_counts function (using COUNT)
CREATE OR REPLACE FUNCTION update_like_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = OLD.comment_id AND type = 'LIKE'
        ) WHERE id = OLD.comment_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle reaction type changes (e.g., LIKE -> LOVE)
        -- Update old comment if comment_id changed (unlikely but safe)
        IF OLD.comment_id IS DISTINCT FROM NEW.comment_id THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions
                WHERE comment_id = OLD.comment_id AND type = 'LIKE'
            ) WHERE id = OLD.comment_id;
        END IF;
        -- Update new/current comment
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = NEW.comment_id AND type = 'LIKE'
        ) WHERE id = NEW.comment_id;
        RETURN NEW;
    ELSE
        -- INSERT
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = NEW.comment_id AND type = 'LIKE'
        ) WHERE id = NEW.comment_id;
        RETURN NEW;
    END IF;
END;
$$;

-- Restore original update_movie_rating_stats function (using AVG and COUNT)
CREATE OR REPLACE FUNCTION update_movie_rating_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE movies SET
            user_avg_rating = COALESCE((SELECT AVG(score)::REAL FROM ratings WHERE movie_id = OLD.movie_id), 0),
            user_rating_count = (SELECT COUNT(*) FROM ratings WHERE movie_id = OLD.movie_id)
        WHERE id = OLD.movie_id;
        RETURN OLD;
    ELSE
        UPDATE movies SET
            user_avg_rating = COALESCE((SELECT AVG(score)::REAL FROM ratings WHERE movie_id = NEW.movie_id), 0),
            user_rating_count = (SELECT COUNT(*) FROM ratings WHERE movie_id = NEW.movie_id)
        WHERE id = NEW.movie_id;
        RETURN NEW;
    END IF;
END;
$$;

-- Recreate original triggers
CREATE TRIGGER reactions_like_count_delete
    AFTER DELETE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_insert
    AFTER INSERT ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_update
    AFTER UPDATE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER ratings_stats_delete
    AFTER DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_insert
    AFTER INSERT ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_update
    AFTER UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

-- Remove rating_sum column
ALTER TABLE movies DROP COLUMN IF EXISTS rating_sum;
