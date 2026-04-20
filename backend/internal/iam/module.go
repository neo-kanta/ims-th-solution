package iam

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/query"
	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// AdminRouteRegistrar allows external modules to mount admin-scoped routes under IAM security.
type AdminRouteRegistrar interface {
	RegisterAdminRoutes(
		r chi.Router,
		permissionChecker middleware.PermissionChecker,
		exportLimiter middleware.RateLimiter,
		exportPolicy middleware.RateLimitPolicy,
	)
}

// Module is the IAM (Identity and Access Management) module.
type Module struct {
	authHandler      *handler.AuthHandler
	mfaHandler       *handler.MFAHandler
	sessionHandler   *handler.SessionHandler
	adminHandler     *handler.AdminHandler
	userRepo         domain.UserRepository
	permsFetcher     domain.PermissionsFetcher
	clock            clock.Clock
	authzSvc         *appservice.AuthorizationService
	sessionSvc       *appservice.SessionService
	adminCIDRs       []string
	auditAdminRoutes AdminRouteRegistrar

	// Rate limiters (one per policy tier)
	globalLimiter       middleware.RateLimiter
	loginLimiter        middleware.RateLimiter
	refreshLimiter      middleware.RateLimiter
	refreshTokenLimiter middleware.RateLimiter
	sensitiveLimiter    middleware.RateLimiter
	adminLimiter        middleware.RateLimiter
	exportLimiter       middleware.RateLimiter

	// Rate limit policies
	globalPolicy       middleware.RateLimitPolicy
	loginPolicy        middleware.RateLimitPolicy
	refreshPolicy      middleware.RateLimitPolicy
	refreshTokenPolicy middleware.RateLimitPolicy
	sensitivePolicy    middleware.RateLimitPolicy
	adminPolicy        middleware.RateLimitPolicy
	exportPolicy       middleware.RateLimitPolicy

	// Auth middleware configuration
	keyProvider       middleware.KeyProvider
	keyID             string
	jwtSecret         string
	jwtSecretPrevious string
}

