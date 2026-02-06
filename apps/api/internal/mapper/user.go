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
