package entity

import (
	"time"

	"github.com/google/uuid"
)

// IAM audit event type constants.
const (
	AuditLoginSuccess    = "LOGIN_SUCCESS"
	AuditLoginFailure    = "LOGIN_FAILURE"
	AuditLogout          = "LOGOUT"
	AuditTokenRefresh    = "TOKEN_REFRESH"
	AuditPasswordChange  = "PASSWORD_CHANGE"
	AuditAccountLocked   = "ACCOUNT_LOCKED"
	AuditAccountUnlocked = "ACCOUNT_UNLOCKED"
	AuditUserCreated     = "USER_CREATED"
	AuditUserUpdated     = "USER_UPDATED"
	AuditUserDeactivated = "USER_DEACTIVATED"
	AuditSessionRevoked  = "SESSION_REVOKED"
	AuditBreachDetected  = "REFRESH_TOKEN_BREACH"
)

// AuditEvent is an immutable record of an IAM action.
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
