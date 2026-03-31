package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	iammod "github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/neo-kanta/ims-th-solution/backend/docs"
)

// @title           IMS Thailand API
// @version         1.0
// @description     This is the API server for the IMS Thailand solution.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Create structured logger
	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("starting IMS backend",
		"env", cfg.Env,
		"port", cfg.Port,
	)

	// 3. Connect to database
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("database connected",
		"host", cfg.DBHost,
		"database", cfg.DBName,
	)

	// 3b. Run migrations
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "/app/migrations"
	}
	if err := database.RunMigrations(cfg.MigrationDSN(), migrationsPath); err != nil {
		logger.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	if err := database.EnsureSecureBootstrap(ctx, pool, cfg.Env); err != nil {
		logger.Error("secure bootstrap guard failed", "error", err)
		os.Exit(1)
	}

	// 4. Wire modules
	iamModule, err := iammod.NewModule(pool, cfg)
	if err != nil {
		logger.Error("failed to initialize IAM module", "error", err)
		os.Exit(1)
	}

	// 5. Build router
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.NewCORS())
	r.Use(middleware.SecureHeaders)

	// Health check (no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(r.Context(), pool); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"database connection failed"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Build key provider for JWT auth middleware (supports rotation)
	keyProvider := middleware.NewKeyProvider(cfg.JWTKeyID, cfg.JWTSecret, cfg.JWTSecretPrevious)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth required)
		iamModule.RegisterPublicRoutes(r)

		// Swagger UI
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/api/v1/swagger/doc.json"),
		))

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(keyProvider, iamModule))

			iamModule.RegisterProtectedRoutes(r)

			// TODO: Mount other module protected routes
			// workflowModule.RegisterRoutes(r)
			// investmentModule.RegisterRoutes(r)
			// permissionsModule.RegisterRoutes(r)
		})
	})

	// 6. Start server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		logger.Info("Swagger UI available at", "url", fmt.Sprintf("http://localhost:%s/api/v1/swagger/index.html", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced shutdown", "error", err)
	}

	logger.Info("server stopped")
}
