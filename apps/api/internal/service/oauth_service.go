package service

import (
	"context"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/oauth"
	"github.com/jackc/pgx/v5/pgtype"
)

type OAuthService struct {
	queries   *db.Queries
	authSvc   *AuthService // We reuse AuthService to issue tokens
	googleMgr *oauth.GoogleManager
}

func NewOAuthService(q *db.Queries, a *AuthService, g *oauth.GoogleManager) *OAuthService {
	return &OAuthService{queries: q, authSvc: a, googleMgr: g}
}

func (s *OAuthService) GetGoogleAuthURL(state string) string {
	return s.googleMgr.GetAuthURL(state)
}

func (s *OAuthService) HandleGoogleCallback(ctx context.Context, code string, meta SessionMeta) (*dto.AuthResponse, error) {
	// 1. Get info from Google API
	gUser, err := s.googleMgr.GetUserInfo(ctx, code)
	if err != nil {
		return nil, err
	}

	// 2. Find or Create user in our DB
	user, err := s.queries.GetUserByEmail(ctx, gUser.Email)
	if err != nil {
		// User doesn't exist, create them without a local password hash.
		user, err = s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:        gUser.Email,
			Username:     gUser.Email, // Simplified for now
			PasswordHash: pgtype.Text{},
		})
		if err != nil {
			return nil, err
		}
	}

	// 3. Check if existing user is banned or suspended
	if user.Status == db.UserStatusBANNED {
		return nil, ErrUserBanned
	}
	if user.Status == db.UserStatusSUSPENDED {
		return nil, ErrUserSuspended
	}

	// 4. Issue tokens with session metadata
	return s.authSvc.IssueTokens(ctx, user, meta)
}
