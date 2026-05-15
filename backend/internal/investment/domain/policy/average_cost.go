// Package policy contains pure-function business rules for the investment
// domain. Functions here MUST NOT touch I/O — they are unit-testable as
// table-driven cases.
package policy

import (
	"github.com/shopspring/decimal"
)

// AverageCostInputs is the data needed to apply a buy-side post to an existing
// average-cost position. All decimal values are expressed in the portfolio's
// base currency.
type AverageCostInputs struct {
	OldQuantity    decimal.Decimal
	OldAverageCost decimal.Decimal // per-unit, in base ccy
	BuyQuantity    decimal.Decimal
	BuyPriceBase   decimal.Decimal // price per unit, in base ccy (already FX'd)
	BuyFeesBase    decimal.Decimal // total fees for the lot, in base ccy
}

// AverageCostResult holds the new projection.
type AverageCostResult struct {
	NewQuantity    decimal.Decimal
	NewAverageCost decimal.Decimal
	NewCostBasis   decimal.Decimal
}

// ApplyBuyAverageCost computes the new average-cost position after a buy.
// Formula:
//
//	new_qty       = old_qty + buy_qty
//	new_avg_cost  = (old_qty * old_avg_cost + buy_qty * buy_price_base + buy_fees_base) / new_qty
//	new_cost_basis = new_qty * new_avg_cost
//
// When old_qty is zero, the formula reduces to:
//
//	new_avg_cost = (buy_qty * buy_price_base + buy_fees_base) / buy_qty
//
// Inputs must be non-negative. buy_qty must be > 0; otherwise the function
// returns the inputs unchanged (caller guards this elsewhere).
func ApplyBuyAverageCost(in AverageCostInputs) AverageCostResult {
	if in.BuyQuantity.Sign() <= 0 {
		return AverageCostResult{
			NewQuantity:    in.OldQuantity,
			NewAverageCost: in.OldAverageCost,
			NewCostBasis:   in.OldQuantity.Mul(in.OldAverageCost),
		}
	}

	oldNotional := in.OldQuantity.Mul(in.OldAverageCost)
	buyNotional := in.BuyQuantity.Mul(in.BuyPriceBase).Add(in.BuyFeesBase)

	newQty := in.OldQuantity.Add(in.BuyQuantity)
	newAvg := oldNotional.Add(buyNotional).Div(newQty)
	newBasis := newQty.Mul(newAvg)

	return AverageCostResult{
		NewQuantity:    newQty,
		NewAverageCost: newAvg,
		NewCostBasis:   newBasis,
	}
}

// SellInputs is the data needed to apply a sell-side post to an average-cost
// position.
type SellInputs struct {
	OldQuantity    decimal.Decimal
	OldAverageCost decimal.Decimal // per-unit, in base ccy
	SellQuantity   decimal.Decimal
	SellPriceBase  decimal.Decimal // per-unit, in base ccy
	SellFeesBase   decimal.Decimal // total fees, in base ccy
}

// SellResult contains both the new projection and the realised P/L.
type SellResult struct {
	NewQuantity     decimal.Decimal
	NewAverageCost  decimal.Decimal
	NewCostBasis    decimal.Decimal
	RealisedPnLBase decimal.Decimal
}

// ApplySellAverageCost computes the new projection after a sell.
// Average-cost stays the same; quantity decreases. When the position is
// fully closed, average_cost resets to zero so a future buy starts from a
// clean slate.
//
// Realised P/L (in base ccy):
//
//	realised = sell_qty * (sell_price_base - old_avg_cost) - sell_fees_base
//
// Caller MUST validate sell_qty <= old_qty BEFORE calling — this function
// asserts via a return-as-zero-position fallback if it doesn't.
func ApplySellAverageCost(in SellInputs) SellResult {
	if in.SellQuantity.Sign() <= 0 {
		return SellResult{
			NewQuantity:    in.OldQuantity,
			NewAverageCost: in.OldAverageCost,
			NewCostBasis:   in.OldQuantity.Mul(in.OldAverageCost),
		}
	}

	newQty := in.OldQuantity.Sub(in.SellQuantity)
	if newQty.Sign() < 0 {
		// Defensive: caller must reject oversells. Return the inputs to avoid
		// silently corrupting the projection.
		return SellResult{
			NewQuantity:    in.OldQuantity,
			NewAverageCost: in.OldAverageCost,
			NewCostBasis:   in.OldQuantity.Mul(in.OldAverageCost),
		}
	}

	avg := in.OldAverageCost
	if newQty.Sign() == 0 {
		avg = decimal.Zero
	}
	newBasis := newQty.Mul(avg)

	priceMinusCost := in.SellPriceBase.Sub(in.OldAverageCost)
	realised := in.SellQuantity.Mul(priceMinusCost).Sub(in.SellFeesBase)

	return SellResult{
		NewQuantity:     newQty,
		NewAverageCost:  avg,
		NewCostBasis:    newBasis,
		RealisedPnLBase: realised,
	}
}
