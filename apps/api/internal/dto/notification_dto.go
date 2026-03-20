package dto

import "time"

type NotificationResponse struct {
	ID        int32                  `json:"id"`
	UserID    int32                  `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content,omitempty"`
	IsRead    bool                   `json:"is_read"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int32                  `json:"total"`
	UnreadCount   int32                  `json:"unread_count"`
	Page          int32                  `json:"page"`
	Limit         int32                  `json:"limit"`
	TotalPages    int32                  `json:"total_pages"`
}

type CreateNotificationRequest struct {
	UserID   int32                  `json:"user_id" binding:"required"`
	Type     string                 `json:"type" binding:"required"`
	Title    string                 `json:"title" binding:"required"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}
