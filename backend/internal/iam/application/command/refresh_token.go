package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/dto"
	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
)

// RefreshTokenCommand handles token refresh with rotation, breach detection,
// idle timeout enforcement, and absolute session lifetime.
type RefreshTokenCommand struct {
	userRepo      domain.UserRepository
	sessionRepo   domain.SessionRepository
	tokenService  *application.TokenService
	auditService  *appservice.AuditService
	clock         clock.Clock
	idleTimeout   time.Duration
}

// NewRefreshTokenCommand creates a RefreshTokenCommand.
func NewRefreshTokenCommand(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenService *application.TokenService,
	auditService *appservice.AuditService,
	clk clock.Clock,
	idleTimeout time.Duration,
) *RefreshTokenCommand {
	return &RefreshTokenCommand{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
		auditService: auditService,
		clock:        clk,
		idleTimeout:  idleTimeout,
	}
}

// RefreshInput is the input for the refresh use case.
type RefreshInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

// Execute validates the refresh token, enforces idle/absolute timeout,
// rotates it, and returns new tokens.
func (c *RefreshTokenCommand) Execute(ctx context.Context, input RefreshInput) (*dto.RefreshResult, error) {
	now := c.clock.Now()
	tokenHash := application.HashRefreshToken(input.RefreshToken)

	// 1. Find session by token hash
	session, err := c.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid refresh token")
	}

	// 2. Breach detection: if token is already revoked, revoke entire family
	if session.IsRevoked {
		slog.Warn("refresh token reuse detected — revoking token family",
			"session_id", session.ID,
			"token_family", session.TokenFamily,
			"user_id", session.UserID,
		)
		if err := c.sessionRepo.RevokeByFamily(ctx, session.TokenFamily); err != nil {
			slog.Error("failed to revoke token family", "error", err)
		}
		c.auditBreachDetected(ctx, session, input)
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "token reuse detected; all sessions revoked")
	}

	// 3. Check expiry
	if !session.IsValid(now) {
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "refresh token expired")
	}

	// 4. Check absolute session lifetime
	if session.IsAbsoluteExpired(now) {
		_ = c.sessionRepo.RevokeByIDWithReason(ctx, session.ID, entity.RevokeReasonAbsoluteTimeout)
		c.auditSessionTimeout(ctx, session, input, entity.AuditSessionAbsTimeout)
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "session has exceeded maximum lifetime; please login again")
	}

	// 5. Check idle timeout
	if c.idleTimeout > 0 && session.IsIdle(now, c.idleTimeout) {
		_ = c.sessionRepo.RevokeByIDWithReason(ctx, session.ID, entity.RevokeReasonIdleTimeout)
		c.auditSessionTimeout(ctx, session, input, entity.AuditSessionIdleTimeout)
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "session expired due to inactivity; please login again")
	}

	// 6. Verify user is still active
	user, err := c.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "user not found")
	}
	if allowed, reason := user.IsLoginAllowed(now); !allowed {
		return nil, apperrors.NewBusinessError(apperrors.CodeForbidden, reason)
	}

	// 7. Revoke old session (mark as rotated)
	if err := c.sessionRepo.RevokeByIDWithReason(ctx, session.ID, entity.RevokeReasonRotation); err != nil {
		return nil, fmt.Errorf("revoking old session: %w", err)
	}

	// 8. Create new refresh token (same family for rotation tracking)
	newRefreshRaw := c.tokenService.GenerateRefreshToken()
	newRefreshHash := application.HashRefreshToken(newRefreshRaw)
	newSession := &entity.Session{
		ID:                uuid.New(),
		UserID:            user.ID,
		RefreshTokenHash:  newRefreshHash,
		TokenFamily:       session.TokenFamily, // same family
		IPAddress:         input.IPAddress,
		UserAgent:         input.UserAgent,
		ExpiresAt:         now.Add(entity.RefreshTokenTTL),
		AbsoluteExpiresAt: session.AbsoluteExpiresAt, // preserve original absolute expiry
		LastActivityAt:    now,
	}
	if err := c.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("creating new session: %w", err)
	}

	// 9. Generate a new access token bound to the active session.
	accessToken, accessExpiresAt, err := c.tokenService.GenerateAccessTokenForSession(user.ID, newSession.ID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	// 10. Audit
	c.auditTokenRefresh(ctx, user.ID, input)

	return &dto.RefreshResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          newRefreshRaw,
		RefreshTokenExpiresAt: newSession.ExpiresAt,
	}, nil
}

func (c *RefreshTokenCommand) auditBreachDetected(ctx context.Context, session *entity.Session, input RefreshInput) {
	c.auditService.Record(ctx, &session.UserID, entity.AuditBreachDetected, "session", session.TokenFamily.String(), input.IPAddress, input.UserAgent, map[string]interface{}{
		"session_id":   session.ID.String(),
		"token_family": session.TokenFamily.String(),
	})
}

func (c *RefreshTokenCommand) auditSessionTimeout(ctx context.Context, session *entity.Session, input RefreshInput, eventType string) {
	c.auditService.Record(ctx, &session.UserID, eventType, "session", session.ID.String(), input.IPAddress, input.UserAgent, nil)
}

func (c *RefreshTokenCommand) auditTokenRefresh(ctx context.Context, userID uuid.UUID, input RefreshInput) {
	c.auditService.Record(ctx, &userID, entity.AuditTokenRefresh, "user", userID.String(), input.IPAddress, input.UserAgent, nil)
}
