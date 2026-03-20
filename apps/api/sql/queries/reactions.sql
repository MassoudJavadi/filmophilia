-- name: AddReactionToComment :one
INSERT INTO reactions (user_id, comment_id, type)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, comment_id)
DO UPDATE SET type = EXCLUDED.type
RETURNING *;

-- name: GetUserReactionOnComment :one
SELECT * FROM reactions
WHERE user_id = $1 AND comment_id = $2;

-- name: RemoveReactionFromComment :exec
DELETE FROM reactions
WHERE user_id = $1 AND comment_id = $2;

-- name: GetCommentReactions :many
SELECT
    r.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM reactions r
JOIN users u ON u.id = r.user_id
WHERE r.comment_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCommentReactionsByType :many
SELECT type, COUNT(*)::INT as count
FROM reactions
WHERE comment_id = $1
GROUP BY type;