// NewModule creates a new IAM module, wiring all dependencies.
// redisClient may be nil when RATE_LIMIT_BACKEND is "memory".
func NewModule(
	pool *pgxpool.Pool,
	cfg *config.AppConfig,
	redisClient *redis.Client,
	auditRecorder auditdomain.Recorder,
	auditAdminRoutes AdminRouteRegistrar,
) (*Module, error) {
	clk := clock.RealClock{}
	if auditRecorder == nil {
		auditRecorder = auditdomain.NopRecorder{}
	}

	// Configure trusted proxies globally
	middleware.SetTrustedProxies(middleware.NewTrustedProxyConfig(cfg.TrustedProxies))

	// Infrastructure
	userRepo := persistence.NewPostgresUserRepository(pool)
	sessionRepo := persistence.NewPostgresSessionRepository(pool)
	mfaRepo := persistence.NewPostgresMFARepository(pool)
	permsFetcher := adapter.NewPermissionsFetcher(pool)
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

	// Lockout policy from config
	lockoutPolicy := entity.LockoutPolicy{
		MaxFailedAttempts: cfg.LoginMaxFailedAttempts,
		LockoutDuration:   cfg.LoginLockoutDuration,
	}
	if lockoutPolicy.MaxFailedAttempts <= 0 {
		lockoutPolicy = entity.DefaultLockoutPolicy()
	}

	// Session policy
	sessionPolicy := command.SessionPolicy{
		IdleTimeout:      cfg.SessionIdleTimeout,
		AbsoluteLifetime: cfg.SessionAbsoluteLife,
		MaxConcurrent:    cfg.SessionMaxConcurrent,
		ConcurrentMode:   cfg.SessionConcurrentMode,
		PasswordMaxAge:   cfg.PasswordMaxAge(),
		LockoutPolicy:    lockoutPolicy,
	}

	// Application services
	loginCmd := command.NewLoginCommand(userRepo, sessionRepo, mfaRepo, permsFetcher, tokenSvc, totpSvc, mfaChallengeSvc, auditRecorder, clk, sessionPolicy, cfg.MFAForcedForAdmin)
	refreshCmd := command.NewRefreshTokenCommand(userRepo, sessionRepo, tokenSvc, auditRecorder, clk, cfg.SessionIdleTimeout)
	logoutCmd := command.NewLogoutCommand(sessionRepo, auditRecorder)
	logoutAllCmd := command.NewLogoutAllCommand(sessionRepo, auditRecorder)
	changePassCmd := command.NewChangePasswordCommand(userRepo, auditRecorder, sessionRepo)
	adminUserCmd := command.NewAdminUserCommand(userRepo, auditRecorder, clk)
	mfaCmd := command.NewMFAEnrollCommand(mfaRepo, userRepo, auditRecorder, totpSvc)
	getMeQry := query.NewGetMeQuery(userRepo, permsFetcher)
	listSessionsQry := query.NewListSessionsQuery(sessionRepo)
	listUsersQry := query.NewListUsersQuery(userRepo)

	// Rate limit policies
	globalPolicy := middleware.RateLimitPolicy{Name: "global", Max: cfg.RateLimitGlobalPerIP, Window: cfg.RateLimitGlobalWindow}
	loginPolicy := middleware.RateLimitPolicy{Name: "login", Max: cfg.RateLimitLoginPerIP, Window: cfg.RateLimitLoginWindow}
	refreshPolicy := middleware.RateLimitPolicy{Name: "refresh", Max: cfg.RateLimitRefreshPerIP, Window: cfg.RateLimitRefreshWindow}
	refreshTokenPolicy := middleware.RateLimitPolicy{Name: "refresh_token", Max: 10, Window: 15 * time.Minute}
	sensitivePolicy := middleware.RateLimitPolicy{Name: "sensitive", Max: cfg.RateLimitSensitiveMax, Window: cfg.RateLimitSensitiveWindow}
	adminPolicy := middleware.RateLimitPolicy{Name: "admin", Max: cfg.RateLimitAdminMax, Window: cfg.RateLimitAdminWindow}
	exportPolicy := middleware.RateLimitPolicy{Name: "export", Max: cfg.RateLimitExportMax, Window: cfg.RateLimitExportWindow}

	// Rate limiters (one per tier — each has its own window/max)
	newLimiter := func(window time.Duration, max int) middleware.RateLimiter {
		if cfg.RateLimitBackend == "redis" && redisClient != nil {
			return middleware.NewRedisRateLimiter(redisClient, window, max)
		}
		return middleware.NewInMemoryRateLimiter(window, max)
	}
	globalLimiter := newLimiter(globalPolicy.Window, globalPolicy.Max)
	loginLimiter := newLimiter(loginPolicy.Window, loginPolicy.Max)
	refreshLimiter := newLimiter(refreshPolicy.Window, refreshPolicy.Max)
	refreshTokenLimiter := newLimiter(refreshTokenPolicy.Window, refreshTokenPolicy.Max)
	sensitiveLimiter := newLimiter(sensitivePolicy.Window, sensitivePolicy.Max)
	adminLimiter := newLimiter(adminPolicy.Window, adminPolicy.Max)
	exportLimiter := newLimiter(exportPolicy.Window, exportPolicy.Max)

	// Per IP+username login limiter (prevents distributed brute force against one account)
	loginUserPolicy := middleware.RateLimitPolicy{Name: "login_user", Max: cfg.RateLimitLoginPerUser, Window: cfg.RateLimitLoginWindow}
	loginUserLimiter := newLimiter(loginUserPolicy.Window, loginUserPolicy.Max)

	// Transport
	authHandler := handler.NewAuthHandler(loginCmd, refreshCmd, logoutCmd, logoutAllCmd, changePassCmd, getMeQry, loginLimiter, loginPolicy, loginUserLimiter, loginUserPolicy, refreshTokenLimiter, refreshTokenPolicy)
	mfaHandler := handler.NewMFAHandler(mfaCmd, cfg.Env == "development" || cfg.Env == "test")
	sessionHandler := handler.NewSessionHandler(listSessionsQry, sessionRepo)
	adminHandler := handler.NewAdminHandler(adminUserCmd, listUsersQry)

	// Create key provider for auth middleware
	keyProvider := middleware.NewKeyProvider(cfg.JWTKeyID, cfg.JWTSecret, cfg.JWTSecretPrevious)

	return &Module{
		authHandler:      authHandler,
		mfaHandler:       mfaHandler,
		sessionHandler:   sessionHandler,
		adminHandler:     adminHandler,
		userRepo:         userRepo,
		permsFetcher:     permsFetcher,
		clock:            clk,
		authzSvc:         authzSvc,
		sessionSvc:       sessionSvc,
		adminCIDRs:       cfg.AdminIPAllowlist,
		auditAdminRoutes: auditAdminRoutes,

		globalLimiter:       globalLimiter,
		loginLimiter:        loginLimiter,
		refreshLimiter:      refreshLimiter,
		refreshTokenLimiter: refreshTokenLimiter,
		sensitiveLimiter:    sensitiveLimiter,
		adminLimiter:        adminLimiter,
		exportLimiter:       exportLimiter,

		globalPolicy:       globalPolicy,
		loginPolicy:        loginPolicy,
		refreshPolicy:      refreshPolicy,
		refreshTokenPolicy: refreshTokenPolicy,
		sensitivePolicy:    sensitivePolicy,
		adminPolicy:        adminPolicy,
		exportPolicy:       exportPolicy,

		keyProvider:       keyProvider,
		keyID:             cfg.JWTKeyID,
		jwtSecret:         cfg.JWTSecret,
		jwtSecretPrevious: cfg.JWTSecretPrevious,
	}, nil
}

