package entity

import (
	"time"

	"github.com/google/uuid"
)

// RefreshTokenTTL is the server-side refresh token validity period.
const RefreshTokenTTL = 7 * 24 * time.Hour // 7 days

// DefaultIdleTimeout is the default inactivity timeout for sessions.
const DefaultIdleTimeout = 30 * time.Minute

// DefaultAbsoluteLifetime is the default absolute session lifetime.
const DefaultAbsoluteLifetime = 24 * time.Hour

// Session revoke reason constants.
const (
	RevokeReasonLogout          = "logout"
	RevokeReasonBreach          = "breach"
	RevokeReasonAdminRevoke     = "admin_revoke"
	RevokeReasonPasswordChange  = "password_change"
	RevokeReasonIdleTimeout     = "idle_timeout"
	RevokeReasonAbsoluteTimeout = "absolute_timeout"
	RevokeReasonConcurrentLimit = "concurrent_limit"
	RevokeReasonRotation        = "rotation"
)

// Session represents a server-side refresh token with rotation tracking.
// Token family enables breach detection: if a rotated-out token is reused,
// all tokens in the family are revoked.
type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	RefreshTokenHash  string
	TokenFamily       uuid.UUID
	IPAddress         string
	UserAgent         string
	IsRevoked         bool
	ExpiresAt         time.Time
	AbsoluteExpiresAt time.Time
	LastActivityAt    time.Time
	DeviceFingerprint string
	RevokeReason      string
	CreatedAt         time.Time
	RotatedAt         *time.Time
}

// IsValid checks if the session is usable (not revoked, not expired).
func (s *Session) IsValid(now time.Time) bool {
	return !s.IsRevoked && now.Before(s.ExpiresAt)
}

// IsIdle checks if the session has been idle beyond the given timeout.
func (s *Session) IsIdle(now time.Time, idleTimeout time.Duration) bool {
	return now.After(s.LastActivityAt.Add(idleTimeout))
}

// IsAbsoluteExpired checks if the session has exceeded its absolute lifetime.
func (s *Session) IsAbsoluteExpired(now time.Time) bool {
	return now.After(s.AbsoluteExpiresAt)
}

// TouchActivity updates the last activity timestamp.
func (s *Session) TouchActivity(now time.Time) {
	s.LastActivityAt = now
}

// Revoke marks this session as revoked.
func (s *Session) Revoke(now time.Time) {
	s.IsRevoked = true
	s.RotatedAt = &now
}

// RevokeWithReason marks this session as revoked with a specific reason.
func (s *Session) RevokeWithReason(now time.Time, reason string) {
	s.IsRevoked = true
	s.RotatedAt = &now
	s.RevokeReason = reason
}
