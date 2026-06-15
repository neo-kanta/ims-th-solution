// Package policy contains pure business-rule functions for the workflow module.
//
// Design constraint: no I/O. Every method receives all data it needs as
// parameters. This keeps the policy layer fast and fully unit-testable without
// mocks or database connections.
package policy

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// TransitionPolicy enforces all guard conditions for workflow state transitions.
type TransitionPolicy struct{}

// NewTransitionPolicy creates a TransitionPolicy.
func NewTransitionPolicy() *TransitionPolicy { return &TransitionPolicy{} }

// ─────────────────────────────────────────────────────────────────────────────
// OPEN_DAY
// ─────────────────────────────────────────────────────────────────────────────

// OpenDayInput carries everything CanOpenDay needs to evaluate.
// All values are resolved by the command handler before calling the policy.
type OpenDayInput struct {
	// IsBusinessDay is false when businessDate falls on a weekend or holiday.
	IsBusinessDay bool

	// HolidayName is set when IsBusinessDay = false (for the error body).
	HolidayName string

	// CurrentDay is the existing workflow__day_states row, or nil when NOT_STARTED.
	CurrentDay *entity.WorkflowDay

	// PrevDay is the workflow__day_states row for the previous business date, or nil.
	PrevDay *entity.WorkflowDay
}

