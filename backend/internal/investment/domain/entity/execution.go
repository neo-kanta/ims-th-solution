package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Execution captures a minimal record of an order send / fill against a
// decision. It is intentionally lightweight — the IMS demo does not integrate
// with a real OMS, so the broker-side bookkeeping is the responsibility of the
// downstream trade confirmation aggregate.
type Execution struct {
	ID         uuid.UUID
	DecisionID uuid.UUID
	// FundID is nil for an execution on a fund-less portfolio.
	FundID             *uuid.UUID
	PortfolioID        uuid.UUID
	InstrumentID       *uuid.UUID
	InstrumentCode     string
	BusinessDate       time.Time
	Side               vo.OrderSide
	OrderedQuantity    *decimal.Decimal
	OrderedAmount      *decimal.Decimal
	ExecutedQuantity   *decimal.Decimal
	ExecutedAmount     *decimal.Decimal
	ExecutionPrice     *decimal.Decimal
	Currency           string
	Status             vo.ExecutionStatus
	TraderUserID       *uuid.UUID
	BrokerReference    string
	ExecutedAt         *time.Time
	CancelledAt        *time.Time
	CancelledBy        *uuid.UUID
	CancellationReason string
	CreatedAt          time.Time
	CreatedBy          uuid.UUID
	UpdatedAt          time.Time
	UpdatedBy          uuid.UUID
}

// CanUpdate reports whether the execution may receive new fills / status
// updates. Cancelled rows are terminal.
func (e *Execution) CanUpdate() bool {
	if e == nil {
		return false
	}
	switch e.Status {
	case vo.ExecutionStatusCancelled:
		return false
	}
	return true
}

// IsTerminal reports whether the execution has reached an end state.
func (e *Execution) IsTerminal() bool {
	if e == nil {
		return false
	}
	switch e.Status {
	case vo.ExecutionStatusExecuted, vo.ExecutionStatusCancelled:
		return true
	}
	return false
}
