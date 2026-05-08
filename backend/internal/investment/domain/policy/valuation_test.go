package policy_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
)

func TestComputePortfolioValuation_Empty(t *testing.T) {
	out := policy.ComputePortfolioValuation(nil)
	assert.True(t, out.MarketValue.IsZero())
	assert.True(t, out.CostBasis.IsZero())
	assert.True(t, out.UnrealisedPnL.IsZero())
	assert.Nil(t, out.ROI)
	assert.False(t, out.HasStaleInputs)
}

func TestComputePortfolioValuation_HappyPath(t *testing.T) {
	holdings := []policy.HoldingValuationInput{
		{
			InstrumentID:         uuid.New(),
			Quantity:             d("100"),
			CostBasisBase:        d("5000"),
			PriceInQuoteCcy:      d("60"),
			QuoteCurrency:        "THB",
			FxRateToValuationCcy: d("1"),
		},
		{
			InstrumentID:         uuid.New(),
			Quantity:             d("50"),
			CostBasisBase:        d("4000"),
			PriceInQuoteCcy:      d("90"),
			QuoteCurrency:        "THB",
			FxRateToValuationCcy: d("1"),
		},
	}
	out := policy.ComputePortfolioValuation(holdings)

	// MV = 100*60 + 50*90 = 6000 + 4500 = 10500
	assert.True(t, d("10500").Equal(out.MarketValue))
	// CB = 5000 + 4000 = 9000
	assert.True(t, d("9000").Equal(out.CostBasis))
	// UPnL = 1500
	assert.True(t, d("1500").Equal(out.UnrealisedPnL))
	// ROI = 1500/9000 = 0.1666...
	assert.NotNil(t, out.ROI)
	expected := d("1500").Div(d("9000"))
	assert.True(t, expected.Equal(*out.ROI))
}

func TestComputePortfolioValuation_StalePropagates(t *testing.T) {
	holdings := []policy.HoldingValuationInput{
		{
			InstrumentID:         uuid.New(),
			Quantity:             d("10"),
			CostBasisBase:        d("100"),
			PriceInQuoteCcy:      d("12"),
			QuoteCurrency:        "THB",
			FxRateToValuationCcy: d("1"),
			IsStale:              true,
		},
	}
	out := policy.ComputePortfolioValuation(holdings)
	assert.True(t, out.HasStaleInputs)
}

func TestComputePortfolioValuation_FX(t *testing.T) {
	holdings := []policy.HoldingValuationInput{
		{
			InstrumentID:         uuid.New(),
			Quantity:             d("10"),
			CostBasisBase:        d("3600"), // already in valuation ccy
			PriceInQuoteCcy:      d("100"),  // USD price
			QuoteCurrency:        "USD",
			FxRateToValuationCcy: d("36"), // 1 USD = 36 THB
		},
	}
	out := policy.ComputePortfolioValuation(holdings)
	// MV = 10 * 100 * 36 = 36000
	assert.True(t, d("36000").Equal(out.MarketValue))
	// UPnL = 36000 - 3600 = 32400
	assert.True(t, d("32400").Equal(out.UnrealisedPnL))
}

func TestComputePortfolioValuation_ROINilWhenZeroCostBasis(t *testing.T) {
	holdings := []policy.HoldingValuationInput{
		{
			InstrumentID:         uuid.New(),
			Quantity:             d("10"),
			CostBasisBase:        decimal.Zero,
			PriceInQuoteCcy:      d("12"),
			QuoteCurrency:        "THB",
			FxRateToValuationCcy: d("1"),
		},
	}
	out := policy.ComputePortfolioValuation(holdings)
	assert.Nil(t, out.ROI, "ROI must be nil when cost_basis is zero")
}

func TestComputePriceSetHash_StableAcrossOrder(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	ps1 := uuid.New()
	ps2 := uuid.New()
	a := []policy.HoldingValuationInput{
		{InstrumentID: id1, PriceSnapshotID: &ps1},
		{InstrumentID: id2, PriceSnapshotID: &ps2},
	}
	b := []policy.HoldingValuationInput{
		{InstrumentID: id2, PriceSnapshotID: &ps2},
		{InstrumentID: id1, PriceSnapshotID: &ps1},
	}
	assert.Equal(t, policy.ComputePriceSetHash(a), policy.ComputePriceSetHash(b),
		"price set hash must be order-independent")
}

func TestComputeAUM(t *testing.T) {
	assert.True(t, d("1500").Equal(policy.ComputeAUM(d("1000"), d("500"))))
}

func TestComputeNAVPerUnit(t *testing.T) {
	assert.True(t, d("10").Equal(policy.ComputeNAVPerUnit(d("100"), d("10"))))
	assert.True(t, decimal.Zero.Equal(policy.ComputeNAVPerUnit(d("100"), decimal.Zero)))
}
