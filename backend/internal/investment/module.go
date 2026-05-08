// Package investment wires the Stock Investment Management module.
//
// Cross-module integrations live behind pkg/contract interfaces — this module
// never imports internal/compliance, internal/iam, or internal/workflow
// directly. Adapters in infrastructure/adapter/ bridge the gap.
package investment

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/infrastructure/persistence"
	invperm "github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module owns the investment module's wire-up: persistence, application
// commands, and HTTP transport.
type Module struct {
	pool *pgxpool.Pool

	// Repositories
	funds       *persistence.PostgresFundRepository
	portfolios  *persistence.PostgresPortfolioRepository
	instruments *persistence.PostgresInstrumentRepository
	positions   *persistence.PostgresPositionRepository
	cash        *persistence.PostgresCashLedgerRepository
	txns        *persistence.PostgresTransactionRepository
	prices      *persistence.PostgresPriceRepository
	valuation   *persistence.PostgresValuationRepository
	taxonomy    *persistence.PostgresTaxonomyRepository

	// Cross-module adapters
	permissionAdapter *adapter.PermissionCheckerAdapter
	auditAdapter      *adapter.AuditLoggerAdapter

	// Application services
	projector    *service.PortfolioProjector
	valuationRun *service.ValuationRunner

	// Application commands
	fundCmd       *command.FundCommandHandler
	portfolioCmd  *command.PortfolioCommandHandler
	instrumentCmd *command.InstrumentCommandHandler
	postTxn       *command.PostTransactionHandler
	reverseTxn    *command.ReverseTransactionHandler
	postPrice     *command.PostPriceSnapshotHandler
	fundAUM       *command.ComputeFundAUMHandler

	submitDecision *command.SubmitDecisionForExecutionHandler

	// Transport
	handler *handler.InvestmentHandler

	// Middleware-side permission checker (uses the IAM port — its
	// HasFunctionPermission signature already matches middleware.PermissionChecker).
	middlewarePerm middleware.PermissionChecker
}

// NewModule constructs the investment module.
//
// Dependencies:
//   - pool:        Postgres pool used by every repository.
//   - compliance:  pre-trade IRG gate (pkg/contract.ComplianceChecker).
//     MUST NOT be nil in production — pre-trade enforcement is a
//     non-negotiable regulatory control.
//   - workflow:    pkg/contract.WorkflowStateProvider — gates ledger posts.
//     MUST NOT be nil in production.
//   - iamPort:     adapter.IAMPermissionPort, satisfied by *iam.Module —
//     used both for object-level data scoping (pkg/contract)
//     and for route-level function-permission middleware.
//   - auditRecorder: audit.Recorder used to emit immutable audit events.
//
// Pass nil for iamPort or auditRecorder only in test contexts; production
// wiring in cmd/server/main.go MUST supply real values.
func NewModule(
	pool *pgxpool.Pool,
	compliance contract.ComplianceChecker,
	workflow contract.WorkflowStateProvider,
	iamPort adapter.IAMPermissionPort,
	auditRecorder auditdomain.Recorder,
) *Module {
	m := &Module{pool: pool}

	// ── Persistence ───────────────────────────────────────────────────────
	m.funds = persistence.NewPostgresFundRepository(pool)
	m.portfolios = persistence.NewPostgresPortfolioRepository(pool)
	m.instruments = persistence.NewPostgresInstrumentRepository(pool)
	m.positions = persistence.NewPostgresPositionRepository(pool)
	m.cash = persistence.NewPostgresCashLedgerRepository(pool)
	m.txns = persistence.NewPostgresTransactionRepository(pool)
	m.prices = persistence.NewPostgresPriceRepository(pool)
	m.valuation = persistence.NewPostgresValuationRepository(pool)
	m.taxonomy = persistence.NewPostgresTaxonomyRepository(pool)

	// ── Adapters ──────────────────────────────────────────────────────────
	m.permissionAdapter = adapter.NewPermissionCheckerAdapter(iamPort)
	m.auditAdapter = adapter.NewAuditLoggerAdapter(auditRecorder)

	// IAMPermissionPort.HasFunctionPermission has the exact signature
	// middleware.PermissionChecker expects, so we can use it directly.
	if iamPort != nil {
		m.middlewarePerm = iamPort
	}

	// ── Application services ──────────────────────────────────────────────
	m.projector = service.NewPortfolioProjectorWithTransactions(m.positions, m.cash, m.txns)
	m.valuationRun = service.NewValuationRunner(
		pool, m.portfolios, m.positions, m.cash, m.prices,
		m.instruments, m.valuation, m.txns, nil,
	)

	// ── Application commands ──────────────────────────────────────────────
	m.fundCmd = command.NewFundCommandHandler(pool, m.funds, m.auditAdapter, nil)
	m.portfolioCmd = command.NewPortfolioCommandHandler(pool, m.portfolios, m.funds, m.auditAdapter, nil)
	m.instrumentCmd = command.NewInstrumentCommandHandler(pool, m.instruments, m.auditAdapter, nil)
	m.postTxn = command.NewPostTransactionHandler(
		pool, m.portfolios, m.funds, m.instruments,
		m.positions, m.txns, m.projector,
		workflow, compliance, m.auditAdapter, nil,
	)
	m.reverseTxn = command.NewReverseTransactionHandler(
		pool, m.txns, m.projector, workflow, m.auditAdapter, nil,
	)
	m.postPrice = command.NewPostPriceSnapshotHandler(pool, m.prices, m.instruments, m.auditAdapter, nil)
	m.fundAUM = command.NewComputeFundAUMHandler(pool, m.funds, m.portfolios, m.valuation, m.auditAdapter, nil)

	// Existing decision-submit pipeline (compliance pre-trade gate).
	// Persistence for the Decision aggregate is not yet implemented; keep
	// the wire-up gated by a non-nil repo when that lands.
	_ = compliance

	// ── Transport ─────────────────────────────────────────────────────────
	m.handler = handler.NewInvestmentHandler(
		m.permissionAdapter,
		m.funds, m.portfolios, m.instruments,
		m.positions, m.cash, m.txns,
		m.prices, m.valuation, m.taxonomy,
		m.fundCmd, m.portfolioCmd, m.instrumentCmd,
		m.postTxn, m.reverseTxn, m.postPrice,
		m.fundAUM,
		m.valuationRun,
	)

	return m
}

