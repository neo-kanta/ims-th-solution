package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
)

// LogoutAllCommand revokes all refresh token sessions globally.
type LogoutAllCommand struct {
	sessionRepo domain.SessionRepository
	auditSvc    *appservice.AuditService
}

// NewLogoutAllCommand creates a LogoutAllCommand.
func NewLogoutAllCommand(sessionRepo domain.SessionRepository, auditSvc *appservice.AuditService) *LogoutAllCommand {
	return &LogoutAllCommand{
		sessionRepo: sessionRepo,
		auditSvc:    auditSvc,
	}
}

// LogoutAllInput is the input for the logout-all use case.
type LogoutAllInput struct {
	UserID    uuid.UUID
	IPAddress string
	UserAgent string
}

// Execute revokes all refresh token sessions globally for the user.
// This is used as a security measure to kill all active device sessions.
func (c *LogoutAllCommand) Execute(ctx context.Context, input LogoutAllInput) error {
	if err := c.sessionRepo.RevokeAllForUser(ctx, input.UserID); err != nil {
		return fmt.Errorf("revoking all sessions: %w", err)
	}

	c.auditSvc.Record(ctx, &input.UserID, "LOGOUT", "user", input.UserID.String(), input.IPAddress, input.UserAgent, map[string]interface{}{"scope": "all_sessions"})

	return nil
}
