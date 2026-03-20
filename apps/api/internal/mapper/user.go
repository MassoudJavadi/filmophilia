package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

// ToUserResponse converts a db.User to dto.UserResponse
func ToUserResponse(user db.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
	}
}

func ToUserProfileResponse(p db.GetUserProfileRow) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		ID:             p.ID,
		Username:       p.Username,
		DisplayName:    p.DisplayName.String,
		AvatarURL:      p.AvatarUrl.String,
		Bio:            p.Bio.String,
		CreatedAt:      p.CreatedAt.Time,
		FollowerCount:  p.FollowerCount,
		FollowingCount: p.FollowingCount,
		RatingCount:    p.RatingCount,
		CommentCount:   p.CommentCount,
	}
}

func ToUserProfileResponseFromUsername(p db.GetUserProfileByUsernameRow) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		ID:             p.ID,
		Username:       p.Username,
		DisplayName:    p.DisplayName.String,
		AvatarURL:      p.AvatarUrl.String,
		Bio:            p.Bio.String,
		CreatedAt:      p.CreatedAt.Time,
		FollowerCount:  p.FollowerCount,
		FollowingCount: p.FollowingCount,
		RatingCount:    p.RatingCount,
		CommentCount:   p.CommentCount,
	}
}

func ToUserSearchResult(u db.SearchUsersRow) dto.UserSearchResult {
	return dto.UserSearchResult{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName.String,
		AvatarURL:   u.AvatarUrl.String,
	}
}

func ToUserSearchResults(rows []db.SearchUsersRow) []dto.UserSearchResult {
	result := make([]dto.UserSearchResult, len(rows))
	for i, r := range rows {
		result[i] = ToUserSearchResult(r)
	}
	return result
}
