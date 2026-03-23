package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
)

var (
	ErrReactionNotFound = errors.New("reaction not found")
)

type ReactionService interface {
	ReactToComment(ctx context.Context, userID int32, commentID int64, reactionType db.ReactionType) (db.Reaction, error)
	RemoveCommentReaction(ctx context.Context, userID int32, commentID int64) error
	GetCommentReactions(ctx context.Context, commentID int64, limit, offset int32) ([]db.GetCommentReactionsRow, error)
	GetCommentReactionCounts(ctx context.Context, commentID int64) ([]db.CountCommentReactionsByTypeRow, error)
	GetUserCommentReaction(ctx context.Context, userID int32, commentID int64) (*db.Reaction, error)
}

type reactionService struct {
	queries *db.Queries
}

func NewReactionService(queries *db.Queries) ReactionService {
	return &reactionService{queries: queries}
}

func (s *reactionService) ReactToComment(ctx context.Context, userID int32, commentID int64, reactionType db.ReactionType) (db.Reaction, error) {
	if _, err := s.queries.GetCommentByID(ctx, commentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Reaction{}, ErrCommentNotFound
		}
		return db.Reaction{}, err
	}

	return s.queries.AddReactionToComment(ctx, db.AddReactionToCommentParams{
		UserID:    userID,
		CommentID: commentID,
		Type:      reactionType,
	})
}

func (s *reactionService) RemoveCommentReaction(ctx context.Context, userID int32, commentID int64) error {
	_, err := s.queries.GetUserReactionOnComment(ctx, db.GetUserReactionOnCommentParams{
		UserID:    userID,
		CommentID: commentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrReactionNotFound
		}
		return err
	}

	return s.queries.RemoveReactionFromComment(ctx, db.RemoveReactionFromCommentParams{
		UserID:    userID,
		CommentID: commentID,
	})
}

func (s *reactionService) GetCommentReactions(ctx context.Context, commentID int64, limit, offset int32) ([]db.GetCommentReactionsRow, error) {
	return s.queries.GetCommentReactions(ctx, db.GetCommentReactionsParams{
		CommentID: commentID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (s *reactionService) GetCommentReactionCounts(ctx context.Context, commentID int64) ([]db.CountCommentReactionsByTypeRow, error) {
	return s.queries.CountCommentReactionsByType(ctx, commentID)
}

func (s *reactionService) GetUserCommentReaction(ctx context.Context, userID int32, commentID int64) (*db.Reaction, error) {
	reaction, err := s.queries.GetUserReactionOnComment(ctx, db.GetUserReactionOnCommentParams{
		UserID:    userID,
		CommentID: commentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &reaction, nil
}
