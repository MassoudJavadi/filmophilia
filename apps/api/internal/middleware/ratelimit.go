package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates a rate limiting middleware.
func RateLimitMiddleware(limiter *ratelimit.RateLimiter, limit int, window time.Duration, keyFunc ratelimit.KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		result := limiter.Check(c.Request.Context(), key, limit, window)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			retryAfter := result.RetryAfter()
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}

		c.Next()
	}
}
