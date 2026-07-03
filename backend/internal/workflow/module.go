// Package workflow wires the Workflow Management module.
//
// Workflow owns the per-contract, per-business-date state machine with four
// checkpoints (DAY_OPEN → MANAGER_APPROVED → TRANSACTION_CLOSED → ACCOUNTING_CLOSED)
// plus the implicit NOT_STARTED baseline. Business-logic rules (business day,
// previous-day dependency, zero-transaction attestation) live in
// domain/policy. Transaction boundaries are managed by the application
// command handlers.
package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/policy"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/jobs"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Compile-time assertion that *Module satisfies the cross-module contract.
var _ contract.WorkflowStateProvider = (*Module)(nil)
var _ contract.WorkflowTradeDayLocker = (*Module)(nil)

// Module owns the workflow state machine, its persistence, and its HTTP transport.
type Module struct {
	handler           *handler.WorkflowHandler
	dailyHandler      *handler.DailyHandler
	dayRepo           *persistence.PostgresWorkflowDayRepository
	logRepo           *persistence.PostgresTransitionLogRepository
	dayScheduler      *jobs.DayScheduler
	stuckDayWatcher   *jobs.StuckDayWatcher
	permissionChecker platformmw.PermissionChecker

	closeTransactions *command.CloseTransactionsHandler
	pool              *pgxpool.Pool

	// contractCatalog resolves business-readable contract codes to internal UUIDs.
	// Injected post-construction via SetContractCatalog so neither module must
	// import the other's internals. Nil until wired; handlers must nil-check.
	contractCatalog contract.ContractCatalog

	// settingRepo manages per-operationType approver configuration.
	settingRepo *persistence.PostgresApprovalSettingRepository
}

// NewModule constructs the workflow module with Postgres-backed repositories
// and Nop port adapters suitable for the PoC.
//
// Swap the Nop adapters for real implementations in later batches:
//   - NopHolidayCalendarAdapter    → reference_data holiday calendar
//   - NopInvestmentQueryAdapter    → investment module transaction summary
//
// postTradeVerifier is consulted by CloseTransactions to enforce the IRG
// post-trade gate. Nil is accepted for early bring-up (dev/tests) but is
// REQUIRED in production wiring.
func NewModule(
	pool *pgxpool.Pool,
	permChecker platformmw.PermissionChecker,
	postTradeVerifier contract.PostTradeVerifier,
) *Module {
	// ── Repositories ─────────────────────────────────────────────────────
	dayRepo := persistence.NewPostgresWorkflowDayRepository(pool)
	logRepo := persistence.NewPostgresTransitionLogRepository(pool)
	approvalRepo := persistence.NewPostgresApprovalRecordRepository(pool)
	settingRepo := persistence.NewPostgresApprovalSettingRepository(pool)
	schedulerRepo := persistence.NewPostgresSchedulerRepository(pool)
	contractSource := persistence.NewPostgresWorkflowSchedulerContractSource(pool)

	// ── Port adapters (NOP stubs for Batch 1) ────────────────────────────
	var calendarPort ports.HolidayCalendarPort = &adapter.NopHolidayCalendarAdapter{}
	var investmentPort ports.InvestmentQueryPort = &adapter.NopInvestmentQueryAdapter{}

	// ── Domain policy (pure, stateless) ──────────────────────────────────
	pol := policy.NewTransitionPolicy()

	// ── Application — queries ────────────────────────────────────────────
	getCurrentState := query.NewGetCurrentStateHandler(dayRepo, calendarPort)
	getHistory := query.NewGetHistoryHandler(logRepo)
	getDailyWorkflow := query.NewGetDailyWorkflowHandler(dayRepo, logRepo, settingRepo, calendarPort)

	// ── Application — commands ───────────────────────────────────────────
	openDay := command.NewOpenDayHandler(pool, dayRepo, logRepo, calendarPort, pol)
	approve := command.NewManagerApprovalHandler(pool, dayRepo, logRepo, approvalRepo, investmentPort, pol)
	cancelDayStart := command.NewCancelDayStartHandler(pool, dayRepo, logRepo, investmentPort, pol)
	cancelApproval := command.NewCancelApprovalHandler(pool, dayRepo, logRepo, approvalRepo, pol)
	closeTransactions := command.NewCloseTransactionsHandler(pool, dayRepo, logRepo, pol, postTradeVerifier)
	cancelTransactionClose := command.NewCancelTransactionCloseHandler(pool, dayRepo, logRepo, pol)
	closeAccounting := command.NewCloseAccountingHandler(pool, dayRepo, logRepo, pol)
	rollbackAccountingClose := command.NewRollbackAccountingCloseHandler(pool, dayRepo, logRepo, pol)
	dayScheduler := jobs.NewDayScheduler(
		schedulerRepo,
		dayRepo,
		contractSource,
		calendarPort,
		openDay,
		clock.RealClock{},
		jobs.DaySchedulerConfig{},
	)
	stuckDayWatcher := jobs.NewStuckDayWatcher(dayRepo, ports.NopOperatorNotifier{}, clock.RealClock{})

	// ── Transport ────────────────────────────────────────────────────────
	h := handler.NewWorkflowHandler(
		getCurrentState, getHistory,
		dayScheduler,
		openDay, approve,
		cancelDayStart, cancelApproval,
		closeTransactions, cancelTransactionClose,
		closeAccounting, rollbackAccountingClose,
		permChecker,
	)

	// Daily handler wired with nil contractCatalog; SetContractCatalog() wires it
	// post-construction from main.go after the investment module is available.
	daily := handler.NewDailyHandler(
		getDailyWorkflow,
		settingRepo,
		logRepo,
		nil, // contractCatalog — injected post-construction via SetContractCatalog
		pool,
		openDay, approve,
		cancelDayStart, cancelApproval,
		closeTransactions, cancelTransactionClose,
		closeAccounting, rollbackAccountingClose,
	)

	return &Module{
		handler:           h,
		dailyHandler:      daily,
		dayRepo:           dayRepo,
		logRepo:           logRepo,
		dayScheduler:      dayScheduler,
		stuckDayWatcher:   stuckDayWatcher,
		permissionChecker: permChecker,
		closeTransactions: closeTransactions,
		pool:              pool,
		settingRepo:       settingRepo,
	}
}

