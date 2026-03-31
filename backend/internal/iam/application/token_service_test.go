package application_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

func TestTokenService_GenerateAndValidate(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc := application.NewTokenService("test-secret-key-at-least-32-bytes", clk)

	userID := uuid.New()

	token, expiresAt, err := svc.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, clk.Now().Add(application.AccessTokenExpiry), expiresAt)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, application.TokenIssuer, claims.Issuer)
}

func TestTokenService_InvalidToken(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc := application.NewTokenService("test-secret", clk)

	_, err := svc.ValidateAccessToken("invalid.token.string")
	assert.Error(t, err)
}

func TestTokenService_WrongSecret(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc1 := application.NewTokenService("secret-one", clk)
	svc2 := application.NewTokenService("secret-two", clk)

	userID := uuid.New()
	token, _, err := svc1.GenerateAccessToken(userID)
	require.NoError(t, err)

	_, err = svc2.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestTokenService_KeyRotation(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}

	oldSvc := application.NewTokenServiceWithKeyRing("k1", "old-secret", "", "", clk)
	userID := uuid.New()
	oldToken, _, err := oldSvc.GenerateAccessToken(userID)
	require.NoError(t, err)

	// New service with rotated key, keeping old key for verification
	newSvc := application.NewTokenServiceWithKeyRing("k2", "new-secret", "k1", "old-secret", clk)

	// Old token should still validate via previous key
	claims, err := newSvc.ValidateAccessToken(oldToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)

	// New token should also validate
	newToken, _, err := newSvc.GenerateAccessToken(userID)
	require.NoError(t, err)

	claims2, err := newSvc.ValidateAccessToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims2.Subject)

	assert.Equal(t, "k2", newSvc.ActiveKID())
	assert.Equal(t, []string{"k2", "k1"}, newSvc.AllKeyIDs())
}

func TestHashRefreshToken(t *testing.T) {
	token := "test-refresh-token-value"
	hash1 := application.HashRefreshToken(token)
	hash2 := application.HashRefreshToken(token)

	assert.Equal(t, hash1, hash2, "same token should produce same hash")
	assert.Len(t, hash1, 64, "SHA-256 hex should be 64 chars")

	different := application.HashRefreshToken("different-token")
	assert.NotEqual(t, hash1, different)
}

func TestTokenService_GenerateAccessTokenForSession(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc := application.NewTokenService("test-secret-key-at-least-32-bytes", clk)

	userID := uuid.New()
	sessionID := uuid.New()

	token, _, err := svc.GenerateAccessTokenForSession(userID, sessionID)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, sessionID.String(), claims.SessionID)
}
