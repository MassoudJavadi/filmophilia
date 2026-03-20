package dto

import "time"

type AddReactionRequest struct {
	Type string `json:"type" binding:"required,oneof=LIKE LOVE LAUGH SAD ANGRY"`
}

type ReactionResponse struct {
	ID          int32     `json:"id"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      int32     `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
}

type ReactionCountResponse struct {
	Type  string `json:"type"`
	Count int32  `json:"count"`
}

type ReactionSummaryResponse struct {
	Counts       []ReactionCountResponse `json:"counts"`
	UserReaction *string                 `json:"user_reaction,omitempty"`
}
