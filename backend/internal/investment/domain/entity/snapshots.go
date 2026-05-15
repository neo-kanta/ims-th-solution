package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PriceSnapshot is an append-only price observation per
// (instrument, business_date, source).
type PriceSnapshot struct {
	ID           uuid.UUID
	InstrumentID uuid.UUID
	BusinessDate time.Time
	Price        decimal.Decimal
	Currency     string
	PriceSource  string
	ProviderRef  string
	IsStale      bool
	StaleReason  string
	CapturedAt   time.Time
	CreatedAt    time.Time
	CreatedBy    uuid.UUID
}

// ValuationSnapshot is an append-only portfolio valuation as of a business date.
// IsIndicative is true for any internally-computed snapshot; only an external
// PAM integration may produce a non-indicative snapshot.
type ValuationSnapshot struct {
	ID           uuid.UUID
	PortfolioID  uuid.UUID
	BusinessDate time.Time
	ValuationCcy string

	MarketValue   decimal.Decimal
	CostBasis     decimal.Decimal
	UnrealisedPnL decimal.Decimal
	RealisedPnL   decimal.Decimal
	ROI           *decimal.Decimal
	AUM           decimal.Decimal
	CashBalance   decimal.Decimal

	PriceSetHash   string
	HasStaleInputs bool
	IsIndicative   bool
	Source         vo.ValuationSource

	CreatedAt time.Time
	CreatedBy uuid.UUID

	HoldingLines []ValuationHoldingLine
}

// ValuationHoldingLine is the per-instrument detail attached to a
// ValuationSnapshot.
type ValuationHoldingLine struct {
	ID                   uuid.UUID
	ValuationSnapshotID  uuid.UUID
	InstrumentID         uuid.UUID
	PriceSnapshotID      *uuid.UUID
	Quantity             decimal.Decimal
	PriceInQuoteCcy      decimal.Decimal
	QuoteCurrency        string
	FxRateToValuationCcy decimal.Decimal
	MarketValue          decimal.Decimal
	CostBasis            decimal.Decimal
	UnrealisedPnL        decimal.Decimal
	IsStale              bool
	CreatedAt            time.Time
}

// NAVSnapshot is produced only when Portfolio.HasUnits is true.
type NAVSnapshot struct {
	ID                  uuid.UUID
	PortfolioID         uuid.UUID
	BusinessDate        time.Time
	TotalUnits          decimal.Decimal
	NAVPerUnit          decimal.Decimal
	ValuationSnapshotID uuid.UUID
	IsIndicative        bool
	CreatedAt           time.Time
	CreatedBy           uuid.UUID
}

// AUMSnapshot records AUM at either portfolio or fund scope.
type AUMSnapshot struct {
	ID           uuid.UUID
	ScopeType    vo.AumScopeType
	ScopeID      uuid.UUID
	BusinessDate time.Time
	AUM          decimal.Decimal
	ValuationCcy string
	Source       vo.ValuationSource
	CreatedAt    time.Time
	CreatedBy    uuid.UUID
}
