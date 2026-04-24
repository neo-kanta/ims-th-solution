package middleware

import (
	"context"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// MFAEnforcer defines the interface for checking MFA requirement.
type MFAEnforcer interface {
	IsMFARequired(ctx context.Context, userID string) (bool, error)
	IsMFAEnabled(ctx context.Context, userID string) (bool, error)
}

// EnforceMFA creates a middleware that requires MFA for sensitive operations.
// Use on routes that handle sensitive operations (investments, approvals, etc.)
func EnforceMFA(enforcer MFAEnforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				httputil.Unauthorized(w, "not authenticated")
				return
			}

			// Check if user has completed MFA setup
			isMFAEnabled, err := enforcer.IsMFAEnabled(r.Context(), claims.Subject)
			if err != nil {
				httputil.InternalError(w, "failed to check MFA status")
				return
			}

			// Check if MFA is required for this user
			isMFARequired, err := enforcer.IsMFARequired(r.Context(), claims.Subject)
			if err != nil {
				httputil.InternalError(w, "failed to check MFA requirement")
				return
			}

			// If MFA is required but not enabled, deny access
			if isMFARequired && !isMFAEnabled {
				httputil.Forbidden(w, "MFA enrollment required")
				return
			}

			// If user is marked as restricted, they must complete MFA before sensitive ops
			if claims.Restricted && !isMFAEnabled {
				httputil.Forbidden(w, "MFA setup required before accessing sensitive operations")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalMFA creates a middleware that checks MFA status but doesn't block.
// Useful for tracking MFA compliance in audit logs.
func OptionalMFA(enforcer MFAEnforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			isMFAEnabled, err := enforcer.IsMFAEnabled(r.Context(), claims.Subject)
			if err == nil {
				ctx := context.WithValue(r.Context(), "mfa_enabled", isMFAEnabled)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