// SetOperatorNotifier replaces the stuck-day watcher's notifier. Wired
// post-construction from main.go so the workflow module does not depend on
// the notification module's types. Passing nil restores the no-op default.
func (m *Module) SetOperatorNotifier(n ports.OperatorNotifier) {
	if m == nil {
		return
	}
	m.stuckDayWatcher = jobs.NewStuckDayWatcher(m.dayRepo, n, clock.RealClock{})
}

// RunStuckDayScanOnce runs one pass of the stuck-day watcher. Exposed for
// tests and ad-hoc triggers; the scheduler ticks call it automatically when
// StartScheduler is in use.
func (m *Module) RunStuckDayScanOnce(ctx context.Context) error {
	if m == nil || m.stuckDayWatcher == nil {
		return nil
	}
	return m.stuckDayWatcher.RunOnce(ctx)
}

// SetConfirmationGate installs the trade-confirmation gate on the
// CloseTransactions handler post-construction. Called from main.go after the
// investment module is wired so we can keep both modules' constructors free of
// each other.
func (m *Module) SetConfirmationGate(g contract.TradeConfirmationGate) {
	if m == nil || m.closeTransactions == nil {
		return
	}
	m.closeTransactions.SetConfirmationGate(g)
}

// SetContractCatalog injects the ContractCatalog resolver post-construction.
// Called from main.go after the investment module is constructed.
// Workflow uses this to resolve contractCode params to internal UUIDs without
// importing investment internals. Also wired into the daily handler.
func (m *Module) SetContractCatalog(c contract.ContractCatalog) {
	if m == nil {
		return
	}
	m.contractCatalog = c
}

// ResolveContractCode is a convenience helper for workflow application handlers.
// Returns an error if no ContractCatalog has been wired or the code is unknown.
func (m *Module) ResolveContractCode(ctx context.Context, contractCode string) (uuid.UUID, error) {
	if m == nil || m.contractCatalog == nil {
		return uuid.Nil, fmt.Errorf("workflow: contract catalog not wired")
	}
	return m.contractCatalog.ResolveContractCode(ctx, contractCode)
}

// RegisterRoutes mounts workflow API routes onto the given authenticated router.
// Expected to be called inside an r.Route("/api/v1", ...) block that already
// has auth middleware applied.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}
	transport.RegisterRoutes(r, m.handler, m.dailyHandler, m.permissionChecker)
}

// RunSchedulerOnce executes one workflow scheduler tick and returns the audit
// run headers that were created.
func (m *Module) RunSchedulerOnce(ctx context.Context) ([]*entity.SchedulerRun, error) {
	if m == nil || m.dayScheduler == nil {
		return nil, fmt.Errorf("workflow scheduler not initialised")
	}
	return m.dayScheduler.RunOnce(ctx)
}

