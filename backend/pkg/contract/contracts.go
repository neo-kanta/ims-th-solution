package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// ContractCatalog resolves business-readable contract codes to internal UUIDs.
//
// Current implementation lives temporarily in the Investment module's
// infrastructure/adapter layer (ContractCatalogAdapter). Future implementations
// may move to a dedicated ReferenceData or ContractMaster module without any
// change to consumers. Workflow and other modules must depend only on this
// interface, never on Investment internals.
type ContractCatalog interface {
	// ResolveContractCode maps a human-readable contractCode (e.g. "ABC") to its
	// internal uuid.UUID. Returns errcode.CodeContractNotFound if no alive fund
	// has that code. Never returns uuid.Nil with a nil error.
	ResolveContractCode(ctx context.Context, contractCode string) (uuid.UUID, error)
}

// WorkflowStateProvider defines the cross-module interface for querying workflow state.
// Implemented by the workflow module; consumed by investment, compliance modules.
//
// Deviation from original: contractID is uuid.UUID (was string). No callers existed yet.
// IsTransactionLocked added per approved design (manager approval locks transactions).
type WorkflowStateProvider interface {
	IsTradeAllowed(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error)
	IsTransactionLocked(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error)
}

// WorkflowStateInvestmentDayStarted is the canonical Phase 2 state name for an open trading day.
const WorkflowStateInvestmentDayStarted = "INVESTMENT_DAY_STARTED"

// WorkflowStateDayOpen is the Phase 1 backward-compat alias for WorkflowStateInvestmentDayStarted.
// Kept until the Phase 2 cleanup migration removes DAY_OPEN from the database.
const WorkflowStateDayOpen = "DAY_OPEN"

// IsWorkflowOpenForTrading reports whether a cross-module workflow state string
// represents an open trading day, accepting both canonical and compat names.
func IsWorkflowOpenForTrading(state string) bool {
	return state == WorkflowStateInvestmentDayStarted || state == WorkflowStateDayOpen
}

// WorkflowDayLock is the minimal workflow state returned after locking a day row.
type WorkflowDayLock struct {
	Exists       bool
	CurrentState string
}

// WorkflowTradeDayLocker locks a workflow day row inside the caller's active
// database transaction. Posting code uses this to prevent workflow transitions
// from moving a day out of DAY_OPEN between the final guard and ledger insert.
type WorkflowTradeDayLocker interface {
	LockTradeDayForPost(ctx context.Context, tx pgx.Tx, contractID uuid.UUID, businessDate time.Time) (*WorkflowDayLock, error)
}

// PermissionChecker defines the cross-module interface for permission verification.
type PermissionChecker interface {
	HasFunctionPermission(userID uuid.UUID, permissionCode string) (bool, error)
	HasDataPermission(userID uuid.UUID, contractID string) (bool, error)
	GetAccessibleContracts(userID uuid.UUID) ([]string, error)
}

// LeaveChecker defines the cross-module interface for querying leave status.
// IAM does NOT own leave state — the leave_delegation module implements this.
type LeaveChecker interface {
	IsOnLeave(userID uuid.UUID, date time.Time) (bool, error)
}

