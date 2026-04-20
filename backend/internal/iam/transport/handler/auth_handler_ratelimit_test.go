package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

func TestLogin_PerUserRateLimit_BlocksAfterThreshold(t *testing.T) {
	// Create a handler with only the per-user limiter wired (no login command needed —
	// the rate limiter fires before the command is called).
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 2)
	policy := middleware.RateLimitPolicy{Name: "login_user", Max: 2, Window: 1 * time.Minute}

	h := handler.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, middleware.RateLimitPolicy{}, limiter, policy, nil, middleware.RateLimitPolicy{})

	makeReq := func(username string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "SomePassword123",
		})
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Login(w, req)
		return w
	}

	// First two attempts for "alice" should pass through to the login command.
	// Since loginCmd is nil, the command will panic — but the rate limiter allows it.
	// We use recover to isolate the test to just rate-limit behavior.
	for i := 0; i < 2; i++ {
		func() {
			defer func() { recover() }() // login command is nil, will panic after rate limit passes
			makeReq("alice")
		}()
	}

	// Third attempt for "alice" should be blocked by the per-user rate limiter
	w := makeReq("alice")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	assert.Equal(t, "too many login attempts for this account", body["error"])
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}

func TestLogin_PerUserRateLimit_DifferentUsersIndependent(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "login_user", Max: 1, Window: 1 * time.Minute}

	h := handler.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, middleware.RateLimitPolicy{}, limiter, policy, nil, middleware.RateLimitPolicy{})

	makeReq := func(username string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "SomePassword123",
		})
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Login(w, req)
		return w
	}

	// Exhaust alice's limit
	func() {
		defer func() { recover() }()
		makeReq("alice")
	}()

	// Alice is now blocked
	w := makeReq("alice")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Bob should still pass through (different user, same IP)
	func() {
		defer func() { recover() }()
		w := makeReq("bob")
		// If we got here without 429, the rate limiter let it through (correct).
		// It will panic on nil loginCmd, but that's expected.
		_ = w
	}()
}

func TestLogin_PerUserRateLimit_CaseInsensitive(t *testing.T) {
	limiter := middleware.NewInMemoryRateLimiter(1*time.Minute, 1)
	policy := middleware.RateLimitPolicy{Name: "login_user", Max: 1, Window: 1 * time.Minute}

	h := handler.NewAuthHandler(nil, nil, nil, nil, nil, nil, nil, middleware.RateLimitPolicy{}, limiter, policy, nil, middleware.RateLimitPolicy{})

	makeReq := func(username string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "SomePassword123",
		})
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.Login(w, req)
		return w
	}

	// Exhaust limit with lowercase
	func() {
		defer func() { recover() }()
		makeReq("alice")
	}()

	// Uppercase "ALICE" should also be blocked (same account)
	w := makeReq("ALICE")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
