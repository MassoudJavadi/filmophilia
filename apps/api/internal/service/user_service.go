package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserBanned         = errors.New("account is banned")
	ErrUserSuspended      = errors.New("account is suspended")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already in use")
	ErrUsernameTaken      = errors.New("username already in use")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
)

type UserService struct {
	queries *db.Queries
}

// NewUserService is a provider for Wire
func NewUserService(q *db.Queries) *UserService {
	return &UserService{queries: q}
}

// Register creates a new user with a hashed password
func (s *UserService) Register(ctx context.Context, req dto.SignupRequest) (db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	})
	if err != nil {
		return db.User{}, err
	}
	return user, nil
}

// Login validates credentials and returns access/refresh tokens
func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is restricted
	switch user.Status {
	case db.UserStatusBANNED:
		return nil, ErrUserBanned
	case db.UserStatusSUSPENDED:
		return nil, ErrUserSuspended
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate Access Token (JWT)
	accessToken, err := s.generateJWT(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token (Opaque String)
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Persist session in database
	_, err = s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:     uuid.New().String(),
		UserID: user.ID,
		RefreshToken: pgtype.Text{
			String: refreshToken,
			Valid:  true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.toUserResponse(user),
	}, nil
}

// RefreshToken handles token rotation: validates old refresh token and issues new ones
func (s *UserService) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	// 1. Validate the refresh token exists in DB
	session, err := s.queries.GetSessionByRefreshToken(ctx, pgtype.Text{
		String: req.RefreshToken,
		Valid:  true,
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 2. Check if the session/refresh token has expired
	if time.Now().After(session.ExpiresAt.Time) {
		_ = s.queries.DeleteSession(ctx, session.ID)
		return nil, ErrInvalidToken
	}

	// 3. Fetch user to get latest role and info
	user, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 4. Generate new tokens (Rotation)
	accessToken, err := s.generateJWT(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	// 5. Update the existing session with new token (Token Rotation)
	err = s.queries.UpdateSession(ctx, db.UpdateSessionParams{
		ID: session.ID,
		RefreshToken: pgtype.Text{
			String: newRefreshToken,
			Valid:  true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  newExpiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         s.toUserResponse(user),
	}, nil
}

// BanUser restricts a user and nukes all their active sessions
func (s *UserService) BanUser(ctx context.Context, userID int32) error {
	err := s.queries.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
		ID:     userID,
		Status: db.UserStatusBANNED,
	})
	if err != nil {
		return err
	}

	// Force logout by deleting all sessions
	return s.queries.DeleteUserSessions(ctx, userID)
}

// generateJWT creates a new JWT for the given user
func (s *UserService) generateJWT(userID int32, role db.Role, expiry time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(expiry).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// toUserResponse maps a DB User to a DTO UserResponse
func (s *UserService) toUserResponse(user db.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
	}
}