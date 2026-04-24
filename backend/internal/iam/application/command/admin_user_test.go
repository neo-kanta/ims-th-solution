package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

type fakeUserRepo struct {
	users map[uuid.UUID]*entity.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uuid.UUID]*entity.User)}
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (*entity.User, error) {
	for _, u := range r.users {
		if u.Username == username && u.DeletedAt == nil {
			return u, nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	u, ok := r.users[id]
	if !ok || u.DeletedAt != nil {
		return nil, nil
	}
	return u, nil
}

func (r *fakeUserRepo) Create(_ context.Context, user *entity.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *entity.User) error {
	r.users[user.ID] = user
	user.Version++
	return nil
}

func (r *fakeUserRepo) SoftDelete(_ context.Context, id uuid.UUID, _ uuid.UUID) error {
	if u, ok := r.users[id]; ok {
		now := time.Now()
		u.DeletedAt = &now
	}
	return nil
}

func (r *fakeUserRepo) List(_ context.Context, _ domain.UserFilter) ([]entity.User, int, error) {
	var users []entity.User
	for _, u := range r.users {
		if u.DeletedAt == nil {
			users = append(users, *u)
		}
	}
	return users, len(users), nil
}

// === Tests ===

func TestSetUserStatus_DisableThenEnable(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{FixedTime: now}

	userID := uuid.New()
	adminID := uuid.New()

	// Simulate a user who had failed login attempts and got auto-locked
	lockUntil := now.Add(30 * time.Minute)
	user := &entity.User{
		ID:                  userID,
		Username:            "testuser",
		Email:               "test@example.com",
		IsActive:            true,
		FailedLoginAttempts: 10,
		LockedUntil:         &lockUntil,
		Version:             1,
	}

	repo := newFakeUserRepo()
	repo.users[userID] = user

	cmd := command.NewAdminUserCommand(repo, auditdomain.NopRecorder{}, clk)

	ctx := context.Background()

	// Step 1: Disable the user
	err := cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "disable",
	})
	require.NoError(t, err)
	assert.False(t, repo.users[userID].IsActive)

	// Step 2: Enable the user
	err = cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "enable",
	})
	require.NoError(t, err)

	// Verify: user is active AND lock is cleared
	u := repo.users[userID]
	assert.True(t, u.IsActive, "user should be active after enable")
	assert.Nil(t, u.LockedUntil, "lock should be cleared after enable")
	assert.Equal(t, 0, u.FailedLoginAttempts, "failed login attempts should be reset after enable")

	// Verify login is allowed
	allowed, reason := u.IsLoginAllowed(now)
	assert.True(t, allowed, "login should be allowed after enable, got reason: %s", reason)
}

func TestSetUserStatus_LockThenUnlock(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{FixedTime: now}

	userID := uuid.New()
	adminID := uuid.New()

	user := &entity.User{
		ID:       userID,
		Username: "testuser",
		IsActive: true,
		Version:  1,
	}

	repo := newFakeUserRepo()
	repo.users[userID] = user

	cmd := command.NewAdminUserCommand(repo, auditdomain.NopRecorder{}, clk)
	ctx := context.Background()

	// Lock the user
	err := cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "lock",
	})
	require.NoError(t, err)
	assert.NotNil(t, repo.users[userID].LockedUntil, "should be locked")
	assert.True(t, repo.users[userID].IsActive, "lock should NOT change is_active")

	// Verify login blocked when locked
	allowed, _ := repo.users[userID].IsLoginAllowed(now)
	assert.False(t, allowed, "login should be blocked when locked")

	// Unlock the user
	err = cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "unlock",
	})
	require.NoError(t, err)
	assert.Nil(t, repo.users[userID].LockedUntil, "should be unlocked")
	assert.True(t, repo.users[userID].IsActive, "unlock should NOT change is_active")
	assert.Equal(t, 0, repo.users[userID].FailedLoginAttempts)

	// Verify login allowed after unlock
	allowed, _ = repo.users[userID].IsLoginAllowed(now)
	assert.True(t, allowed, "login should be allowed after unlock")
}

func TestSetUserStatus_UnlockDoesNotEnableDisabledUser(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{FixedTime: now}

	userID := uuid.New()
	adminID := uuid.New()
	lockUntil := now.Add(30 * time.Minute)

	user := &entity.User{
		ID:                  userID,
		Username:            "testuser",
		IsActive:            false, // disabled
		LockedUntil:         &lockUntil,
		FailedLoginAttempts: 10,
		Version:             1,
	}

	repo := newFakeUserRepo()
	repo.users[userID] = user

	cmd := command.NewAdminUserCommand(repo, auditdomain.NopRecorder{}, clk)
	ctx := context.Background()

	// Unlock should clear lock but NOT enable the user
	err := cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "unlock",
	})
	require.NoError(t, err)

	u := repo.users[userID]
	assert.False(t, u.IsActive, "unlock should NOT enable a disabled user")
	assert.Nil(t, u.LockedUntil, "lock should be cleared")
	assert.Equal(t, 0, u.FailedLoginAttempts)

	// Login still blocked because user is disabled
	allowed, reason := u.IsLoginAllowed(now)
	assert.False(t, allowed, "login should still be blocked for disabled user")
	assert.Contains(t, reason, "deactivated")
}

func TestSetUserStatus_InvalidAction(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{FixedTime: now}

	userID := uuid.New()
	adminID := uuid.New()

	user := &entity.User{
		ID:       userID,
		Username: "testuser",
		IsActive: true,
		Version:  1,
	}

	repo := newFakeUserRepo()
	repo.users[userID] = user

	cmd := command.NewAdminUserCommand(repo, auditdomain.NopRecorder{}, clk)
	ctx := context.Background()

	err := cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  adminID,
		TargetID: userID,
		Action:   "invalid",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestSetUserStatus_UserNotFound(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	clk := clock.FixedClock{FixedTime: now}

	repo := newFakeUserRepo()
	cmd := command.NewAdminUserCommand(repo, auditdomain.NopRecorder{}, clk)
	ctx := context.Background()

	err := cmd.SetUserStatus(ctx, command.SetUserStatusInput{
		AdminID:  uuid.New(),
		TargetID: uuid.New(),
		Action:   "disable",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
