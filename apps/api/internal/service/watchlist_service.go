package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrAlreadyInWatchlist = errors.New("movie already in watchlist")
	ErrNotInWatchlist     = errors.New("movie not in watchlist")
)

type WatchlistFilter string

const (
	WatchlistFilterAll       WatchlistFilter = "all"
	WatchlistFilterUnwatched WatchlistFilter = "unwatched"
	WatchlistFilterWatched   WatchlistFilter = "watched"
)

type WatchlistService interface {
	Add(ctx context.Context, userID, movieID int32, notes *string, rank *float32) (db.Watchlist, error)
	Get(ctx context.Context, userID, movieID int32) (db.Watchlist, error)
	List(ctx context.Context, userID int32, filter WatchlistFilter, limit, offset int32) ([]db.GetUserWatchlistRow, error)
	Count(ctx context.Context, userID int32) (db.CountUserWatchlistRow, error)
	Update(ctx context.Context, userID, movieID int32, notes *string, rank *float32) (db.Watchlist, error)
	MarkWatched(ctx context.Context, userID, movieID int32) (db.Watchlist, error)
	MarkUnwatched(ctx context.Context, userID, movieID int32) (db.Watchlist, error)
	Remove(ctx context.Context, userID, movieID int32) error
	IsInWatchlist(ctx context.Context, userID, movieID int32) (bool, error)
}

type watchlistService struct {
	queries *db.Queries
}

func NewWatchlistService(queries *db.Queries) WatchlistService {
	return &watchlistService{queries: queries}
}

func (s *watchlistService) Add(ctx context.Context, userID, movieID int32, notes *string, rank *float32) (db.Watchlist, error) {
	if _, err := s.queries.GetMovieByID(ctx, movieID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Watchlist{}, ErrMovieNotFound
		}
		return db.Watchlist{}, err
	}

	exists, err := s.queries.IsInWatchlist(ctx, db.IsInWatchlistParams{
		UserID:  userID,
		MovieID: movieID,
	})
	if err != nil {
		return db.Watchlist{}, err
	}
	if exists {
		return db.Watchlist{}, ErrAlreadyInWatchlist
	}

	params := db.AddToWatchlistParams{
		UserID:  userID,
		MovieID: movieID,
	}
	if notes != nil {
		params.Notes = pgtype.Text{String: *notes, Valid: true}
	}
	if rank != nil {
		params.Column4 = *rank
	}

	return s.queries.AddToWatchlist(ctx, params)
}

func (s *watchlistService) Get(ctx context.Context, userID, movieID int32) (db.Watchlist, error) {
	entry, err := s.queries.GetWatchlistEntry(ctx, db.GetWatchlistEntryParams{
		UserID:  userID,
		MovieID: movieID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Watchlist{}, ErrNotInWatchlist
		}
		return db.Watchlist{}, err
	}
	return entry, nil
}

func (s *watchlistService) List(ctx context.Context, userID int32, filter WatchlistFilter, limit, offset int32) ([]db.GetUserWatchlistRow, error) {
	params := db.GetUserWatchlistParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	switch filter {
	case WatchlistFilterUnwatched:
		params.Column2 = true // unwatched only
	case WatchlistFilterWatched:
		params.Column3 = true // watched only
	}

	return s.queries.GetUserWatchlist(ctx, params)
}

func (s *watchlistService) Count(ctx context.Context, userID int32) (db.CountUserWatchlistRow, error) {
	return s.queries.CountUserWatchlist(ctx, userID)
}

func (s *watchlistService) Update(ctx context.Context, userID, movieID int32, notes *string, rank *float32) (db.Watchlist, error) {
	if _, err := s.Get(ctx, userID, movieID); err != nil {
		return db.Watchlist{}, err
	}

	params := db.UpdateWatchlistEntryParams{
		UserID:  userID,
		MovieID: movieID,
	}
	if notes != nil {
		params.Notes = pgtype.Text{String: *notes, Valid: true}
	}
	if rank != nil {
		params.RankPosition = pgtype.Float4{Float32: *rank, Valid: true}
	}

	return s.queries.UpdateWatchlistEntry(ctx, params)
}

func (s *watchlistService) MarkWatched(ctx context.Context, userID, movieID int32) (db.Watchlist, error) {
	if _, err := s.Get(ctx, userID, movieID); err != nil {
		return db.Watchlist{}, err
	}

	return s.queries.MarkAsWatched(ctx, db.MarkAsWatchedParams{
		UserID:  userID,
		MovieID: movieID,
	})
}

func (s *watchlistService) MarkUnwatched(ctx context.Context, userID, movieID int32) (db.Watchlist, error) {
	if _, err := s.Get(ctx, userID, movieID); err != nil {
		return db.Watchlist{}, err
	}

	return s.queries.MarkAsUnwatched(ctx, db.MarkAsUnwatchedParams{
		UserID:  userID,
		MovieID: movieID,
	})
}

func (s *watchlistService) Remove(ctx context.Context, userID, movieID int32) error {
	if _, err := s.Get(ctx, userID, movieID); err != nil {
		return err
	}

	return s.queries.RemoveFromWatchlist(ctx, db.RemoveFromWatchlistParams{
		UserID:  userID,
		MovieID: movieID,
	})
}

func (s *watchlistService) IsInWatchlist(ctx context.Context, userID, movieID int32) (bool, error) {
	return s.queries.IsInWatchlist(ctx, db.IsInWatchlistParams{
		UserID:  userID,
		MovieID: movieID,
	})
}
