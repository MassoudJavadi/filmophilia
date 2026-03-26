package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds essential security headers to all responses.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip restrictive CSP for Swagger UI (needs to load its own assets)
		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable HSTS (1 year, include subdomains, allow preload)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy - restrictive default for API
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		c.Next()
	}
}
