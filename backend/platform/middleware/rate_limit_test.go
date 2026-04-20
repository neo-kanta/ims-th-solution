package middleware_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// ---------- InMemoryRateLimiter unit tests ----------

func TestInMemoryRateLimiter_AllowsWithinLimit(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 3)

	for i := 0; i < 3; i++ {
		allowed, count, _, err := limiter.Allow(context.Background(), "key1")
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
		assert.Equal(t, i+1, count)
	}
}

func TestInMemoryRateLimiter_BlocksOverLimit(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 2)

	// Use up the limit
	_, _, _, _ = limiter.Allow(context.Background(), "key1")
	_, _, _, _ = limiter.Allow(context.Background(), "key1")

	// Third request should be blocked
	allowed, count, _, err := limiter.Allow(context.Background(), "key1")
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 3, count)
}

func TestInMemoryRateLimiter_SeparateKeys(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)

	allowed1, _, _, _ := limiter.Allow(context.Background(), "key1")
	allowed2, _, _, _ := limiter.Allow(context.Background(), "key2")

	assert.True(t, allowed1)
	assert.True(t, allowed2, "different keys should have independent limits")
}

func TestInMemoryRateLimiter_GetAttempts(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 10)

	_, _, _, _ = limiter.Allow(context.Background(), "key1")
	_, _, _, _ = limiter.Allow(context.Background(), "key1")

	count, err := limiter.GetAttempts(context.Background(), "key1")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = limiter.GetAttempts(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestInMemoryRateLimiter_Reset(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)

	_, _, _, _ = limiter.Allow(context.Background(), "key1")
	err := limiter.Reset(context.Background(), "key1")
	require.NoError(t, err)

	// Should be allowed again after reset
	allowed, count, _, _ := limiter.Allow(context.Background(), "key1")
	assert.True(t, allowed)
	assert.Equal(t, 1, count)
}

// ---------- RateLimit middleware tests ----------

func newOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestRateLimit_Returns429WhenExceeded(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 2)
	policy := middleware.RateLimitPolicy{Name: "test", Max: 2, Window: 1 * time.Minute}
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

func TestRateLimit_ReturnsProperHeaders_OnAllowedRequest(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 10)
	policy := middleware.RateLimitPolicy{Name: "test", Max: 10, Window: 1 * time.Minute}
	handler := middleware.RateLimit(limiter, policy)(newOKHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.5.5.5:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "9", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	assert.Equal(t, "test", w.Header().Get("X-RateLimit-Policy"))
}

func TestRateLimit_ReturnsRetryAfter_On429(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(5*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "test", Max: 1, Window: 5 * time.Minute}
	handler := middleware.RateLimit(limiter, policy)(newOKHandler())

	// Exhaust limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "6.6.6.6:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Trigger 429
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "6.6.6.6:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	retryAfterStr := w2.Header().Get("Retry-After")
	retryAfterVal, _ := strconv.Atoi(retryAfterStr)
	assert.True(t, retryAfterVal > 0 && retryAfterVal <= 301, "Retry-After should be between 1 and 301, got %d", retryAfterVal)
	assert.Equal(t, "0", w2.Header().Get("X-RateLimit-Remaining"))

	// Verify response body
	var body map[string]interface{}
	err := json.NewDecoder(w2.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "rate limit exceeded", body["error"])
}

func TestRateLimit_DifferentIPs_IndependentLimits(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "test", Max: 1, Window: 1 * time.Minute}
	handler := middleware.RateLimit(limiter, policy)(newOKHandler())

	// IP A uses its one request
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// IP B should still be allowed
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

// ---------- RateLimitByUser middleware tests ----------

