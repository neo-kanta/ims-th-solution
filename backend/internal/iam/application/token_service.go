package application

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/your-org/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/your-org/ims-th-solution/backend/platform/middleware"
)

const (
	// DefaultTokenExpiry is the default JWT token validity duration.
	DefaultTokenExpiry = 8 * time.Hour

	// RefreshTokenExpiry is the refresh window — tokens can be refreshed within this period.
	RefreshTokenExpiry = 24 * time.Hour
)

// TokenService handles JWT token generation and validation for authenticated users.
type TokenService struct {
	secret []byte
	expiry time.Duration
}

// NewTokenService creates a new token service with the given signing secret.
func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		expiry: DefaultTokenExpiry,
	}
}

// GenerateToken creates a signed JWT for the given user.
// The token includes user identity and group membership as claims.
func (s *TokenService) GenerateToken(user *entity.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.expiry)

	claims := &middleware.UserClaims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Groups:      user.Groups,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}

	return signed, expiresAt, nil
}

// RefreshToken generates a new token for an existing authenticated user.
// This extends the session without requiring re-authentication.
func (s *TokenService) RefreshToken(user *entity.User) (string, time.Time, error) {
	return s.GenerateToken(user)
}
