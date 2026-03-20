-- name: UpsertRating :one
INSERT INTO ratings (user_id, movie_id, score)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, movie_id)
DO UPDATE SET score = EXCLUDED.score, updated_at = NOW()
RETURNING *;

-- name: GetUserRatingForMovie :one
SELECT * FROM ratings
WHERE user_id = $1 AND movie_id = $2;

-- name: GetMovieRatings :many
SELECT
    r.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM ratings r
JOIN users u ON u.id = r.user_id
WHERE r.movie_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserRatings :many
SELECT
    r.*,
    m.title,
    m.slug,
    m.poster_url
FROM ratings r
JOIN movies m ON m.id = r.movie_id
WHERE r.user_id = $1
ORDER BY r.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteRating :exec
DELETE FROM ratings
WHERE user_id = $1 AND movie_id = $2;

-- name: UpdateMovieRatingStats :exec
UPDATE movies
SET
    user_avg_rating = (SELECT AVG(r.score)::REAL FROM ratings r WHERE r.movie_id = $1),
    user_rating_count = (SELECT COUNT(*)::INT FROM ratings r WHERE r.movie_id = $1),
    updated_at = NOW()
WHERE id = $1;

-- name: GetMovieByID :one
SELECT * FROM movies WHERE id = $1;