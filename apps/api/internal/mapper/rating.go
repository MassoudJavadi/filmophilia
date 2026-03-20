package mapper

import (
	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
)

func ToRatingResponse(r db.Rating) dto.RatingResponse {
	return dto.RatingResponse{
		ID:        r.ID,
		UserID:    r.UserID,
		MovieID:   r.MovieID,
		Score:     r.Score,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func ToRatingWithUserResponse(r db.GetMovieRatingsRow) dto.RatingWithUserResponse {
	return dto.RatingWithUserResponse{
		ID:          r.ID,
		Score:       r.Score,
		CreatedAt:   r.CreatedAt.Time,
		UserID:      r.UserID,
		Username:    r.Username,
		DisplayName: r.DisplayName.String,
		AvatarURL:   r.AvatarUrl.String,
	}
}

func ToRatingWithUserResponses(rows []db.GetMovieRatingsRow) []dto.RatingWithUserResponse {
	result := make([]dto.RatingWithUserResponse, len(rows))
	for i, r := range rows {
		result[i] = ToRatingWithUserResponse(r)
	}
	return result
}

func ToRatingWithMovieResponse(r db.GetUserRatingsRow) dto.RatingWithMovieResponse {
	return dto.RatingWithMovieResponse{
		ID:        r.ID,
		Score:     r.Score,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
		MovieID:   r.MovieID,
		Title:     r.Title,
		Slug:      r.Slug,
		PosterURL: r.PosterUrl.String,
	}
}

func ToRatingWithMovieResponses(rows []db.GetUserRatingsRow) []dto.RatingWithMovieResponse {
	result := make([]dto.RatingWithMovieResponse, len(rows))
	for i, r := range rows {
		result[i] = ToRatingWithMovieResponse(r)
	}
	return result
}
