-- ============================================================
-- Rollback: Drop entire initial schema
-- ============================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS users_status_change ON users;
DROP TRIGGER IF EXISTS reactions_like_count_delete ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_insert ON reactions;
DROP TRIGGER IF EXISTS ratings_stats_delete ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_update ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_insert ON ratings;
DROP TRIGGER IF EXISTS comments_updated_at ON comments;
DROP TRIGGER IF EXISTS reviews_updated_at ON reviews;
DROP TRIGGER IF EXISTS ratings_updated_at ON ratings;
DROP TRIGGER IF EXISTS persons_updated_at ON persons;
DROP TRIGGER IF EXISTS movies_updated_at ON movies;
DROP TRIGGER IF EXISTS users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS log_user_status_change();
DROP FUNCTION IF EXISTS update_like_counts();
DROP FUNCTION IF EXISTS update_movie_rating_stats();
DROP FUNCTION IF EXISTS update_updated_at();

-- Drop tables (reverse order of creation for FK dependencies)
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS reactions;
DROP TABLE IF EXISTS watchlists;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS credits;
DROP TABLE IF EXISTS movie_genres;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS movies;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_status_logs;
DROP TABLE IF EXISTS users;

-- Drop enums
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS entity_type;
DROP TYPE IF EXISTS department;
DROP TYPE IF EXISTS reaction_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS role;

-- Drop extensions
DROP EXTENSION IF EXISTS pg_trgm;
