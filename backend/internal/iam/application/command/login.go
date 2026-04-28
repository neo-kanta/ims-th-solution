package command

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	auditentity "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/dto"
	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// SessionPolicy holds configurable session limits.
type SessionPolicy struct {
	IdleTimeout      time.Duration
	AbsoluteLifetime time.Duration
	MaxConcurrent    int
	ConcurrentMode   string
	PasswordMaxAge   time.Duration
	LockoutPolicy    entity.LockoutPolicy
}

// LoginCommand handles user authentication with lockout, MFA, session policy, and audit.
type LoginCommand struct {
	userRepo          domain.UserRepository
	sessionRepo       domain.SessionRepository
	mfaRepo           domain.MFARepository
	permsFetcher      domain.PermissionsFetcher
	tokenService      *application.TokenService
	totpService       *application.TOTPService
	mfaChallengeSvc   *appservice.MFAChallengeService
	auditService      auditdomain.Recorder
	clock             clock.Clock
	sessionPolicy     SessionPolicy
	mfaForcedForAdmin bool
}

// NewLoginCommand creates a LoginCommand with its dependencies.
func NewLoginCommand(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	mfaRepo domain.MFARepository,
	permsFetcher domain.PermissionsFetcher,
	tokenService *application.TokenService,
	totpService *application.TOTPService,
	mfaChallengeSvc *appservice.MFAChallengeService,
	auditService auditdomain.Recorder,
	clk clock.Clock,
	policy SessionPolicy,
	mfaForcedForAdmin bool,
) *LoginCommand {
	if auditService == nil {
		auditService = auditdomain.NopRecorder{}
	}
	return &LoginCommand{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		mfaRepo:           mfaRepo,
		permsFetcher:      permsFetcher,
		tokenService:      tokenService,
		totpService:       totpService,
		mfaChallengeSvc:   mfaChallengeSvc,
		auditService:      auditService,
		clock:             clk,
		sessionPolicy:     policy,
		mfaForcedForAdmin: mfaForcedForAdmin,
	}
}

// LoginInput is the input for the login use case.
type LoginInput struct {
	Username     string
	Password     string
	TOTPCode     string
	RecoveryCode string
	MFAToken     string
	IPAddress    string
	UserAgent    string
}

