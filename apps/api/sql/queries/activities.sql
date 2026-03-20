-- name: CreateActivity :one
-- Create a new activity entry
INSERT INTO activities (user_id, action, entity_type, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserActivities :many
-- Get activities for a specific user (user's own feed)
SELECT a.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM activities a
JOIN users u ON u.id = a.user_id
WHERE a.user_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFollowingActivities :many
-- Get activities from users that current user follows (social feed)
SELECT a.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM activities a
JOIN users u ON u.id = a.user_id
JOIN follows f ON f.following_id = a.user_id
WHERE f.follower_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActivitiesByEntity :many
-- Get all activities for a specific entity (e.g., all activities for a movie)
SELECT a.*,
    u.username,
    u.display_name,
    u.avatar_url
FROM activities a
JOIN users u ON u.id = a.user_id
WHERE a.entity_type = $1 AND a.entity_id = $2
ORDER BY a.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountUserActivities :one
-- Count total activities for a user
SELECT COUNT(*)::INT as count
FROM activities
WHERE user_id = $1;

-- name: CountFollowingActivities :one
-- Count total activities from users that current user follows
SELECT COUNT(*)::INT as count
FROM activities a
JOIN follows f ON f.following_id = a.user_id
WHERE f.follower_id = $1;

-- name: DeleteUserActivities :exec
-- Delete all activities for a user (cleanup)
DELETE FROM activities
WHERE user_id = $1;

-- name: DeleteActivityByEntity :exec
-- Delete activities related to a specific entity (when entity is deleted)
DELETE FROM activities
WHERE entity_type = $1 AND entity_id = $2;
