package dto

import "time"

type FollowUserResponse struct {
	ID          int32     `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	FollowedAt  time.Time `json:"followed_at"`
}

type FollowStatsResponse struct {
	FollowerCount  int32 `json:"follower_count"`
	FollowingCount int32 `json:"following_count"`
}
