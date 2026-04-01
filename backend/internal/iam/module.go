package iam

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/query"
	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module is the IAM (Identity and Access Management) module.
type Module struct {
	authHandler    *handler.AuthHandler
	mfaHandler     *handler.MFAHandler
	sessionHandler *handler.SessionHandler
	auditHandler   *handler.AuditHandler
	adminHandler   *handler.AdminHandler
	userRepo       domain.UserRepository
	permsFetcher   domain.PermissionsFetcher
	clock          clock.Clock
	rateLimiter    *middleware.RateLimiter
	authzSvc       *appservice.AuthorizationService
	sessionSvc     *appservice.SessionService
	adminCIDRs     []string
}

// NewModule creates a new IAM module, wiring all dependencies.
func NewModule(pool *pgxpool.Pool, cfg *config.AppConfig) (*Module, error) {
	clk := clock.RealClock{}

	// Infrastructure
	userRepo := persistence.NewPostgresUserRepository(pool)
	sessionRepo := persistence.NewPostgresSessionRepository(pool)
	auditRepo := persistence.NewPostgresAuditRepository(pool)
	mfaRepo := persistence.NewPostgresMFARepository(pool)
	permsFetcher := adapter.NewPermissionsFetcher(pool)
	auditSvc := appservice.NewAuditService(auditRepo)
	authzSvc := appservice.NewAuthorizationService(permsFetcher)
	sessionSvc := appservice.NewSessionService(sessionRepo)

	// TOTP service
	totpSvc, err := application.NewTOTPService(cfg.MFAEncryptionKey, cfg.MFAIssuerName)
	if err != nil {
		return nil, err
	}
	mfaChallengeSvc := appservice.NewMFAChallengeService(cfg.JWTSecret, clk)

	// Token service with key rotation support
	var tokenSvc *application.TokenService
	if cfg.JWTSecretPrevious != "" {
		tokenSvc = application.NewTokenServiceWithKeyRing(cfg.JWTKeyID, cfg.JWTSecret, cfg.JWTKeyIDPrevious, cfg.JWTSecretPrevious, clk)
	} else {
		tokenSvc = application.NewTokenServiceWithKeyRing(cfg.JWTKeyID, cfg.JWTSecret, "", "", clk)
	}

	// Session policy
	sessionPolicy := command.SessionPolicy{
		IdleTimeout:      cfg.SessionIdleTimeout,
		AbsoluteLifetime: cfg.SessionAbsoluteLife,
		MaxConcurrent:    cfg.SessionMaxConcurrent,
		ConcurrentMode:   cfg.SessionConcurrentMode,
		PasswordMaxAge:   cfg.PasswordMaxAge(),
	}

	// Application services
	loginCmd := command.NewLoginCommand(userRepo, sessionRepo, mfaRepo, permsFetcher, tokenSvc, totpSvc, mfaChallengeSvc, auditSvc, clk, sessionPolicy, cfg.MFAForcedForAdmin)
	refreshCmd := command.NewRefreshTokenCommand(userRepo, sessionRepo, tokenSvc, auditSvc, clk, cfg.SessionIdleTimeout)
	logoutCmd := command.NewLogoutCommand(sessionRepo, auditSvc)
	logoutAllCmd := command.NewLogoutAllCommand(sessionRepo, auditSvc)
	changePassCmd := command.NewChangePasswordCommand(userRepo, auditSvc, sessionRepo)
	adminUserCmd := command.NewAdminUserCommand(userRepo, auditSvc, clk)
	mfaCmd := command.NewMFAEnrollCommand(mfaRepo, userRepo, auditSvc, totpSvc)
	getMeQry := query.NewGetMeQuery(userRepo, permsFetcher)
	listSessionsQry := query.NewListSessionsQuery(sessionRepo)
	listAuditQry := query.NewListAuditEventsQuery(auditRepo)

	// Rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitLoginPerIP, cfg.RateLimitLoginWindow)

	// Transport
	authHandler := handler.NewAuthHandler(loginCmd, refreshCmd, logoutCmd, logoutAllCmd, changePassCmd, getMeQry)
	mfaHandler := handler.NewMFAHandler(mfaCmd)
	sessionHandler := handler.NewSessionHandler(listSessionsQry, sessionRepo)
	auditHandler := handler.NewAuditHandler(listAuditQry)
	adminHandler := handler.NewAdminHandler(adminUserCmd)

	return &Module{
		authHandler:    authHandler,
		mfaHandler:     mfaHandler,
		sessionHandler: sessionHandler,
		auditHandler:   auditHandler,
		adminHandler:   adminHandler,
		userRepo:       userRepo,
		permsFetcher:   permsFetcher,
		clock:          clk,
		rateLimiter:    rateLimiter,
		authzSvc:       authzSvc,
		sessionSvc:     sessionSvc,
		adminCIDRs:     cfg.AdminIPAllowlist,
	}, nil
}

