package service

import (
	"context"
	"errors"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/mapper"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
)

type AuthService struct {
	queries *db.Queries
	jwt     *token.JWTManager
}

func NewAuthService(q *db.Queries, j *token.JWTManager) *AuthService {
	return &AuthService{queries: q, jwt: j}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	return s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	})
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status == db.UserStatusBANNED {
		return nil, errors.New("user is banned")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	session, err := s.queries.GetSessionByRefreshToken(ctx, pgtype.Text{String: refreshToken, Valid: true})
	if err != nil || time.Now().After(session.ExpiresAt.Time) {
		return nil, ErrInvalidToken
	}

	user, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Token Rotation: Delete old session and issue new one
	_ = s.queries.DeleteSession(ctx, session.ID)
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.queries.DeleteSessionByRefreshToken(ctx, pgtype.Text{String: refreshToken, Valid: true})
}

// Helper to bundle token issuance
func (s *AuthService) issueTokens(ctx context.Context, user db.User) (*dto.AuthResponse, error) {
	access, err := s.jwt.Generate(user.ID, string(user.Role), 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refresh := uuid.New().String()
	_, err = s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		RefreshToken: pgtype.Text{String: refresh, Valid: true},
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	})

	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         mapper.ToUserResponse(user),
	}, err
}