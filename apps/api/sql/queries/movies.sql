-- ============================================================
-- PERSONS QUERIES
-- ============================================================

-- name: UpsertPerson :one
-- Insert a person or update their name if the slug already exists
INSERT INTO persons (name, slug)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- ============================================================
-- MOVIES QUERIES
-- ============================================================

-- name: CreateMovie :one
-- Create a new movie record with external ratings
INSERT INTO movies (
    title, slug, overview, poster_url, release_date, runtime, 
    imdb_id, tmdb_id, imdb_rating, rotten_tomatoes, metacritic_score
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: ListMovies :many
-- Get a paginated list of movies ordered by creation date
SELECT * FROM movies
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetMovieBySlug :one
-- Get a single movie by its unique slug
SELECT * FROM movies
WHERE slug = $1 LIMIT 1;

-- name: SearchMovies :many
-- Simple search by title or slug
SELECT * FROM movies
WHERE title ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%'
ORDER BY imdb_rating DESC
LIMIT $2 OFFSET $3;

-- name: AdvancedSearchMovies :many
-- Advanced search with filters: query, genres, year range, rating range, sort
SELECT DISTINCT m.*
FROM movies m
LEFT JOIN movie_genres mg ON mg.movie_id = m.id
WHERE
    -- Text search (optional)
    (sqlc.narg(query)::TEXT IS NULL OR m.title ILIKE '%' || sqlc.narg(query)::TEXT || '%')
    -- Genre filter (optional, array of genre IDs)
    AND (sqlc.narg(genre_ids)::INT[] IS NULL OR mg.genre_id = ANY(sqlc.narg(genre_ids)::INT[]))
    -- Year range (optional)
    AND (sqlc.narg(year_from)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) >= sqlc.narg(year_from)::INT)
    AND (sqlc.narg(year_to)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) <= sqlc.narg(year_to)::INT)
    -- IMDB rating range (optional, scale 0-10)
    AND (sqlc.narg(imdb_min)::NUMERIC IS NULL OR m.imdb_rating >= sqlc.narg(imdb_min)::NUMERIC)
    AND (sqlc.narg(imdb_max)::NUMERIC IS NULL OR m.imdb_rating <= sqlc.narg(imdb_max)::NUMERIC)
    -- User rating range (optional, scale 1-10)
    AND (sqlc.narg(user_rating_min)::REAL IS NULL OR m.user_avg_rating >= sqlc.narg(user_rating_min)::REAL)
    AND (sqlc.narg(user_rating_max)::REAL IS NULL OR m.user_avg_rating <= sqlc.narg(user_rating_max)::REAL)
    -- Rotten Tomatoes range (optional, scale 0-100)
    AND (sqlc.narg(rt_min)::INT IS NULL OR m.rotten_tomatoes >= sqlc.narg(rt_min)::INT)
    AND (sqlc.narg(rt_max)::INT IS NULL OR m.rotten_tomatoes <= sqlc.narg(rt_max)::INT)
    -- Metacritic range (optional, scale 0-100)
    AND (sqlc.narg(metacritic_min)::INT IS NULL OR m.metacritic_score >= sqlc.narg(metacritic_min)::INT)
    AND (sqlc.narg(metacritic_max)::INT IS NULL OR m.metacritic_score <= sqlc.narg(metacritic_max)::INT)
ORDER BY
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'title_asc' THEN m.title END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'title_desc' THEN m.title END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'release_date_asc' THEN m.release_date END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'release_date_desc' THEN m.release_date END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'imdb_rating_asc' THEN m.imdb_rating END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'imdb_rating_desc' THEN m.imdb_rating END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'user_rating_asc' THEN m.user_avg_rating END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'user_rating_desc' THEN m.user_avg_rating END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'rotten_tomatoes_asc' THEN m.rotten_tomatoes END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'rotten_tomatoes_desc' THEN m.rotten_tomatoes END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'metacritic_asc' THEN m.metacritic_score END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'metacritic_desc' THEN m.metacritic_score END DESC,
    m.imdb_rating DESC NULLS LAST,
    m.release_date DESC NULLS LAST
LIMIT $1 OFFSET $2;

-- name: CountAdvancedSearchMovies :one
-- Count total results for advanced search (for pagination)
SELECT COUNT(DISTINCT m.id)::INT as total
FROM movies m
LEFT JOIN movie_genres mg ON mg.movie_id = m.id
WHERE
    (sqlc.narg(query)::TEXT IS NULL OR m.title ILIKE '%' || sqlc.narg(query)::TEXT || '%')
    AND (sqlc.narg(genre_ids)::INT[] IS NULL OR mg.genre_id = ANY(sqlc.narg(genre_ids)::INT[]))
    AND (sqlc.narg(year_from)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) >= sqlc.narg(year_from)::INT)
    AND (sqlc.narg(year_to)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) <= sqlc.narg(year_to)::INT)
    AND (sqlc.narg(imdb_min)::NUMERIC IS NULL OR m.imdb_rating >= sqlc.narg(imdb_min)::NUMERIC)
    AND (sqlc.narg(imdb_max)::NUMERIC IS NULL OR m.imdb_rating <= sqlc.narg(imdb_max)::NUMERIC)
    AND (sqlc.narg(user_rating_min)::REAL IS NULL OR m.user_avg_rating >= sqlc.narg(user_rating_min)::REAL)
    AND (sqlc.narg(user_rating_max)::REAL IS NULL OR m.user_avg_rating <= sqlc.narg(user_rating_max)::REAL)
    AND (sqlc.narg(rt_min)::INT IS NULL OR m.rotten_tomatoes >= sqlc.narg(rt_min)::INT)
    AND (sqlc.narg(rt_max)::INT IS NULL OR m.rotten_tomatoes <= sqlc.narg(rt_max)::INT)
    AND (sqlc.narg(metacritic_min)::INT IS NULL OR m.metacritic_score >= sqlc.narg(metacritic_min)::INT)
    AND (sqlc.narg(metacritic_max)::INT IS NULL OR m.metacritic_score <= sqlc.narg(metacritic_max)::INT);

-- ============================================================
-- CREDITS QUERIES
-- ============================================================

-- name: CreateCredit :exec
-- Create a link between a movie and a person (cast/crew)
INSERT INTO credits (movie_id, person_id, department, role, character)
VALUES ($1, $2, $3, $4, $5);