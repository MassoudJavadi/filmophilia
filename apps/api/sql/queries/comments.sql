-- name: CreateComment :one
INSERT INTO comments (user_id, movie_id, parent_id, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCommentWithUser :one
SELECT
    c.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: GetMovieComments :many
SELECT
    c.*,
    u.username,
    u.display_name,
    u.avatar_url,
    (SELECT COUNT(*) FROM comments r WHERE r.parent_id = c.id AND r.deleted_at IS NULL)::INT as reply_count
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.movie_id = $1
  AND c.parent_id IS NULL
  AND c.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetCommentReplies :many
SELECT
    c.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.parent_id = $1 AND c.deleted_at IS NULL
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;

-- name: GetUserComments :many
SELECT
    c.*,
    m.title as movie_title,
    m.slug as movie_slug
FROM comments c
JOIN movies m ON m.id = c.movie_id
WHERE c.user_id = $1 AND c.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
UPDATE comments
SET content = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteComment :exec
UPDATE comments
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountMovieComments :one
SELECT COUNT(*)::INT as count
FROM comments
WHERE movie_id = $1 AND parent_id IS NULL AND deleted_at IS NULL;

-- name: CountCommentReplies :one
SELECT COUNT(*)::INT as count
FROM comments
WHERE parent_id = $1 AND deleted_at IS NULL;
