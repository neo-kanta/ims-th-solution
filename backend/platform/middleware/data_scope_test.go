package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type mockDataPermissionChecker struct {
	allowed bool
	err     error
}

func (m *mockDataPermissionChecker) HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error) {
	return m.allowed, m.err
}

func TestRequireDataPermission(t *testing.T) {
	tests := []struct {
		name           string
		claims         *middleware.UserClaims
		scopeID        string
		allowed        bool
		expectedStatus int
	}{
		{name: "not authenticated", claims: nil, scopeID: "FUND-1", allowed: false, expectedStatus: http.StatusUnauthorized},
		{name: "missing scope", claims: &middleware.UserClaims{}, scopeID: "", allowed: false, expectedStatus: http.StatusBadRequest},
		{name: "forbidden", claims: &middleware.UserClaims{RegisteredClaims: middleware.UserClaims{}.RegisteredClaims}, scopeID: "FUND-1", allowed: false, expectedStatus: http.StatusForbidden},
		{name: "allowed", claims: &middleware.UserClaims{}, scopeID: "FUND-1", allowed: true, expectedStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockDataPermissionChecker{allowed: tt.allowed}
			handler := middleware.RequireDataPermission(checker, "contract_id")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/contracts/"+tt.scopeID, nil)
			if tt.scopeID != "" {
				routeCtx := chi.NewRouteContext()
				routeCtx.URLParams.Add("contract_id", tt.scopeID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			}
			if tt.claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, tt.claims))
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
