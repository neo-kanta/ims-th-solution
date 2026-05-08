package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CashMovement is an immutable cash ledger row. Each movement is denominated
// in a single currency; multi-currency portfolios accumulate one balance row
// per currency.
type CashMovement struct {
	ID            uuid.UUID
	PortfolioID   uuid.UUID
	Currency      string
	Amount        decimal.Decimal
	BusinessDate  time.Time
	TransactionID *uuid.UUID
	MovementType  string

	CreatedAt time.Time
	CreatedBy uuid.UUID
}

// CashBalance is the current projection of cash for a (portfolio, currency).
type CashBalance struct {
	ID               uuid.UUID
	PortfolioID      uuid.UUID
	Currency         string
	Balance          decimal.Decimal
	LastMovementID   *uuid.UUID
	LastBusinessDate *time.Time
	Version          int
	UpdatedAt        time.Time
}
