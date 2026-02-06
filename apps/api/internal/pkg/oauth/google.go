package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	envGoogleClientID     = "GOOGLE_CLIENT_ID"
	envGoogleClientSecret = "GOOGLE_CLIENT_SECRET"
	envGoogleRedirectURL  = "GOOGLE_REDIRECT_URL"

	scopeEmail   = "https://www.googleapis.com/auth/userinfo.email"
	scopeProfile = "https://www.googleapis.com/auth/userinfo.profile"

	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type GoogleManager struct {
	config *oauth2.Config
}

func NewGoogleManager() *GoogleManager {
	return &GoogleManager{
		config: &oauth2.Config{
			ClientID:     os.Getenv(envGoogleClientID),
			ClientSecret: os.Getenv(envGoogleClientSecret),
			RedirectURL:  os.Getenv(envGoogleRedirectURL),
			Scopes:       []string{scopeEmail, scopeProfile},
			Endpoint:     google.Endpoint,
		},
	}
}

func (m *GoogleManager) GetAuthURL(state string) string {
	return m.config.AuthCodeURL(state)
}

func (m *GoogleManager) GetUserInfo(ctx context.Context, code string) (*GoogleUser, error) {
	// Exchange code for token
	token, err := m.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	// Fetch user info from Google
	resp, err := http.Get(googleUserInfoURL + "?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer resp.Body.Close()

	var gUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		return nil, err
	}

	return &gUser, nil
}