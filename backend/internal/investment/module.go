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
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/query"
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
	funds                  *persistence.PostgresFundRepository
	portfolios             *persistence.PostgresPortfolioRepository
	portfolioStatusHistory *persistence.PostgresPortfolioStatusHistoryRepository
	instruments            *persistence.PostgresInstrumentRepository
	positions              *persistence.PostgresPositionRepository
	cash                   *persistence.PostgresCashLedgerRepository
	txns                   *persistence.PostgresTransactionRepository
	prices                 *persistence.PostgresPriceRepository
	valuation              *persistence.PostgresValuationRepository
	taxonomy               *persistence.PostgresTaxonomyRepository
	research               *persistence.PostgresResearchReportRepository
	decisions              *persistence.PostgresDecisionRepository
	decisionLines          *persistence.PostgresDecisionLineRepository
	executions             *persistence.PostgresExecutionRepository
	confirmations          *persistence.PostgresTradeConfirmationRepository
	confirmationImports    *persistence.PostgresTradeConfirmationImportRepository

	// Cross-module adapters
	permissionAdapter *adapter.PermissionCheckerAdapter
	auditAdapter      *adapter.AuditLoggerAdapter

	// Application services
	projector    *service.PortfolioProjector
	valuationRun *service.ValuationRunner
	intraday     *service.IntradayValuationService

	// Application commands
	fundCmd       *command.FundCommandHandler
	portfolioCmd  *command.PortfolioCommandHandler
	instrumentCmd *command.InstrumentCommandHandler
	postTxn       *command.PostTransactionHandler
	reverseTxn    *command.ReverseTransactionHandler
	postPrice     *command.PostPriceSnapshotHandler
	fundAUM       *command.ComputeFundAUMHandler

	// Application queries
	fundNAVQuery   *query.GetLatestFundNAVHandler
	fundAllocQuery *query.GetFundAllocationHandler
	fundNAVHistory *query.GetFundNAVHistoryHandler

	submitDecision        *command.SubmitDecisionForExecutionHandler
	researchCmd           *command.ResearchReportCommandHandler
	decisionCmd           *command.DecisionCommandHandler
	decisionBatchCmd      *command.DecisionBatchApprovalHandler
	executionCmd          *command.ExecutionCommandHandler
	confirmationCmd       *command.TradeConfirmationCommandHandler
	confirmationImportCmd *command.ConfirmationBatchImportHandler

	// Transport
	handler             *handler.InvestmentHandler
	researchHandler     *handler.ResearchReportHandler
	decisionHandler     *handler.DecisionHandler
	executionHandler    *handler.ExecutionHandler
	confirmationHandler *handler.TradeConfirmationHandler
	intradayHandler     *handler.IntradayValuationHandler

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
	m.portfolioStatusHistory = persistence.NewPostgresPortfolioStatusHistoryRepository(pool)
	m.instruments = persistence.NewPostgresInstrumentRepository(pool)
	m.positions = persistence.NewPostgresPositionRepository(pool)
	m.cash = persistence.NewPostgresCashLedgerRepository(pool)
	m.txns = persistence.NewPostgresTransactionRepository(pool)
	m.prices = persistence.NewPostgresPriceRepository(pool)
	m.valuation = persistence.NewPostgresValuationRepository(pool)
	m.taxonomy = persistence.NewPostgresTaxonomyRepository(pool)
	m.research = persistence.NewPostgresResearchReportRepository(pool)
	m.decisions = persistence.NewPostgresDecisionRepository(pool)
	m.decisionLines = persistence.NewPostgresDecisionLineRepository(pool)
	m.executions = persistence.NewPostgresExecutionRepository(pool)
	m.confirmations = persistence.NewPostgresTradeConfirmationRepository(pool)
	m.confirmationImports = persistence.NewPostgresTradeConfirmationImportRepository(pool)

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
	m.portfolioCmd.SetStatusHistoryRepository(m.portfolioStatusHistory)
	m.instrumentCmd = command.NewInstrumentCommandHandler(pool, m.instruments, m.auditAdapter, nil)
	m.postTxn = command.NewPostTransactionHandler(
		pool, m.portfolios, m.funds, m.instruments,
		m.positions, m.cash, m.txns, m.projector,
		workflow, compliance, m.auditAdapter, nil,
	)
	m.reverseTxn = command.NewReverseTransactionHandler(
		pool, m.txns, m.projector, workflow, m.auditAdapter, nil,
	)
	m.postPrice = command.NewPostPriceSnapshotHandler(pool, m.prices, m.instruments, m.auditAdapter, nil)
	m.fundAUM = command.NewComputeFundAUMHandler(pool, m.funds, m.portfolios, m.valuation, m.auditAdapter, nil)
	m.researchCmd = command.NewResearchReportCommandHandler(pool, m.research, m.auditAdapter, nil)
	m.decisionCmd = command.NewDecisionCommandHandler(pool, m.decisions, m.research, workflow, m.auditAdapter, nil)
	m.decisionBatchCmd = command.NewDecisionBatchApprovalHandler(m.decisions, nil)
	if iamPort != nil {
		m.decisionBatchCmd.SetPermissionChecker(iamPort)
	}
	m.executionCmd = command.NewExecutionCommandHandler(pool, m.decisions, m.executions, m.auditAdapter, nil)
	m.confirmationCmd = command.NewTradeConfirmationCommandHandler(pool, m.executions, m.confirmations, m.auditAdapter, nil)
	m.confirmationImportCmd = command.NewConfirmationBatchImportHandler(pool, m.executions, m.confirmations, m.confirmationImports, m.auditAdapter, nil)

	// ── Application queries ───────────────────────────────────────────────
	m.fundNAVQuery = query.NewGetLatestFundNAVHandler(m.funds, m.portfolios, m.cash, m.valuation)
	m.fundAllocQuery = query.NewGetFundAllocationHandler(
		m.funds, m.portfolios, m.valuation, m.cash, pool,
	)
	m.fundNAVHistory = query.NewGetFundNAVHistoryHandler(m.funds, m.portfolios, m.valuation)

	// Wire the pre-trade compliance checker and fund repository into the decision
	// command handler. Both are optional at construction — nil guards inside the
	// handler skip the checks, preserving test compatibility.
	if compliance != nil {
		m.decisionCmd.SetComplianceChecker(compliance)
	}
	m.decisionCmd.SetFundRepository(m.funds)
	m.decisionCmd.SetPortfolioRepository(m.portfolios)
	if workflow != nil {
		m.executionCmd.SetWorkflowStateProvider(workflow)
	}
	if compliance != nil {
		m.executionCmd.SetComplianceChecker(compliance)
	}
	m.executionCmd.SetPortfolioRepository(m.portfolios)

	// ── Transport ─────────────────────────────────────────────────────────
	// Intraday valuation handler ships without a quote provider; main.go
	// wires it via SetMarketQuoteProvider once the market_data module is up
	// (avoids a circular construction dependency).
	m.intradayHandler = handler.NewIntradayValuationHandler(nil, m.auditAdapter)

	m.handler = handler.NewInvestmentHandler(
		m.permissionAdapter,
		m.funds, m.portfolios, m.instruments,
		m.positions, m.cash, m.txns,
		m.prices, m.valuation, m.taxonomy,
		m.fundCmd, m.portfolioCmd, m.instrumentCmd,
		m.postTxn, m.reverseTxn, m.postPrice,
		m.fundAUM,
		m.valuationRun,
		m.fundNAVQuery,
		m.fundAllocQuery,
		m.fundNAVHistory,
	)
	m.researchHandler = handler.NewResearchReportHandler(m.research, m.researchCmd)
	m.decisionHandler = handler.NewDecisionHandler(m.decisions, m.decisionCmd)
	m.decisionHandler.SetDecisionLineRepository(m.decisionLines)
	m.decisionHandler.SetBatchApprovalHandler(m.decisionBatchCmd)
	m.decisionHandler.SetPortfolioRepository(m.portfolios)
	m.executionHandler = handler.NewExecutionHandler(m.executions, m.executionCmd)
	m.executionHandler.SetDecisionRepository(m.decisions)
	m.executionHandler.SetPortfolioRepository(m.portfolios)
	m.confirmationHandler = handler.NewTradeConfirmationHandler(m.confirmations, m.confirmationCmd)
	m.confirmationHandler.SetBatchImportHandler(m.confirmationImportCmd)
	m.confirmationHandler.SetExecutionRepository(m.executions)
	m.confirmationHandler.SetPortfolioRepository(m.portfolios)

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
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerSimulate))
			r.Post("/portfolios/{id}/transactions/simulate", h.SimulateTransaction)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerPost))
			r.Post("/portfolios/{id}/transactions", h.PostTransaction)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerReverse))
			r.Post("/portfolios/{id}/transactions/{txnId}/reverse", h.ReverseTransaction)
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
			r.Get("/funds/{id}/nav/latest", h.GetLatestFundNAV)
			r.Get("/funds/{id}/allocation", h.GetFundAllocation)
			r.Get("/funds/{id}/nav-history", h.GetFundNAVHistory)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeValuationRun))
			r.Post("/portfolios/{id}/valuations/run", h.RunValuation)
		})

		// ── Intraday (live market data) ─────────────────────────────────
		// Live valuation pulled from a market-data provider port. Official
		// accounting NAV/AUM tables are NOT mutated by these routes —
		// values are computed on the fly and surfaced alongside the
		// official numbers for the dashboard. Gracefully degrades to
		// cached snapshots when the provider chain is down.
		if m.intradayHandler != nil {
			ih := m.intradayHandler
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeValuationView))
				r.Get("/funds/{id}/holdings/valuation", ih.GetFundHoldingsValuation)
				r.Get("/funds/{id}/market-data/status", ih.GetFundMarketDataStatus)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeValuationRun))
				r.Post("/funds/{id}/market-data/refresh", ih.RefreshFundMarketData)
			})
		}

		// ── Fund AUM aggregation ────────────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeFundAUMCompute))
			r.Post("/funds/{id}/aum/compute", h.ComputeFundAUM)
		})

		// ── Decisions (CRUD + lifecycle + batch approval) ──────────────
		if m.decisionHandler != nil {
			dh := m.decisionHandler
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionView))
				r.Get("/decisions", dh.ListDecisions)
				r.Get("/decisions/{id}", dh.GetDecision)
				r.Get("/decisions/{id}/details", dh.GetDecisionWithLines)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionApprove))
				r.Get("/decisions/approval-items", dh.ListApprovalItems)
				r.Post("/decisions/batch-approve", dh.BatchApproveDecisions)
				r.Post("/decisions/batch-reject", dh.BatchRejectDecisions)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionManage))
				r.Post("/decisions", dh.CreateDecision)
				r.Put("/decisions/{id}", dh.UpdateDecision)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionSubmit))
				r.Post("/decisions/{id}/submit", dh.SubmitDecision)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionCancel))
				r.Post("/decisions/{id}/cancel", dh.CancelDecision)
			})
		}

		// ── Executions ──────────────────────────────────────────────────
		if m.executionHandler != nil {
			eh := m.executionHandler
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeExecutionView))
				r.Get("/executions", eh.ListExecutions)
				r.Get("/executions/{id}", eh.GetExecution)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeExecutionManage))
				r.Post("/executions", eh.CreateExecution)
				r.Post("/executions/{id}/fill", eh.FillExecution)
				r.Post("/executions/{id}/cancel", eh.CancelExecution)
			})
		}

		// ── Trade confirmations ─────────────────────────────────────────
		if m.confirmationHandler != nil {
			ch := m.confirmationHandler
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeConfirmationView))
				r.Get("/trade-confirmations", ch.ListConfirmations)
				r.Get("/trade-confirmations/{id}", ch.GetConfirmation)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeConfirmationManage))
				r.Post("/trade-confirmations", ch.RecordConfirmation)
				r.Post("/trade-confirmations/{id}/resolve", ch.ResolveConfirmation)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeConfirmationImport))
				r.Post("/trade-confirmations/batch", ch.ImportBatch)
			})
		}

		// ── Research reports (CRUD + simple submit/cancel) ──────────────
		if m.researchHandler != nil {
			rh := m.researchHandler
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchView))
				r.Get("/research-reports", rh.ListResearchReports)
				r.Get("/research-reports/{id}", rh.GetResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchCreate))
				r.Post("/research-reports", rh.CreateResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchUpdate))
				r.Put("/research-reports/{id}", rh.UpdateResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchDelete))
				r.Delete("/research-reports/{id}", rh.DeleteResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchSubmit))
				r.Post("/research-reports/{id}/submit", rh.SubmitResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchCancelSubmit))
				r.Post("/research-reports/{id}/cancel-submit", rh.CancelSubmitResearchReport)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeResearchInvalidate))
				r.Post("/research-reports/{id}/invalidate", rh.InvalidateResearchReport)
			})
		}
	})
}

