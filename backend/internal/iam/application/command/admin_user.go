package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
)

// AdminUserCommand handles all administrative actions for users.
type AdminUserCommand struct {
	userRepo  domain.UserRepository
	auditSvc  *appservice.AuditService
	clock     clock.Clock
}

// NewAdminUserCommand creates a new AdminUserCommand.
func NewAdminUserCommand(userRepo domain.UserRepository, auditSvc *appservice.AuditService, clock clock.Clock) *AdminUserCommand {
	return &AdminUserCommand{
		userRepo:  userRepo,
		auditSvc:  auditSvc,
		clock:     clock,
	}
}

// CreateUserInput payload.
type CreateUserInput struct {
	AdminID     uuid.UUID
	Username    string
	DisplayName string
	Email       string
	Password    string
	IPAddress   string
	UserAgent   string
}

// CreateUser handles the administrative creation of a new user.
// It validates password strength, hashes credentials, and enforces a mandatory
// password change on the user's first login.
func (c *AdminUserCommand) CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error) {
	// 1. Validate password strength and hash
	hash, err := valueobject.HashPassword(input.Password)
	if err != nil {
		return uuid.Nil, apperrors.NewBusinessError("bad_request", err.Error())
	}

	// 2. Build User Entity
	// Note: Duplicate username rely on repository/db constraints for race-condition safety.
	now := c.clock.Now()
	newID := uuid.New()
	user := &entity.User{
		ID:                  newID,
		Username:            input.Username,
		DisplayName:         input.DisplayName,
		Email:               input.Email,
		PasswordHash:        hash,
		IsActive:            true,
		ForcePasswordChange: true,
		PasswordChangedAt:   &now,
		CreatedAt:           now,
		UpdatedAt:           now,
		CreatedBy:           &input.AdminID,
		UpdatedBy:           &input.AdminID,
	}

	// 3. Save
	if err := c.userRepo.Create(ctx, user); err != nil {
		return uuid.Nil, fmt.Errorf("creating user: %w", err)
	}

	// 4. Audit
	c.recordAudit(ctx, input.AdminID, entity.AuditUserCreated, newID.String(), input.IPAddress, input.UserAgent, nil)

	return newID, nil
}

// SetUserStatusInput toggles active/locked.
type SetUserStatusInput struct {
	AdminID   uuid.UUID
	TargetID  uuid.UUID
	Action    string // "disable", "enable", "lock", "unlock"
	IPAddress string
	UserAgent string
}

// SetUserStatus performs lifecycle toggles (enable, disable, lock, unlock) on a user account.
func (c *AdminUserCommand) SetUserStatus(ctx context.Context, input SetUserStatusInput) error {
	user, err := c.userRepo.FindByID(ctx, input.TargetID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return apperrors.NewNotFoundError("user", input.TargetID.String())
	}

	now := c.clock.Now()
	var eventType string

	switch input.Action {
	case "disable":
		user.IsActive = false
		eventType = entity.AuditUserDeactivated
	case "enable":
		user.IsActive = true
		eventType = entity.AuditUserUpdated
	case "lock":
		lockUntil := now.AddDate(100, 0, 0)
		user.LockedUntil = &lockUntil
		eventType = entity.AuditAccountLocked
	case "unlock":
		user.LockedUntil = nil
		user.FailedLoginAttempts = 0
		eventType = entity.AuditAccountUnlocked
	default:
		return apperrors.NewBusinessError("bad_request", "invalid action")
	}

	user.UpdatedAt = now
	user.UpdatedBy = &input.AdminID

	if err := c.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user status: %w", err)
	}

	c.recordAudit(ctx, input.AdminID, eventType, user.ID.String(), input.IPAddress, input.UserAgent, map[string]interface{}{"action": input.Action})

	return nil
}

// ResetPasswordInput for admin forced reset.
type ResetPasswordInput struct {
	AdminID     uuid.UUID
	TargetID    uuid.UUID
	NewPassword string
	IPAddress   string
	UserAgent   string
}

// ResetPassword allows an administrator to forcefully reset a user's password.
// This also triggers the ForcePasswordChange flag for security.
func (c *AdminUserCommand) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	user, err := c.userRepo.FindByID(ctx, input.TargetID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return apperrors.NewNotFoundError("user", input.TargetID.String())
	}

	hash, err := valueobject.HashPassword(input.NewPassword)
	if err != nil {
		return apperrors.NewBusinessError("bad_request", err.Error())
	}

	now := c.clock.Now()
	user.PasswordHash = hash
	user.ForcePasswordChange = true
	user.PasswordChangedAt = &now
	user.UpdatedAt = now
	user.UpdatedBy = &input.AdminID

	if err := c.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user password: %w", err)
	}

	c.recordAudit(ctx, input.AdminID, entity.AuditPasswordChange, user.ID.String(), input.IPAddress, input.UserAgent, map[string]interface{}{"forced_by_admin": true})

	return nil
}

// recordAudit is a private helper to persist administrative audit events.
func (c *AdminUserCommand) recordAudit(ctx context.Context, adminID uuid.UUID, eventType string, targetID string, ip, ua string, meta map[string]interface{}) {
	c.auditSvc.Record(ctx, &adminID, eventType, "user", targetID, ip, ua, meta)
}
