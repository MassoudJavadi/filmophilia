package dto

import "time"

type ActivityResponse struct {
	ID          int64                  `json:"id"`
	UserID      int32                  `json:"user_id"`
	Username    string                 `json:"username"`
	DisplayName string                 `json:"display_name,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	Action      string                 `json:"action"`
	EntityType  string                 `json:"entity_type"`
	EntityID    int64                  `json:"entity_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ActivityFeedResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Total      int32              `json:"total"`
	Page       int32              `json:"page"`
	Limit      int32              `json:"limit"`
	TotalPages int32              `json:"total_pages"`
}

type CreateActivityRequest struct {
	Action     string                 `json:"action" binding:"required"`
	EntityType string                 `json:"entity_type" binding:"required"`
	EntityID   int64                  `json:"entity_id" binding:"required"`
	Metadata   map[string]interface{} `json:"metadata"`
}
