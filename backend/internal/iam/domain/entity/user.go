package entity

import (
	"time"

	"github.com/google/uuid"
)

// DefaultMaxFailedLoginAttempts is the default number of failed attempts before account lockout.
// Override at runtime via LockoutPolicy.
const DefaultMaxFailedLoginAttempts = 10

// DefaultLockoutDuration is the default lockout window.
// Override at runtime via LockoutPolicy.
const DefaultLockoutDuration = 30 * time.Minute

// LockoutPolicy holds configurable lockout thresholds.
type LockoutPolicy struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

// DefaultLockoutPolicy returns the default lockout policy.
func DefaultLockoutPolicy() LockoutPolicy {
	return LockoutPolicy{
		MaxFailedAttempts: DefaultMaxFailedLoginAttempts,
		LockoutDuration:   DefaultLockoutDuration,
	}
}

// User is the aggregate root for identity and access management.
type User struct {
	ID                  uuid.UUID
	Username            string
	DisplayName         string
	Email               string
	PasswordHash        string
	IsActive            bool
	ForcePasswordChange bool
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	PasswordChangedAt   *time.Time
	Version             int
	Groups              []string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CreatedBy           *uuid.UUID
	UpdatedBy           *uuid.UUID
	DeletedAt           *time.Time
}

// IsLoginAllowed checks all business constraints (active, not deleted, not locked)
// to determine if a user can proceed with authentication.
func (u *User) IsLoginAllowed(now time.Time) (bool, string) {
	if !u.IsActive {
		return false, "account is deactivated"
	}
	if u.DeletedAt != nil {
		return false, "account not found"
	}
	if u.IsLocked(now) {
		return false, "account is temporarily locked due to too many failed attempts"
	}
	return true, ""
}

// IsPasswordExpired checks if the password has exceeded the max age.
// A maxAge of 0 means password expiration is disabled.
func (u *User) IsPasswordExpired(now time.Time, maxAge time.Duration) bool {
	if maxAge <= 0 {
		return false
	}
	if u.PasswordChangedAt == nil {
		return true // never changed, treat as expired
	}
	return now.After(u.PasswordChangedAt.Add(maxAge))
}

// IsLocked returns true if the account is currently within a lockout window.
func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && now.Before(*u.LockedUntil)
}

// RecordFailedLogin increments the lockout state machine.
// If the threshold is reached, it sets a future LockedUntil timestamp.
// Uses the provided policy; pass DefaultLockoutPolicy() if no override is needed.
func (u *User) RecordFailedLogin(now time.Time, policy ...LockoutPolicy) {
	p := DefaultLockoutPolicy()
	if len(policy) > 0 {
		p = policy[0]
	}
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= p.MaxFailedAttempts {
		lockUntil := now.Add(p.LockoutDuration)
		u.LockedUntil = &lockUntil
	}
}

// RecordSuccessfulLogin clears the lockout state machine and updates the last login timestamp.
func (u *User) RecordSuccessfulLogin(now time.Time) {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
}

// RecordPasswordChange updates the password change timestamp.
func (u *User) RecordPasswordChange(now time.Time) {
	u.PasswordChangedAt = &now
	u.ForcePasswordChange = false
}
