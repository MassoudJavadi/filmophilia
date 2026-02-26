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

-- ============================================================
-- CREDITS QUERIES
-- ============================================================

-- name: CreateCredit :exec
-- Create a link between a movie and a person (cast/crew)
INSERT INTO credits (movie_id, person_id, department, role, character)
VALUES ($1, $2, $3, $4, $5);