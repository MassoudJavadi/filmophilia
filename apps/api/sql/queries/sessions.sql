-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: GetSessionByRefreshToken :one
SELECT * FROM sessions WHERE refresh_token = $1 LIMIT 1;

-- name: UpdateSession :exec
UPDATE sessions 
SET refresh_token = $2, expires_at = $3 
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();


-- name: DeleteSessionByRefreshToken :exec
DELETE FROM sessions
WHERE refresh_token = $1;