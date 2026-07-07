package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// TradeConfirmation reconciles the broker-confirmed quantities/prices against
// a previous Execution. The aggregate is consumed by the workflow closing
// gate: an execution without a MATCHED or REVIEWED confirmation blocks
// transaction closing for the business date.
type TradeConfirmation struct {
	ID                uuid.UUID
	ExecutionID       uuid.UUID
	DecisionID        uuid.UUID
	FundID            uuid.UUID
	PortfolioID       uuid.UUID
	BusinessDate      time.Time
	ConfirmedQuantity *decimal.Decimal
	ConfirmedAmount   *decimal.Decimal
	ConfirmedPrice    *decimal.Decimal
	Currency          string
	BrokerReference   string
	ImportBatchID     *uuid.UUID
	Status            vo.TradeConfirmationStatus
	DiscrepancyReason string
	ReviewedAt        *time.Time
	ReviewedBy        *uuid.UUID
	CreatedAt         time.Time
	CreatedBy         uuid.UUID
	UpdatedAt         time.Time
	UpdatedBy         uuid.UUID
}

// IsResolved reports whether the confirmation is in a state acceptable to
// the workflow closing gate (MATCHED or REVIEWED with reason).
func (c *TradeConfirmation) IsResolved() bool {
	if c == nil {
		return false
	}
	switch c.Status {
	case vo.TradeConfirmationMatched, vo.TradeConfirmationReviewed:
		return true
	}
	return false
}
