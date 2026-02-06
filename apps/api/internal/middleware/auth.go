package middleware

import (
	"net/http"
	"strings"

	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/token"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwt *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwt.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", int32(claims["sub"].(float64)))
		c.Set("user_role", claims["role"].(string))

		c.Next()
	}
}
