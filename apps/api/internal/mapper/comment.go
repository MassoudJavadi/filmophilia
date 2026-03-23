package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToCommentResponse(c db.GetMovieCommentsRow) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:          c.ID,
		MovieID:     c.MovieID,
		Content:     c.Content,
		LikeCount:   c.LikeCount.Int32,
		CreatedAt:   c.CreatedAt.Time,
		UpdatedAt:   c.UpdatedAt.Time,
		UserID:      c.UserID,
		Username:    c.Username,
		DisplayName: c.DisplayName.String,
		AvatarURL:   c.AvatarUrl.String,
		ReplyCount:  c.ReplyCount,
	}
	if c.ParentID.Valid {
		parentID := c.ParentID.Int64
		resp.ParentID = &parentID
	}
	return resp
}

func ToCommentResponses(rows []db.GetMovieCommentsRow) []dto.CommentResponse {
	result := make([]dto.CommentResponse, len(rows))
	for i, r := range rows {
		result[i] = ToCommentResponse(r)
	}
	return result
}

func ToReplyResponse(c db.GetCommentRepliesRow) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:          c.ID,
		MovieID:     c.MovieID,
		Content:     c.Content,
		LikeCount:   c.LikeCount.Int32,
		CreatedAt:   c.CreatedAt.Time,
		UpdatedAt:   c.UpdatedAt.Time,
		UserID:      c.UserID,
		Username:    c.Username,
		DisplayName: c.DisplayName.String,
		AvatarURL:   c.AvatarUrl.String,
	}
	if c.ParentID.Valid {
		parentID := c.ParentID.Int64
		resp.ParentID = &parentID
	}
	return resp
}

func ToReplyResponses(rows []db.GetCommentRepliesRow) []dto.CommentResponse {
	result := make([]dto.CommentResponse, len(rows))
	for i, r := range rows {
		result[i] = ToReplyResponse(r)
	}
	return result
}

func ToCommentWithUserResponse(c db.GetCommentWithUserRow) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:          c.ID,
		MovieID:     c.MovieID,
		Content:     c.Content,
		LikeCount:   c.LikeCount.Int32,
		CreatedAt:   c.CreatedAt.Time,
		UpdatedAt:   c.UpdatedAt.Time,
		UserID:      c.UserID,
		Username:    c.Username,
		DisplayName: c.DisplayName.String,
		AvatarURL:   c.AvatarUrl.String,
	}
	if c.ParentID.Valid {
		parentID := c.ParentID.Int64
		resp.ParentID = &parentID
	}
	return resp
}

func ToCommentWithMovieResponse(c db.GetUserCommentsRow) dto.CommentWithMovieResponse {
	resp := dto.CommentWithMovieResponse{
		ID:         c.ID,
		MovieID:    c.MovieID,
		MovieTitle: c.MovieTitle,
		MovieSlug:  c.MovieSlug,
		Content:    c.Content,
		LikeCount:  c.LikeCount.Int32,
		CreatedAt:  c.CreatedAt.Time,
		UpdatedAt:  c.UpdatedAt.Time,
	}
	if c.ParentID.Valid {
		parentID := c.ParentID.Int64
		resp.ParentID = &parentID
	}
	return resp
}

func ToCommentWithMovieResponses(rows []db.GetUserCommentsRow) []dto.CommentWithMovieResponse {
	result := make([]dto.CommentWithMovieResponse, len(rows))
	for i, r := range rows {
		result[i] = ToCommentWithMovieResponse(r)
	}
	return result
}
