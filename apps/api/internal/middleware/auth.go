package middleware

import (
	"net/http"
	"strings"

	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/MassoudJavadi/filmophilia/api/internal/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwt *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.AbortError(c, http.StatusUnauthorized, "AUTH_HEADER_REQUIRED", "authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.AbortError(c, http.StatusUnauthorized, "INVALID_AUTH_HEADER", "invalid authorization header format")
			return
		}

		claims, err := jwt.Verify(parts[1])
		if err != nil {
			response.AbortError(c, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token")
			return
		}

		c.Set("user_id", int32(claims["sub"].(float64)))
		c.Set("user_role", claims["role"].(string))

		c.Next()
	}
}
