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

func ToReactionCounts(rows []db.CountCommentReactionsByTypeRow) []dto.ReactionCountResponse {
	result := make([]dto.ReactionCountResponse, len(rows))
	for i, r := range rows {
		result[i] = dto.ReactionCountResponse{
			Type:  string(r.Type),
			Count: r.Count,
		}
	}
	return result
}