// RegisterRoutesV2 mounts the investment module's Portfolio V2 HTTP routes
// (docs/api/portfolio-v2-api-ddd.md) onto the given authenticated router.
// Expected to be called inside an r.Route("/api/v2", ...) block that already
// has auth middleware applied. Additive only — V1 routes registered by
// RegisterRoutes are untouched and keep serving.
func (m *Module) RegisterRoutesV2(r chi.Router) {
	if m == nil || m.handler == nil || m.middlewarePerm == nil {
		return
	}
	pc := m.middlewarePerm
	h := m.handler

	r.Route("/portfolios", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodePortfolioView))
			r.Get("/{portfolioCode}", h.GetPortfolioByCode)
			r.Get("/{portfolioCode}/holdings", h.GetHoldingsByCode)
			r.Get("/{portfolioCode}/cash", h.GetCashByCode)
			r.Get("/{portfolioCode}/transactions", h.ListTransactionsByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeValuationView))
			r.Get("/{portfolioCode}/valuations", h.ListValuationsByCode)
			r.Get("/{portfolioCode}/valuations/latest", h.GetLatestValuationByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerSimulate))
			r.Post("/{portfolioCode}/transactions/simulate", h.SimulateTransactionByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerPost))
			r.Post("/{portfolioCode}/transactions", h.PostTransactionByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, invperm.CodeLedgerReverse))
			r.Post("/{portfolioCode}/transactions/{transactionId}/reverse", h.ReverseTransactionByCode)
		})

		// ── Decisions (Milestone 5) ────────────────────────────────────
		if dh := m.decisionHandler; dh != nil {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionView))
				r.Get("/{portfolioCode}/decisions", dh.ListDecisionsByCode)
				r.Get("/{portfolioCode}/decisions/{decisionId}", dh.GetDecisionByCode)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionManage))
				r.Post("/{portfolioCode}/decisions", dh.CreateDecisionByCode)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionSubmit))
				r.Post("/{portfolioCode}/decisions/{decisionId}/submit", dh.SubmitDecisionByCode)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeDecisionCancel))
				r.Post("/{portfolioCode}/decisions/{decisionId}/cancel", dh.CancelDecisionByCode)
			})
		}

		// ── Executions (Milestone 5) ────────────────────────────────────
		if eh := m.executionHandler; eh != nil {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeExecutionManage))
				r.Post("/{portfolioCode}/decisions/{decisionId}/executions", eh.CreateExecutionByCode)
				r.Post("/{portfolioCode}/executions/{executionId}/fill", eh.FillExecutionByCode)
				r.Post("/{portfolioCode}/executions/{executionId}/cancel", eh.CancelExecutionByCode)
			})
		}

		// ── Trade confirmations (Milestone 5) ───────────────────────────
		if ch := m.confirmationHandler; ch != nil {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequirePermission(pc, invperm.CodeConfirmationManage))
				r.Post("/{portfolioCode}/executions/{executionId}/confirmations", ch.RecordConfirmationByCode)
				r.Post("/{portfolioCode}/confirmations/{confirmationId}/resolve", ch.ResolveConfirmationByCode)
			})
		}

		// ── Portfolio Compliance V2 ──────────────────────────────────────
		// Permission codes are the compliance module's own IRG_* catalog
		// (see internal/compliance/module.go's RegisterRoutes for the V1
		// mirror) — passed as literal strings rather than an invperm
		// constant so investment does not import compliance's internal
		// permission package.
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, "IRG_VIEW_RULES"))
			r.Get("/{portfolioCode}/compliance/rules", h.ListRulesByCode)
			r.Get("/{portfolioCode}/compliance/breaches", h.ListBreachesByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, "IRG_EDIT_BINDING"))
			r.Post("/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings", h.BindRuleByCode)
			r.Delete("/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}", h.DeactivateRuleBindingByCode)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(pc, "WORKFLOW_EXECUTE"))
			r.Post("/{portfolioCode}/compliance/checks/pre-trade", h.RunPreTradeCheckByCode)
			r.Post("/{portfolioCode}/compliance/checks/post-trade", h.RunPostTradeCheckByCode)
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

