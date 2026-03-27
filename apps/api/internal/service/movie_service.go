package service

import (
	"context"
	"math/big"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/jackc/pgx/v5/pgtype"
)

// Helper to convert float64 to pgtype.Numeric
func float64ToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	// Convert float to big.Int by multiplying by 100 (2 decimal places)
	scaled := int64(*f * 100)
	return pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -2,
		Valid: true,
	}
}

type MovieService interface {
	GetMovies(ctx context.Context, limit, offset int32) ([]db.ListMoviesRow, error)
	GetMovie(ctx context.Context, slug string) (db.Movie, error)
	Search(ctx context.Context, query string, limit, offset int32) ([]db.Movie, error)
	AdvancedSearch(ctx context.Context, req dto.AdvancedSearchRequest) ([]db.Movie, int32, error)
}

type movieService struct {
	queries *db.Queries
}

func NewMovieService(queries *db.Queries) MovieService {
	return &movieService{queries: queries}
}

func (s *movieService) GetMovies(ctx context.Context, limit, offset int32) ([]db.ListMoviesRow, error) {
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

func (s *movieService) AdvancedSearch(ctx context.Context, req dto.AdvancedSearchRequest) ([]db.Movie, int32, error) {
	// Set defaults
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Build search params
	params := db.AdvancedSearchMoviesParams{
		Limit:  limit,
		Offset: offset,
	}

	// Query filter
	if req.Query != "" {
		params.Query = pgtype.Text{String: req.Query, Valid: true}
	}

	// Genre filter
	if len(req.GenreIDs) > 0 {
		params.GenreIds = req.GenreIDs
	}

	// Year range
	if req.YearFrom != nil {
		params.YearFrom = pgtype.Int4{Int32: *req.YearFrom, Valid: true}
	}
	if req.YearTo != nil {
		params.YearTo = pgtype.Int4{Int32: *req.YearTo, Valid: true}
	}

	// IMDB rating range
	params.ImdbMin = float64ToNumeric(req.ImdbMin)
	params.ImdbMax = float64ToNumeric(req.ImdbMax)

	// User rating range
	if req.UserRatingMin != nil {
		params.UserRatingMin = pgtype.Float4{Float32: *req.UserRatingMin, Valid: true}
	}
	if req.UserRatingMax != nil {
		params.UserRatingMax = pgtype.Float4{Float32: *req.UserRatingMax, Valid: true}
	}

	// Rotten Tomatoes range
	if req.RTMin != nil {
		params.RtMin = pgtype.Int4{Int32: *req.RTMin, Valid: true}
	}
	if req.RTMax != nil {
		params.RtMax = pgtype.Int4{Int32: *req.RTMax, Valid: true}
	}

	// Metacritic range
	if req.MetacriticMin != nil {
		params.MetacriticMin = pgtype.Int4{Int32: *req.MetacriticMin, Valid: true}
	}
	if req.MetacriticMax != nil {
		params.MetacriticMax = pgtype.Int4{Int32: *req.MetacriticMax, Valid: true}
	}

	// Sort
	if req.SortBy != "" && dto.ValidSortOptions[req.SortBy] {
		params.SortBy = pgtype.Text{String: req.SortBy, Valid: true}
	}

	// Execute search
	movies, err := s.queries.AdvancedSearchMovies(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	countParams := db.CountAdvancedSearchMoviesParams{
		Query:         params.Query,
		GenreIds:      params.GenreIds,
		YearFrom:      params.YearFrom,
		YearTo:        params.YearTo,
		ImdbMin:       params.ImdbMin,
		ImdbMax:       params.ImdbMax,
		UserRatingMin: params.UserRatingMin,
		UserRatingMax: params.UserRatingMax,
		RtMin:         params.RtMin,
		RtMax:         params.RtMax,
		MetacriticMin: params.MetacriticMin,
		MetacriticMax: params.MetacriticMax,
	}

	total, err := s.queries.CountAdvancedSearchMovies(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}
