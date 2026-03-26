package ratelimit

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Config holds rate limit configuration.
type Config struct {
	SignupLimit       int
	LoginLimit        int
	SearchLimit       int
	CommentReplyLimit int
	PublicReadLimit   int
	ProtectedWrite    int
	ProtectedRead     int
}

// DefaultConfig returns the default rate limit configuration.
func DefaultConfig() Config {
	return Config{
		SignupLimit:       5,
		LoginLimit:        10,
		SearchLimit:       60,
		CommentReplyLimit: 30,
		PublicReadLimit:   100,
		ProtectedWrite:    100,
		ProtectedRead:     200,
	}
}

// LoadFromEnv loads rate limit configuration from environment variables.
func LoadFromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("RATE_LIMIT_SIGNUP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SignupLimit = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_LOGIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.LoginLimit = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_SEARCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SearchLimit = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_COMMENT_REPLY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.CommentReplyLimit = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_PUBLIC_READ"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.PublicReadLimit = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_PROTECTED_WRITE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.ProtectedWrite = n
		}
	}
	if v := os.Getenv("RATE_LIMIT_PROTECTED_READ"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.ProtectedRead = n
		}
	}

	return cfg
}

// Window constants for rate limiting.
const (
	HourWindow = time.Hour
)

// NewRedisClient creates a Redis client from environment configuration.
func NewRedisClient() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// Fallback to simple address if not a URL
		opt = &redis.Options{
			Addr: redisURL,
		}
	}

	return redis.NewClient(opt)
}

// GetClientIP extracts the client IP from the request.
// It checks X-Forwarded-For, X-Real-IP, then falls back to ClientIP.
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For first (common for proxies/load balancers)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs; take the first one
		if first, _, found := strings.Cut(xff, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP (common for nginx)
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to Gin's ClientIP
	return c.ClientIP()
}

// KeyFunc is a function that extracts a rate limit key from a request.
type KeyFunc func(c *gin.Context) string

// IPKey returns a key function that uses the client IP.
func IPKey(prefix string) KeyFunc {
	return func(c *gin.Context) string {
		return prefix + ":" + GetClientIP(c)
	}
}

// UserKey returns a key function that uses the authenticated user ID.
// Falls back to IP if user is not authenticated.
func UserKey(prefix string) KeyFunc {
	return func(c *gin.Context) string {
		if userID, exists := c.Get("user_id"); exists {
			return prefix + ":user:" + strconv.Itoa(int(userID.(int32)))
		}
		// Fall back to IP if not authenticated
		return prefix + ":ip:" + GetClientIP(c)
	}
}
