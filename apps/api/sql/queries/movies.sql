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

-- ============================================================
-- CREDITS QUERIES
-- ============================================================

-- name: CreateCredit :exec
-- Create a link between a movie and a person (cast/crew)
INSERT INTO credits (movie_id, person_id, department, role, character)
VALUES ($1, $2, $3, $4, $5);