func TestRateLimitByUser_KeysByUserClaims(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "sensitive", Max: 1, Window: 1 * time.Minute}
	handler := middleware.RateLimitByUser(limiter, policy)(newOKHandler())

	// Simulate authenticated request by injecting claims into context
	claims := &middleware.UserClaims{}
	claims.Subject = "user-123"

	req := httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Same user, second request should be blocked
	req2 := httptest.NewRequest("POST", "/", nil)
	req2.RemoteAddr = "10.0.0.2:1234" // different IP, same user
	req2 = req2.WithContext(context.WithValue(req2.Context(), middleware.UserContextKey, claims))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitByUser_FallsBackToIP_WhenUnauthenticated(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "sensitive", Max: 1, Window: 1 * time.Minute}
	handler := middleware.RateLimitByUser(limiter, policy)(newOKHandler())

	// No claims in context
	req := httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Same IP again
	req2 := httptest.NewRequest("POST", "/", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

// ---------- Trusted proxy / GetClientIP tests ----------

func TestGetClientIP_UsesRemoteAddr_WhenNoTrustedProxy(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	ip := middleware.GetClientIP(req)
	assert.Equal(t, "1.2.3.4", ip, "should ignore X-Forwarded-For without trusted proxy")
}

func TestGetClientIP_TrustsXForwardedFor_WhenProxyIsTrusted(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig([]string{"192.168.1.0/24"}))
	defer middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.10:5678"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 192.168.1.10")

	ip := middleware.GetClientIP(req)
	assert.Equal(t, "203.0.113.50", ip, "should use first IP from X-Forwarded-For")
}

func TestGetClientIP_TrustsXRealIP_WhenProxyIsTrusted(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig([]string{"172.16.0.1"}))
	defer middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "172.16.0.1:5678"
	req.Header.Set("X-Real-IP", "203.0.113.99")

	ip := middleware.GetClientIP(req)
	assert.Equal(t, "203.0.113.99", ip)
}

func TestGetClientIP_IgnoresSpoofedHeaders_FromUntrustedIP(t *testing.T) {
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig([]string{"10.0.0.0/8"}))
	defer middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(nil))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.50:5678" // not in 10.0.0.0/8
	req.Header.Set("X-Forwarded-For", "1.1.1.1")

	ip := middleware.GetClientIP(req)
	assert.Equal(t, "203.0.113.50", ip, "should ignore X-Forwarded-For from untrusted source")
}

// ---------- TrustedProxyConfig tests ----------

func TestTrustedProxyConfig_IsTrusted(t *testing.T) {
	config := middleware.NewTrustedProxyConfig([]string{"10.0.0.0/8", "172.16.0.1"})

	tests := []struct {
		ip      string
		trusted bool
	}{
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.16.0.2", false},
		{"1.2.3.4", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			parsed := net.ParseIP(tt.ip)
			assert.Equal(t, tt.trusted, config.IsTrusted(parsed))
		})
	}
}

func TestTrustedProxyConfig_EmptyIsNeverTrusted(t *testing.T) {
	config := middleware.NewTrustedProxyConfig(nil)
	parsed := net.ParseIP("10.0.0.1")
	assert.False(t, config.IsTrusted(parsed))
}

// ---------- RateLimitByKey with custom key function ----------

func TestRateLimitByKey_CustomKeyFunction(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "custom", Max: 1, Window: 1 * time.Minute}

	// Key by a custom header
	keyFunc := func(r *http.Request) string {
		return "custom:" + r.Header.Get("X-API-Key")
	}
	handler := middleware.RateLimitByKey(limiter, policy, keyFunc)(newOKHandler())

	// First request with API key A
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("X-API-Key", "key-a")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request with same API key A — blocked
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-API-Key", "key-a")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// Request with API key B — allowed (different key)
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("X-API-Key", "key-b")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// ---------- Remaining counter accuracy ----------

func TestRateLimit_RemainingCountIsAccurate(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 5)
	policy := middleware.RateLimitPolicy{Name: "test", Max: 5, Window: 1 * time.Minute}
	handler := middleware.RateLimit(limiter, policy)(newOKHandler())

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		remaining, _ := strconv.Atoi(w.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, 5-(i+1), remaining, "remaining after request %d", i+1)
	}
}