// RegisterPublicRoutes mounts open endpoints that do NOT require authentication.
func (m *Module) RegisterPublicRoutes(r chi.Router) {
	// Login with rate limiting
	r.Group(func(r chi.Router) {
		r.Use(middleware.LoginRateLimit(m.rateLimiter))
		r.Post("/auth/login", m.authHandler.Login)
	})
	r.Post("/auth/refresh", m.authHandler.RefreshToken)
}

// RegisterProtectedRoutes mounts endpoints that require a valid JWT.
func (m *Module) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/auth/me", m.authHandler.GetMe)
	r.Post("/auth/logout", m.authHandler.Logout)
	r.Post("/auth/logout-all", m.authHandler.LogoutAll)
	r.Post("/auth/change-password", m.authHandler.ChangePassword)

	// MFA endpoints
	r.Post("/auth/mfa/enroll", m.mfaHandler.Enroll)
	r.Post("/auth/mfa/verify", m.mfaHandler.Verify)
	r.Post("/auth/mfa/disable", m.mfaHandler.Disable)
	r.Get("/auth/mfa/status", m.mfaHandler.Status)

	// Session management
	r.Get("/auth/sessions", m.sessionHandler.ListMySessions)
	r.Post("/auth/sessions/{id}/revoke", m.sessionHandler.RevokeSession)

	// Admin operations require IAM_ADMIN permission
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.IPAllowlist(m.adminCIDRs))

		r.With(middleware.RequirePermission(m, "IAM_USER_CREATE")).Post("/users", m.adminHandler.CreateUser)
		r.With(middleware.RequirePermission(m, "IAM_USER_DEACTIVATE")).Post("/users/{id}/disable", m.adminHandler.DisableUser)
		r.With(middleware.RequirePermission(m, "IAM_USER_DEACTIVATE")).Post("/users/{id}/enable", m.adminHandler.EnableUser)
		r.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/lock", m.adminHandler.LockUser)
		r.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/unlock", m.adminHandler.UnlockUser)
		r.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/reset-password", m.adminHandler.ResetPassword)

		// Admin session management
		r.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Get("/users/{id}/sessions", m.sessionHandler.AdminListUserSessions)
		r.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/sessions/{id}/revoke", m.sessionHandler.AdminRevokeSession)

		// Audit log query + export
		r.With(middleware.RequirePermission(m, "IAM_AUDIT_VIEW")).Get("/audit", m.auditHandler.ListAuditEvents)
		r.With(middleware.RequirePermission(m, "IAM_AUDIT_VIEW")).Get("/audit/export", m.auditHandler.ExportAuditCSV)
	})
}

// IsUserActive implements the platform's UserStatusChecker interface.
func (m *Module) IsUserActive(ctx context.Context, userID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	user, err := m.userRepo.FindByID(ctx, uid)
	if err != nil || user == nil {
		return false, nil
	}
	allowed, _ := user.IsLoginAllowed(m.clock.Now())
	return allowed, nil
}

// HasFunctionPermission implements the platform's PermissionChecker interface.
func (m *Module) HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	return m.authzSvc.HasFunctionPermission(ctx, uid, code)
}

// HasDataPermission implements data-scope authorization checks for other modules/middleware.
func (m *Module) HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	return m.authzSvc.HasDataPermission(ctx, uid, scopeID)
}

// GetAccessibleContracts returns all effective contract scopes for the given user.
func (m *Module) GetAccessibleContracts(ctx context.Context, userID string) ([]string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return m.authzSvc.GetAccessibleScopes(ctx, uid)
}

// ValidateActiveSession ensures the JWT is still backed by an active server-side session.
func (m *Module) ValidateActiveSession(ctx context.Context, userID string, sessionID string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return false, err
	}
	return m.sessionSvc.ValidateActiveSession(ctx, uid, sid, m.clock.Now())
}

// TouchSessionActivity updates the last-activity marker for an authenticated session.
func (m *Module) TouchSessionActivity(ctx context.Context, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return err
	}
	return m.sessionSvc.TouchSessionActivity(ctx, sid)
}
