package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// redisTestClient returns a connected Redis client or skips the test.
// Set REDIS_TEST_ADDR to run these tests (e.g., REDIS_TEST_ADDR=localhost:6379).
func redisTestClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR not set; skipping Redis integration test")
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not reachable at %s: %v", addr, err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// flushTestKeys deletes test keys to prevent interference between tests.
func flushTestKeys(t *testing.T, client *redis.Client, keys ...string) {
	t.Helper()
	ctx := context.Background()
	for _, k := range keys {
		client.Del(ctx, k)
	}
}

func TestRedisRateLimiter_AllowsWithinLimit(t *testing.T) {
	client := redisTestClient(t)
	key := "test:redis:allow_within:" + t.Name()
	flushTestKeys(t, client, key)
	t.Cleanup(func() { flushTestKeys(t, client, key) })

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 3)

	for i := 0; i < 3; i++ {
		allowed, count, _, err := limiter.Allow(context.Background(), key)
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
		assert.Equal(t, i+1, count)
	}
}

func TestRedisRateLimiter_BlocksOverLimit(t *testing.T) {
	client := redisTestClient(t)
	key := "test:redis:block_over:" + t.Name()
	flushTestKeys(t, client, key)
	t.Cleanup(func() { flushTestKeys(t, client, key) })

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 2)

	_, _, _, _ = limiter.Allow(context.Background(), key)
	_, _, _, _ = limiter.Allow(context.Background(), key)

	allowed, count, _, err := limiter.Allow(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 3, count)
}

func TestRedisRateLimiter_SeparateKeys(t *testing.T) {
	client := redisTestClient(t)
	key1 := "test:redis:sep1:" + t.Name()
	key2 := "test:redis:sep2:" + t.Name()
	flushTestKeys(t, client, key1, key2)
	t.Cleanup(func() { flushTestKeys(t, client, key1, key2) })

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 1)

	allowed1, _, _, _ := limiter.Allow(context.Background(), key1)
	allowed2, _, _, _ := limiter.Allow(context.Background(), key2)

	assert.True(t, allowed1)
	assert.True(t, allowed2, "different keys should have independent limits")
}

func TestRedisRateLimiter_GetAttempts(t *testing.T) {
	client := redisTestClient(t)
	key := "test:redis:attempts:" + t.Name()
	flushTestKeys(t, client, key)
	t.Cleanup(func() { flushTestKeys(t, client, key) })

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 10)

	_, _, _, _ = limiter.Allow(context.Background(), key)
	_, _, _, _ = limiter.Allow(context.Background(), key)

	count, err := limiter.GetAttempts(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = limiter.GetAttempts(context.Background(), "test:redis:nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRedisRateLimiter_Reset(t *testing.T) {
	client := redisTestClient(t)
	key := "test:redis:reset:" + t.Name()
	flushTestKeys(t, client, key)
	t.Cleanup(func() { flushTestKeys(t, client, key) })

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 1)

	_, _, _, _ = limiter.Allow(context.Background(), key)
	err := limiter.Reset(context.Background(), key)
	require.NoError(t, err)

	allowed, count, _, _ := limiter.Allow(context.Background(), key)
	assert.True(t, allowed)
	assert.Equal(t, 1, count)
}

func TestRedisRateLimiter_TTLExpiry(t *testing.T) {
	client := redisTestClient(t)
	key := "test:redis:ttl:" + t.Name()
	flushTestKeys(t, client, key)
	t.Cleanup(func() { flushTestKeys(t, client, key) })

	// Use a very short window
	limiter := middleware.NewRedisRateLimiter(client, 1*time.Second, 1)

	allowed, _, _, err := limiter.Allow(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Second request should be blocked
	allowed, _, _, err = limiter.Allow(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for TTL to expire
	time.Sleep(1200 * time.Millisecond)

	// Should be allowed again
	allowed, count, _, err := limiter.Allow(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 1, count)
}

func TestRedisRateLimiter_MiddlewareIntegration(t *testing.T) {
	client := redisTestClient(t)
	// Use unique prefix to avoid collision
	prefix := "test:redis:mw:" + t.Name()
	ctx := context.Background()
	// Pre-clean any keys with this prefix
	iter := client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		client.Del(ctx, iter.Val())
	}

	limiter := middleware.NewRedisRateLimiter(client, 1*time.Minute, 2)
	policy := middleware.RateLimitPolicy{Name: prefix, Max: 2, Window: 1 * time.Minute}
	handler := middleware.RateLimit(limiter, policy)(newOKHandler())

	// First two requests should pass
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d", i+1)
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
