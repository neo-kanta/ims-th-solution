package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// DecisionLine is one instrument line within a BASKET_ORDER, REBALANCE, or
// SWITCH decision. SINGLE_ORDER decisions have no lines — their instrument,
// side, and quantity live directly on the parent Decision.
type DecisionLine struct {
	ID             uuid.UUID
	DecisionID     uuid.UUID
	LineNumber     int
	InstrumentID   *uuid.UUID
	InstrumentCode string
	ProductType    vo.DecisionProductType
	Side           vo.OrderSide
	Quantity       *decimal.Decimal
	Amount         *decimal.Decimal
	TargetWeight   *decimal.Decimal
	LimitPrice     *decimal.Decimal
	Currency       string
	Notes          string
	CreatedAt      time.Time
	CreatedBy      uuid.UUID
	UpdatedAt      time.Time
	UpdatedBy      uuid.UUID
}