// SetApprovalStatusProvider wires the read-side approval lookup into transport
// handlers (so the research-report response can render PENDING_LEVEL_N from
// the live approval stage, and the decision approval-items grid can show stage
// and approver names). Wired post-construction to avoid a circular dependency.
func (m *Module) SetApprovalStatusProvider(p contract.ApprovalStatusProvider) {
	if m == nil {
		return
	}
	if m.researchHandler != nil {
		m.researchHandler.SetApprovalStatusProvider(p)
	}
	if m.decisionHandler != nil {
		m.decisionHandler.SetApprovalStatusProvider(p)
	}
}

// SetApprovalBatchActor wires the approval batch-action actor into the decision
// batch approval command handler. Wired post-construction to avoid a circular
// construction dependency.
func (m *Module) SetApprovalBatchActor(a contract.ApprovalBatchActor) {
	if m == nil {
		return
	}
	if m.decisionBatchCmd != nil {
		m.decisionBatchCmd.SetBatchActor(a)
	}
	if m.decisionHandler != nil {
		m.decisionHandler.SetBatchApprovalHandler(m.decisionBatchCmd)
	}
}

// SetApprovalSubmitter connects the generic Approval Module so that submitting a
// research report (and investment decision) creates a real approval request.
// Wired post-construction in cmd/server/main.go to avoid a circular
// construction dependency.
func (m *Module) SetApprovalSubmitter(s contract.ApprovalSubmitter) {
	if m == nil {
		return
	}
	if m.researchCmd != nil {
		m.researchCmd.SetApprovalSubmitter(s)
	}
	if m.decisionCmd != nil {
		m.decisionCmd.SetApprovalSubmitter(s)
	}
}

