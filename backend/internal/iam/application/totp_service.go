package application

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

const (
	totpPeriod      = 30 // seconds
	totpDigits      = 6  // OTP digit count
	totpSecretSize  = 20 // bytes (160-bit, standard for SHA1 TOTP)
	totpSkew        = 1  // allow +/- 1 time step for clock drift
	recoveryCodeLen = 8  // characters per recovery code
)

// TOTPService handles TOTP secret generation, OTP validation, and recovery codes.
type TOTPService struct {
	encKey     []byte // AES-256 encryption key (32 bytes)
	issuerName string
}

// NewTOTPService creates a TOTP service with the given AES-256 encryption key (hex-encoded).
func NewTOTPService(encKeyHex string, issuerName string) (*TOTPService, error) {
	key, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid MFA encryption key hex: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("MFA encryption key must be 32 bytes (64 hex chars), got %d", len(key))
	}
	return &TOTPService{encKey: key, issuerName: issuerName}, nil
}

// GenerateSecret creates a new TOTP secret and returns it encrypted + the provisioning URI.
func (s *TOTPService) GenerateSecret(username string) (encryptedSecret string, provisioningURI string, err error) {
	secret := make([]byte, totpSecretSize)
	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return "", "", fmt.Errorf("generating TOTP secret: %w", err)
	}

	// Encrypt the secret for storage
	encrypted, err := s.encrypt(secret)
	if err != nil {
		return "", "", fmt.Errorf("encrypting TOTP secret: %w", err)
	}

	// Build provisioning URI (otpauth://totp/...)
	b32Secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		s.issuerName, username, b32Secret, s.issuerName, totpDigits, totpPeriod)

	return encrypted, uri, nil
}

// ValidateCode validates a TOTP code against an encrypted secret.
func (s *TOTPService) ValidateCode(encryptedSecret string, code string) (bool, error) {
	secret, err := s.decrypt(encryptedSecret)
	if err != nil {
		return false, fmt.Errorf("decrypting TOTP secret: %w", err)
	}

	now := time.Now().UTC()
	counter := uint64(now.Unix()) / uint64(totpPeriod)

	// Check current time step and neighbors (for clock skew)
	for i := -totpSkew; i <= totpSkew; i++ {
		expected := generateTOTP(secret, counter+uint64(i))
		if expected == code {
			return true, nil
		}
	}

	return false, nil
}

// GenerateCurrentCode returns the current TOTP code for a stored encrypted secret.
// This is intended for development/test tooling and must not be exposed in production flows.
func (s *TOTPService) GenerateCurrentCode(encryptedSecret string) (string, error) {
	secret, err := s.decrypt(encryptedSecret)
	if err != nil {
		return "", fmt.Errorf("decrypting TOTP secret: %w", err)
	}

	now := time.Now().UTC()
	counter := uint64(now.Unix()) / uint64(totpPeriod)
	return generateTOTP(secret, counter), nil
}

// GenerateRecoveryCodes creates a set of single-use recovery codes.
// Returns the plaintext codes (to show the user) and hashed codes (for storage).
func (s *TOTPService) GenerateRecoveryCodes(userID uuid.UUID) (plaintextCodes []string, hashedCodes []entity.MFARecoveryCode, err error) {
	for i := 0; i < entity.RecoveryCodeCount; i++ {
		code, err := generateSecureCode(recoveryCodeLen)
		if err != nil {
			return nil, nil, fmt.Errorf("generating recovery code: %w", err)
		}
		plaintextCodes = append(plaintextCodes, code)
		hashedCodes = append(hashedCodes, entity.MFARecoveryCode{
			ID:       uuid.New(),
			UserID:   userID,
			CodeHash: hashRecoveryCode(code),
		})
	}
	return plaintextCodes, hashedCodes, nil
}

// VerifyRecoveryCode checks a plaintext code against a list of hashed codes.
// Returns the matching code entity if found.
func (s *TOTPService) VerifyRecoveryCode(code string, storedCodes []entity.MFARecoveryCode) *entity.MFARecoveryCode {
	hash := hashRecoveryCode(code)
	for i := range storedCodes {
		if !storedCodes[i].IsUsed && storedCodes[i].CodeHash == hash {
			return &storedCodes[i]
		}
	}
	return nil
}

// encrypt uses AES-256-GCM to encrypt data.
func (s *TOTPService) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// decrypt uses AES-256-GCM to decrypt data.
func (s *TOTPService) decrypt(ciphertextHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// generateTOTP generates a TOTP value per RFC 6238 / RFC 4226.
func generateTOTP(secret []byte, counter uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	code = code % uint32(math.Pow10(totpDigits))

	return fmt.Sprintf("%0*d", totpDigits, code)
}

// generateSecureCode creates a random alphanumeric code.
func generateSecureCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// hashRecoveryCode returns a SHA-256 hex hash of a recovery code (case-insensitive).
func hashRecoveryCode(code string) string {
	h := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(code))))
	return hex.EncodeToString(h[:])
}
