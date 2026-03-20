package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToReactionResponse(r db.GetCommentReactionsRow) dto.ReactionResponse {
	return dto.ReactionResponse{
		ID:          r.ID,
		Type:        string(r.Type),
		CreatedAt:   r.CreatedAt.Time,
		UserID:      r.UserID,
		Username:    r.Username,
		DisplayName: r.DisplayName.String,
		AvatarURL:   r.AvatarUrl.String,
	}
}

func ToReactionResponses(rows []db.GetCommentReactionsRow) []dto.ReactionResponse {
	result := make([]dto.ReactionResponse, len(rows))
	for i, r := range rows {
		result[i] = ToReactionResponse(r)
	}
	return result
}

func ToReviewReactionResponse(r db.GetReviewReactionsRow) dto.ReactionResponse {
	return dto.ReactionResponse{
		ID:          r.ID,
		Type:        string(r.Type),
		CreatedAt:   r.CreatedAt.Time,
		UserID:      r.UserID,
		Username:    r.Username,
		DisplayName: r.DisplayName.String,
		AvatarURL:   r.AvatarUrl.String,
	}
}

func ToReviewReactionResponses(rows []db.GetReviewReactionsRow) []dto.ReactionResponse {
	result := make([]dto.ReactionResponse, len(rows))
	for i, r := range rows {
		result[i] = ToReviewReactionResponse(r)
	}
	return result
}

func ToCommentReactionCounts(rows []db.CountCommentReactionsByTypeRow) []dto.ReactionCountResponse {
	result := make([]dto.ReactionCountResponse, len(rows))
	for i, r := range rows {
		result[i] = dto.ReactionCountResponse{
			Type:  string(r.Type),
			Count: r.Count,
		}
	}
	return result
}

func ToReviewReactionCounts(rows []db.CountReviewReactionsByTypeRow) []dto.ReactionCountResponse {
	result := make([]dto.ReactionCountResponse, len(rows))
	for i, r := range rows {
		result[i] = dto.ReactionCountResponse{
			Type:  string(r.Type),
			Count: r.Count,
		}
	}
	return result
}
