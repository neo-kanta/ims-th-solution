package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PortfolioTransaction is an immutable ledger row.
// Once persisted, no field is mutable — corrections are made by posting a new
// row of type REVERSAL referencing this one.
type PortfolioTransaction struct {
	ID          uuid.UUID
	PortfolioID uuid.UUID
	// FundID is nil for a transaction on a fund-less portfolio.
	FundID          *uuid.UUID
	InstrumentID    *uuid.UUID
	TransactionType vo.TransactionType
	Side            *vo.OrderSide

	Quantity        *decimal.Decimal
	Price           *decimal.Decimal
	Currency        string
	GrossAmount     decimal.Decimal
	Fees            decimal.Decimal
	NetAmount       decimal.Decimal
	RealisedPnLBase decimal.Decimal
	FxRateToBase    *decimal.Decimal

	BusinessDate   time.Time
	SettlementDate *time.Time

	SourceDecisionID      *uuid.UUID
	SourceExecutionID     *uuid.UUID
	ReversesTransactionID *uuid.UUID
	ExternalRef           string
	Reason                string
	Status                vo.TransactionStatus

	CreatedAt time.Time
	CreatedBy uuid.UUID
}

// IsReversal returns true when this row reverses an earlier transaction.
func (t *PortfolioTransaction) IsReversal() bool {
	return t != nil && t.TransactionType == vo.TransactionTypeReversal
}
