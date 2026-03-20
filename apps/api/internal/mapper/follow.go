package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToFollowUserResponse(f db.GetFollowersRow) dto.FollowUserResponse {
	return dto.FollowUserResponse{
		ID:          f.ID,
		Username:    f.Username,
		DisplayName: f.DisplayName.String,
		AvatarURL:   f.AvatarUrl.String,
		Bio:         f.Bio.String,
		FollowedAt:  f.FollowedAt.Time,
	}
}

func ToFollowUserResponses(rows []db.GetFollowersRow) []dto.FollowUserResponse {
	result := make([]dto.FollowUserResponse, len(rows))
	for i, r := range rows {
		result[i] = ToFollowUserResponse(r)
	}
	return result
}

func ToFollowingUserResponse(f db.GetFollowingRow) dto.FollowUserResponse {
	return dto.FollowUserResponse{
		ID:          f.ID,
		Username:    f.Username,
		DisplayName: f.DisplayName.String,
		AvatarURL:   f.AvatarUrl.String,
		Bio:         f.Bio.String,
		FollowedAt:  f.FollowedAt.Time,
	}
}

func ToFollowingUserResponses(rows []db.GetFollowingRow) []dto.FollowUserResponse {
	result := make([]dto.FollowUserResponse, len(rows))
	for i, r := range rows {
		result[i] = ToFollowingUserResponse(r)
	}
	return result
}
