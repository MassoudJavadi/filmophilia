-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
VALUES ($1, 'sha256:' || encode(digest($2, 'sha256'), 'hex'), $3, $4, $5)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: GetSessionByRefreshToken :one
SELECT * FROM sessions
WHERE refresh_token = 'sha256:' || encode(digest($1, 'sha256'), 'hex')
LIMIT 1;

-- name: UpdateSession :exec
UPDATE sessions 
SET refresh_token = 'sha256:' || encode(digest($2, 'sha256'), 'hex'), expires_at = $3 
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();


-- name: DeleteSessionByRefreshToken :exec
DELETE FROM sessions
WHERE refresh_token = 'sha256:' || encode(digest($1, 'sha256'), 'hex');
