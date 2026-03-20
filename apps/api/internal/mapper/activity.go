package mapper

import (
	"encoding/json"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToActivityResponse(a db.GetUserActivitiesRow) dto.ActivityResponse {
	resp := dto.ActivityResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		Username:   a.Username,
		Action:     a.Action,
		EntityType: string(a.EntityType),
		EntityID:   a.EntityID,
		CreatedAt:  a.CreatedAt.Time,
	}

	if a.DisplayName.Valid {
		resp.DisplayName = a.DisplayName.String
	}
	if a.AvatarUrl.Valid {
		resp.AvatarURL = a.AvatarUrl.String
	}

	// Parse metadata JSONB
	if len(a.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(a.Metadata, &metadata); err == nil {
			resp.Metadata = metadata
		}
	}

	return resp
}

func ToActivityResponses(activities []db.GetUserActivitiesRow) []dto.ActivityResponse {
	result := make([]dto.ActivityResponse, len(activities))
	for i, a := range activities {
		result[i] = ToActivityResponse(a)
	}
	return result
}

func ToFollowingActivityResponse(a db.GetFollowingActivitiesRow) dto.ActivityResponse {
	resp := dto.ActivityResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		Username:   a.Username,
		Action:     a.Action,
		EntityType: string(a.EntityType),
		EntityID:   a.EntityID,
		CreatedAt:  a.CreatedAt.Time,
	}

	if a.DisplayName.Valid {
		resp.DisplayName = a.DisplayName.String
	}
	if a.AvatarUrl.Valid {
		resp.AvatarURL = a.AvatarUrl.String
	}

	if len(a.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(a.Metadata, &metadata); err == nil {
			resp.Metadata = metadata
		}
	}

	return resp
}

func ToFollowingActivityResponses(activities []db.GetFollowingActivitiesRow) []dto.ActivityResponse {
	result := make([]dto.ActivityResponse, len(activities))
	for i, a := range activities {
		result[i] = ToFollowingActivityResponse(a)
	}
	return result
}

func ToEntityActivityResponse(a db.GetActivitiesByEntityRow) dto.ActivityResponse {
	resp := dto.ActivityResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		Username:   a.Username,
		Action:     a.Action,
		EntityType: string(a.EntityType),
		EntityID:   a.EntityID,
		CreatedAt:  a.CreatedAt.Time,
	}

	if a.DisplayName.Valid {
		resp.DisplayName = a.DisplayName.String
	}
	if a.AvatarUrl.Valid {
		resp.AvatarURL = a.AvatarUrl.String
	}

	if len(a.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(a.Metadata, &metadata); err == nil {
			resp.Metadata = metadata
		}
	}

	return resp
}

func ToEntityActivityResponses(activities []db.GetActivitiesByEntityRow) []dto.ActivityResponse {
	result := make([]dto.ActivityResponse, len(activities))
	for i, a := range activities {
		result[i] = ToEntityActivityResponse(a)
	}
	return result
}
