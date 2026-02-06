package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey string
	issuer    string
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secretKey: secret,
		issuer:    "filmophilia",
	}
}

// Generate creates a new JWT for a specific user
func (m *JWTManager) Generate(userID int32, role string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(duration).Unix(),
		"iat":  time.Now().Unix(),
		"iss":  m.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// Verify validates the token and returns the claims
func (m *JWTManager) Verify(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}