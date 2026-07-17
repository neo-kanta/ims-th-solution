package contract

import (
	"context"
	"encoding/json"
	"errors"
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

// ComplianceStatus reports whether configured controls produced the verdict.
// LIVE investment flows fail closed on NOT_CONFIGURED or UNAVAILABLE; other
// portfolio types preserve their existing policy at the owning boundary.
type ComplianceStatus string

const (
	ComplianceStatusEvaluated     ComplianceStatus = "COMPLIANCE_EVALUATED"
	ComplianceStatusNotConfigured ComplianceStatus = "COMPLIANCE_NOT_CONFIGURED"
	ComplianceStatusUnavailable   ComplianceStatus = "COMPLIANCE_UNAVAILABLE"
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
	Status         ComplianceStatus
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

// ErrInvalidProposedOrder signals that a ProposedOrderCheck failed input
// validation (missing/invalid field) rather than an infrastructure failure.
// The compliance module wraps its internal validation error into this
// sentinel before returning across the module boundary, so callers can
// classify failures with errors.Is without importing internal/compliance —
// same pattern as the Portfolio Compliance V2 binding sentinels below.
var ErrInvalidProposedOrder = errors.New("compliance: invalid proposed order request")

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

// ─────────────────────────────────────────────────────────────────────────────
// Portfolio Compliance V2 — portfolio-scoped compliance checks + rule
// binding administration, keyed by portfolio_id alone (fund_id optional).
// Implemented by the compliance module; consumed by investment's Portfolio V2
// ({portfolioCode}) HTTP handlers so investment never imports compliance
// internals and compliance never resolves portfolio codes itself.
// ─────────────────────────────────────────────────────────────────────────────

// PortfolioPostTradeRequest runs a post-trade compliance scan scoped to a
// single portfolio. FundID is optional — uuid.Nil means the portfolio has no
// fund_id and the scan runs at GLOBAL + PORTFOLIO scope only (no CONTRACT
// scope bindings apply).
type PortfolioPostTradeRequest struct {
	PortfolioID  uuid.UUID
	FundID       uuid.UUID // optional; uuid.Nil if the portfolio has no fund
	BusinessDate time.Time
	Actor        string
}

// PortfolioRuleBindingView describes an existing binding of a rule instance
// to a portfolio.
type PortfolioRuleBindingView struct {
	BindingID     uuid.UUID  `json:"binding_id"`
	Severity      string     `json:"severity"`
	Priority      int        `json:"priority"`
	IsActive      bool       `json:"is_active"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
}

// PortfolioRuleCatalogEntry is one rule instance in the catalog, annotated
// with its binding to a specific portfolio when one exists (Binding is nil
// when the rule instance is not bound to that portfolio).
type PortfolioRuleCatalogEntry struct {
	RuleInstanceID uuid.UUID `json:"rule_instance_id"`
	RuleTypeID     string    `json:"rule_type_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	IsActive       bool      `json:"is_active"`
	// Parameters is the rule instance's current configured parameter set
	// (e.g. {"asset_class":"EQUITY","max_percent_nav":60}), so the portfolio
	// settings UI can render thresholds without a second round-trip.
	Parameters json.RawMessage           `json:"parameters,omitempty" swaggertype:"object"`
	Binding    *PortfolioRuleBindingView `json:"binding,omitempty"`
}

// PortfolioRuleBindingRequest binds a rule instance to a portfolio
// (scope_type=PORTFOLIO, scope_id=PortfolioID).
type PortfolioRuleBindingRequest struct {
	PortfolioID    uuid.UUID
	RuleInstanceID uuid.UUID
	Severity       string
	Priority       int // 0 = use the compliance module's default
	EffectiveFrom  time.Time
	EffectiveTo    *time.Time
	ActorID        uuid.UUID
}

// PortfolioBreachFilter filters ListPortfolioBreaches.
type PortfolioBreachFilter struct {
	Status     *string
	RuleTypeID string
	DateFrom   *time.Time
	DateTo     *time.Time
	Offset     int
	Limit      int
}

// PortfolioBreachView is a compact breach summary for the portfolio-code API.
type PortfolioBreachView struct {
	BreachID     uuid.UUID `json:"breach_id"`
	RuleTypeID   string    `json:"rule_type_id"`
	Severity     string    `json:"severity"`
	Verdict      string    `json:"verdict"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	BusinessDate time.Time `json:"business_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// Portfolio Compliance V2 error sentinels. The compliance module wraps its
// internal domain errors into one of these before returning across the
// module boundary (see internal/compliance/transport/portfolio_contract_adapter.go),
// so investment's V2 handlers can classify failures with errors.Is without
// importing internal/compliance.
var (
	ErrPortfolioRuleNotFound     = errors.New("compliance: rule instance not found")
	ErrPortfolioRuleInactive     = errors.New("compliance: rule instance is not active")
	ErrPortfolioBindingDuplicate = errors.New("compliance: an active binding already exists for this rule on this portfolio")
	ErrPortfolioBindingInvalid   = errors.New("compliance: invalid rule binding request")
	ErrPortfolioBindingNotFound  = errors.New("compliance: rule binding not found")
)

// PortfolioComplianceContract is the full Portfolio Compliance V2 surface:
// pre/post-trade checks plus rule catalog and binding administration, all
// keyed by portfolio_id with fund_id optional. Implemented by the compliance
// module's portfolio contract adapter (transport/portfolio_contract_adapter.go).
type PortfolioComplianceContract interface {
	ComplianceChecker
	ComplianceSimulator

	RunPortfolioPostTradeCheck(ctx context.Context, req PortfolioPostTradeRequest) (*ProposedOrderResult, error)

	ListPortfolioRules(ctx context.Context, portfolioID uuid.UUID) ([]PortfolioRuleCatalogEntry, error)
	BindPortfolioRule(ctx context.Context, req PortfolioRuleBindingRequest) (*PortfolioRuleBindingView, error)
	DeactivatePortfolioRuleBinding(ctx context.Context, portfolioID, bindingID uuid.UUID) error

	ListPortfolioBreaches(ctx context.Context, portfolioID uuid.UUID, filter PortfolioBreachFilter) ([]PortfolioBreachView, error)
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
