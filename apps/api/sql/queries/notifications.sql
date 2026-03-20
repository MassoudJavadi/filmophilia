-- name: CreateNotification :one
-- Create a new notification
INSERT INTO notifications (user_id, type, title, content, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserNotifications :many
-- Get all notifications for a user with pagination
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotifications :many
-- Get unread notifications for a user
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = FALSE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetNotificationByID :one
-- Get a specific notification
SELECT * FROM notifications
WHERE id = $1;

-- name: CountUserNotifications :one
-- Count total notifications for a user
SELECT COUNT(*)::INT as count
FROM notifications
WHERE user_id = $1;

-- name: CountUnreadNotifications :one
-- Count unread notifications for a user
SELECT COUNT(*)::INT as count
FROM notifications
WHERE user_id = $1 AND is_read = FALSE;

-- name: MarkNotificationAsRead :exec
-- Mark a specific notification as read
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :exec
-- Mark all notifications as read for a user
UPDATE notifications
SET is_read = TRUE
WHERE user_id = $1 AND is_read = FALSE;

-- name: DeleteNotification :exec
-- Delete a specific notification
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteAllNotifications :exec
-- Delete all notifications for a user
DELETE FROM notifications
WHERE user_id = $1;

-- name: DeleteOldNotifications :exec
-- Delete notifications older than specified date (cleanup)
DELETE FROM notifications
WHERE created_at < $1;