// RegisterRoutes mounts the investment module's HTTP routes onto the given
// authenticated router. Expected to be called inside an r.Route("/api/v1", ...)
// block that already has auth middleware applied.
//
// Each route group is gated by a function permission code via
// middleware.RequirePermission.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil || m.middlewarePerm == nil {
		return
	}
	pc := m.middlewarePerm
	h := m.handler

	r.Route("/investment", func(r chi.Router) {
		// ── Reference / taxonomy ────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeReferenceView))
			r.Get("/reference/asset-classes", h.ListAssetClasses)
			r.Get("/reference/asset-subtypes", h.ListAssetSubtypes)
			r.Get("/reference/sectors", h.ListSectors)
			r.Get("/reference/fund-categories", h.ListFundCategories)
			r.Get("/reference/regions", h.ListRegions)
			r.Get("/reference/countries", h.ListCountries)
			r.Get("/reference/investment-styles", h.ListInvestmentStyles)
		})

		// ── Funds ───────────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeFundView))
			r.Get("/funds", h.ListFunds)
			r.Get("/funds/{id}", h.GetFund)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeFundManage))
			r.Post("/funds", h.CreateFund)
			r.Put("/funds/{id}", h.UpdateFund)
			r.Delete("/funds/{id}", h.DeleteFund)
		})

		// ── Portfolios ──────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodePortfolioView))
			r.Get("/portfolios", h.ListPortfolios)
			r.Get("/portfolios/{id}", h.GetPortfolio)
			r.Get("/portfolios/{id}/holdings", h.ListHoldings)
			r.Get("/portfolios/{id}/transactions", h.ListTransactions)
			r.Get("/portfolios/{id}/cash", h.ListCashBalances)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodePortfolioManage))
			r.Post("/portfolios", h.CreatePortfolio)
			r.Put("/portfolios/{id}", h.UpdatePortfolio)
			r.Delete("/portfolios/{id}", h.DeletePortfolio)
		})

		// ── Instruments ─────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeInstrumentView))
			r.Get("/instruments", h.ListInstruments)
			r.Get("/instruments/{id}", h.GetInstrument)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeInstrumentManage))
			r.Post("/instruments", h.CreateInstrument)
			r.Put("/instruments/{id}", h.UpdateInstrument)
		})

		// ── Ledger (post / reverse) ─────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerPost))
			r.Post("/portfolios/{id}/transactions", h.PostTransaction)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerReverse))
			r.Post("/portfolios/{id}/transactions/{txn_id}/reverse", h.ReverseTransaction)
		})

		// ── Prices ──────────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodePricePost))
			r.Post("/instruments/{id}/prices", h.PostPriceSnapshot)
		})

		// ── Valuation ───────────────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeValuationView))
			r.Get("/portfolios/{id}/valuations", h.ListValuations)
			r.Get("/portfolios/{id}/valuations/latest", h.GetLatestValuation)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeValuationRun))
			r.Post("/portfolios/{id}/valuations/run", h.RunValuation)
		})

		// ── Fund AUM aggregation ────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeFundAUMCompute))
			r.Post("/funds/{id}/aum/compute", h.ComputeFundAUM)
		})
	})
}

// SubmitDecisionHandler returns the Decision-submit command handler. Kept
// for the existing pre-trade tests; persistence wiring will land in a
// follow-up batch once the Decision repo is implemented.
func (m *Module) SubmitDecisionHandler() *command.SubmitDecisionForExecutionHandler {
	if m == nil {
		return nil
	}
	return m.submitDecision
}