// SetApprovalCanceller connects the approval canceller so that cancelling a
// research report or investment decision also terminates any in-flight approval
// request, keeping the two states in sync.
func (m *Module) SetApprovalCanceller(c contract.ApprovalCanceller) {
	if m == nil {
		return
	}
	if m.researchCmd != nil {
		m.researchCmd.SetApprovalCanceller(c)
	}
	if m.decisionCmd != nil {
		m.decisionCmd.SetApprovalCanceller(c)
	}
}

// SetPortfolioComplianceAdmin wires the Portfolio Compliance V2 contract
// (checks, rule catalog, binding administration, breach listing) into the
// V2 {portfolioCode} compliance handlers. Wired post-construction in
// cmd/server/main.go — mirrors SetApprovalSubmitter's circular-dependency
// avoidance.
func (m *Module) SetPortfolioComplianceAdmin(c contract.PortfolioComplianceContract) {
	if m == nil || m.handler == nil {
		return
	}
	m.handler.SetPortfolioComplianceAdmin(c)
}

// ResearchReportSubjectValidator returns the approval subject validator for
// the RESEARCH_REPORT subject type. The approval engine calls it inside the
// approve/reject transaction to verify the report is still SUBMITTED and has
// not been cancelled or invalidated concurrently.
func (m *Module) ResearchReportSubjectValidator() contract.ApprovalSubjectValidator {
	if m == nil || m.research == nil {
		return nil
	}
	return adapter.NewResearchReportSubjectValidator(m.research)
}

