package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(5, 1*time.Minute)

	for i := 0; i < 5; i++ {
		blocked := rl.RecordAndCheck("192.168.1.1")
		assert.False(t, blocked, "request %d should not be blocked", i+1)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(3, 1*time.Minute)

	for i := 0; i < 3; i++ {
		rl.RecordAndCheck("192.168.1.1")
	}

	blocked := rl.RecordAndCheck("192.168.1.1")
	assert.True(t, blocked, "4th request should be blocked")
}

func TestRateLimiter_SeparateKeysAreIndependent(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 1*time.Minute)

	rl.RecordAndCheck("192.168.1.1")
	rl.RecordAndCheck("192.168.1.1")

	blocked := rl.RecordAndCheck("192.168.1.2")
	assert.False(t, blocked, "different IP should not be affected")
}

func TestRateLimiter_ResetClearsState(t *testing.T) {
	rl := middleware.NewRateLimiter(2, 1*time.Minute)

	rl.RecordAndCheck("192.168.1.1")
	rl.RecordAndCheck("192.168.1.1")

	rl.Reset("192.168.1.1")

	blocked := rl.RecordAndCheck("192.168.1.1")
	assert.False(t, blocked, "after reset, requests should be allowed again")
}

func TestLoginRateLimit_Returns429WhenBlocked(t *testing.T) {
	rl := middleware.NewRateLimiter(1, 1*time.Minute)

	handler := middleware.LoginRateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request — allowed
	req1 := httptest.NewRequest("POST", "/auth/login", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request — blocked (same RemoteAddr so same key)
	req2 := httptest.NewRequest("POST", "/auth/login", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Equal(t, "60", w2.Header().Get("Retry-After"))
}

func TestLoginRateLimit_UsesXForwardedFor(t *testing.T) {
	rl := middleware.NewRateLimiter(1, 1*time.Minute)

	handler := middleware.LoginRateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with X-Forwarded-For
	req1 := httptest.NewRequest("POST", "/auth/login", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.50")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request same forwarded IP — blocked
	req2 := httptest.NewRequest("POST", "/auth/login", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.50")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Different forwarded IP — allowed
	req3 := httptest.NewRequest("POST", "/auth/login", nil)
	req3.Header.Set("X-Forwarded-For", "203.0.113.51")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}
