package service

import (
	"context"
	"errors"
	"strings"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
)

var (
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrNotFollowing     = errors.New("not following this user")
)

type FollowService interface {
	Follow(ctx context.Context, followerID, followingID int32) error
	Unfollow(ctx context.Context, followerID, followingID int32) error
	IsFollowing(ctx context.Context, followerID, followingID int32) (bool, error)
	GetFollowers(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowersRow, error)
	GetFollowing(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowingRow, error)
	GetStats(ctx context.Context, userID int32) (followerCount, followingCount int32, err error)
}

type followService struct {
	queries *db.Queries
}

func NewFollowService(queries *db.Queries) FollowService {
	return &followService{queries: queries}
}

func (s *followService) Follow(ctx context.Context, followerID, followingID int32) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}

	// Check if user to follow exists and is active
	if _, err := s.queries.GetUserProfile(ctx, followingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	_, err := s.queries.FollowUser(ctx, db.FollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return ErrAlreadyFollowing
		}
		return err
	}

	return nil
}

func (s *followService) Unfollow(ctx context.Context, followerID, followingID int32) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}

	isFollowing, err := s.queries.IsFollowing(ctx, db.IsFollowingParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		return err
	}
	if !isFollowing {
		return ErrNotFollowing
	}

	return s.queries.UnfollowUser(ctx, db.UnfollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *followService) IsFollowing(ctx context.Context, followerID, followingID int32) (bool, error) {
	return s.queries.IsFollowing(ctx, db.IsFollowingParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *followService) GetFollowers(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowersRow, error) {
	return s.queries.GetFollowers(ctx, db.GetFollowersParams{
		FollowingID: userID,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *followService) GetFollowing(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowingRow, error) {
	return s.queries.GetFollowing(ctx, db.GetFollowingParams{
		FollowerID: userID,
		Limit:      limit,
		Offset:     offset,
	})
}

func (s *followService) GetStats(ctx context.Context, userID int32) (followerCount, followingCount int32, err error) {
	followerCount, err = s.queries.CountFollowers(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	followingCount, err = s.queries.CountFollowing(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	return followerCount, followingCount, nil
}
