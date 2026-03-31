package valueobject

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// MinPasswordLength per NIST 800-63B Rev 4.
const MinPasswordLength = 8

// weakPasswords is a small built-in blocklist.
// In production, load from a file or external service (e.g., HaveIBeenPwned k-Anonymity API).
var weakPasswords = map[string]struct{}{
	"password":  {},
	"12345678":  {},
	"123456789": {},
	"qwerty123": {},
	"admin123":  {},
	"letmein":   {},
	"welcome1":  {},
	"changeme":  {},
}

// ValidatePasswordStrength checks password against NIST 800-63B rules:
// - Minimum 8 characters
// - No composition rules (no forced uppercase/special chars)
// - Check against blocklist
func ValidatePasswordStrength(plaintext string) error {
	if len(plaintext) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	lower := strings.ToLower(plaintext)
	if _, blocked := weakPasswords[lower]; blocked {
		return fmt.Errorf("this password is too common; please choose a different one")
	}
	return nil
}

// HashPassword hashes a plaintext password using bcrypt with cost 12.
func HashPassword(plaintext string) (string, error) {
	if err := ValidatePasswordStrength(plaintext); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil if they match.
func CheckPassword(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
