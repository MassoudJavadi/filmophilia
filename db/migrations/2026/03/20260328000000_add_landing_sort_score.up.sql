-- Add pre-computed landing sort score for performant pagination
ALTER TABLE movies ADD COLUMN landing_sort_score NUMERIC;

-- Create index for fast sorting on landing page
CREATE INDEX movies_landing_sort_idx ON movies (landing_sort_score DESC NULLS LAST, created_at DESC);

-- Function to calculate the landing sort score
CREATE OR REPLACE FUNCTION calculate_landing_sort_score(
    p_user_avg_rating REAL,
    p_imdb_rating NUMERIC,
    p_rotten_tomatoes INT,
    p_metacritic_score INT
) RETURNS NUMERIC
    LANGUAGE plpgsql IMMUTABLE
AS $$
DECLARE
    community_rating NUMERIC;
    rating_count INT;
BEGIN
    -- Count how many community ratings exist
    rating_count := (CASE WHEN p_imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN p_rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN p_metacritic_score IS NOT NULL THEN 1 ELSE 0 END);

    -- Calculate community rating average (normalized to 0-10 scale)
    IF rating_count > 0 THEN
        community_rating := (
            COALESCE(p_imdb_rating, 0) +
            COALESCE(p_rotten_tomatoes::NUMERIC / 10.0, 0) +
            COALESCE(p_metacritic_score::NUMERIC / 10.0, 0)
        ) / rating_count;
    END IF;

    -- Return weighted score based on what's available
    IF p_user_avg_rating IS NOT NULL AND rating_count > 0 THEN
        -- 90% user rating + 10% community rating
        RETURN (p_user_avg_rating::NUMERIC * 0.9) + (community_rating * 0.1);
    ELSIF p_user_avg_rating IS NOT NULL THEN
        RETURN p_user_avg_rating::NUMERIC;
    ELSIF rating_count > 0 THEN
        RETURN community_rating;
    ELSE
        RETURN 0;
    END IF;
END;
$$;

-- Backfill existing movies
UPDATE movies
SET landing_sort_score = calculate_landing_sort_score(
    user_avg_rating,
    imdb_rating,
    rotten_tomatoes,
    metacritic_score
);

-- Trigger function to update landing_sort_score when ratings change
CREATE OR REPLACE FUNCTION update_movie_landing_sort_score() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    NEW.landing_sort_score := calculate_landing_sort_score(
        NEW.user_avg_rating,
        NEW.imdb_rating,
        NEW.rotten_tomatoes,
        NEW.metacritic_score
    );
    RETURN NEW;
END;
$$;

-- Trigger on movies table for direct rating updates
CREATE TRIGGER trg_movies_landing_sort_score
    BEFORE INSERT OR UPDATE OF user_avg_rating, imdb_rating, rotten_tomatoes, metacritic_score
    ON movies
    FOR EACH ROW
    EXECUTE FUNCTION update_movie_landing_sort_score();
