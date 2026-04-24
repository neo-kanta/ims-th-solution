package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// UserClaims represents the minimal JWT claims for an authenticated user.
type UserClaims struct {
	SessionID  string `json:"sid,omitempty"`
	Restricted bool   `json:"rst,omitempty"`
	jwt.RegisteredClaims
}

// UserStatusChecker verifies if a user is still active and permitted to use their session.
type UserStatusChecker interface {
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// SessionAccessChecker verifies that the access token is still backed by an active server-side session.
type SessionAccessChecker interface {
	ValidateActiveSession(ctx context.Context, userID string, sessionID string) (bool, error)
	TouchSessionActivity(ctx context.Context, sessionID string) error
}

// KeyProvider resolves JWT signing keys, supporting key rotation via kid header.
type KeyProvider interface {
	ResolveKey(kid string) ([]byte, error)
}

// staticKeyProvider wraps a single secret for backwards compatibility.
type staticKeyProvider struct {
	secret []byte
}

func (s *staticKeyProvider) ResolveKey(_ string) ([]byte, error) {
	return s.secret, nil
}

// multiKeyProvider supports active + previous keys.
type multiKeyProvider struct {
	activeKID    string
	activeSecret []byte
	prevSecret   []byte
}

func (m *multiKeyProvider) ResolveKey(kid string) ([]byte, error) {
	if kid == "" || kid == m.activeKID {
		return m.activeSecret, nil
	}
	if m.prevSecret != nil {
		return m.prevSecret, nil
	}
	return m.activeSecret, nil
}

// NewKeyProvider creates a key provider. If previousSecret is empty, single-key mode.
func NewKeyProvider(activeKID string, activeSecret string, previousSecret string) KeyProvider {
	if previousSecret == "" {
		return &staticKeyProvider{secret: []byte(activeSecret)}
	}
	return &multiKeyProvider{
		activeKID:    activeKID,
		activeSecret: []byte(activeSecret),
		prevSecret:   []byte(previousSecret),
	}
}

// Auth returns a middleware that validates JWT tokens with iss/aud enforcement and checks active status.
func Auth(keyProvider KeyProvider, checker UserStatusChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims := &UserClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				kid, _ := token.Header["kid"].(string)
				return keyProvider.ResolveKey(kid)
			},
				jwt.WithIssuer("ims-th"),
				jwt.WithAudience("ims-th-api"),
				jwt.WithExpirationRequired(),
			)

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			if checker != nil {
				isActive, err := checker.IsUserActive(r.Context(), claims.Subject)
				if err != nil || !isActive {
					http.Error(w, `{"error":"account suspended or deleted"}`, http.StatusForbidden)
					return
				}
			}

			if sessionChecker, ok := checker.(SessionAccessChecker); ok && claims.SessionID != "" {
				isActive, err := sessionChecker.ValidateActiveSession(r.Context(), claims.Subject, claims.SessionID)
				if err != nil || !isActive {
					http.Error(w, `{"error":"session revoked or expired"}`, http.StatusUnauthorized)
					return
				}
				if err := sessionChecker.TouchSessionActivity(r.Context(), claims.SessionID); err != nil {
					http.Error(w, `{"error":"failed to refresh session activity"}`, http.StatusInternalServerError)
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims extracts user claims from the request context.
// Returns nil if no user claims are present.
func GetUserClaims(ctx context.Context) *UserClaims {
	claims, _ := ctx.Value(UserContextKey).(*UserClaims)
	return claims
}