// DecisionSubjectValidator returns the approval subject validator for the
// INVESTMENT_DECISION subject type. The approval engine calls it inside the
// approve/reject transaction to verify the decision is still PENDING_APPROVAL.
func (m *Module) DecisionSubjectValidator() contract.ApprovalSubjectValidator {
	if m == nil || m.decisions == nil {
		return nil
	}
	return adapter.NewDecisionSubjectValidator(m.decisions)
}

// ComplianceReleaseSubjectValidator returns the approval subject validator for
// the COMPLIANCE_RELEASE subject type. Verifies the decision is still in
// PENDING_COMPLIANCE_RELEASE when the approval engine acts on it.
func (m *Module) ComplianceReleaseSubjectValidator() contract.ApprovalSubjectValidator {
	if m == nil || m.decisions == nil {
		return nil
	}
	return adapter.NewComplianceReleaseSubjectValidator(m.decisions)
}

// ApprovalSubjectCallback returns the callback the Approval Module invokes when
// a research-report approval reaches a final decision (drives report status).
func (m *Module) ApprovalSubjectCallback() contract.ApprovalSubjectCallback {
	if m == nil || m.researchCmd == nil {
		return nil
	}
	return adapter.NewResearchApprovalCallback(m.researchCmd)
}

// DecisionApprovalSubjectCallback returns the callback the Approval Module
// invokes when an INVESTMENT_DECISION approval reaches a final decision.
func (m *Module) DecisionApprovalSubjectCallback() contract.ApprovalSubjectCallback {
	if m == nil || m.decisionCmd == nil {
		return nil
	}
	return adapter.NewDecisionApprovalCallback(m.decisionCmd)
}

