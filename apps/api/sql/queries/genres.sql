-- name: GetAllGenres :many
SELECT * FROM genres
ORDER BY name;

-- name: GetGenreByID :one
SELECT * FROM genres WHERE id = $1;

-- name: GetGenreBySlug :one
SELECT * FROM genres WHERE slug = $1;

-- name: GetMoviesByGenre :many
SELECT
    m.*
FROM movies m
JOIN movie_genres mg ON mg.movie_id = m.id
WHERE mg.genre_id = $1
ORDER BY m.release_date DESC NULLS LAST
LIMIT $2 OFFSET $3;

-- name: CountMoviesByGenre :one
SELECT COUNT(*)::INT as count
FROM movie_genres
WHERE genre_id = $1;

-- name: GetGenresForMovie :many
SELECT g.*
FROM genres g
JOIN movie_genres mg ON mg.genre_id = g.id
WHERE mg.movie_id = $1
ORDER BY g.name;

-- name: GetMoviesByGenreSlug :many
SELECT
    m.*
FROM movies m
JOIN movie_genres mg ON mg.movie_id = m.id
JOIN genres g ON g.id = mg.genre_id
WHERE g.slug = $1
ORDER BY m.release_date DESC NULLS LAST
LIMIT $2 OFFSET $3;
