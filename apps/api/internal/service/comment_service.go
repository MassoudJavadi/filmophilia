package service

import (
	"context"
	"errors"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrNotCommentOwner  = errors.New("not the comment owner")
	ErrParentNotFound   = errors.New("parent comment not found")
	ErrCannotReplyReply = errors.New("cannot reply to a reply")
)

type CommentService interface {
	Create(ctx context.Context, userID, movieID int32, parentID *int32, content string) (db.GetCommentWithUserRow, error)
	GetByID(ctx context.Context, commentID int32) (db.GetCommentWithUserRow, error)
	GetMovieComments(ctx context.Context, movieID, limit, offset int32) ([]db.GetMovieCommentsRow, error)
	GetReplies(ctx context.Context, parentID, limit, offset int32) ([]db.GetCommentRepliesRow, error)
	GetUserComments(ctx context.Context, userID, limit, offset int32) ([]db.GetUserCommentsRow, error)
	Update(ctx context.Context, userID, commentID int32, content string) (db.GetCommentWithUserRow, error)
	Delete(ctx context.Context, userID, commentID int32) error
	CountMovieComments(ctx context.Context, movieID int32) (int32, error)
}

type commentService struct {
	queries *db.Queries
}

func NewCommentService(queries *db.Queries) CommentService {
	return &commentService{queries: queries}
}

func (s *commentService) Create(ctx context.Context, userID, movieID int32, parentID *int32, content string) (db.GetCommentWithUserRow, error) {
	if _, err := s.queries.GetMovieByID(ctx, movieID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetCommentWithUserRow{}, ErrMovieNotFound
		}
		return db.GetCommentWithUserRow{}, err
	}

	params := db.CreateCommentParams{
		UserID:  userID,
		MovieID: movieID,
		Content: content,
	}

	if parentID != nil {
		parent, err := s.queries.GetCommentByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return db.GetCommentWithUserRow{}, ErrParentNotFound
			}
			return db.GetCommentWithUserRow{}, err
		}
		if parent.ParentID.Valid {
			return db.GetCommentWithUserRow{}, ErrCannotReplyReply
		}
		if parent.MovieID != movieID {
			return db.GetCommentWithUserRow{}, ErrParentNotFound
		}
		params.ParentID = pgtype.Int4{Int32: *parentID, Valid: true}
	}

	comment, err := s.queries.CreateComment(ctx, params)
	if err != nil {
		return db.GetCommentWithUserRow{}, err
	}

	return s.queries.GetCommentWithUser(ctx, comment.ID)
}

func (s *commentService) GetByID(ctx context.Context, commentID int32) (db.GetCommentWithUserRow, error) {
	comment, err := s.queries.GetCommentWithUser(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetCommentWithUserRow{}, ErrCommentNotFound
		}
		return db.GetCommentWithUserRow{}, err
	}
	return comment, nil
}

func (s *commentService) GetMovieComments(ctx context.Context, movieID, limit, offset int32) ([]db.GetMovieCommentsRow, error) {
	return s.queries.GetMovieComments(ctx, db.GetMovieCommentsParams{
		MovieID: movieID,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *commentService) GetReplies(ctx context.Context, parentID, limit, offset int32) ([]db.GetCommentRepliesRow, error) {
	return s.queries.GetCommentReplies(ctx, db.GetCommentRepliesParams{
		ParentID: pgtype.Int4{Int32: parentID, Valid: true},
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *commentService) GetUserComments(ctx context.Context, userID, limit, offset int32) ([]db.GetUserCommentsRow, error) {
	return s.queries.GetUserComments(ctx, db.GetUserCommentsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *commentService) Update(ctx context.Context, userID, commentID int32, content string) (db.GetCommentWithUserRow, error) {
	comment, err := s.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetCommentWithUserRow{}, ErrCommentNotFound
		}
		return db.GetCommentWithUserRow{}, err
	}

	if comment.UserID != userID {
		return db.GetCommentWithUserRow{}, ErrNotCommentOwner
	}

	if _, err := s.queries.UpdateComment(ctx, db.UpdateCommentParams{
		ID:      commentID,
		Content: content,
	}); err != nil {
		return db.GetCommentWithUserRow{}, err
	}

	return s.queries.GetCommentWithUser(ctx, commentID)
}

func (s *commentService) Delete(ctx context.Context, userID, commentID int32) error {
	comment, err := s.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return ErrNotCommentOwner
	}

	return s.queries.SoftDeleteComment(ctx, commentID)
}

func (s *commentService) CountMovieComments(ctx context.Context, movieID int32) (int32, error) {
	return s.queries.CountMovieComments(ctx, movieID)
}
