package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval"
	approvaladapter "github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
	chatprovider "github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/provider"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment"
	investsvc "github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	marketdata "github.com/neo-kanta/ims-th-solution/backend/internal/market_data"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification"
	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions"
	referencedata "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
	"github.com/neo-kanta/ims-th-solution/backend/platform/metrics"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"

	_ "github.com/neo-kanta/ims-th-solution/backend/docs"
	_ "github.com/neo-kanta/ims-th-solution/backend/docs/v2"
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

	// Module
	complianceModule := compliance.NewModule(pool, iamModule)
	workflowModule := workflow.NewModule(pool, iamModule, complianceModule.ContractAdapter())
	investmentModule := investment.NewModule(
		pool,
		complianceModule.ContractAdapter(),
		workflowModule,
		iamModule,
		auditModule.Recorder(),
	)
	referenceDataModule := referencedata.NewModule(pool)
	marketDataModule := marketdata.NewModule(pool, cfg, redisClient, referenceDataModule.Resolver())
	integrationModule := integration.NewModule(
		pool,
		iamModule,
		investmentModule.ValuationSummaryProvider(cfg.ReportingCurrency, marketDataModule.QuoteProvider()),
	)
	permissionsModule := permissions.NewModule(pool, iamModule)
	notificationModule := notification.NewModule(pool, cfg, iamModule)
	approvalModule := approval.NewModule(pool, iamModule, auditModule.Recorder(), notificationModule.ApprovalNotifier(), approvaladapter.NewPostgresDelegateResolver(pool), nil)
	watchlistModule := watchlist.NewModule(watchlist.Dependencies{
		Pool:             pool,
		SecurityResolver: referenceDataModule.Resolver(),
		QuoteProvider:    marketDataModule.QuoteProvider(),
		PortfolioScope:   investmentModule.PortfolioScopeResolver(),
		Notifier:         notificationModule.WatchlistAlertNotifier(),
		AuditRecorder:    auditModule.Recorder(),
		IAM:              iamModule,
	})

	// Chat module. Builds the configured LLM provider, persists sessions +
	// messages, and connects the MCP client through which ALL business data
	// is reached. If the active provider is misconfigured (e.g. missing API
	// key) we log and disable the /chat endpoint — the rest of the API stays
	// up. Chat must not import other modules' internals; it sees the audit
	// Recorder and a structural IAM permission port only.
	imsMCPBin := cfg.ChatMCPIMSBin
	if imsMCPBin == "" {
		if exe, exeErr := os.Executable(); exeErr == nil {
			imsMCPBin = filepath.Join(filepath.Dir(exe), "ims-mcp")
		} else {
			imsMCPBin = "ims-mcp"
		}
	}
	chatModule, chatErr := chat.NewModule(pool, chat.Config{
		Provider: chatprovider.Config{
			ActiveProvider:  valueobject.ProviderID(cfg.LLMProvider),
			AnthropicAPIKey: cfg.AnthropicAPIKey,
			AnthropicModel:  cfg.AnthropicModel,
		},
		MaxTokensPerTurn: cfg.ChatMaxTokensPerTurn,
		WriteEnabled:     cfg.ChatWriteEnabled,
		MCPConfigPath:    cfg.ChatMCPServersConfig,
		MCPIMSBinPath:    imsMCPBin,
		IMSAPIBaseURL:    cfg.IMSAPIBaseURL,
	}, auditModule.Recorder(), iamModule)
	if chatErr != nil {
		slog.Warn("Chat module disabled", "error", chatErr, "provider", cfg.LLMProvider)
	}
	if chatModule != nil {
		defer func() { _ = chatModule.Close() }()
	}
	investmentModule.SetPortfolioComplianceAdmin(complianceModule.PortfolioContractAdapter())
	investmentModule.SetApprovalSubmitter(approvalModule)
	investmentModule.SetApprovalStatusProvider(approvalModule)
	investmentModule.SetApprovalBatchActor(approvalModule)
	investmentModule.SetApprovalCanceller(approvalModule)
	approvalModule.RegisterSubjectCallback("RESEARCH_REPORT", investmentModule.ApprovalSubjectCallback())
	approvalModule.RegisterSubjectCallback("INVESTMENT_DECISION", investmentModule.DecisionApprovalSubjectCallback())
	approvalModule.RegisterSubjectCallback("COMPLIANCE_RELEASE", investmentModule.ComplianceReleaseSubjectCallback())
	// PORTFOLIO approval is deferred (Option A): PortfolioApprovalCallback exists but is not
	// registered until the subject access adapter supports PORTFOLIO (see docs/handoff/approval-subject-access-port.md).
	approvalModule.RegisterSubjectValidator("RESEARCH_REPORT", investmentModule.ResearchReportSubjectValidator())
	approvalModule.RegisterSubjectValidator("INVESTMENT_DECISION", investmentModule.DecisionSubjectValidator())
	approvalModule.RegisterSubjectValidator("COMPLIANCE_RELEASE", investmentModule.ComplianceReleaseSubjectValidator())

	// Register per-subject-type access ports so the approval engine enforces
	// object-level authorisation without inspecting business-specific keys.
	// COMPLIANCE_RELEASE additionally requires the INVESTMENT_COMPLIANCE_RELEASE_APPROVE
	// function permission (enforced inside the accessor, not here).
	// PORTFOLIO is intentionally excluded until its access port is implemented.
	investSubjectAccessor := investmentModule.SubjectAccessor(iamModule)
	if investSubjectAccessor != nil {
		approvalModule.RegisterSubjectAccessPort("RESEARCH_REPORT", investSubjectAccessor)
		approvalModule.RegisterSubjectAccessPort("INVESTMENT_DECISION", investSubjectAccessor)
		approvalModule.RegisterSubjectAccessPort("COMPLIANCE_RELEASE", investSubjectAccessor)
	} else {
		slog.Warn("investment subject accessor is nil; approval subject access ports not registered — approval reads will fail closed")
	}

	// Install the trade-confirmation gate on the workflow CloseTransactions
	// handler so close-day refuses to advance while broker confirmations are
	// pending or unresolved mismatches remain.
	workflowModule.SetConfirmationGate(investmentModule.TradeConfirmationGate())

	// Wire the contract catalog so workflow can resolve ?contractCode= query
	// params to internal UUIDs without importing investment internals.
	// Current implementation lives in investment temporarily; future move to
	// ReferenceData/ContractMaster only requires changing this single line.
	workflowModule.SetContractCatalog(investmentModule.ContractCatalog())

	// Wire the cross-module market-data quote provider into the investment
	// module's intraday valuation service. Done post-construction so neither
	// module imports the other's internal/ package.
	if quote := marketDataModule.QuoteProvider(); quote != nil {
		investmentModule.SetMarketQuoteProvider(quote, investsvc.IntradayConfig{
			StaleAfter: cfg.MarketDataStaleAfter,
		})
	}

	// Wire the stuck-day watcher's operator notifier so the workflow scheduler
	// raises an in-app notification when a day sits in DAY_OPEN or
	// MANAGER_APPROVED past the configured cutoff.
	workflowModule.SetOperatorNotifier(notificationModule.WorkflowStuckDayNotifier())

	healthHandler := NewHealthHandler(pool, redisClient, chatModule)

	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	defer stopScheduler()
	workflowModule.StartScheduler(schedulerCtx, time.Hour)
	notificationModule.StartWorker(schedulerCtx)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.NewCORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.SecureHeaders)
	r.Use(iamModule.GlobalRateLimitMiddleware())

	r.Get("/health", healthHandler.Get)

	// Prometheus metrics (includes chat_* observability). Mount on a network
	// the operator considers safe to scrape; no secrets are exposed.
	r.Handle("/metrics", metrics.Handler())

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Portfolio V2 has its own Swagger spec (basePath /api/v2) because
	// Swagger 2.0 only supports one basePath per spec — see
	// cmd/server/swagger_v2_docs.go for why.
	r.Get("/swagger/v2/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/v2/doc.json"),
		httpSwagger.InstanceName("v2"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		iamModule.SetupRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(iamModule.AuthMiddleware())

			complianceModule.RegisterRoutes(r)
			workflowModule.RegisterRoutes(r)
			investmentModule.RegisterRoutes(r)
			referenceDataModule.RegisterRoutes(r)
			marketDataModule.RegisterRoutes(r)
			integrationModule.RegisterRoutes(r)
			permissionsModule.RegisterRoutes(r)
			approvalModule.RegisterRoutes(r)
			notificationModule.RegisterRoutes(r)
			watchlistModule.RegisterRoutes(r, iamModule)
			if chatModule != nil {
				chatModule.RegisterRoutes(r)
			}
		})
	})

	// Portfolio V2 (docs/api/portfolio-v2-api-ddd.md): additive, portfolio-code
	// routes alongside the V1 API above. Only the investment module exposes V2
	// routes so far — this grows as other modules migrate to portfolio-code
	// identity. Documented at /swagger/v2/*, not /swagger/* (see
	// swagger_v2_docs.go).
	r.Route("/api/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(iamModule.AuthMiddleware())

			investmentModule.RegisterRoutesV2(r)
		})
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

		stopScheduler()

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
	chatModule  *chat.Module
}

type HealthResponse struct {
	Status   string       `json:"status"`
	Database string       `json:"database"`
	Redis    string       `json:"redis"`
	Chat     *chat.Health `json:"chat,omitempty"`
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client, chatModule *chat.Module) *HealthHandler {
	return &HealthHandler{
		pool:        pool,
		redisClient: redisClient,
		chatModule:  chatModule,
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
	if h.chatModule != nil {
		ch := h.chatModule.HealthSnapshot(r.Context())
		resp.Chat = &ch
	}

	if dbErr != nil || redisStatus == "unhealthy" {
		resp.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Chat/MCP readiness is informational and does not flip the overall status:
	// the rest of the API stays healthy even when chat is disabled or MCP is down.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("health: encode response failed", "error", err)
	}
}

func boolToHealth(ok bool) string {
	if ok {
		return "ok"
	}
	return "unhealthy"
}
