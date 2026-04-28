package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WorkflowStateProvider defines the cross-module interface for querying workflow state.
// Implemented by the workflow module; consumed by investment, compliance modules.
//
// Deviation from original: contractID is uuid.UUID (was string). No callers existed yet.
// IsTransactionLocked added per approved design (manager approval locks transactions).
type WorkflowStateProvider interface {
	IsTradeAllowed(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error)
	IsTransactionLocked(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (bool, error)
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
type AuditLogger interface {
	LogAction(entry AuditEntry) error
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
