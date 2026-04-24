package application_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
)

const testEncKeyHex = "0000000000000000000000000000000000000000000000000000000000000000"

func TestTOTPService_GenerateSecret(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	encrypted, uri, err := svc.GenerateSecret("testuser")
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.Contains(t, uri, "otpauth://totp/TestIMS:testuser")
	assert.Contains(t, uri, "secret=")
	assert.Contains(t, uri, "algorithm=SHA1")
	assert.Contains(t, uri, "digits=6")
}

func TestTOTPService_InvalidEncryptionKey(t *testing.T) {
	_, err := application.NewTOTPService("not-hex", "Test")
	assert.Error(t, err)

	_, err = application.NewTOTPService("0000", "Test")
	assert.Error(t, err)
}

func TestTOTPService_ValidateInvalidCode(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	encrypted, _, err := svc.GenerateSecret("testuser")
	require.NoError(t, err)

	// An obviously wrong code format
	valid, err := svc.ValidateCode(encrypted, "abcdef")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestTOTPService_RecoveryCodes(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	userID := uuid.New()
	plaintextCodes, hashedCodes, err := svc.GenerateRecoveryCodes(userID)
	require.NoError(t, err)
	assert.Len(t, plaintextCodes, 10)
	assert.Len(t, hashedCodes, 10)

	// Each code should be 8 chars
	for _, code := range plaintextCodes {
		assert.Len(t, code, 8)
	}

	// Verify codes match
	for i, code := range plaintextCodes {
		assert.Equal(t, userID, hashedCodes[i].UserID)
		matched := svc.VerifyRecoveryCode(code, hashedCodes)
		assert.NotNil(t, matched, "recovery code %d should match", i)
	}

	// Invalid code should not match
	matched := svc.VerifyRecoveryCode("invalid1", hashedCodes)
	assert.Nil(t, matched)
}

func TestTOTPService_RecoveryCodeCaseInsensitive(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	userID := uuid.New()
	plaintextCodes, hashedCodes, err := svc.GenerateRecoveryCodes(userID)
	require.NoError(t, err)

	upperCode := strings.ToUpper(plaintextCodes[0])
	matched := svc.VerifyRecoveryCode(upperCode, hashedCodes)
	assert.NotNil(t, matched, "uppercase recovery code should match")
}

func TestTOTPService_UsedRecoveryCodeDoesNotMatch(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	userID := uuid.New()
	plaintextCodes, hashedCodes, err := svc.GenerateRecoveryCodes(userID)
	require.NoError(t, err)

	// Mark first code as used
	hashedCodes[0].IsUsed = true

	matched := svc.VerifyRecoveryCode(plaintextCodes[0], hashedCodes)
	assert.Nil(t, matched, "used recovery code should not match")
}

func TestTOTPService_GenerateCurrentCode(t *testing.T) {
	svc, err := application.NewTOTPService(testEncKeyHex, "TestIMS")
	require.NoError(t, err)

	encrypted, _, err := svc.GenerateSecret("testuser")
	require.NoError(t, err)

	code, err := svc.GenerateCurrentCode(encrypted)
	require.NoError(t, err)
	assert.Len(t, code, 6)

	valid, err := svc.ValidateCode(encrypted, code)
	require.NoError(t, err)
	assert.True(t, valid)
}
