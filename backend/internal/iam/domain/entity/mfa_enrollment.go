package entity

import (
	"time"

	"github.com/google/uuid"
)

// MFA audit event type constants.
const (
	AuditMFAEnrolled        = "MFA_ENROLLED"
	AuditMFAVerified        = "MFA_VERIFIED"
	AuditMFAEnabled         = "MFA_ENABLED"
	AuditMFADisabled        = "MFA_DISABLED"
	AuditMFAChallengeOK     = "MFA_CHALLENGE_SUCCESS"
	AuditMFAChallengeFail   = "MFA_CHALLENGE_FAILURE"
	AuditMFARecoveryUsed    = "MFA_RECOVERY_CODE_USED"
	AuditSessionIdleTimeout = "SESSION_IDLE_TIMEOUT"
	AuditSessionAbsTimeout  = "SESSION_ABSOLUTE_TIMEOUT"
	AuditSessionConcurrent  = "SESSION_CONCURRENT_EVICTED"
	AuditRateLimitBlocked   = "RATE_LIMIT_BLOCKED"
	AuditPasswordExpired    = "PASSWORD_EXPIRED"
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
