package middleware

import (
	"context"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// PermissionChecker interface to verify explicit function codes.
type PermissionChecker interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
}

// RequirePermission enforces a function permission code on a route.
func RequirePermission(checker PermissionChecker, code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				httputil.Unauthorized(w, "not authenticated")
				return
			}

			hasPerm, err := checker.HasFunctionPermission(r.Context(), claims.Subject, code)
			if err != nil {
				httputil.InternalError(w, "failed to verify permissions")
				return
			}
			if !hasPerm {
				httputil.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