// Execute authenticates the user, enforces MFA if enrolled, applies session policy,
// and returns JWT + refresh token + profile + permissions.
func (c *LoginCommand) Execute(ctx context.Context, input LoginInput) (*dto.LoginResult, error) {
	now := c.clock.Now()

	user, err := c.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		c.auditLoginFailure(ctx, nil, input, "user not found")
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid username or password")
	}

	if allowed, reason := user.IsLoginAllowed(now); !allowed {
		c.auditLoginFailure(ctx, &user.ID, input, reason)
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid username or password")
	}

	if err := valueobject.CheckPassword(user.PasswordHash, input.Password); err != nil {
		user.RecordFailedLogin(now, c.sessionPolicy.LockoutPolicy)
		if updateErr := c.userRepo.Update(ctx, user); updateErr != nil {
			slog.Error("failed to update user after failed login", "error", updateErr, "user_id", user.ID)
		}
		if user.IsLocked(now) {
			c.auditAccountLocked(ctx, user, input)
		}
		c.auditLoginFailure(ctx, &user.ID, input, "invalid password")
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid username or password")
	}

	functions, err := c.permsFetcher.GetUserFunctionPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching function permissions: %w", err)
	}

	mfaEnrollment, err := c.mfaRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		slog.Error("failed to check MFA enrollment", "error", err, "user_id", user.ID)
	}

	restrictedSession := false

	if mfaEnrollment != nil && mfaEnrollment.IsActive() {
		if input.TOTPCode == "" && input.RecoveryCode == "" {
			challengeToken, _, err := c.mfaChallengeSvc.Issue(user.ID, user.Username)
			if err != nil {
				return nil, fmt.Errorf("issuing MFA challenge: %w", err)
			}
			return &dto.LoginResult{
				MFARequired: true,
				MFAToken:    challengeToken,
			}, nil
		}

		if input.MFAToken == "" {
			c.auditMFAFailure(ctx, user, input)
			return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "missing MFA challenge token")
		}
		if err := c.mfaChallengeSvc.Validate(input.MFAToken, user.ID, user.Username); err != nil {
			c.auditMFAFailure(ctx, user, input)
			return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid MFA challenge token")
		}

		if input.TOTPCode != "" {
			valid, verr := c.totpService.ValidateCode(mfaEnrollment.SecretEncrypted, input.TOTPCode)
			if verr != nil || !valid {
				c.auditMFAFailure(ctx, user, input)
				return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid MFA code")
			}
			c.recordAudit(ctx, &user.ID, auditentity.AuditMFAChallengeOK, "user", user.ID.String(), input, nil)
		} else if input.RecoveryCode != "" {
			codes, err := c.mfaRepo.FindUnusedRecoveryCodes(ctx, user.ID)
			if err != nil {
				return nil, fmt.Errorf("finding recovery codes: %w", err)
			}
			matched := c.totpService.VerifyRecoveryCode(input.RecoveryCode, codes)
			if matched == nil {
				c.auditMFAFailure(ctx, user, input)
				return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid recovery code")
			}
			if err := c.mfaRepo.UseRecoveryCode(ctx, matched.ID); err != nil {
				slog.Error("failed to mark recovery code as used", "error", err)
			}
			c.recordAudit(ctx, &user.ID, auditentity.AuditMFARecoveryUsed, "user", user.ID.String(), input, nil)
		}
	}

	// Privileged users without MFA get restricted tokens until they enroll.
	// This enforces MFA for admin accounts when MFA_FORCED_FOR_ADMIN=true.
	if c.mfaForcedForAdmin && (mfaEnrollment == nil || !mfaEnrollment.IsActive()) && len(functions) > 0 {
		restrictedSession = true
	}

	forceChange := user.ForcePasswordChange
	if c.sessionPolicy.PasswordMaxAge > 0 && user.IsPasswordExpired(now, c.sessionPolicy.PasswordMaxAge) {
		forceChange = true
		c.recordAudit(ctx, &user.ID, auditentity.AuditPasswordExpired, "user", user.ID.String(), input, nil)
	}

	if c.sessionPolicy.MaxConcurrent > 0 {
		activeCount, err := c.sessionRepo.CountActiveForUser(ctx, user.ID)
		if err != nil {
			slog.Error("failed to count active sessions", "error", err, "user_id", user.ID)
		} else if activeCount >= c.sessionPolicy.MaxConcurrent {
			if c.sessionPolicy.ConcurrentMode == "reject" {
				return nil, apperrors.NewBusinessError(apperrors.CodeForbidden, "maximum concurrent sessions reached; please logout from another device")
			}
			if err := c.sessionRepo.RevokeOldestForUser(ctx, user.ID, entity.RevokeReasonConcurrentLimit); err != nil {
				slog.Error("failed to evict oldest session", "error", err)
			}
		}
	}

	refreshTokenRaw := c.tokenService.GenerateRefreshToken()
	refreshTokenHash := application.HashRefreshToken(refreshTokenRaw)
	session := &entity.Session{
		ID:                uuid.New(),
		UserID:            user.ID,
		RefreshTokenHash:  refreshTokenHash,
		TokenFamily:       uuid.New(),
		IPAddress:         input.IPAddress,
		UserAgent:         input.UserAgent,
		ExpiresAt:         now.Add(entity.RefreshTokenTTL),
		AbsoluteExpiresAt: now.Add(c.sessionPolicy.AbsoluteLifetime),
		LastActivityAt:    now,
	}
	if err := c.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	accessToken, accessExpiresAt, err := c.tokenService.GenerateAccessTokenForSessionWithProfile(
		user.ID, session.ID, restrictedSession, user.Username, user.Groups,
	)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	user.RecordSuccessfulLogin(now)
	if err := c.userRepo.Update(ctx, user); err != nil {
		slog.Error("failed to update user after successful login", "error", err, "user_id", user.ID)
	}

	contracts, err := c.permsFetcher.GetUserDataPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching data permissions: %w", err)
	}

	c.auditLoginSuccess(ctx, user, input)

	return &dto.LoginResult{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshTokenRaw,
		RefreshTokenExpiresAt: session.ExpiresAt,
		ForcePasswordChange:   forceChange,
		MFAEnrollmentRequired: restrictedSession,
		RestrictedSession:     restrictedSession,
		User: dto.UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			Groups:      user.Groups,
		},
		Permissions: dto.UserPermissions{
			Functions: functions,
			Contracts: contracts,
		},
	}, nil
}

// auditLoginSuccess records a successful login event.
func (c *LoginCommand) auditLoginSuccess(ctx context.Context, user *entity.User, input LoginInput) {
	c.recordAudit(ctx, &user.ID, auditentity.AuditLoginSuccess, "user", user.ID.String(), input, nil)
}

// auditLoginFailure records a failed login attempt.
func (c *LoginCommand) auditLoginFailure(ctx context.Context, userID *uuid.UUID, input LoginInput, reason string) {
	targetID := ""
	if userID != nil {
		targetID = userID.String()
	}
	c.recordAudit(ctx, userID, auditentity.AuditLoginFailure, "user", targetID, input, map[string]interface{}{"reason": reason})
}

// auditAccountLocked records an account lockout event.
func (c *LoginCommand) auditAccountLocked(ctx context.Context, user *entity.User, input LoginInput) {
	c.recordAudit(ctx, &user.ID, auditentity.AuditAccountLocked, "user", user.ID.String(), input, map[string]interface{}{
		"failed_attempts": user.FailedLoginAttempts,
	})
}

// auditMFAFailure records a failed MFA challenge.
func (c *LoginCommand) auditMFAFailure(ctx context.Context, user *entity.User, input LoginInput) {
	c.recordAudit(ctx, &user.ID, auditentity.AuditMFAChallengeFail, "user", user.ID.String(), input, nil)
}

func (c *LoginCommand) recordAudit(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID string, input LoginInput, extra map[string]interface{}) {
	c.auditService.Record(ctx, actorID, eventType, targetType, targetID, input.IPAddress, input.UserAgent, extra)
}

// ExtractIPAddress extracts the client IP from the HTTP request.
// Delegates to middleware.GetClientIP which respects trusted proxy configuration.
func ExtractIPAddress(r *http.Request) string {
	return middleware.GetClientIP(r)
}
