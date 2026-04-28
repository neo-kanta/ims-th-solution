// Package entity holds aggregates/entities for the investment domain.
package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Decision is the aggregate root representing a trader's proposed trade
// ("decision report"). It moves through the 4-step flow (analysis → decision →
// execution → review). The SubmitForExecution command is what promotes it from
// DRAFT to SUBMITTED after the IRG pre-trade check passes.
type Decision struct {
	ID           uuid.UUID
	PortfolioID  uuid.UUID
	ContractID   uuid.UUID
	Ticker       string
	Side         vo.OrderSide
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	Currency     string
	Exchange     string
	BusinessDate time.Time
	Status       vo.DecisionStatus
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UpdatedBy    uuid.UUID

	// ComplianceCheckGroupID stores the IRG check group that last evaluated
	// this decision. Persisted for audit; ties the decision back to the breach
	// override workflow.
	ComplianceCheckGroupID *uuid.UUID
}

// CanSubmitForExecution enforces the state-machine precondition: only DRAFT
// decisions may enter the pre-trade gate. Already-BLOCKED decisions must be
// resurrected via an explicit re-draft — not re-submitted.
func (d *Decision) CanSubmitForExecution() bool {
	return d != nil && d.Status == vo.DecisionStatusDraft
}
