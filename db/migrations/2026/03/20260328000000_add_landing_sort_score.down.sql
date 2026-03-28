DROP TRIGGER IF EXISTS trg_movies_landing_sort_score ON movies;
DROP FUNCTION IF EXISTS update_movie_landing_sort_score();
DROP FUNCTION IF EXISTS calculate_landing_sort_score(REAL, NUMERIC, INT, INT);
DROP INDEX IF EXISTS movies_landing_sort_idx;
ALTER TABLE movies DROP COLUMN IF EXISTS landing_sort_score;
