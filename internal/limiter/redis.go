package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(redisURL string) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	client := redis.NewClient(opts)

	// Ping to ensure connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RateLimiter{client: client}, nil
}

// Allow checks if the given apiKey exceeds the rpm (Requests Per Minute) limit.
// Uses a basic Redis-backed counter with 1-minute expiration as a simple Token Bucket approximation.
func (rl *RateLimiter) Allow(ctx context.Context, apiKey string, rpm int) (bool, error) {
	if rpm <= 0 {
		return false, nil // 0 means blocked
	}

	key := fmt.Sprintf("rate_limit:%s", apiKey)

	// Increment the counter
	val, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// If this is the first request in the window, set expiration to 60 seconds
	if val == 1 {
		rl.client.Expire(ctx, key, time.Minute)
	}

	if int(val) > rpm {
		return false, nil // Limit exceeded
	}

	return true, nil
}
