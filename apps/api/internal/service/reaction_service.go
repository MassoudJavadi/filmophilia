package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrReactionNotFound = errors.New("reaction not found")
	ErrInvalidTarget    = errors.New("invalid reaction target")
)

type ReactionService interface {
	// Comment reactions
	ReactToComment(ctx context.Context, userID, commentID int32, reactionType db.ReactionType) (db.Reaction, error)
	RemoveCommentReaction(ctx context.Context, userID, commentID int32) error
	GetCommentReactions(ctx context.Context, commentID, limit, offset int32) ([]db.GetCommentReactionsRow, error)
	GetCommentReactionCounts(ctx context.Context, commentID int32) ([]db.CountCommentReactionsByTypeRow, error)
	GetUserCommentReaction(ctx context.Context, userID, commentID int32) (*db.Reaction, error)

	// Review reactions (for future use)
	ReactToReview(ctx context.Context, userID, reviewID int32, reactionType db.ReactionType) (db.Reaction, error)
	RemoveReviewReaction(ctx context.Context, userID, reviewID int32) error
	GetReviewReactions(ctx context.Context, reviewID, limit, offset int32) ([]db.GetReviewReactionsRow, error)
	GetReviewReactionCounts(ctx context.Context, reviewID int32) ([]db.CountReviewReactionsByTypeRow, error)
	GetUserReviewReaction(ctx context.Context, userID, reviewID int32) (*db.Reaction, error)
}

type reactionService struct {
	queries *db.Queries
}

func NewReactionService(queries *db.Queries) ReactionService {
	return &reactionService{queries: queries}
}

// Comment reactions

func (s *reactionService) ReactToComment(ctx context.Context, userID, commentID int32, reactionType db.ReactionType) (db.Reaction, error) {
	if _, err := s.queries.GetCommentByID(ctx, commentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Reaction{}, ErrCommentNotFound
		}
		return db.Reaction{}, err
	}

	return s.queries.AddReactionToComment(ctx, db.AddReactionToCommentParams{
		UserID:    userID,
		CommentID: pgtype.Int4{Int32: commentID, Valid: true},
		Type:      reactionType,
	})
}

func (s *reactionService) RemoveCommentReaction(ctx context.Context, userID, commentID int32) error {
	_, err := s.queries.GetUserReactionOnComment(ctx, db.GetUserReactionOnCommentParams{
		UserID:    userID,
		CommentID: pgtype.Int4{Int32: commentID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrReactionNotFound
		}
		return err
	}

	return s.queries.RemoveReactionFromComment(ctx, db.RemoveReactionFromCommentParams{
		UserID:    userID,
		CommentID: pgtype.Int4{Int32: commentID, Valid: true},
	})
}

func (s *reactionService) GetCommentReactions(ctx context.Context, commentID, limit, offset int32) ([]db.GetCommentReactionsRow, error) {
	return s.queries.GetCommentReactions(ctx, db.GetCommentReactionsParams{
		CommentID: pgtype.Int4{Int32: commentID, Valid: true},
		Limit:     limit,
		Offset:    offset,
	})
}

func (s *reactionService) GetCommentReactionCounts(ctx context.Context, commentID int32) ([]db.CountCommentReactionsByTypeRow, error) {
	return s.queries.CountCommentReactionsByType(ctx, pgtype.Int4{Int32: commentID, Valid: true})
}

func (s *reactionService) GetUserCommentReaction(ctx context.Context, userID, commentID int32) (*db.Reaction, error) {
	reaction, err := s.queries.GetUserReactionOnComment(ctx, db.GetUserReactionOnCommentParams{
		UserID:    userID,
		CommentID: pgtype.Int4{Int32: commentID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &reaction, nil
}

// Review reactions

func (s *reactionService) ReactToReview(ctx context.Context, userID, reviewID int32, reactionType db.ReactionType) (db.Reaction, error) {
	// TODO: Add GetReviewByID query when reviews are implemented
	return s.queries.AddReactionToReview(ctx, db.AddReactionToReviewParams{
		UserID:   userID,
		ReviewID: pgtype.Int4{Int32: reviewID, Valid: true},
		Type:     reactionType,
	})
}

func (s *reactionService) RemoveReviewReaction(ctx context.Context, userID, reviewID int32) error {
	_, err := s.queries.GetUserReactionOnReview(ctx, db.GetUserReactionOnReviewParams{
		UserID:   userID,
		ReviewID: pgtype.Int4{Int32: reviewID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrReactionNotFound
		}
		return err
	}

	return s.queries.RemoveReactionFromReview(ctx, db.RemoveReactionFromReviewParams{
		UserID:   userID,
		ReviewID: pgtype.Int4{Int32: reviewID, Valid: true},
	})
}

func (s *reactionService) GetReviewReactions(ctx context.Context, reviewID, limit, offset int32) ([]db.GetReviewReactionsRow, error) {
	return s.queries.GetReviewReactions(ctx, db.GetReviewReactionsParams{
		ReviewID: pgtype.Int4{Int32: reviewID, Valid: true},
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *reactionService) GetReviewReactionCounts(ctx context.Context, reviewID int32) ([]db.CountReviewReactionsByTypeRow, error) {
	return s.queries.CountReviewReactionsByType(ctx, pgtype.Int4{Int32: reviewID, Valid: true})
}

func (s *reactionService) GetUserReviewReaction(ctx context.Context, userID, reviewID int32) (*db.Reaction, error) {
	reaction, err := s.queries.GetUserReactionOnReview(ctx, db.GetUserReactionOnReviewParams{
		UserID:   userID,
		ReviewID: pgtype.Int4{Int32: reviewID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &reaction, nil
}
