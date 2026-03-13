package valueobject

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(plaintext string) (string, error) {
	if len(plaintext) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil if they match, or an error if they don't.
func CheckPassword(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