// AuditLogger defines the cross-module interface for recording audit events.
//
// LogAction is the legacy fire-and-forget path: implementations log failures
// internally and return nil. Use it for low-risk CRUD events where audit
// loss is preferable to operational disruption.
//
// LogActionStrict is the synchronous, error-returning path required for
// financial-grade actions where audit loss is unacceptable. The caller MUST
// check the returned error and decide whether to retry, alert, or fail the
// operation.
type AuditLogger interface {
	LogAction(entry AuditEntry) error
	LogActionStrict(ctx context.Context, entry AuditEntry) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Compliance (IRG) cross-module contracts
// ─────────────────────────────────────────────────────────────────────────────

// ComplianceVerdict mirrors the compliance module's pipeline verdict without
// forcing consumers to import compliance internals. Callers compare against the
// exported constants below.
type ComplianceVerdict string

const (
	ComplianceVerdictPass  ComplianceVerdict = "PASS"
	ComplianceVerdictWarn  ComplianceVerdict = "WARN"
	ComplianceVerdictBlock ComplianceVerdict = "BLOCK"
)

// ComplianceOrderSide is the direction of a proposed trade.
type ComplianceOrderSide string

const (
	ComplianceOrderSideBuy  ComplianceOrderSide = "BUY"
	ComplianceOrderSideSell ComplianceOrderSide = "SELL"
)

// ProposedOrderCheck is the request submitted to the IRG pre-trade pipeline.
// All fields are required unless stated otherwise.
type ProposedOrderCheck struct {
	// CheckGroupID is an optional idempotency key. When uuid.Nil the compliance
	// module generates one.
	CheckGroupID uuid.UUID
	PortfolioID  uuid.UUID
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        string // originating user ID or "system"

	OrderID  uuid.UUID
	Ticker   string
	Side     ComplianceOrderSide
	Quantity decimal.Decimal
	Price    decimal.Decimal
	Fees     decimal.Decimal
	Currency string
	Exchange string
}

// ProposedOrderBreach is a compact breach summary returned to non-compliance callers.
// It carries the minimum detail required to surface a rejection to the end user
// and to pivot back into the full compliance module for drill-down.
type ProposedOrderBreach struct {
	BreachID    uuid.UUID         `json:"breach_id"`
	RuleTypeID  string            `json:"rule_type_id"`
	Verdict     ComplianceVerdict `json:"verdict"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	Overridable bool              `json:"overridable"`
}

// ProposedOrderResult is the pre-trade verdict returned to the OMS layer.
// CheckGroupID is always set — even on BLOCK — so the caller can cite it for
// override workflows and audit trails.
type ProposedOrderResult struct {
	CheckGroupID   uuid.UUID
	Verdict        ComplianceVerdict
	RulesEvaluated int
	Breaches       []ProposedOrderBreach
}

// ComplianceChecker is the cross-module contract for pre-trade IRG evaluation.
// Implemented by the compliance module; consumed by investment (OMS) command handlers.
// The implementation MUST be synchronous and idempotent with respect to CheckGroupID.
type ComplianceChecker interface {
	CheckProposedOrder(ctx context.Context, req ProposedOrderCheck) (*ProposedOrderResult, error)
}

// ComplianceSimulator is an optional extension for callers that need the same
// pre-trade verdict without persisting compliance check records or breaches.
type ComplianceSimulator interface {
	SimulateProposedOrder(ctx context.Context, req ProposedOrderCheck) (*ProposedOrderResult, error)
}

// PostTradeBreach captures a single persistent breach discovered by the
// post-trade sweep. The workflow module uses these entries to refuse a
// TRANSACTION_CLOSED transition when at least one BLOCK-level breach exists.
type PostTradeBreach struct {
	BreachID   uuid.UUID         `json:"breach_id"`
	RuleTypeID string            `json:"rule_type_id"`
	Verdict    ComplianceVerdict `json:"verdict"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
}

// PostTradeVerificationResult is the outcome of a portfolio-level post-trade scan.
// HasBlockingBreach is a convenience boolean the workflow gate consults directly.
type PostTradeVerificationResult struct {
	CheckGroupID      uuid.UUID
	Verdict           ComplianceVerdict
	HasBlockingBreach bool
	Breaches          []PostTradeBreach
}

// PostTradeVerifier is the cross-module contract used by the workflow module to
// ensure no outstanding BLOCK-level breaches exist before transitioning a
// business day to TRANSACTION_CLOSED. Implemented by the compliance module.
type PostTradeVerifier interface {
	RunPostTradeVerification(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*PostTradeVerificationResult, error)
}

// AuditEntry represents a single audit log record.
type AuditEntry struct {
	ActorID      string      `json:"actor_id"`
	Action       string      `json:"action"`
	Module       string      `json:"module"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Details      interface{} `json:"details,omitempty"`
	BusinessDate time.Time   `json:"business_date"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Trade confirmation gate (workflow ⇄ investment)
// ─────────────────────────────────────────────────────────────────────────────

// ConfirmationGateResult is the outcome of the closing-time gate. The workflow
// CloseTransactions handler refuses the transition while any execution is
// still PENDING confirmation or has a MISMATCHED row without a reason.
type ConfirmationGateResult struct {
	PendingCount    int
	UnresolvedCount int
	Reasons         []string
}

// HasBlocker reports whether the gate result blocks transaction closing.
func (g *ConfirmationGateResult) HasBlocker() bool {
	if g == nil {
		return false
	}
	return g.PendingCount > 0 || g.UnresolvedCount > 0
}

// TradeConfirmationGate is implemented by the investment module and consumed
// by the workflow CloseTransactions handler. Returning a non-nil error makes
// the close handler fail closed.
type TradeConfirmationGate interface {
	EvaluateClose(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*ConfirmationGateResult, error)
}
