package dto

import "time"

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentResponse struct {
	ID          int32     `json:"id"`
	MovieID     int32     `json:"movie_id"`
	ParentID    *int32    `json:"parent_id,omitempty"`
	Content     string    `json:"content"`
	LikeCount   int32     `json:"like_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      int32     `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	ReplyCount  int32     `json:"reply_count,omitempty"`
}

type CommentWithMovieResponse struct {
	ID         int32     `json:"id"`
	MovieID    int32     `json:"movie_id"`
	MovieTitle string    `json:"movie_title"`
	MovieSlug  string    `json:"movie_slug"`
	ParentID   *int32    `json:"parent_id,omitempty"`
	Content    string    `json:"content"`
	LikeCount  int32     `json:"like_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
