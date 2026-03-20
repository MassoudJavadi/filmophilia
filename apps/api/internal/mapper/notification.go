package mapper

import (
	"encoding/json"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToNotificationResponse(n db.Notification) dto.NotificationResponse {
	resp := dto.NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      string(n.Type),
		Title:     n.Title,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Time,
	}

	if n.Content.Valid {
		resp.Content = n.Content.String
	}

	// Parse metadata JSONB
	if len(n.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(n.Metadata, &metadata); err == nil {
			resp.Metadata = metadata
		}
	}

	return resp
}

func ToNotificationResponses(notifications []db.Notification) []dto.NotificationResponse {
	result := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = ToNotificationResponse(n)
	}
	return result
}