// CanOpenDay evaluates all guard conditions for the OPEN_DAY action.
// Returns a typed domain error on any violation; nil means the transition is allowed.
func (p *TransitionPolicy) CanOpenDay(in OpenDayInput) error {
	// Guard 1 — business day
	if !in.IsBusinessDay {
		msg := "not a Thai business day"
		if in.HolidayName != "" {
			msg = fmt.Sprintf("not a business day: %s", in.HolidayName)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_NOT_BUSINESS_DAY",
			AttemptedAction: string(vo.ActionOpenDay),
			CurrentState:    string(vo.StateNotStarted),
			Reason:          msg,
			Details:         map[string]any{"holidayName": in.HolidayName},
		}
	}

	// Guard 2 — current state must be NOT_STARTED (nil row)
	if in.CurrentDay != nil {
		return &domain.ErrWorkflowDayExists{
			ContractID:   in.CurrentDay.ContractID.String(),
			BusinessDate: in.CurrentDay.BusinessDate.Format("2006-01-02"),
			CurrentState: string(in.CurrentDay.CurrentState),
		}
	}

	// Guard 3 — previous business day must not be in a blocking state.
	// Both canonical (INVESTMENT_DAY_STARTED, MANAGER_APPROVED_END_OF_DAY) and
	// Phase 1 compat (DAY_OPEN, MANAGER_APPROVED) names are covered by helpers.
	if in.PrevDay != nil {
		switch {
		case in.PrevDay.CurrentState.IsOpenForTrading():
			return &domain.ErrInvalidTransition{
				Code:            "WORKFLOW_PREVIOUS_DAY_NOT_APPROVED",
				AttemptedAction: string(vo.ActionOpenDay),
				CurrentState:    string(vo.StateNotStarted),
				Reason: fmt.Sprintf(
					"previous business day %s is still open for trading; "+
						"manager approval must be completed before opening today",
					in.PrevDay.BusinessDate.Format("2006-01-02"),
				),
				Details: map[string]any{
					"previousDate":  in.PrevDay.BusinessDate.Format("2006-01-02"),
					"previousState": string(in.PrevDay.CurrentState),
				},
			}
		case in.PrevDay.CurrentState.IsManagerApproved():
			return &domain.ErrInvalidTransition{
				Code:            "WORKFLOW_PREVIOUS_DAY_NOT_TRANSACTION_CLOSED",
				AttemptedAction: string(vo.ActionOpenDay),
				CurrentState:    string(vo.StateNotStarted),
				Reason: fmt.Sprintf(
					"previous business day %s is manager-approved; "+
						"transaction closing must be confirmed before opening today",
					in.PrevDay.BusinessDate.Format("2006-01-02"),
				),
				Details: map[string]any{
					"previousDate":  in.PrevDay.BusinessDate.Format("2006-01-02"),
					"previousState": string(in.PrevDay.CurrentState),
				},
			}
		}
		// TRANSACTION_CLOSED, ACCOUNTING_CLOSED, and NOT_STARTED (gap day) are all fine.
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// APPROVE (Manager Approval)
// ─────────────────────────────────────────────────────────────────────────────

// ApprovalInput carries everything CanApprove needs to evaluate.
type ApprovalInput struct {
	// CurrentDay is the existing workflow day record (must not be nil).
	CurrentDay *entity.WorkflowDay

	// ActorID is the user attempting to approve. Used to enforce
	// maker-checker (the approver must differ from the day-opener).
	ActorID uuid.UUID

	// TransactionCount is the number of investment transactions for this date.
	// Resolved from the InvestmentQueryPort before entering the transaction.
	TransactionCount int

	// HasPendingUnreviewed is true when at least one transaction has not yet been
	// reviewed by the fund manager.
	HasPendingUnreviewed bool

	// ZeroTransactionAttestation must be true when TransactionCount == 0.
	ZeroTransactionAttestation bool

	// AttestationReason is required when ZeroTransactionAttestation is true.
	// Must be at least 30 characters (enforced here).
	AttestationReason string
}

const minAttestationReasonLen = 30

// CanApprove evaluates all guard conditions for the APPROVE action.
func (p *TransitionPolicy) CanApprove(in ApprovalInput) error {
	// Guard 0 — maker-checker: the approver must differ from the user
	// who opened the day. Evaluated before the state check so that a
	// self-approval attempt is rejected even when other inputs are off.
	if in.CurrentDay != nil && in.CurrentDay.OpenedBy != nil &&
		*in.CurrentDay.OpenedBy == in.ActorID {
		return &domain.ErrSelfApprovalForbidden{
			ContractID:   in.CurrentDay.ContractID.String(),
			BusinessDate: in.CurrentDay.BusinessDate.Format("2006-01-02"),
			ActorID:      in.ActorID.String(),
		}
	}

	// Guard 1 — current state must be open for trading (DAY_OPEN or INVESTMENT_DAY_STARTED).
	if in.CurrentDay == nil || !in.CurrentDay.CurrentState.IsOpenForTrading() {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionApprove),
			CurrentState:    state,
			Reason:          "manager approval requires the investment day to be open",
		}
	}

	// Guard 2 — no pending unreviewed transactions
	if in.HasPendingUnreviewed {
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_APPROVAL_PENDING_TRANSACTIONS",
			AttemptedAction: string(vo.ActionApprove),
			CurrentState:    string(in.CurrentDay.CurrentState),
			Reason:          "one or more transactions have not yet been reviewed; review or void them before approving",
		}
	}

	// Guard 3 — zero-transaction day attestation
	if in.TransactionCount == 0 {
		if !in.ZeroTransactionAttestation {
			return &domain.ErrInvalidTransition{
				Code:            "WORKFLOW_APPROVAL_ZERO_TRANSACTION_UNATTESTED",
				AttemptedAction: string(vo.ActionApprove),
				CurrentState:    string(in.CurrentDay.CurrentState),
				Reason: "no transactions found for this business date; " +
					"set zeroTransactionAttestation=true and provide attestationReason (≥30 chars) to proceed",
			}
		}
		if len(in.AttestationReason) < minAttestationReasonLen {
			return &domain.ErrInvalidTransition{
				Code:            "WORKFLOW_APPROVAL_ATTESTATION_REASON_TOO_SHORT",
				AttemptedAction: string(vo.ActionApprove),
				CurrentState:    string(in.CurrentDay.CurrentState),
				Reason: fmt.Sprintf(
					"attestationReason must be at least %d characters (got %d)",
					minAttestationReasonLen, len(in.AttestationReason),
				),
			}
		}
	}

	return nil
}

const minCancelReasonLen = 20
const maxRecloseCount = 10

// ─────────────────────────────────────────────────────────────────────────────
// CANCEL_DAY_START
// ─────────────────────────────────────────────────────────────────────────────

type CancelDayStartInput struct {
	CurrentDay       *entity.WorkflowDay
	Reason           string
	TransactionCount int
}

func (p *TransitionPolicy) CanCancelDayStart(in CancelDayStartInput) error {
	if in.CurrentDay == nil || !in.CurrentDay.CurrentState.IsOpenForTrading() {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionCancelDayStart),
			CurrentState:    state,
			Reason:          "cancelling day start requires the investment day to be open",
		}
	}
	if err := requireReason(in.Reason, vo.ActionCancelDayStart, in.CurrentDay.CurrentState); err != nil {
		return err
	}
	if in.TransactionCount > 0 {
		return &domain.ErrCancelBlockedByTransactions{
			ContractID:       in.CurrentDay.ContractID.String(),
			BusinessDate:     in.CurrentDay.BusinessDate.Format("2006-01-02"),
			TransactionCount: in.TransactionCount,
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CANCEL_APPROVAL
// ─────────────────────────────────────────────────────────────────────────────

type CancelApprovalInput struct {
	CurrentDay *entity.WorkflowDay
	Reason     string
}

func (p *TransitionPolicy) CanCancelApproval(in CancelApprovalInput) error {
	if in.CurrentDay == nil || !in.CurrentDay.CurrentState.IsManagerApproved() {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionCancelApproval),
			CurrentState:    state,
			Reason:          "revoking approval requires the day to be manager-approved",
		}
	}
	return requireReason(in.Reason, vo.ActionCancelApproval, in.CurrentDay.CurrentState)
}

