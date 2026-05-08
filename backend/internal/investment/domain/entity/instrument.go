package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Instrument is a tradable security in the master data set.
// Asset-class-specific fields live in Attributes (JSONB) — this keeps the
// schema stable when adding bond/derivative support later.
type Instrument struct {
	ID              uuid.UUID
	PrimaryTicker   string
	Name            string
	AssetClassID    uuid.UUID
	AssetSubtypeID  uuid.UUID
	Currency        string
	CountryID       uuid.UUID
	RegionID        *uuid.UUID
	PrimaryExchange string
	SectorID        *uuid.UUID
	FundCategoryID  *uuid.UUID
	LotSize         int
	TickSize        *decimal.Decimal
	IsTradable      bool
	Status          vo.InstrumentStatus
	Attributes      map[string]any

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedAt *time.Time
}

// IsTradableNow returns true if a new transaction may reference this instrument.
func (i *Instrument) IsTradableNow() bool {
	return i != nil && i.DeletedAt == nil && i.IsTradable && i.Status == vo.InstrumentStatusActive
}

// InstrumentIdentifier maps an instrument to an external identifier (ISIN,
// CUSIP, BBG ticker, vendor symbol, etc.).
type InstrumentIdentifier struct {
	ID           uuid.UUID
	InstrumentID uuid.UUID
	IDType       string
	IDValue      string
	ProviderCode string
	IsPrimary    bool

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
}
