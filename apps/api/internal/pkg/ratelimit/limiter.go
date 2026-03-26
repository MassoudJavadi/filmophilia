package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Result contains the result of a rate limit check.
type Result struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// RetryAfter returns the number of seconds until the rate limit resets.
func (r Result) RetryAfter() int {
	if r.Allowed {
		return 0
	}
	seconds := int(time.Until(r.ResetAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

// RateLimiter provides rate limiting functionality backed by Redis.
type RateLimiter struct {
	store *Store
}

// NewRateLimiter creates a new rate limiter with Redis backend.
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		store: NewStore(client),
	}
}

// Check checks if a request is allowed for the given key within the limit and window.
func (rl *RateLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) Result {
	allowed, remaining, resetAt := rl.store.Allow(ctx, key, limit, window)
	return Result{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}
}
