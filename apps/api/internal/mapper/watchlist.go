package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToWatchlistEntryResponse(w db.Watchlist) dto.WatchlistEntryResponse {
	resp := dto.WatchlistEntryResponse{
		ID:           w.ID,
		MovieID:      w.MovieID,
		Notes:        w.Notes.String,
		RankPosition: w.RankPosition.Float32,
		CreatedAt:    w.CreatedAt.Time,
	}
	if w.WatchedAt.Valid {
		resp.WatchedAt = &w.WatchedAt.Time
	}
	return resp
}

func ToWatchlistItemResponse(w db.GetUserWatchlistRow) dto.WatchlistItemResponse {
	resp := dto.WatchlistItemResponse{
		ID:            w.ID,
		MovieID:       w.MovieID,
		Title:         w.Title,
		Slug:          w.Slug,
		PosterURL:     w.PosterUrl.String,
		Runtime:       w.Runtime.Int32,
		UserAvgRating: w.UserAvgRating.Float32,
		Notes:         w.Notes.String,
		RankPosition:  w.RankPosition.Float32,
		CreatedAt:     w.CreatedAt.Time,
	}
	if w.ReleaseDate.Valid {
		resp.ReleaseDate = w.ReleaseDate.Time.Format("2006-01-02")
	}
	if w.WatchedAt.Valid {
		resp.WatchedAt = &w.WatchedAt.Time
	}
	return resp
}

func ToWatchlistItemResponses(rows []db.GetUserWatchlistRow) []dto.WatchlistItemResponse {
	result := make([]dto.WatchlistItemResponse, len(rows))
	for i, r := range rows {
		result[i] = ToWatchlistItemResponse(r)
	}
	return result
}

func ToWatchlistCountResponse(c db.CountUserWatchlistRow) dto.WatchlistCountResponse {
	return dto.WatchlistCountResponse{
		Unwatched: c.Unwatched,
		Watched:   c.Watched,
		Total:     c.Total,
	}
}
