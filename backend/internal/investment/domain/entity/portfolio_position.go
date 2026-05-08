package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PortfolioPosition is the current projection of holdings for a single
// (portfolio, instrument) pair. It is recomputable from the immutable ledger;
// the stored row exists for query speed and is guarded by an optimistic
// `Version` field.
type PortfolioPosition struct {
	ID                uuid.UUID
	PortfolioID       uuid.UUID
	InstrumentID      uuid.UUID
	Quantity          decimal.Decimal
	AverageCost       decimal.Decimal
	CostBasis         decimal.Decimal
	LastTransactionID *uuid.UUID
	LastBusinessDate  *time.Time
	Version           int
	UpdatedAt         time.Time
}

// IsZero reports whether the position is fully closed.
func (p *PortfolioPosition) IsZero() bool {
	return p == nil || p.Quantity.Sign() == 0
}
