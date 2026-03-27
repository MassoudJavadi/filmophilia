-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $2 WHERE id = $1;

-- name: GetUserProfile :one
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.bio,
    u.created_at,
    u.follower_count,
    u.following_count,
    u.rating_count,
    u.comment_count
FROM users u
WHERE u.id = $1 AND u.status = 'ACTIVE';

-- name: GetUserProfileByUsername :one
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.bio,
    u.created_at,
    u.follower_count,
    u.following_count,
    u.rating_count,
    u.comment_count
FROM users u
WHERE u.username = $1 AND u.status = 'ACTIVE';

-- name: UpdateUserProfile :one
UPDATE users
SET
    display_name = COALESCE($2, display_name),
    avatar_url = COALESCE($3, avatar_url),
    bio = COALESCE($4, bio),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SearchUsers :many
SELECT
    id,
    username,
    display_name,
    avatar_url
FROM users
WHERE status = 'ACTIVE'
  AND (username ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%')
ORDER BY
    CASE WHEN username ILIKE $1 || '%' THEN 0 ELSE 1 END,
    username
LIMIT $2 OFFSET $3;
