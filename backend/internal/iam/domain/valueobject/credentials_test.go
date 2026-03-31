package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{"valid password", "MyStr0ngP@ss", false, ""},
		{"too short", "short", true, "at least 8"},
		{"exactly 8 chars", "12345678", true, "too common"},
		{"blocked password", "password", true, "too common"},
		{"blocked case-insensitive", "PASSWORD", true, "too common"},
		{"good 8 char password", "g00dP@ss", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := valueobject.ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPassword_And_CheckPassword(t *testing.T) {
	plaintext := "securePassword123"

	hash, err := valueobject.HashPassword(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, plaintext, hash)

	// Correct password matches
	err = valueobject.CheckPassword(hash, plaintext)
	assert.NoError(t, err)

	// Wrong password does not match
	err = valueobject.CheckPassword(hash, "wrongPassword")
	assert.Error(t, err)
}

func TestHashPassword_RejectsWeakPassword(t *testing.T) {
	_, err := valueobject.HashPassword("short")
	assert.Error(t, err)
}
