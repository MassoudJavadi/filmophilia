-- Rename existing columns to be more specific
ALTER TABLE movies RENAME COLUMN average_rating TO user_avg_rating;
ALTER TABLE movies RENAME COLUMN rating_count TO user_rating_count;

-- Add new external rating columns
ALTER TABLE movies ADD COLUMN imdb_rating DECIMAL(3, 1);
ALTER TABLE movies ADD COLUMN rotten_tomatoes INT;
ALTER TABLE movies ADD COLUMN metacritic_score INT;
ALTER TABLE movies ADD COLUMN letterboxd_rating DECIMAL(3, 1);
