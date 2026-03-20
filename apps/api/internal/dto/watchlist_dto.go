package dto

import "time"

type AddToWatchlistRequest struct {
	Notes *string  `json:"notes,omitempty"`
	Rank  *float32 `json:"rank,omitempty"`
}

type UpdateWatchlistRequest struct {
	Notes *string  `json:"notes,omitempty"`
	Rank  *float32 `json:"rank,omitempty"`
}

type WatchlistEntryResponse struct {
	ID           int32      `json:"id"`
	MovieID      int32      `json:"movie_id"`
	Notes        string     `json:"notes,omitempty"`
	RankPosition float32    `json:"rank_position"`
	WatchedAt    *time.Time `json:"watched_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type WatchlistItemResponse struct {
	ID            int32      `json:"id"`
	MovieID       int32      `json:"movie_id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	PosterURL     string     `json:"poster_url,omitempty"`
	ReleaseDate   string     `json:"release_date,omitempty"`
	Runtime       int32      `json:"runtime,omitempty"`
	UserAvgRating float32    `json:"user_avg_rating,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	RankPosition  float32    `json:"rank_position"`
	WatchedAt     *time.Time `json:"watched_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type WatchlistCountResponse struct {
	Unwatched int32 `json:"unwatched"`
	Watched   int32 `json:"watched"`
	Total     int32 `json:"total"`
}
