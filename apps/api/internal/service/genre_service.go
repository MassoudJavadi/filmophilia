package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
)

var (
	ErrGenreNotFound = errors.New("genre not found")
)

type GenreService interface {
	GetAll(ctx context.Context) ([]db.Genre, error)
	GetByID(ctx context.Context, id int32) (db.Genre, error)
	GetBySlug(ctx context.Context, slug string) (db.Genre, error)
	GetMoviesByGenre(ctx context.Context, genreID, limit, offset int32) ([]db.Movie, error)
	GetMoviesByGenreSlug(ctx context.Context, slug string, limit, offset int32) ([]db.Movie, error)
	GetGenresForMovie(ctx context.Context, movieID int32) ([]db.Genre, error)
	CountMoviesByGenre(ctx context.Context, genreID int32) (int32, error)
}

type genreService struct {
	queries *db.Queries
}

func NewGenreService(queries *db.Queries) GenreService {
	return &genreService{queries: queries}
}

func (s *genreService) GetAll(ctx context.Context) ([]db.Genre, error) {
	return s.queries.GetAllGenres(ctx)
}

func (s *genreService) GetByID(ctx context.Context, id int32) (db.Genre, error) {
	genre, err := s.queries.GetGenreByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Genre{}, ErrGenreNotFound
		}
		return db.Genre{}, err
	}
	return genre, nil
}

func (s *genreService) GetBySlug(ctx context.Context, slug string) (db.Genre, error) {
	genre, err := s.queries.GetGenreBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Genre{}, ErrGenreNotFound
		}
		return db.Genre{}, err
	}
	return genre, nil
}

func (s *genreService) GetMoviesByGenre(ctx context.Context, genreID, limit, offset int32) ([]db.Movie, error) {
	return s.queries.GetMoviesByGenre(ctx, db.GetMoviesByGenreParams{
		GenreID: genreID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *genreService) GetMoviesByGenreSlug(ctx context.Context, slug string, limit, offset int32) ([]db.Movie, error) {
	return s.queries.GetMoviesByGenreSlug(ctx, db.GetMoviesByGenreSlugParams{
		Slug:   slug,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *genreService) GetGenresForMovie(ctx context.Context, movieID int32) ([]db.Genre, error) {
	return s.queries.GetGenresForMovie(ctx, movieID)
}

func (s *genreService) CountMoviesByGenre(ctx context.Context, genreID int32) (int32, error) {
	return s.queries.CountMoviesByGenre(ctx, genreID)
}
