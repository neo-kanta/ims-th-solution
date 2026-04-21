package spi

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// CheckInput is the request context passed to every rule evaluator.
type CheckInput struct {
	CheckGroupID uuid.UUID      `json:"check_group_id"`
	Timing       vo.CheckTiming `json:"timing"`
	BusinessDate time.Time      `json:"business_date"`
	Actor        string         `json:"actor"` // user ID or "system"

	PortfolioID uuid.UUID `json:"portfolio_id"`
	ContractID  uuid.UUID `json:"contract_id"`

	// ProposedOrder is set for pre-trade checks. Nil for periodic/post-trade portfolio scans.
	ProposedOrder *ProposedOrder `json:"proposed_order,omitempty"`
}

// ProposedOrder represents the order being evaluated.
type ProposedOrder struct {
	OrderID  uuid.UUID       `json:"order_id"`
	Ticker   string          `json:"ticker"`
	Side     vo.OrderSide    `json:"side"`
	Quantity decimal.Decimal `json:"quantity"`
	Price    decimal.Decimal `json:"price"`
	Currency string          `json:"currency"`
	Exchange string          `json:"exchange"`
}

// TradeValue returns Quantity * Price.
func (o *ProposedOrder) TradeValue() decimal.Decimal {
	return o.Quantity.Mul(o.Price)
}
