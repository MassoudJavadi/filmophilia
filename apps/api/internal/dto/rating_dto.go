package dto

import "time"

type RateMovieRequest struct {
	Score int32 `json:"score" binding:"required,min=1,max=10"`
}

type RatingResponse struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	MovieID   int32     `json:"movie_id"`
	Score     int32     `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RatingWithUserResponse struct {
	ID          int32     `json:"id"`
	Score       int32     `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      int32     `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
}

type RatingWithMovieResponse struct {
	ID        int32     `json:"id"`
	Score     int32     `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MovieID   int32     `json:"movie_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	PosterURL string    `json:"poster_url,omitempty"`
}
