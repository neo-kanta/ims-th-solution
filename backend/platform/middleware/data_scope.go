package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// DataPermissionChecker verifies access to a contract/fund or other data scope.
type DataPermissionChecker interface {
	HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error)
}

// RequireDataPermission enforces data-scope access using a URL parameter or query parameter.
func RequireDataPermission(checker DataPermissionChecker, paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				httputil.Unauthorized(w, "not authenticated")
				return
			}

			scopeID := chi.URLParam(r, paramName)
			if scopeID == "" {
				scopeID = r.URL.Query().Get(paramName)
			}
			if scopeID == "" {
				httputil.BadRequest(w, "missing required data scope")
				return
			}

			allowed, err := checker.HasDataPermission(r.Context(), claims.Subject, scopeID)
			if err != nil {
				httputil.InternalError(w, "failed to verify data permissions")
				return
			}
			if !allowed {
				httputil.Forbidden(w, "insufficient data scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
