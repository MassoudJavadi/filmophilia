-- name: AddToWatchlist :one
INSERT INTO watchlists (user_id, movie_id, notes, rank_position)
VALUES ($1, $2, $3, COALESCE($4, (SELECT COALESCE(MAX(rank_position), 0) + 1 FROM watchlists WHERE user_id = $1)))
RETURNING *;

-- name: GetWatchlistEntry :one
SELECT * FROM watchlists
WHERE user_id = $1 AND movie_id = $2;

-- name: GetUserWatchlist :many
SELECT
    w.*,
    m.title,
    m.slug,
    m.poster_url,
    m.release_date,
    m.runtime,
    m.user_avg_rating
FROM watchlists w
JOIN movies m ON m.id = w.movie_id
WHERE w.user_id = $1
  AND ($2::bool IS FALSE OR w.watched_at IS NULL)
  AND ($3::bool IS FALSE OR w.watched_at IS NOT NULL)
ORDER BY w.rank_position ASC
LIMIT $4 OFFSET $5;

-- name: CountUserWatchlist :one
SELECT
    COUNT(*) FILTER (WHERE watched_at IS NULL)::INT as unwatched,
    COUNT(*) FILTER (WHERE watched_at IS NOT NULL)::INT as watched,
    COUNT(*)::INT as total
FROM watchlists
WHERE user_id = $1;

-- name: UpdateWatchlistEntry :one
UPDATE watchlists
SET
    notes = COALESCE($3, notes),
    rank_position = COALESCE($4, rank_position),
    updated_at = NOW()
WHERE user_id = $1 AND movie_id = $2
RETURNING *;

-- name: MarkAsWatched :one
UPDATE watchlists
SET watched_at = NOW()
WHERE user_id = $1 AND movie_id = $2
RETURNING *;

-- name: MarkAsUnwatched :one
UPDATE watchlists
SET watched_at = NULL
WHERE user_id = $1 AND movie_id = $2
RETURNING *;

-- name: RemoveFromWatchlist :exec
DELETE FROM watchlists
WHERE user_id = $1 AND movie_id = $2;

-- name: IsInWatchlist :one
SELECT EXISTS(
    SELECT 1 FROM watchlists
    WHERE user_id = $1 AND movie_id = $2
) as in_watchlist;

-- name: ReorderWatchlistItem :exec
UPDATE watchlists
SET rank_position = $3
WHERE user_id = $1 AND movie_id = $2;
