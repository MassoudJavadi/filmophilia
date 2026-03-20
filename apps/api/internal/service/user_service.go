package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
	GetProfile(ctx context.Context, userID int32) (db.GetUserProfileRow, error)
	GetProfileByUsername(ctx context.Context, username string) (db.GetUserProfileByUsernameRow, error)
	UpdateProfile(ctx context.Context, userID int32, displayName, avatarURL, bio *string) (db.User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int32) ([]db.SearchUsersRow, error)
}

type userService struct {
	queries *db.Queries
}

func NewUserService(queries *db.Queries) UserService {
	return &userService{queries: queries}
}

func (s *userService) GetProfile(ctx context.Context, userID int32) (db.GetUserProfileRow, error) {
	profile, err := s.queries.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetUserProfileRow{}, ErrUserNotFound
		}
		return db.GetUserProfileRow{}, err
	}
	return profile, nil
}

func (s *userService) GetProfileByUsername(ctx context.Context, username string) (db.GetUserProfileByUsernameRow, error) {
	profile, err := s.queries.GetUserProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetUserProfileByUsernameRow{}, ErrUserNotFound
		}
		return db.GetUserProfileByUsernameRow{}, err
	}
	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID int32, displayName, avatarURL, bio *string) (db.User, error) {
	params := db.UpdateUserProfileParams{
		ID: userID,
	}

	if displayName != nil {
		params.DisplayName = pgtype.Text{String: *displayName, Valid: true}
	}
	if avatarURL != nil {
		params.AvatarUrl = pgtype.Text{String: *avatarURL, Valid: true}
	}
	if bio != nil {
		params.Bio = pgtype.Text{String: *bio, Valid: true}
	}

	return s.queries.UpdateUserProfile(ctx, params)
}

func (s *userService) SearchUsers(ctx context.Context, query string, limit, offset int32) ([]db.SearchUsersRow, error) {
	return s.queries.SearchUsers(ctx, db.SearchUsersParams{
		Column1: pgtype.Text{String: query, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
}
