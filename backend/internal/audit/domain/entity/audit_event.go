package entity

import (
	"time"

	"github.com/google/uuid"
)

// Audit event type constants for security- and business-critical traceability.
const (
	AuditLoginSuccess       = "LOGIN_SUCCESS"
	AuditLoginFailure       = "LOGIN_FAILURE"
	AuditLogout             = "LOGOUT"
	AuditTokenRefresh       = "TOKEN_REFRESH"
	AuditPasswordChange     = "PASSWORD_CHANGE"
	AuditPasswordExpired    = "PASSWORD_EXPIRED"
	AuditAccountLocked      = "ACCOUNT_LOCKED"
	AuditAccountUnlocked    = "ACCOUNT_UNLOCKED"
	AuditUserCreated        = "USER_CREATED"
	AuditUserUpdated        = "USER_UPDATED"
	AuditUserActivated      = "USER_ACTIVATED"
	AuditUserDeactivated    = "USER_DEACTIVATED"
	AuditSessionRevoked     = "SESSION_REVOKED"
	AuditSessionIdleTimeout = "SESSION_IDLE_TIMEOUT"
	AuditSessionAbsTimeout  = "SESSION_ABSOLUTE_TIMEOUT"
	AuditSessionConcurrent  = "SESSION_CONCURRENT_EVICTED"
	AuditBreachDetected     = "REFRESH_TOKEN_BREACH"
	AuditRateLimitBlocked   = "RATE_LIMIT_BLOCKED"
	AuditMFAEnrolled        = "MFA_ENROLLED"
	AuditMFAVerified        = "MFA_VERIFIED"
	AuditMFAEnabled         = "MFA_ENABLED"
	AuditMFADisabled        = "MFA_DISABLED"
	AuditMFAChallengeOK     = "MFA_CHALLENGE_SUCCESS"
	AuditMFAChallengeFail   = "MFA_CHALLENGE_FAILURE"
	AuditMFARecoveryUsed    = "MFA_RECOVERY_CODE_USED"
	AuditLogViewed          = "AUDIT_LOG_VIEWED"
	AuditLogExported        = "AUDIT_LOG_EXPORTED"
)

// AuditEvent is an immutable record of a sensitive system action.
type AuditEvent struct {
	ID         uuid.UUID
	ActorID    *uuid.UUID
	EventType  string
	TargetType string
	TargetID   string
	IPAddress  string
	UserAgent  string
	Metadata   map[string]interface{}
	CreatedAt  time.Time
}
