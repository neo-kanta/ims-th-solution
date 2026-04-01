package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type mockStatusChecker struct {
	active bool
}

func (m *mockStatusChecker) IsUserActive(ctx context.Context, userID string) (bool, error) {
	return m.active, nil
}

type mockSessionChecker struct {
	mockStatusChecker
	sessionActive bool
	touched       bool
}

func (m *mockSessionChecker) ValidateActiveSession(ctx context.Context, userID string, sessionID string) (bool, error) {
	return m.sessionActive, nil
}

func (m *mockSessionChecker) TouchSessionActivity(ctx context.Context, sessionID string) error {
	m.touched = true
	return nil
}

func TestAuth_RejectsRevokedSession(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	tokenSvc := application.NewTokenService("test-secret", clk)
	userID := uuid.New()
	sessionID := uuid.New()
	token, _, err := tokenSvc.GenerateAccessTokenForSession(userID, sessionID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	keyProvider := middleware.NewKeyProvider("k1", "test-secret", "")
	checker := &mockSessionChecker{
		mockStatusChecker: mockStatusChecker{active: true},
		sessionActive:     false,
	}

	handler := middleware.Auth(keyProvider, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuth_TouchesActiveSession(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	tokenSvc := application.NewTokenService("test-secret", clk)
	userID := uuid.New()
	sessionID := uuid.New()
	token, _, err := tokenSvc.GenerateAccessTokenForSession(userID, sessionID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	keyProvider := middleware.NewKeyProvider("k1", "test-secret", "")
	checker := &mockSessionChecker{
		mockStatusChecker: mockStatusChecker{active: true},
		sessionActive:     true,
	}

	handler := middleware.Auth(keyProvider, checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !checker.touched {
		t.Fatal("expected session activity to be touched")
	}
}