// StartScheduler starts the hourly workflow scheduler loop until ctx is done.
// In addition to opening business days, the loop also fires the stuck-day
// watcher each tick so operators get an in-app notification when a day is
// left in DAY_OPEN or MANAGER_APPROVED past the configured cutoff.
func (m *Module) StartScheduler(ctx context.Context, interval time.Duration) {
	if m == nil || m.dayScheduler == nil {
		return
	}
	m.dayScheduler.Start(ctx, interval)
	m.startStuckDayWatcher(ctx, interval)
}

// startStuckDayWatcher runs a parallel ticker for the stuck-day scan. Kept
// off the OpenDay tick path so a scan failure cannot wedge the open-day
// flow, and vice-versa.
func (m *Module) startStuckDayWatcher(ctx context.Context, interval time.Duration) {
	if m == nil || m.stuckDayWatcher == nil {
		return
	}
	if interval <= 0 {
		interval = time.Hour
	}
	go func() {
		// First scan immediately so demos do not have to wait for the next
		// tick interval.
		if err := m.stuckDayWatcher.RunOnce(ctx); err != nil {
			fmt.Printf("stuck-day watcher first run failed: %v\n", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.stuckDayWatcher.RunOnce(ctx); err != nil {
					fmt.Printf("stuck-day watcher run failed: %v\n", err)
				}
			}
		}
	}()
}

// ─────────────────────────────────────────────────────────────────────────────
// contract workflow guard implementations
//
// Exposed to other modules (investment, compliance) via pkg/contract.
// Read-side methods do not lock. LockTradeDayForPost uses SELECT FOR UPDATE
// inside the caller's transaction for the final ledger-post guard.
// ─────────────────────────────────────────────────────────────────────────────

// IsTradeAllowed reports whether investment trading is permitted for the given
// contract/date. Trading is allowed while the day is open (INVESTMENT_DAY_STARTED
// or the Phase 1 compat alias DAY_OPEN); once the manager approves, transactions
// are locked.
func (m *Module) IsTradeAllowed(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error) {
	if m == nil || m.dayRepo == nil {
		return false, fmt.Errorf("workflow module not initialised")
	}
	day, err := m.dayRepo.GetByBusinessDate(ctx, businessDate)
	if err != nil {
		return false, fmt.Errorf("workflow: reading day state: %w", err)
	}
	if day == nil {
		// Not started → trading is not allowed.
		return false, nil
	}
	return day.CurrentState.IsOpenForTrading(), nil
}

// IsTransactionLocked reports whether transactions for the given contract/date
// are locked. Transactions lock at MANAGER_APPROVED and remain locked for every
// subsequent state (TRANSACTION_CLOSED, ACCOUNTING_CLOSED). A missing row is
// treated as "not locked" — the investment module should fall back to
// IsTradeAllowed for the positive check.
func (m *Module) IsTransactionLocked(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error) {
	if m == nil || m.dayRepo == nil {
		return false, fmt.Errorf("workflow module not initialised")
	}
	day, err := m.dayRepo.GetByBusinessDate(ctx, businessDate)
	if err != nil {
		return false, fmt.Errorf("workflow: reading day state: %w", err)
	}
	if day == nil {
		return false, nil
	}
	return day.IsTransactionLocked(), nil
}

// LockTradeDayForPost locks the workflow day row inside the caller's
// transaction and returns its current state. The investment posting flow then
// enforces DAY_OPEN while the row lock is held.
func (m *Module) LockTradeDayForPost(
	ctx context.Context,
	tx pgx.Tx,
	contractID uuid.UUID,
	businessDate time.Time,
) (*contract.WorkflowDayLock, error) {
	if m == nil || m.dayRepo == nil {
		return nil, fmt.Errorf("workflow module not initialised")
	}
	if tx == nil {
		return nil, fmt.Errorf("workflow: transaction is required")
	}
	day, err := m.dayRepo.GetForUpdateByBusinessDate(ctx, tx, businessDate)
	if err != nil {
		return nil, fmt.Errorf("workflow: locking day state: %w", err)
	}
	if day == nil {
		return &contract.WorkflowDayLock{Exists: false}, nil
	}
	return &contract.WorkflowDayLock{
		Exists:       true,
		CurrentState: string(day.CurrentState),
	}, nil
}
