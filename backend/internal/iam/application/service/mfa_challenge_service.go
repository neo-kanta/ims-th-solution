package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

const mfaChallengePurpose = "iam_mfa_login"

// MFAChallengeClaims are the claims for a short-lived MFA login challenge.
type MFAChallengeClaims struct {
	UserID   string `json:"uid"`
	Username string `json:"username"`
	Purpose  string `json:"purpose"`
	jwt.RegisteredClaims
}

// MFAChallengeService issues and validates short-lived MFA login challenges.
type MFAChallengeService struct {
	secret []byte
	clock  clock.Clock
	ttl    time.Duration
}

// NewMFAChallengeService creates a new MFA challenge service.
func NewMFAChallengeService(secret string, clk clock.Clock) *MFAChallengeService {
	return &MFAChallengeService{
		secret: []byte(secret),
		clock:  clk,
		ttl:    5 * time.Minute,
	}
}

// Issue creates a short-lived challenge token that binds the MFA step to a prior password check.
func (s *MFAChallengeService) Issue(userID uuid.UUID, username string) (string, time.Time, error) {
	now := s.clock.Now()
	expiresAt := now.Add(s.ttl)
	claims := &MFAChallengeClaims{
		UserID:   userID.String(),
		Username: username,
		Purpose:  mfaChallengePurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ims-th",
			Audience:  jwt.ClaimStrings{"ims-th-mfa"},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "mfa"

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing MFA challenge token: %w", err)
	}
	return signed, expiresAt, nil
}

// Validate ensures the MFA challenge is authentic, unexpired, and belongs to the same user.
func (s *MFAChallengeService) Validate(tokenString string, expectedUserID uuid.UUID, expectedUsername string) error {
	claims := &MFAChallengeClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	},
		jwt.WithIssuer("ims-th"),
		jwt.WithAudience("ims-th-mfa"),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return fmt.Errorf("validating MFA challenge token: %w", err)
	}
	if !token.Valid {
		return fmt.Errorf("invalid MFA challenge token")
	}
	if claims.Purpose != mfaChallengePurpose {
		return fmt.Errorf("invalid MFA challenge purpose")
	}
	if claims.UserID != expectedUserID.String() || claims.Username != expectedUsername {
		return fmt.Errorf("MFA challenge does not match the authenticated user")
	}
	return nil
}
