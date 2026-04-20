package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	auditentity "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
)

// LogoutCommand revokes a refresh token session on logout.
type LogoutCommand struct {
	sessionRepo domain.SessionRepository
	auditSvc    auditdomain.Recorder
}

// NewLogoutCommand creates a LogoutCommand.
func NewLogoutCommand(sessionRepo domain.SessionRepository, auditSvc auditdomain.Recorder) *LogoutCommand {
	if auditSvc == nil {
		auditSvc = auditdomain.NopRecorder{}
	}
	return &LogoutCommand{
		sessionRepo: sessionRepo,
		auditSvc:    auditSvc,
	}
}

// LogoutInput is the input for the logout use case.
type LogoutInput struct {
	RefreshToken string
	UserID       uuid.UUID
	IPAddress    string
	UserAgent    string
}

// Execute revokes the refresh token session.
func (c *LogoutCommand) Execute(ctx context.Context, input LogoutInput) error {
	tokenHash := application.HashRefreshToken(input.RefreshToken)

	session, err := c.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("finding session: %w", err)
	}

	if session != nil && !session.IsRevoked {
		if err := c.sessionRepo.RevokeByID(ctx, session.ID); err != nil {
			return fmt.Errorf("revoking session: %w", err)
		}
	}

	// Audit logout
	c.auditSvc.Record(ctx, &input.UserID, auditentity.AuditLogout, "user", input.UserID.String(), input.IPAddress, input.UserAgent, nil)

	return nil
}
