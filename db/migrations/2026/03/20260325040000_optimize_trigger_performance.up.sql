-- Add rating_sum column to track sum for efficient average calculation
ALTER TABLE movies ADD COLUMN rating_sum BIGINT DEFAULT 0;

-- Populate rating_sum with current values
UPDATE movies
SET rating_sum = COALESCE((SELECT SUM(score) FROM ratings WHERE movie_id = movies.id), 0);

-- Drop old triggers
DROP TRIGGER IF EXISTS reactions_like_count_delete ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_insert ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_update ON reactions;
DROP TRIGGER IF EXISTS ratings_stats_delete ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_insert ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_update ON ratings;

-- Optimized update_like_counts function using incremental arithmetic
CREATE OR REPLACE FUNCTION update_like_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Decrement like_count if the deleted reaction was a LIKE
        IF OLD.type = 'LIKE' THEN
            UPDATE comments
            SET like_count = GREATEST(like_count - 1, 0)
            WHERE id = OLD.comment_id;
        END IF;
        RETURN OLD;
    ELSIF TG_OP = 'INSERT' THEN
        -- Increment like_count if the new reaction is a LIKE
        IF NEW.type = 'LIKE' THEN
            UPDATE comments
            SET like_count = like_count + 1
            WHERE id = NEW.comment_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle reaction type changes (e.g., LIKE -> LOVE or LOVE -> LIKE)
        IF OLD.type = 'LIKE' AND NEW.type != 'LIKE' THEN
            -- Changed from LIKE to something else: decrement
            UPDATE comments
            SET like_count = GREATEST(like_count - 1, 0)
            WHERE id = OLD.comment_id;
        ELSIF OLD.type != 'LIKE' AND NEW.type = 'LIKE' THEN
            -- Changed from something else to LIKE: increment
            UPDATE comments
            SET like_count = like_count + 1
            WHERE id = NEW.comment_id;
        END IF;

        -- Handle comment_id changes (rare but possible)
        IF OLD.comment_id IS DISTINCT FROM NEW.comment_id THEN
            -- Remove from old comment
            IF OLD.type = 'LIKE' THEN
                UPDATE comments
                SET like_count = GREATEST(like_count - 1, 0)
                WHERE id = OLD.comment_id;
            END IF;
            -- Add to new comment
            IF NEW.type = 'LIKE' THEN
                UPDATE comments
                SET like_count = like_count + 1
                WHERE id = NEW.comment_id;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
END;
$$;

-- Optimized update_movie_rating_stats function using incremental arithmetic
CREATE OR REPLACE FUNCTION update_movie_rating_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Decrement count and subtract score from sum
        UPDATE movies
        SET
            user_rating_count = GREATEST(user_rating_count - 1, 0),
            rating_sum = GREATEST(rating_sum - OLD.score, 0),
            user_avg_rating = CASE
                WHEN user_rating_count - 1 > 0
                THEN (rating_sum - OLD.score)::REAL / (user_rating_count - 1)
                ELSE 0
            END
        WHERE id = OLD.movie_id;
        RETURN OLD;
    ELSIF TG_OP = 'INSERT' THEN
        -- Increment count and add score to sum
        UPDATE movies
        SET
            user_rating_count = user_rating_count + 1,
            rating_sum = rating_sum + NEW.score,
            user_avg_rating = (rating_sum + NEW.score)::REAL / (user_rating_count + 1)
        WHERE id = NEW.movie_id;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle score changes
        IF OLD.score != NEW.score THEN
            UPDATE movies
            SET
                rating_sum = rating_sum - OLD.score + NEW.score,
                user_avg_rating = CASE
                    WHEN user_rating_count > 0
                    THEN (rating_sum - OLD.score + NEW.score)::REAL / user_rating_count
                    ELSE 0
                END
            WHERE id = NEW.movie_id;
        END IF;

        -- Handle movie_id changes (rare but possible due to unique constraint)
        IF OLD.movie_id != NEW.movie_id THEN
            -- Remove from old movie
            UPDATE movies
            SET
                user_rating_count = GREATEST(user_rating_count - 1, 0),
                rating_sum = GREATEST(rating_sum - OLD.score, 0),
                user_avg_rating = CASE
                    WHEN user_rating_count - 1 > 0
                    THEN (rating_sum - OLD.score)::REAL / (user_rating_count - 1)
                    ELSE 0
                END
            WHERE id = OLD.movie_id;

            -- Add to new movie
            UPDATE movies
            SET
                user_rating_count = user_rating_count + 1,
                rating_sum = rating_sum + NEW.score,
                user_avg_rating = (rating_sum + NEW.score)::REAL / (user_rating_count + 1)
            WHERE id = NEW.movie_id;
        END IF;
        RETURN NEW;
    END IF;
END;
$$;

-- Recreate triggers with optimized functions
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
