package iam

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/ims-th-solution/backend/internal/iam/application"
	"github.com/your-org/ims-th-solution/backend/internal/iam/application/command"
	"github.com/your-org/ims-th-solution/backend/internal/iam/application/query"
	"github.com/your-org/ims-th-solution/backend/internal/iam/infrastructure/adapter"
	"github.com/your-org/ims-th-solution/backend/internal/iam/infrastructure/persistence"
	"github.com/your-org/ims-th-solution/backend/internal/iam/transport/handler"
	"github.com/your-org/ims-th-solution/backend/platform/middleware"
)

// Module is the IAM (Identity and Access Management) module.
// It provides authentication endpoints: login, me, and token refresh.
type Module struct {
	authHandler *handler.AuthHandler
	jwtSecret   string
}

// NewModule creates a new IAM module, wiring all dependencies.
func NewModule(pool *pgxpool.Pool, jwtSecret string) *Module {
	// Infrastructure
	userRepo := persistence.NewPostgresUserRepository(pool)
	permsFetcher := adapter.NewPermissionsFetcher(pool)

	// Application services
	tokenSvc := application.NewTokenService(jwtSecret)
	loginCmd := command.NewLoginCommand(userRepo, tokenSvc, permsFetcher)
	getMeQry := query.NewGetMeQuery(userRepo)

	// Transport
	authHandler := handler.NewAuthHandler(loginCmd, getMeQry, tokenSvc, userRepo)

	return &Module{
		authHandler: authHandler,
		jwtSecret:   jwtSecret,
	}
}

// RegisterPublicRoutes mounts routes that do NOT require authentication.
// These are mounted outside the auth middleware group.
func (m *Module) RegisterPublicRoutes(r chi.Router) {
	r.Post("/auth/login", m.authHandler.Login)
}

// RegisterProtectedRoutes mounts routes that require authentication.
// These are mounted inside the auth middleware group.
func (m *Module) RegisterProtectedRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(m.jwtSecret))
		r.Get("/auth/me", m.authHandler.GetMe)
		r.Post("/auth/refresh", m.authHandler.RefreshToken)
	})
}
