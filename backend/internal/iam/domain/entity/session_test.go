package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

func TestSession_IsValid(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		session entity.Session
		valid   bool
	}{
		{
			name: "valid session",
			session: entity.Session{
				ExpiresAt: now.Add(1 * time.Hour),
				IsRevoked: false,
			},
			valid: true,
		},
		{
			name: "expired session",
			session: entity.Session{
				ExpiresAt: now.Add(-1 * time.Hour),
				IsRevoked: false,
			},
			valid: false,
		},
		{
			name: "revoked session",
			session: entity.Session{
				ExpiresAt: now.Add(1 * time.Hour),
				IsRevoked: true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.session.IsValid(now))
		})
	}
}

func TestSession_Revoke(t *testing.T) {
	now := time.Now().UTC()
	session := entity.Session{
		ID:        uuid.New(),
		ExpiresAt: now.Add(1 * time.Hour),
	}

	session.Revoke(now)
	assert.True(t, session.IsRevoked)
	assert.NotNil(t, session.RotatedAt)
	assert.Equal(t, now, *session.RotatedAt)
}

func TestSession_RevokeWithReason(t *testing.T) {
	now := time.Now().UTC()
	session := entity.Session{
		ID:        uuid.New(),
		ExpiresAt: now.Add(1 * time.Hour),
	}

	session.RevokeWithReason(now, entity.RevokeReasonIdleTimeout)
	assert.True(t, session.IsRevoked)
	assert.Equal(t, entity.RevokeReasonIdleTimeout, session.RevokeReason)
}

func TestSession_IsIdle(t *testing.T) {
	now := time.Now().UTC()
	timeout := 30 * time.Minute

	tests := []struct {
		name           string
		lastActivityAt time.Time
		idle           bool
	}{
		{
			name:           "active session (5 min ago)",
			lastActivityAt: now.Add(-5 * time.Minute),
			idle:           false,
		},
		{
			name:           "idle session (31 min ago)",
			lastActivityAt: now.Add(-31 * time.Minute),
			idle:           true,
		},
		{
			name:           "exactly at timeout boundary",
			lastActivityAt: now.Add(-30 * time.Minute),
			idle:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := entity.Session{
				LastActivityAt: tt.lastActivityAt,
			}
			assert.Equal(t, tt.idle, session.IsIdle(now, timeout))
		})
	}
}

func TestSession_IsAbsoluteExpired(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name              string
		absoluteExpiresAt time.Time
		expired           bool
	}{
		{
			name:              "not expired",
			absoluteExpiresAt: now.Add(1 * time.Hour),
			expired:           false,
		},
		{
			name:              "expired",
			absoluteExpiresAt: now.Add(-1 * time.Hour),
			expired:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := entity.Session{
				AbsoluteExpiresAt: tt.absoluteExpiresAt,
			}
			assert.Equal(t, tt.expired, session.IsAbsoluteExpired(now))
		})
	}
}

func TestSession_TouchActivity(t *testing.T) {
	now := time.Now().UTC()
	session := entity.Session{
		LastActivityAt: now.Add(-10 * time.Minute),
	}

	session.TouchActivity(now)
	assert.Equal(t, now, session.LastActivityAt)
}