// ─────────────────────────────────────────────────────────────────────────────
// CLOSE_TRANSACTIONS
// ─────────────────────────────────────────────────────────────────────────────

type CloseTransactionsInput struct {
	CurrentDay *entity.WorkflowDay
}

func (p *TransitionPolicy) CanCloseTransactions(in CloseTransactionsInput) error {
	if in.CurrentDay == nil || !in.CurrentDay.CurrentState.IsManagerApproved() {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionCloseTransactions),
			CurrentState:    state,
			Reason:          "closing transactions requires the day to be manager-approved",
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CANCEL_TRANSACTION_CLOSE
// ─────────────────────────────────────────────────────────────────────────────

type CancelTransactionCloseInput struct {
	CurrentDay *entity.WorkflowDay
	Reason     string
}

func (p *TransitionPolicy) CanCancelTransactionClose(in CancelTransactionCloseInput) error {
	if in.CurrentDay == nil || in.CurrentDay.CurrentState != vo.StateTransactionClosed {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionCancelTransactionClose),
			CurrentState:    state,
			Reason:          "cancelling transaction close requires the day to be in TRANSACTION_CLOSED state",
		}
	}
	return requireReason(in.Reason, vo.ActionCancelTransactionClose, in.CurrentDay.CurrentState)
}

// ─────────────────────────────────────────────────────────────────────────────
// CLOSE_ACCOUNTING
// ─────────────────────────────────────────────────────────────────────────────

type CloseAccountingInput struct {
	CurrentDay *entity.WorkflowDay
}

func (p *TransitionPolicy) CanCloseAccounting(in CloseAccountingInput) error {
	if in.CurrentDay == nil || in.CurrentDay.CurrentState != vo.StateTransactionClosed {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionCloseAccounting),
			CurrentState:    state,
			Reason:          "closing accounting requires the day to be in TRANSACTION_CLOSED state",
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ROLLBACK_ACCOUNTING_CLOSE
// ─────────────────────────────────────────────────────────────────────────────

type RollbackAccountingCloseInput struct {
	CurrentDay *entity.WorkflowDay
	Reason     string
}

// CanRollbackAccountingClose mirrors the DB CHECK constraint chk_wf_reclose_count
// so the operator gets a typed domain error instead of a SQL CHECK violation
// when the rollback budget is exhausted.
func (p *TransitionPolicy) CanRollbackAccountingClose(in RollbackAccountingCloseInput) error {
	if in.CurrentDay == nil || in.CurrentDay.CurrentState != vo.StateAccountingClosed {
		state := string(vo.StateNotStarted)
		if in.CurrentDay != nil {
			state = string(in.CurrentDay.CurrentState)
		}
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_INVALID_TRANSITION",
			AttemptedAction: string(vo.ActionRollbackAccountingClose),
			CurrentState:    state,
			Reason:          "rolling back accounting close requires the day to be in ACCOUNTING_CLOSED state",
		}
	}
	if in.CurrentDay.RecloseCount >= maxRecloseCount {
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_RECLOSE_LIMIT_EXCEEDED",
			AttemptedAction: string(vo.ActionRollbackAccountingClose),
			CurrentState:    string(in.CurrentDay.CurrentState),
			Reason: fmt.Sprintf(
				"reclose_count is %d (max %d); manual intervention required",
				in.CurrentDay.RecloseCount, maxRecloseCount,
			),
			Details: map[string]any{
				"recloseCount": in.CurrentDay.RecloseCount,
				"maxAllowed":   maxRecloseCount,
			},
		}
	}
	return requireReason(in.Reason, vo.ActionRollbackAccountingClose, in.CurrentDay.CurrentState)
}

func requireReason(reason string, action vo.WorkflowAction, currentState vo.WorkflowState) error {
	if len(reason) < minCancelReasonLen {
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_REASON_REQUIRED",
			AttemptedAction: string(action),
			CurrentState:    string(currentState),
			Reason: fmt.Sprintf(
				"a reason of at least %d characters is required for %s (got %d)",
				minCancelReasonLen, action, len(reason),
			),
		}
	}
	return nil
}
