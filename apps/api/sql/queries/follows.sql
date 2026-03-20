-- name: FollowUser :one
INSERT INTO follows (follower_id, following_id)
VALUES ($1, $2)
RETURNING *;

-- name: UnfollowUser :exec
DELETE FROM follows
WHERE follower_id = $1 AND following_id = $2;

-- name: IsFollowing :one
SELECT EXISTS(
    SELECT 1 FROM follows
    WHERE follower_id = $1 AND following_id = $2
) as is_following;

-- name: GetFollowers :many
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.bio,
    f.created_at as followed_at
FROM follows f
JOIN users u ON u.id = f.follower_id
WHERE f.following_id = $1 AND u.status = 'ACTIVE'
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetFollowing :many
SELECT
    u.id,
    u.username,
    u.display_name,
    u.avatar_url,
    u.bio,
    f.created_at as followed_at
FROM follows f
JOIN users u ON u.id = f.following_id
WHERE f.follower_id = $1 AND u.status = 'ACTIVE'
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountFollowers :one
SELECT COUNT(*)::INT as count
FROM follows f
JOIN users u ON u.id = f.follower_id
WHERE f.following_id = $1 AND u.status = 'ACTIVE';

-- name: CountFollowing :one
SELECT COUNT(*)::INT as count
FROM follows f
JOIN users u ON u.id = f.following_id
WHERE f.follower_id = $1 AND u.status = 'ACTIVE';
