package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// rateLimitLua atomically increments a counter and sets TTL on first increment.
// Returns {count, ttl} where ttl is seconds remaining until the key expires.
var rateLimitLua = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('TTL', KEYS[1])
return {count, ttl}
`)

// RedisRateLimiter is a fixed-window, distributed rate limiter backed by Redis.
// Uses an atomic Lua script for INCR+EXPIRE to prevent race conditions.
// Suitable for multi-instance production deployments.
type RedisRateLimiter struct {
	client  *redis.Client
	window  time.Duration
	maxReqs int
}

// NewRedisRateLimiter creates a Redis-backed rate limiter.
func NewRedisRateLimiter(client *redis.Client, window time.Duration, maxReqs int) RateLimiter {
	return &RedisRateLimiter{
		client:  client,
		window:  window,
		maxReqs: maxReqs,
	}
}

// Allow checks if a request is allowed and atomically increments the counter.
// Returns (allowed, currentCount, retryAfterSeconds, error).
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, int, int, error) {
	windowSec := int(r.window.Seconds())
	if windowSec < 1 {
		windowSec = 1
	}

	result, err := rateLimitLua.Run(ctx, r.client, []string{key}, windowSec).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("redis rate limit script: %w", err)
	}

	vals := result.([]interface{})
	count := int(vals[0].(int64))
	ttl := int(vals[1].(int64))
	if ttl < 0 {
		ttl = windowSec
	}

	return count <= r.maxReqs, count, ttl, nil
}

// GetAttempts returns the current counter value for a key.
func (r *RedisRateLimiter) GetAttempts(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("redis GET: %w", err)
	}
	return val, nil
}

// Reset clears the counter for a key.
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL: %w", err)
	}
	return nil
}
