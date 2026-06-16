// Package errcode is the single source of truth for the closed set of
// API-visible error codes emitted by the IMS backend.
//
// Every typed domain error must map to one — and only one — of the codes
// declared here. The platform/httputil envelope writer surfaces the code
// to the frontend; tests under cmd/server and each module's transport
// package assert the mapping is total.
//
// Naming rules:
//   - UPPER_SNAKE_CASE.
//   - No prefixes. The code stands alone; do not introduce module-scoped
//     codes (use a richer code rather than IAM_ / WORKFLOW_ prefixes when
//     the domain genuinely needs disambiguation).
//   - Stable forever. Renaming a code is a frontend-breaking change.
package errcode

// API-visible error codes.
//
// Phase 6 declares the full Phase 1+ surface so module authors can pick a
// code from this list without inventing new strings. Codes that have no
// matching typed error today are kept for the Phase that will introduce
// them; tests guard against accidental removal.
const (
	// ─── Validation / shape ──────────────────────────────────────────────
	CodeInvalidRequest = "INVALID_REQUEST"

	// ─── Resource lookup ─────────────────────────────────────────────────
	CodeInstrumentNotFound  = "INSTRUMENT_NOT_FOUND"
	CodeFundNotFound        = "FUND_NOT_FOUND"
	CodePortfolioNotFound   = "PORTFOLIO_NOT_FOUND"
	CodeTransactionNotFound = "TRANSACTION_NOT_FOUND"
	CodeDecisionNotFound    = "DECISION_NOT_FOUND"
	CodeContractNotFound    = "CONTRACT_NOT_FOUND"

	// ─── Concurrency / state ─────────────────────────────────────────────
	CodeVersionMismatch  = "VERSION_MISMATCH"
	CodeCodeAlreadyTaken = "CODE_ALREADY_EXISTS"
	CodeDecisionNotDraft = "DECISION_NOT_DRAFT"

	// ─── Decision lifecycle ──────────────────────────────────────────────
	CodeDecisionLifecycle   = "DECISION_LIFECYCLE_INVALID"
	CodeDecisionReferenceBad = "DECISION_REPORT_REFERENCE_INVALID"

	// ─── Research report lifecycle ───────────────────────────────────────
	CodeResearchReportInvalidateBlocked = "RESEARCH_REPORT_INVALIDATE_BLOCKED"

	// ─── Execution / confirmation ────────────────────────────────────────
	CodeExecutionNotFound      = "EXECUTION_NOT_FOUND"
	CodeExecutionLifecycle     = "EXECUTION_LIFECYCLE_INVALID"
	CodeConfirmationNotFound   = "TRADE_CONFIRMATION_NOT_FOUND"
	CodeConfirmationLifecycle  = "TRADE_CONFIRMATION_LIFECYCLE_INVALID"
	CodeConfirmationMismatch   = "TRADE_CONFIRMATION_MISMATCH"
	CodeClosePendingConfirm    = "CLOSE_PENDING_CONFIRMATIONS"

	// ─── Position / posting violations ───────────────────────────────────
	CodeOversell                = "OVERSELL"
	CodeWorkflowTradeNotAllowed = "WORKFLOW_TRADE_NOT_ALLOWED"
	CodeWorkflowLocked          = "WORKFLOW_LOCKED"
	CodeComplianceRejected      = "COMPLIANCE_REJECTED"
	CodePostTradeBlocked        = "POST_TRADE_BLOCKED"
	CodeInstrumentNotMapped     = "INSTRUMENT_NOT_MAPPED"
	CodePriceCurrencyMismatch   = "PRICE_CURRENCY_MISMATCH"
	CodeUnitisedNotSupported    = "UNITISED_NOT_SUPPORTED"

	// ─── Reversal / lifecycle ────────────────────────────────────────────
	CodeAlreadyReversed       = "ALREADY_REVERSED"
	CodeCannotReverseReversal = "CANNOT_REVERSE_REVERSAL"

	// ─── Lifecycle blocking conditions ───────────────────────────────────
	CodeFundHasActivePortfolios  = "FUND_HAS_ACTIVE_PORTFOLIOS"
	CodePortfolioHasOpenActivity = "PORTFOLIO_HAS_OPEN_ACTIVITY"
	CodeIncompleteFundValuation  = "INCOMPLETE_FUND_VALUATION"

	// ─── Authorization ───────────────────────────────────────────────────
	CodeForbidden = "FORBIDDEN"

	// ─── Catch-all ───────────────────────────────────────────────────────
	CodeInternal = "INTERNAL_ERROR"
)

// All returns every declared code, sorted, for use in tests that assert
// the registry is complete.
func All() []string {
	return []string{
		CodeAlreadyReversed,
		CodeCannotReverseReversal,
		CodeClosePendingConfirm,
		CodeCodeAlreadyTaken,
		CodeComplianceRejected,
		CodeConfirmationLifecycle,
		CodeConfirmationMismatch,
		CodeConfirmationNotFound,
		CodeContractNotFound,
		CodeDecisionLifecycle,
		CodeDecisionNotDraft,
		CodeDecisionNotFound,
		CodeDecisionReferenceBad,
		CodeExecutionLifecycle,
		CodeExecutionNotFound,
		CodeForbidden,
		CodeFundHasActivePortfolios,
		CodeFundNotFound,
		CodeIncompleteFundValuation,
		CodeInstrumentNotFound,
		CodeInstrumentNotMapped,
		CodeInternal,
		CodeInvalidRequest,
		CodeOversell,
		CodePortfolioHasOpenActivity,
		CodePortfolioNotFound,
		CodePostTradeBlocked,
		CodePriceCurrencyMismatch,
		CodeResearchReportInvalidateBlocked,
		CodeTransactionNotFound,
		CodeUnitisedNotSupported,
		CodeVersionMismatch,
		CodeWorkflowLocked,
		CodeWorkflowTradeNotAllowed,
	}
}