// GlobalRateLimitMiddleware returns the global per-IP rate limiter middleware.
// Intended to be applied at the top-level router in main.go.
func (m *Module) GlobalRateLimitMiddleware() func(http.Handler) http.Handler {
	return middleware.RateLimit(m.globalLimiter, m.globalPolicy)
}

// SetupRoutes configures all IAM routes with proper middleware scoping.
// Rate limiting is applied per tier:
//   - PUBLIC: /auth/login (login limiter), /auth/refresh (refresh limiter)
//   - PROTECTED: auth, MFA, session endpoints (JWT required; sensitive ops get stricter limits)
//   - ADMIN: admin operations (JWT + IP allowlist + permission checks + admin limiter)
func (m *Module) SetupRoutes(r chi.Router) {
	// ==== PUBLIC ROUTES (no authentication required) ====
	// Login with login-specific rate limiting
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(m.loginLimiter, m.loginPolicy))
		r.Post("/auth/login", m.authHandler.Login)
	})

	// Refresh with refresh-specific rate limiting
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit(m.refreshLimiter, m.refreshPolicy))
		r.Post("/auth/refresh", m.authHandler.RefreshToken)
	})

	// ==== PROTECTED ROUTES (JWT required) ====
	protectedRouter := chi.NewRouter()
	protectedRouter.Use(middleware.Auth(m.keyProvider, m))

	// Standard auth endpoints (covered by global limiter only)
	protectedRouter.Get("/auth/me", m.authHandler.GetMe)
	protectedRouter.Post("/auth/logout", m.authHandler.Logout)
	protectedRouter.Post("/auth/logout-all", m.authHandler.LogoutAll)

	// Sensitive auth endpoints (stricter per-user rate limiting)
	protectedRouter.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitByUser(m.sensitiveLimiter, m.sensitivePolicy))
		r.Post("/auth/change-password", m.authHandler.ChangePassword)
	})

	// MFA endpoints — verify and disable are sensitive (TOTP brute-force risk)
	protectedRouter.Post("/auth/mfa/enroll", m.mfaHandler.Enroll)
	protectedRouter.Get("/auth/mfa/status", m.mfaHandler.Status)
	protectedRouter.Get("/auth/mfa/dev/totp-code", m.mfaHandler.DevTOTPCode)
	protectedRouter.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitByUser(m.sensitiveLimiter, m.sensitivePolicy))
		r.Post("/auth/mfa/verify", m.mfaHandler.Verify)
		r.Post("/auth/mfa/disable", m.mfaHandler.Disable)
	})

	// Session management
	protectedRouter.Get("/auth/sessions", m.sessionHandler.ListMySessions)
	protectedRouter.Post("/auth/sessions/{id}/revoke", m.sessionHandler.RevokeSession)

	// ==== ADMIN ROUTES (JWT + IP allowlist + admin rate limiter + permission checks) ====
	protectedRouter.Route("/admin", func(adminRouter chi.Router) {
		adminRouter.Use(middleware.IPAllowlist(m.adminCIDRs))
		adminRouter.Use(middleware.RateLimitByUser(m.adminLimiter, m.adminPolicy))

		// User management
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_VIEW")).Get("/users", m.adminHandler.ListUsers)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_CREATE")).Post("/users", m.adminHandler.CreateUser)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_DEACTIVATE")).Post("/users/{id}/disable", m.adminHandler.DisableUser)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_DEACTIVATE")).Post("/users/{id}/enable", m.adminHandler.EnableUser)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/lock", m.adminHandler.LockUser)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/unlock", m.adminHandler.UnlockUser)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/users/{id}/reset-password", m.adminHandler.ResetPassword)

		// Admin session management
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Get("/users/{id}/sessions", m.sessionHandler.AdminListUserSessions)
		adminRouter.With(middleware.RequirePermission(m, "IAM_USER_UPDATE")).Post("/sessions/{id}/revoke", m.sessionHandler.AdminRevokeSession)

		if m.auditAdminRoutes != nil {
			m.auditAdminRoutes.RegisterAdminRoutes(adminRouter, m, m.exportLimiter, m.exportPolicy)
		}
	})

	// Mount protected router
	r.Mount("/", protectedRouter)
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
