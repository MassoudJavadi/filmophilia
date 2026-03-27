package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrMovieNotFound  = errors.New("movie not found")
	ErrRatingNotFound = errors.New("rating not found")
)

type RatingService interface {
	RateMovie(ctx context.Context, userID, movieID, score int32) (db.Rating, error)
	GetUserRating(ctx context.Context, userID, movieID int32) (db.Rating, error)
	GetMovieRatings(ctx context.Context, movieID, limit, offset int32) ([]db.GetMovieRatingsRow, error)
	GetUserRatings(ctx context.Context, userID, limit, offset int32) ([]db.GetUserRatingsRow, error)
	DeleteRating(ctx context.Context, userID, movieID int32) error
}

type ratingService struct {
	queries *db.Queries
}

func NewRatingService(queries *db.Queries) RatingService {
	return &ratingService{queries: queries}
}

func (s *ratingService) RateMovie(ctx context.Context, userID, movieID, score int32) (db.Rating, error) {
	if _, err := s.queries.GetMovieByID(ctx, movieID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Rating{}, ErrMovieNotFound
		}
		return db.Rating{}, err
	}

	rating, err := s.queries.UpsertRating(ctx, db.UpsertRatingParams{
		UserID:  pgtype.Int4{Int32: userID, Valid: true},
		MovieID: movieID,
		Score:   score,
	})
	if err != nil {
		return db.Rating{}, err
	}

	if err := s.queries.UpdateMovieRatingStats(ctx, movieID); err != nil {
		return db.Rating{}, err
	}

	return rating, nil
}

func (s *ratingService) GetUserRating(ctx context.Context, userID, movieID int32) (db.Rating, error) {
	rating, err := s.queries.GetUserRatingForMovie(ctx, db.GetUserRatingForMovieParams{
		UserID:  pgtype.Int4{Int32: userID, Valid: true},
		MovieID: movieID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Rating{}, ErrRatingNotFound
		}
		return db.Rating{}, err
	}
	return rating, nil
}

func (s *ratingService) GetMovieRatings(ctx context.Context, movieID, limit, offset int32) ([]db.GetMovieRatingsRow, error) {
	return s.queries.GetMovieRatings(ctx, db.GetMovieRatingsParams{
		MovieID: movieID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *ratingService) GetUserRatings(ctx context.Context, userID, limit, offset int32) ([]db.GetUserRatingsRow, error) {
	return s.queries.GetUserRatings(ctx, db.GetUserRatingsParams{
		UserID: pgtype.Int4{Int32: userID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (s *ratingService) DeleteRating(ctx context.Context, userID, movieID int32) error {
	if _, err := s.GetUserRating(ctx, userID, movieID); err != nil {
		return err
	}

	if err := s.queries.DeleteRating(ctx, db.DeleteRatingParams{
		UserID:  pgtype.Int4{Int32: userID, Valid: true},
		MovieID: movieID,
	}); err != nil {
		return err
	}

	return s.queries.UpdateMovieRatingStats(ctx, movieID)
}
