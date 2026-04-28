package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// WorkflowDay is the core aggregate of the Workflow Management module.
//
// Aggregate key: (ContractID, BusinessDate).
// One row in workflow__day_states per contract per calendar date.
//
// BusinessDate is always a DATE in Asia/Bangkok calendar semantics.
// It is parsed from an explicit YYYY-MM-DD string provided by the operator;
// it is NEVER derived from time.Now() or the server clock.
type WorkflowDay struct {
	ID           uuid.UUID
	ContractID   uuid.UUID
	BusinessDate time.Time // DATE column; year/month/day in Bangkok calendar

	// State machine position
	CurrentState         vo.WorkflowState
	TransactionsLockedAt *time.Time // set simultaneously with ManagerApprovedAt
	PendingReclose       bool
	RecloseCount         int

	// Per-stage completion timestamps (nil until stage is reached)
	OpenedAt            *time.Time
	OpenedBy            *uuid.UUID
	ManagerApprovedAt   *time.Time
	ManagerApprovedBy   *uuid.UUID
	TransactionClosedAt *time.Time
	TransactionClosedBy *uuid.UUID
	AccountingClosedAt  *time.Time
	AccountingClosedBy  *uuid.UUID

	// Optimistic locking — incremented on every UPDATE
	Version int

	// Standard audit columns (UTC)
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// IsTransactionLocked reports whether new investment transactions are blocked.
// Transactions are locked from Manager Approval onwards.
func (d *WorkflowDay) IsTransactionLocked() bool {
	return d.TransactionsLockedAt != nil
}

// AllowedActions returns the set of actions that can structurally be taken
// from the current state. Does not check permissions or guard conditions.
func (d *WorkflowDay) AllowedActions() []vo.WorkflowAction {
	return vo.AllowedActions(d.CurrentState)
}
