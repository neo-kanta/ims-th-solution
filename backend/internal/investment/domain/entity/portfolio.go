package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Portfolio is the unit of investment management owned by a Fund.
// A portfolio aggregates positions, cash balances, transactions, and
// valuation history. tax_lot_method is a forward-compatibility placeholder;
// all current math uses average-cost.
type Portfolio struct {
	ID                uuid.UUID
	FundID            uuid.UUID
	Code              string
	Name              string
	Description       string
	BaseCurrency      string
	ValuationCurrency string
	StrategyCode      string
	StyleID           *uuid.UUID
	ManagerUserID     *uuid.UUID
	Benchmark         string
	RiskProfile       vo.RiskProfile
	InceptionDate     time.Time
	Status            vo.PortfolioStatus
	HasUnits          bool
	TaxLotMethod      vo.TaxLotMethod
	Version           int

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedAt *time.Time
}

// IsActive reports whether the portfolio accepts new ledger activity.
func (p *Portfolio) IsActive() bool {
	return p != nil && p.DeletedAt == nil && p.Status == vo.PortfolioStatusActive
}
