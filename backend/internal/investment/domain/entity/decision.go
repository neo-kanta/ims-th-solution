// Package entity holds aggregates/entities for the investment domain.
package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Decision is the aggregate root for a trader's proposed trade (step 2 of the
// 4-step flow: analysis → decision → execution → review).
//
// Persistence lives in infrastructure/persistence/decision_repository.go and
// matches database/migrations/20260601000001_investment__create_decisions_*.up.sql
// extended by 20260615000001_investment__add_decision_basket_fields.up.sql.
//
// Status transitions
//
//	DRAFT             — newly created, content still editable.
//	PENDING_APPROVAL  — submitted to the approval engine (via INVESTMENT_DECISION
//	                    process). Awaiting reviewer action.
//	APPROVED          — approval engine emitted a final APPROVED decision.
//	REJECTED          — approval engine rejected the request. Terminal.
//	CANCELLED         — submitter (or admin) withdrew the decision. Terminal.
//	READY_FOR_EXECUTION — backed by an active execution row.
//	EXECUTED          — execution + confirmation reconciled; terminal.
type Decision struct {
	ID             uuid.UUID
	DecisionNumber string
	FundID         uuid.UUID
	PortfolioID    uuid.UUID

	// InstrumentID and InstrumentCode are nil/empty for BASKET_ORDER, REBALANCE,
	// and SWITCH headers — the per-instrument detail lives in DecisionLines.
	InstrumentID   *uuid.UUID
	InstrumentCode string
	BusinessDate   time.Time

	// Side is empty string for basket/rebalance/switch headers.
	// The persistence layer maps empty string to SQL NULL.
	Side       vo.OrderSide
	Quantity   *decimal.Decimal
	Amount     *decimal.Decimal
	LimitPrice *decimal.Decimal
	Currency   string
	Exchange   string

	// Classification fields added in 20260615000001.
	DecisionType DecisionType
	ProcessType  vo.DecisionProcessType
	ProductType  vo.DecisionProductType
	StrategyCode string
	AmendmentNo  int

	ResearchReportID  *uuid.UUID
	ResearchReportNo  string
	Rationale         string
	Status            vo.DecisionLifecycleStatus
	ApprovalRequestID *uuid.UUID
	// ComplianceReleaseApprovalRequestID is the approval request ID for the
	// COMPLIANCE_RELEASE stage, kept separate from ApprovalRequestID so both
	// approval stages remain traceable via the API.
	ComplianceReleaseApprovalRequestID *uuid.UUID
	ApprovalStatus                     string

	ComplianceCheckGroupID *uuid.UUID
	SubmitterUserID        uuid.UUID
	SubmittedAt            *time.Time
	CancelledAt            *time.Time
	CancelledBy            *uuid.UUID
	CancellationReason     string
	ReadyForExecutionAt    *time.Time
	CreatedAt              time.Time
	CreatedBy              uuid.UUID
	UpdatedAt              time.Time
	UpdatedBy              uuid.UUID

	// Lines is populated by repository queries that join decision_lines.
	// Nil for SINGLE_ORDER decisions that have no child lines.
	Lines []*DecisionLine
}

// DecisionType is an alias to the value-object type for field clarity.
type DecisionType = vo.DecisionType

// CanEdit reports whether free-form content fields (quantity, rationale, etc.)
// may still be edited. After submission the decision is locked to mutation
// and only the approval engine + execution can change its state.
func (d *Decision) CanEdit() bool {
	if d == nil {
		return false
	}
	return d.Status == vo.DecisionLifecycleDraft
}

// CanSubmit reports whether the decision may be submitted for approval.
func (d *Decision) CanSubmit() bool {
	return d != nil && d.Status == vo.DecisionLifecycleDraft
}

// CanCancel reports whether the submitter may cancel the decision.
// We allow cancellation while in DRAFT (trivial) or while still PENDING_APPROVAL
// — once the decision is APPROVED/READY_FOR_EXECUTION cancellation must go
// through a compensating workflow not modelled in this phase.
func (d *Decision) CanCancel() bool {
	if d == nil {
		return false
	}
	switch d.Status {
	case vo.DecisionLifecycleDraft, vo.DecisionLifecyclePendingApproval,
		vo.DecisionLifecyclePendingComplianceRelease:
		return true
	}
	return false
}

// CanMarkReadyForExecution reports whether the decision is in a state from
// which an execution row may be opened.
func (d *Decision) CanMarkReadyForExecution() bool {
	return d != nil && d.Status == vo.DecisionLifecycleApproved
}

// CanSubmitForExecution preserves the pre-existing pre-trade gate API so the
// SubmitDecisionForExecutionHandler keeps compiling. In this phase we treat
// APPROVED as the gate input — the legacy command path (with the IRG
// pre-trade pipeline) still requires DRAFT, so we delegate to it.
func (d *Decision) CanSubmitForExecution() bool {
	return d != nil && d.Status == vo.DecisionLifecycleDraft
}
