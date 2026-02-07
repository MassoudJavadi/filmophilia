-- Rollback changes
ALTER TABLE movies DROP COLUMN IF EXISTS imdb_rating;
ALTER TABLE movies DROP COLUMN IF EXISTS rotten_tomatoes;
ALTER TABLE movies DROP COLUMN IF EXISTS metacritic_score;
ALTER TABLE movies DROP COLUMN IF EXISTS letterboxd_rating;

ALTER TABLE movies RENAME COLUMN user_avg_rating TO average_rating;
ALTER TABLE movies RENAME COLUMN user_rating_count TO rating_count;
