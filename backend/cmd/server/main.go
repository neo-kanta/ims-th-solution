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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"

	_ "github.com/neo-kanta/ims-th-solution/backend/docs"
)

// @title           IMS Thailand API
// @version         1.0.0
// @description     Enterprise Investment Management System API with financial-grade security
// @termsOfService  https://example.com/terms
// @contact.name    Support Team
// @contact.email   support@example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @basePath        /api/v1
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer Token. Format: "Bearer <token>"
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.EnsureSecureBootstrap(ctx, pool, cfg.Env); err != nil {
		slog.Error("Secure bootstrap check failed", "error", err)
		os.Exit(1)
	}

	var redisClient *redis.Client
	if cfg.RateLimitBackend == "redis" {
		rc, err := database.NewRedisClient(ctx, cfg)
		if err != nil {
			slog.Error("Failed to connect to Redis (required for RATE_LIMIT_BACKEND=redis)", "error", err)
			os.Exit(1)
		}
		slog.Info("Redis Action...")
		redisClient = rc
		defer redisClient.Close()
	}

	auditModule := audit.NewModule(pool)

	iamModule, err := iam.NewModule(pool, cfg, redisClient, auditModule.Recorder(), auditModule)
	if err != nil {
		slog.Error("Failed to initialize IAM module", "error", err)
		os.Exit(1)
	}
	// Compliance / IRG module
	complianceModule := compliance.NewModule(pool, iamModule)

	healthHandler := NewHealthHandler(pool, redisClient)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.NewCORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.SecureHeaders)
	r.Use(iamModule.GlobalRateLimitMiddleware())

	r.Get("/health", healthHandler.Get)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		iamModule.SetupRoutes(r)
		complianceModule.RegisterRoutes(r)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("Starting server", "address", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// HealthHandler handles system-level utility endpoints.
type HealthHandler struct {
	pool        *pgxpool.Pool
	redisClient *redis.Client
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		pool:        pool,
		redisClient: redisClient,
	}
}

// Get handles GET /health.
// @Summary Health Check
// @Description Check backend, database, and Redis health status
// @Tags System
// @Produce json
// @Success 200 {object} HealthResponse
// @Success 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbErr := database.HealthCheck(r.Context(), h.pool)
	var redisStatus string
	if h.redisClient != nil {
		if err := database.RedisHealthCheck(r.Context(), h.redisClient); err != nil {
			redisStatus = "unhealthy"
		} else {
			redisStatus = "ok"
		}
	} else {
		redisStatus = "not_configured"
	}

	resp := HealthResponse{
		Status:   "ok",
		Database: boolToHealth(dbErr == nil),
		Redis:    redisStatus,
	}

	if dbErr != nil || redisStatus == "unhealthy" {
		resp.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintf(w, `{"status":"%s","database":"%s","redis":"%s"}`, resp.Status, resp.Database, resp.Redis)
}

func boolToHealth(ok bool) string {
	if ok {
		return "ok"
	}
	return "unhealthy"
}
