package service

import (
	"context"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type MovieService interface {
	GetMovies(ctx context.Context, limit, offset int32) ([]db.Movie, error)
	GetMovie(ctx context.Context, slug string) (db.Movie, error)
	Search(ctx context.Context, query string, limit, offset int32) ([]db.Movie, error)
}

type movieService struct {
	queries *db.Queries
}

func NewMovieService(queries *db.Queries) MovieService {
	return &movieService{queries: queries}
}

func (s *movieService) GetMovies(ctx context.Context, limit, offset int32) ([]db.Movie, error) {
	return s.queries.ListMovies(ctx, db.ListMoviesParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *movieService) GetMovie(ctx context.Context, slug string) (db.Movie, error) {
	return s.queries.GetMovieBySlug(ctx, slug)
}

func (s *movieService) Search(ctx context.Context, query string, limit, offset int32) ([]db.Movie, error) {
	return s.queries.SearchMovies(ctx, db.SearchMoviesParams{
		Column1: pgtype.Text{String: query, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
}
