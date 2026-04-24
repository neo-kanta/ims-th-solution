package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type mockPermissionChecker struct {
	hasPerm bool
	err     error
}

func (m *mockPermissionChecker) HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error) {
	return m.hasPerm, m.err
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		claims         *middleware.UserClaims
		mockHasPerm    bool
		expectedStatus int
	}{
		{
			name:           "Not Authenticated",
			claims:         nil,
			mockHasPerm:    false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Sufficient Permissions",
			claims:         &middleware.UserClaims{},
			mockHasPerm:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Insufficient Permissions",
			claims:         &middleware.UserClaims{},
			mockHasPerm:    false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Restricted Session",
			claims:         &middleware.UserClaims{Restricted: true},
			mockHasPerm:    true,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockPermissionChecker{hasPerm: tt.mockHasPerm}
			mw := middleware.RequirePermission(checker, "TEST_CODE")

			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.claims != nil {
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, tt.claims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Result().StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Result().StatusCode)
			}
		})
	}
}
