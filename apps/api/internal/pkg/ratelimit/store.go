package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store is a Redis-backed store for rate limit counters.
type Store struct {
	client *redis.Client
}

// NewStore creates a new Redis-backed rate limit store.
func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

// Allow checks if a request is allowed and updates the counter.
// Returns whether the request is allowed, remaining requests, and when the window resets.
func (s *Store) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time) {
	now := time.Now()
	windowSeconds := int64(window.Seconds())
	currentWindowID := now.Unix() / windowSeconds
	resetAt = time.Unix((currentWindowID+1)*windowSeconds, 0)

	// Build Redis key with window ID to ensure automatic expiry alignment
	redisKey := fmt.Sprintf("ratelimit:%s:%d", key, currentWindowID)

	// Use INCR to atomically increment the counter
	count, err := s.client.Incr(ctx, redisKey).Result()
	if err != nil {
		// On Redis error, allow the request (fail open)
		return true, limit - 1, resetAt
	}

	// Set expiry on first request in window
	if count == 1 {
		s.client.Expire(ctx, redisKey, window+time.Second)
	}

	if int(count) > limit {
		return false, 0, resetAt
	}

	return true, limit - int(count), resetAt
}
