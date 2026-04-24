package application

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

const (
	AccessTokenExpiry = 15 * time.Minute
	TokenIssuer       = "ims-th"
	TokenAudience     = "ims-th-api"
)

// SigningKey represents a single key in the key ring.
type SigningKey struct {
	KID    string // key identifier
	Secret []byte // raw key material
}

// AccessTokenClaims are the minimal JWT claims (no permissions, no groups).
type AccessTokenClaims struct {
	SessionID  string `json:"sid,omitempty"`
	Restricted bool   `json:"rst,omitempty"`
	jwt.RegisteredClaims
}

// TokenService handles JWT access token generation and validation.
// Supports key rotation via a key ring: one active signing key + previous keys for verification.
type TokenService struct {
	activeKey    SigningKey
	previousKeys []SigningKey // keys still valid for verification during rotation window
	clock        clock.Clock
}

// NewTokenService creates a token service with a single signing key (backwards compatible).
func NewTokenService(secret string, clk clock.Clock) *TokenService {
	return &TokenService{
		activeKey: SigningKey{KID: "k1", Secret: []byte(secret)},
		clock:     clk,
	}
}

// NewTokenServiceWithKeyRing creates a token service with key rotation support.
// activeKID/activeSecret is the current signing key.
// previousKID/previousSecret (if non-empty) is still accepted for verification during rotation.
func NewTokenServiceWithKeyRing(activeKID string, activeSecret string, previousKID string, previousSecret string, clk clock.Clock) *TokenService {
	ts := &TokenService{
		activeKey: SigningKey{KID: activeKID, Secret: []byte(activeSecret)},
		clock:     clk,
	}
	if previousSecret != "" {
		kid := previousKID
		if kid == "" {
			kid = "prev"
		}
		ts.previousKeys = append(ts.previousKeys, SigningKey{KID: kid, Secret: []byte(previousSecret)})
	}
	return ts
}

// GenerateAccessToken creates a short-lived JWT (15m) with kid header for key rotation.
func (s *TokenService) GenerateAccessToken(userID uuid.UUID) (string, time.Time, error) {
	return s.GenerateAccessTokenForSessionWithRestriction(userID, uuid.Nil, false)
}

// GenerateAccessTokenForSession creates a short-lived JWT bound to a server-side session when provided.
func (s *TokenService) GenerateAccessTokenForSession(userID uuid.UUID, sessionID uuid.UUID) (string, time.Time, error) {
	return s.GenerateAccessTokenForSessionWithRestriction(userID, sessionID, false)
}

// GenerateAccessTokenForSessionWithRestriction creates a short-lived JWT with optional restricted-session semantics.
func (s *TokenService) GenerateAccessTokenForSessionWithRestriction(userID uuid.UUID, sessionID uuid.UUID, restricted bool) (string, time.Time, error) {
	now := s.clock.Now()
	expiresAt := now.Add(AccessTokenExpiry)

	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    TokenIssuer,
			Audience:  jwt.ClaimStrings{TokenAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
	}
	if sessionID != uuid.Nil {
		claims.SessionID = sessionID.String()
	}
	claims.Restricted = restricted

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = s.activeKey.KID

	signed, err := token.SignedString(s.activeKey.Secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}

	return signed, expiresAt, nil
}

// ValidateAccessToken parses and validates a JWT string.
func (s *TokenService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	// Key function that supports rotation: tries active key, then previous keys
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Check kid header if present
		kid, _ := token.Header["kid"].(string)

		if kid == "" || kid == s.activeKey.KID {
			return s.activeKey.Secret, nil
		}

		// Try previous keys by kid match
		for _, pk := range s.previousKeys {
			if kid == pk.KID {
				return pk.Secret, nil
			}
		}

		// Fallback: return nil to signal unknown kid rather than guessing wrong key
		return nil, fmt.Errorf("unknown key ID: %s", kid)
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc,
		jwt.WithIssuer(TokenIssuer),
		jwt.WithAudience(TokenAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ActiveKID returns the current signing key ID.
func (s *TokenService) ActiveKID() string {
	return s.activeKey.KID
}

// AllKeyIDs returns all key IDs (active + previous) for debugging/admin.
func (s *TokenService) AllKeyIDs() []string {
	kids := []string{s.activeKey.KID}
	for _, pk := range s.previousKeys {
		kids = append(kids, pk.KID)
	}
	return kids
}

// GenerateRefreshToken creates a high-entropy cryptographically random string (2x UUIDv4).
func (s *TokenService) GenerateRefreshToken() string {
	return uuid.New().String() + uuid.New().String()
}

// HashRefreshToken returns the SHA-256 hex digest of a refresh token.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
