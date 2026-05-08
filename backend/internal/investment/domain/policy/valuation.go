package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// HoldingValuationInput is the per-instrument data fed into the valuation pass.
type HoldingValuationInput struct {
	InstrumentID         uuid.UUID
	PriceSnapshotID      *uuid.UUID
	Quantity             decimal.Decimal
	CostBasisBase        decimal.Decimal // already in valuation/base ccy
	PriceInQuoteCcy      decimal.Decimal
	QuoteCurrency        string
	FxRateToValuationCcy decimal.Decimal
	IsStale              bool
}

// HoldingValuationOutput is the per-instrument output of the valuation pass.
type HoldingValuationOutput struct {
	HoldingValuationInput
	MarketValue   decimal.Decimal
	UnrealisedPnL decimal.Decimal
}

// PortfolioValuationOutput aggregates a valuation pass at the portfolio level.
type PortfolioValuationOutput struct {
	MarketValue    decimal.Decimal
	CostBasis      decimal.Decimal
	UnrealisedPnL  decimal.Decimal
	ROI            *decimal.Decimal
	HasStaleInputs bool
	Holdings       []HoldingValuationOutput
}

// ComputePortfolioValuation aggregates per-holding values into a snapshot
// total. Cash is added by the caller (kept separate so this function stays a
// pure aggregator over holdings).
//
// Rules:
//   market_value_per_holding = quantity * (price_in_quote_ccy * fx_to_valuation_ccy)
//   unrealised_pnl_per_holding = market_value - cost_basis
//   portfolio.market_value = sum holdings.market_value
//   portfolio.cost_basis   = sum holdings.cost_basis
//   portfolio.unrealised   = portfolio.market_value - portfolio.cost_basis
//   portfolio.roi          = portfolio.unrealised / portfolio.cost_basis (nil when cost_basis = 0)
//   portfolio.has_stale_inputs = any holding.is_stale
func ComputePortfolioValuation(holdings []HoldingValuationInput) PortfolioValuationOutput {
	totalMV := decimal.Zero
	totalCB := decimal.Zero
	staleSeen := false

	out := make([]HoldingValuationOutput, 0, len(holdings))

	for _, h := range holdings {
		priceInValCcy := h.PriceInQuoteCcy.Mul(h.FxRateToValuationCcy)
		mv := h.Quantity.Mul(priceInValCcy)
		upnl := mv.Sub(h.CostBasisBase)

		out = append(out, HoldingValuationOutput{
			HoldingValuationInput: h,
			MarketValue:           mv,
			UnrealisedPnL:         upnl,
		})

		totalMV = totalMV.Add(mv)
		totalCB = totalCB.Add(h.CostBasisBase)
		if h.IsStale {
			staleSeen = true
		}
	}

	totalUPnL := totalMV.Sub(totalCB)

	var roi *decimal.Decimal
	if totalCB.Sign() > 0 {
		r := totalUPnL.Div(totalCB)
		roi = &r
	}

	return PortfolioValuationOutput{
		MarketValue:    totalMV,
		CostBasis:      totalCB,
		UnrealisedPnL:  totalUPnL,
		ROI:            roi,
		HasStaleInputs: staleSeen,
		Holdings:       out,
	}
}

// ComputePriceSetHash produces a stable SHA-256 hash of the (instrument_id,
// price_snapshot_id) pairs used in a valuation. The hash is recorded on the
// snapshot so a future audit can prove which prices were used.
//
// Holdings with no price snapshot (PriceSnapshotID == nil) are included as
// "<instrument>:none" so a missing-price set is distinguishable from any
// real set.
func ComputePriceSetHash(holdings []HoldingValuationInput) string {
	pairs := make([]string, 0, len(holdings))
	for _, h := range holdings {
		ps := "none"
		if h.PriceSnapshotID != nil {
			ps = h.PriceSnapshotID.String()
		}
		pairs = append(pairs, h.InstrumentID.String()+":"+ps)
	}
	sort.Strings(pairs)

	hasher := sha256.New()
	for _, p := range pairs {
		_, _ = hasher.Write([]byte(p))
		_, _ = hasher.Write([]byte{0x1F}) // unit separator
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// ComputeAUM returns the AUM = market_value + cash_in_valuation_ccy.
func ComputeAUM(marketValue, cashInValuationCcy decimal.Decimal) decimal.Decimal {
	return marketValue.Add(cashInValuationCcy)
}

// ComputeNAVPerUnit returns aum / total_units. Caller MUST guard
// total_units > 0; this function returns zero when units are non-positive.
func ComputeNAVPerUnit(aum, totalUnits decimal.Decimal) decimal.Decimal {
	if totalUnits.Sign() <= 0 {
		return decimal.Zero
	}
	return aum.Div(totalUnits)
}
