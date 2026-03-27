ALTER TABLE movies
ALTER COLUMN user_avg_rating SET DEFAULT 0;

UPDATE movies
SET user_avg_rating = 0
WHERE user_avg_rating IS NULL;

CREATE OR REPLACE FUNCTION update_movie_rating_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
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
        UPDATE movies
        SET
            user_rating_count = user_rating_count + 1,
            rating_sum = rating_sum + NEW.score,
            user_avg_rating = (rating_sum + NEW.score)::REAL / (user_rating_count + 1)
        WHERE id = NEW.movie_id;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
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

        IF OLD.movie_id != NEW.movie_id THEN
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
