package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

func TestUser_IsLoginAllowed(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		user     entity.User
		allowed  bool
		contains string
	}{
		{
			name:    "active user can login",
			user:    entity.User{IsActive: true},
			allowed: true,
		},
		{
			name:     "inactive user cannot login",
			user:     entity.User{IsActive: false},
			allowed:  false,
			contains: "deactivated",
		},
		{
			name: "locked user cannot login",
			user: func() entity.User {
				lockUntil := now.Add(10 * time.Minute)
				return entity.User{IsActive: true, LockedUntil: &lockUntil}
			}(),
			allowed:  false,
			contains: "locked",
		},
		{
			name: "user with expired lock can login",
			user: func() entity.User {
				lockUntil := now.Add(-1 * time.Minute)
				return entity.User{IsActive: true, LockedUntil: &lockUntil}
			}(),
			allowed: true,
		},
		{
			name: "soft-deleted user cannot login",
			user: func() entity.User {
				deletedAt := now.Add(-1 * time.Hour)
				return entity.User{IsActive: true, DeletedAt: &deletedAt}
			}(),
			allowed:  false,
			contains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, reason := tt.user.IsLoginAllowed(now)
			assert.Equal(t, tt.allowed, allowed)
			if !tt.allowed {
				assert.Contains(t, reason, tt.contains)
			}
		})
	}
}

func TestUser_RecordFailedLogin(t *testing.T) {
	now := time.Now().UTC()

	t.Run("increments failed attempts", func(t *testing.T) {
		user := entity.User{IsActive: true}
		user.RecordFailedLogin(now)
		assert.Equal(t, 1, user.FailedLoginAttempts)
		assert.Nil(t, user.LockedUntil)
	})

	t.Run("locks account at max attempts with default policy", func(t *testing.T) {
		user := entity.User{IsActive: true, FailedLoginAttempts: entity.DefaultMaxFailedLoginAttempts - 1}
		user.RecordFailedLogin(now)
		assert.Equal(t, entity.DefaultMaxFailedLoginAttempts, user.FailedLoginAttempts)
		assert.NotNil(t, user.LockedUntil)
		assert.True(t, user.LockedUntil.After(now))
	})

	t.Run("locks account at custom policy threshold", func(t *testing.T) {
		policy := entity.LockoutPolicy{MaxFailedAttempts: 3, LockoutDuration: 10 * time.Minute}
		user := entity.User{IsActive: true, FailedLoginAttempts: 2}
		user.RecordFailedLogin(now, policy)
		assert.Equal(t, 3, user.FailedLoginAttempts)
		assert.NotNil(t, user.LockedUntil)
		expected := now.Add(10 * time.Minute)
		assert.Equal(t, expected, *user.LockedUntil)
	})
}

func TestUser_RecordSuccessfulLogin(t *testing.T) {
	now := time.Now().UTC()
	lockUntil := now.Add(10 * time.Minute)

	user := entity.User{
		IsActive:            true,
		FailedLoginAttempts: 5,
		LockedUntil:         &lockUntil,
	}
	user.RecordSuccessfulLogin(now)

	assert.Equal(t, 0, user.FailedLoginAttempts)
	assert.Nil(t, user.LockedUntil)
	assert.NotNil(t, user.LastLoginAt)
	assert.Equal(t, now, *user.LastLoginAt)
}

func TestUser_IsPasswordExpired(t *testing.T) {
	now := time.Now().UTC()
	maxAge := 90 * 24 * time.Hour // 90 days

	tests := []struct {
		name              string
		passwordChangedAt *time.Time
		maxAge            time.Duration
		expired           bool
	}{
		{
			name:              "disabled when maxAge is 0",
			passwordChangedAt: nil,
			maxAge:            0,
			expired:           false,
		},
		{
			name:              "expired when never changed",
			passwordChangedAt: nil,
			maxAge:            maxAge,
			expired:           true,
		},
		{
			name: "not expired when recently changed",
			passwordChangedAt: func() *time.Time {
				t := now.Add(-10 * 24 * time.Hour) // 10 days ago
				return &t
			}(),
			maxAge:  maxAge,
			expired: false,
		},
		{
			name: "expired when changed beyond max age",
			passwordChangedAt: func() *time.Time {
				t := now.Add(-91 * 24 * time.Hour) // 91 days ago
				return &t
			}(),
			maxAge:  maxAge,
			expired: true,
		},
		{
			name: "not expired at exact boundary",
			passwordChangedAt: func() *time.Time {
				t := now.Add(-maxAge) // exactly at boundary
				return &t
			}(),
			maxAge:  maxAge,
			expired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := entity.User{
				IsActive:          true,
				PasswordChangedAt: tt.passwordChangedAt,
			}
			assert.Equal(t, tt.expired, user.IsPasswordExpired(now, tt.maxAge))
		})
	}
}

func TestUser_RecordPasswordChange(t *testing.T) {
	now := time.Now().UTC()
	user := entity.User{
		IsActive:            true,
		ForcePasswordChange: true,
	}

	user.RecordPasswordChange(now)

	assert.NotNil(t, user.PasswordChangedAt)
	assert.Equal(t, now, *user.PasswordChangedAt)
	assert.False(t, user.ForcePasswordChange, "ForcePasswordChange should be cleared")
}
