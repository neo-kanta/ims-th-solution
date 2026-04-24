package entity

import (
	"time"

	"github.com/google/uuid"
)

// MFAEnrollment represents a user's MFA enrollment (currently TOTP only).
type MFAEnrollment struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	MFAType         string // "totp"
	SecretEncrypted string // encrypted TOTP secret
	IsVerified      bool
	IsEnabled       bool
	CreatedAt       time.Time
	VerifiedAt      *time.Time
	DisabledAt      *time.Time
}

// IsActive returns true if MFA is enrolled, verified, and enabled.
func (m *MFAEnrollment) IsActive() bool {
	return m.IsVerified && m.IsEnabled
}

// MFARecoveryCode represents a hashed single-use recovery code.
type MFARecoveryCode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	IsUsed    bool
	UsedAt    *time.Time
	CreatedAt time.Time
}

// RecoveryCodeCount is the number of recovery codes generated per enrollment.
const RecoveryCodeCount = 10