// ComplianceReleaseSubjectCallback returns the callback the Approval Module
// invokes when a COMPLIANCE_RELEASE approval reaches a final decision.
// Approved → decision is re-submitted to INVESTMENT_DECISION approval.
// Rejected → decision is cancelled.
func (m *Module) ComplianceReleaseSubjectCallback() contract.ApprovalSubjectCallback {
	if m == nil || m.decisionCmd == nil {
		return nil
	}
	return adapter.NewComplianceReleaseCallback(m.decisionCmd)
}

// PortfolioApprovalCallback returns the callback the Approval Module invokes
// when a PORTFOLIO_ONBOARDING approval request reaches a final decision
// (approved → ACTIVE, rejected → REJECTED).
func (m *Module) PortfolioApprovalCallback() contract.ApprovalSubjectCallback {
	if m == nil || m.portfolioCmd == nil {
		return nil
	}
	return adapter.NewPortfolioApprovalCallback(m.portfolioCmd)
}

// SetMarketQuoteProvider wires the cross-module quote provider into the
// intraday valuation service. Called from cmd/server/main.go after the
// market_data module has been constructed so we don't take an internal/
// import of market_data here. Idempotent and nil-safe.
//
// MarketQuoteIntradayConfig (StaleAfter, etc.) is sourced from platform/config
// in main.go and passed through.
func (m *Module) SetMarketQuoteProvider(quotes contract.MarketQuoteProvider, cfg service.IntradayConfig) {
	if m == nil || m.intradayHandler == nil {
		return
	}
	svc := service.NewIntradayValuationService(
		m.pool,
		quotes,
		m.funds,
		m.portfolios,
		m.cash,
		m.valuation,
		cfg,
	)
	m.intraday = svc
	m.intradayHandler.SetService(svc)
	// Allocation reuses the same resolved market values as the holdings
	// mark-to-market view so the two pages never disagree for the same
	// fund/business_date (see GetFundAllocationHandler.SetIntradayService).
	if m.fundAllocQuery != nil {
		m.fundAllocQuery.SetIntradayService(svc)
	}
}

// TradeConfirmationGate returns the adapter the workflow CloseTransactions
// handler consults to refuse closing while confirmations are pending.
func (m *Module) TradeConfirmationGate() contract.TradeConfirmationGate {
	if m == nil || m.executions == nil || m.confirmations == nil {
		return nil
	}
	return adapter.NewConfirmationGateAdapter(m.executions, m.confirmations)
}

// SubjectAccessor returns the contract.SubjectAccessor for the investment
// module's subject types (RESEARCH_REPORT, INVESTMENT_DECISION). Register it
// with the approval module via ApprovalModule.RegisterSubjectAccessPort in
// cmd/server/main.go for each subject type.
func (m *Module) SubjectAccessor(iamPort adapter.IAMPermissionPort) contract.SubjectAccessor {
	if m == nil || m.research == nil || m.decisions == nil {
		return nil
	}
	return adapter.NewInvestmentSubjectAccessor(m.research, m.decisions, iamPort)
}

// ContractCatalog returns an implementation of contract.ContractCatalog backed
// by the investment module's fund repository. Workflow and other modules inject
// this to resolve business-readable contractCodes to internal UUIDs without
// importing investment internals.
//
// This implementation is temporary: contract reference data may later move to
// a ReferenceData/ContractMaster module, at which point only the wiring in
// main.go changes — Workflow keeps the same ContractCatalog interface.
func (m *Module) ContractCatalog() contract.ContractCatalog {
	if m == nil || m.funds == nil {
		return nil
	}
	return adapter.NewContractCatalogAdapter(m.funds)
}

// PortfolioScopeResolver returns an adapter consumed by the watchlist module for
// portfolio data-permission checks and descriptor hydration.
func (m *Module) PortfolioScopeResolver() contract.PortfolioScopeResolver {
	if m == nil || m.portfolios == nil || m.funds == nil {
		return nil
	}
	return adapter.NewPortfolioScopeAdapter(m.portfolios, m.funds)
}

// ValuationSummaryProvider returns an adapter consumed by the integration
// module's dashboard valuation-summary endpoint to aggregate today's AUM and
// P&L across a scoped set of funds.
func (m *Module) ValuationSummaryProvider(
	reportingCurrency string,
	quotes contract.MarketQuoteProvider,
) contract.ValuationSummaryProvider {
	if m == nil || m.funds == nil || m.portfolios == nil || m.valuation == nil {
		return nil
	}
	return adapter.NewValuationSummaryAdapter(m.funds, m.portfolios, m.valuation, reportingCurrency, quotes)
}
