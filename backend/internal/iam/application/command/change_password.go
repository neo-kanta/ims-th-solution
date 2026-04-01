package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
)

// ChangePasswordCommand allows a user to update their own password.
type ChangePasswordCommand struct {
	userRepo    domain.UserRepository
	auditSvc    *appservice.AuditService
	sessionRepo domain.SessionRepository
}

// NewChangePasswordCommand creates a new ChangePasswordCommand.
func NewChangePasswordCommand(userRepo domain.UserRepository, auditSvc *appservice.AuditService, sessionRepo domain.SessionRepository) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		userRepo:    userRepo,
		auditSvc:    auditSvc,
		sessionRepo: sessionRepo,
	}
}

// ChangePasswordInput is the input map for changing a password.
type ChangePasswordInput struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
	IPAddress   string
	UserAgent   string
}

// Execute validates old credentials, replaces the password hash, and revokes
// all active sessions across all devices to ensure total credential invalidation.
func (c *ChangePasswordCommand) Execute(ctx context.Context, input ChangePasswordInput) error {
	user, err := c.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return apperrors.NewBusinessError(apperrors.CodeUnauthorized, "user not found")
	}

	// 1. Verify old password
	if err := valueobject.CheckPassword(user.PasswordHash, input.OldPassword); err != nil {
		return apperrors.NewBusinessError(apperrors.CodeUnauthorized, "incorrect old password")
	}

	// 2. Hash new password (this validates NIST constraints)
	newHash, err := valueobject.HashPassword(input.NewPassword)
	if err != nil {
		return apperrors.NewBusinessError("bad_request", err.Error())
	}

	user.PasswordHash = newHash
	user.RecordPasswordChange(time.Now().UTC())

	// 3. Save
	if err := c.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("updating user password: %w", err)
	}

	// 4. Revoke all sessions (forces re-login across devices)
	if err := c.sessionRepo.RevokeAllForUserWithReason(ctx, user.ID, entity.RevokeReasonPasswordChange); err != nil {
		slog.Error("failed to revoke sessions after password change", "error", err, "user_id", user.ID)
	}

	// 5. Audit
	c.auditSvc.Record(ctx, &input.UserID, entity.AuditPasswordChange, "user", input.UserID.String(), input.IPAddress, input.UserAgent, nil)

	return nil
}
