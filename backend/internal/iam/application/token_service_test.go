package application_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/your-org/ims-th-solution/backend/internal/iam/application"
	"github.com/your-org/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/your-org/ims-th-solution/backend/platform/middleware"
)

func TestTokenService_GenerateToken(t *testing.T) {
	secret := "test-secret-key-12345"
	tokenService := application.NewTokenService(secret)

	user := &entity.User{
		ID:          uuid.New(),
		Username:    "jdoe",
		DisplayName: "John Doe",
		Groups:      []string{"Trader", "Manager"},
	}

	// 1. Generate token
	tokenStr, expiresAt, err := tokenService.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
	assert.True(t, expiresAt.After(time.Now()))

	// 2. Parse and validate token (simulating what middleware.Auth does)
	claims := &middleware.UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	require.NoError(t, err)
	assert.True(t, token.Valid)

	// 3. Verify claims
	assert.Equal(t, user.ID.String(), claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.DisplayName, claims.DisplayName)
	assert.ElementsMatch(t, user.Groups, claims.Groups)
	assert.Equal(t, user.ID.String(), claims.Subject)
}
