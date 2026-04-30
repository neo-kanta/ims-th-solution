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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
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

// Module owns the workflow state machine, its persistence, and its HTTP transport.
type Module struct {
	handler           *handler.WorkflowHandler
	dayRepo           *persistence.PostgresWorkflowDayRepository
	dayScheduler      *jobs.DayScheduler
	permissionChecker platformmw.PermissionChecker
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

	return &Module{
		handler:           h,
		dayRepo:           dayRepo,
		dayScheduler:      dayScheduler,
		permissionChecker: permChecker,
	}
}

// RegisterRoutes mounts workflow API routes onto the given authenticated router.
// Expected to be called inside an r.Route("/api/v1", ...) block that already
// has auth middleware applied.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}
	transport.RegisterRoutes(r, m.handler, m.permissionChecker)
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
func (m *Module) StartScheduler(ctx context.Context, interval time.Duration) {
	if m == nil || m.dayScheduler == nil {
		return
	}
	m.dayScheduler.Start(ctx, interval)
}

// ─────────────────────────────────────────────────────────────────────────────
// contract.WorkflowStateProvider implementation
//
// Exposed to other modules (investment, compliance) via pkg/contract.
// Both methods read the current workflow__day_states row without locking.
// ─────────────────────────────────────────────────────────────────────────────

// IsTradeAllowed reports whether investment trading is permitted for the given
// contract/date. Trading is allowed only while the day is in DAY_OPEN state;
// once the manager approves, transactions are locked.
func (m *Module) IsTradeAllowed(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error) {
	if m == nil || m.dayRepo == nil {
		return false, fmt.Errorf("workflow module not initialised")
	}
	day, err := m.dayRepo.GetByContractDate(ctx, contractID, businessDate)
	if err != nil {
		return false, fmt.Errorf("workflow: reading day state: %w", err)
	}
	if day == nil {
		// Not started → trading is not allowed.
		return false, nil
	}
	return day.CurrentState == vo.StateDayOpen, nil
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
	day, err := m.dayRepo.GetByContractDate(ctx, contractID, businessDate)
	if err != nil {
		return false, fmt.Errorf("workflow: reading day state: %w", err)
	}
	if day == nil {
		return false, nil
	}
	return day.IsTransactionLocked(), nil
}
