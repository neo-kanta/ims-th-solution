package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// UserContextKey is the context key for the authenticated user.
	UserContextKey contextKey = "user"
)

// UserClaims represents the JWT claims for an authenticated user.
type UserClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
	jwt.RegisteredClaims
}

// Auth returns a middleware that validates JWT tokens from the Authorization header.
// It extracts user claims and stores them in the request context.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
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
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Store user claims in context for downstream handlers
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
