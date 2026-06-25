package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// d is a tiny ergonomic helper for the test table — keeps the literals
// readable without dragging in the longer decimal.RequireFromString name.
// Named d, not dec, because the existing portfolio_projector_test already
// owns the dec identifier in the same package.
func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestPickProviderSymbolPrefersConfiguredPrimary(t *testing.T) {
	p := rawPosition{
		ticker:         "AOT",
		primaryExchange: "SET",
		providerYahoo:  "AOT.BK",
		providerAlpha:  "AOT.XYZ",
	}

	require.Equal(t, "AOT.XYZ", pickProviderSymbol(p, "alpha_vantage"))
	require.Equal(t, "AOT.BK", pickProviderSymbol(p, "yahoo"))
}

func TestPickProviderSymbolFallsBackToOtherMappingWhenPrimaryMissing(t *testing.T) {
	p := rawPosition{
		ticker:         "AOT",
		primaryExchange: "SET",
		providerYahoo:  "AOT.BK",
	}

	// Primary is alpha_vantage but we have no alpha symbol — fall back to
	// the yahoo mapping rather than the deterministic suffix so the user's
	// curated mapping wins.
	require.Equal(t, "AOT.BK", pickProviderSymbol(p, "alpha_vantage"))
}

func TestPickProviderSymbolDerivesSETSuffixWhenUnmapped(t *testing.T) {
	p := rawPosition{ticker: "AOT", primaryExchange: "SET"}
	require.Equal(t, "AOT.BK", pickProviderSymbol(p, "alpha_vantage"))
}

func TestPickProviderSymbolReturnsTickerWhenNoExchangeKnown(t *testing.T) {
	p := rawPosition{ticker: "AAPL"}
	require.Equal(t, "AAPL", pickProviderSymbol(p, "yahoo"))
}

func TestPickProviderSymbolReturnsEmptyWhenTickerMissing(t *testing.T) {
	p := rawPosition{}
	require.Empty(t, pickProviderSymbol(p, "yahoo"))
}

// -----------------------------------------------------------------------------
// assembleValuation — calculation, missing units, stale fallback, no-crash
// -----------------------------------------------------------------------------

func newAOT() rawPosition {
	return rawPosition{
		instrumentID:   uuid.New(),
		ticker:         "AOT",
		name:           "Airports of Thailand PCL",
		assetClassCode: "EQUITY",
		assetClassName: "Equity",
		currency:       "THB",
		quantity:       d("5300000"),
		averageCost:    d("61.50"),
		costBasis:      d("325950000"),
	}
}

func TestAssembleValuationComputesMarketValueAndPnL(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{
		Symbol:      "AOT.BK",
		Provider:    "yahoo",
		Price:       d("65.25"),
		Currency:    "THB",
		EffectiveAt: time.Now().UTC(),
	}

	result := &IntradayValuationResult{
		ValuationCcy: "THB",
		CashBalance:  d("73935000"),
		OfficialAUM:  d("1284500000"),
		HasOfficial:  true,
	}

	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)

	require.Len(t, result.Positions, 1)
	row := result.Positions[0]
	require.True(t, d("345825000").Equal(row.MarketValue), "market value = %s", row.MarketValue)
	require.True(t, d("19875000").Equal(row.UnrealisedPnL), "unrealised pnl = %s", row.UnrealisedPnL)
	require.NotNil(t, row.UnrealisedPnLPct)
	require.True(t, d("6.0976").Equal(*row.UnrealisedPnLPct), "unrealised pct = %s", row.UnrealisedPnLPct)

	// Estimated AUM = total market value + cash
	require.True(t, d("419760000").Equal(result.EstimatedAUM), "estimated aum = %s", result.EstimatedAUM)
	require.False(t, result.IsStale)
	require.True(t, result.HasLivePrices)
}

func TestAssembleValuationComputesEstimatedNAVWhenUnitsOutstanding(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{
		Symbol:      "AOT.BK",
		Provider:    "yahoo",
		Price:       d("65.25"),
		Currency:    "THB",
		EffectiveAt: time.Now().UTC(),
	}
	units := d("100000000")
	result := &IntradayValuationResult{
		ValuationCcy:     "THB",
		CashBalance:      d("73935000"),
		OfficialAUM:      d("1284500000"),
		HasOfficial:      true,
		UnitsOutstanding: &units,
	}

	assembleValuation(
		result, true,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)

	require.NotNil(t, result.EstimatedNAVPerUnit)
	// 419 760 000 / 100 000 000 = 4.1976
	require.True(t, d("4.1976").Equal(*result.EstimatedNAVPerUnit),
		"estimated nav per unit = %s", result.EstimatedNAVPerUnit)
}

func TestAssembleValuationSkipsEstimatedNAVWhenUnitsMissing(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{Price: d("65.25"), Currency: "THB", EffectiveAt: time.Now().UTC()}

	result := &IntradayValuationResult{
		ValuationCcy: "THB",
		CashBalance:  d("0"),
		HasOfficial:  true,
		// Note: HasUnits=true but UnitsOutstanding is nil — must not fabricate a NAV.
	}
	assembleValuation(
		result, true,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)
	require.Nil(t, result.EstimatedNAVPerUnit,
		"estimated NAV must be nil when units outstanding is missing")
}

func TestAssembleValuationCarriesAtBookWhenNoQuote(t *testing.T) {
	pos := newAOT() // averageCost 61.50, no lastOfficialPrice
	result := &IntradayValuationResult{
		ValuationCcy: "THB",
		CashBalance:  d("0"),
	}
	// Empty quote map simulates a provider-wide failure.
	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{},
		nil, nil, false, true,
		alwaysFresh,
	)

	require.True(t, result.IsStale, "result must surface stale when no quotes returned")
	require.Len(t, result.Positions, 1)
	row := result.Positions[0]
	require.True(t, row.IsStale)
	require.Equal(t, "no live quote — valued at carrying cost", row.StaleReason)
	// Carried at average cost → market value equals cost basis, no fake P&L.
	require.True(t, d("325950000").Equal(row.MarketValue), "market value = %s", row.MarketValue)
	require.True(t, row.UnrealisedPnL.IsZero(), "unrealised pnl = %s", row.UnrealisedPnL)
	// Estimated AUM is the carried book value + cash — NOT zero, NOT a crash.
	require.True(t, d("325950000").Equal(result.EstimatedAUM), "estimated aum = %s", result.EstimatedAUM)
}

func TestAssembleValuationCarriesAtLastOfficialPriceWhenNoQuote(t *testing.T) {
	pos := newAOT()
	pos.lastOfficialPrice = d("63.00") // last accounting close
	result := &IntradayValuationResult{ValuationCcy: "THB", CashBalance: d("0")}

	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{},
		nil, nil, false, true,
		alwaysFresh,
	)

	row := result.Positions[0]
	require.True(t, row.IsStale)
	require.Equal(t, "no live quote — valued at last official price", row.StaleReason)
	// 5,300,000 * 63.00 = 333,900,000
	require.True(t, d("333900000").Equal(row.MarketValue), "market value = %s", row.MarketValue)
	require.True(t, d("333900000").Equal(result.EstimatedAUM), "estimated aum = %s", result.EstimatedAUM)
}

func TestAssembleValuationZeroWhenNoQuoteAndNoBasis(t *testing.T) {
	pos := rawPosition{
		instrumentID:   uuid.New(),
		ticker:         "TH-LBXX",
		assetClassCode: "FIXED_INCOME",
		quantity:       d("1000"),
		// averageCost & lastOfficialPrice both zero → nothing to carry at.
	}
	result := &IntradayValuationResult{ValuationCcy: "THB", CashBalance: d("0")}

	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{},
		nil, nil, false, true,
		alwaysFresh,
	)

	require.True(t, result.IsStale)
	require.Equal(t, "no provider quote available", result.Positions[0].StaleReason)
	require.True(t, result.EstimatedAUM.IsZero(), "estimated aum = %s", result.EstimatedAUM)
}

func TestAssembleValuationMarksStaleQuoteAsStale(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{
		Price:       d("65.25"),
		Currency:    "THB",
		EffectiveAt: time.Now().UTC(),
		Stale:       true,
		StaleReason: "alpha_vantage rate limited; serving last known price",
	}

	result := &IntradayValuationResult{ValuationCcy: "THB"}
	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, false, true,
		alwaysFresh,
	)

	require.True(t, result.IsStale)
	require.Equal(t, "alpha_vantage rate limited; serving last known price",
		result.Positions[0].StaleReason)
}

func TestAssembleValuationDeltaPctVsLastClose(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{Price: d("65.25"), Currency: "THB", EffectiveAt: time.Now().UTC()}
	result := &IntradayValuationResult{
		ValuationCcy: "THB",
		CashBalance:  d("73935000"),
		OfficialAUM:  d("1284500000"),
		HasOfficial:  true,
	}
	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)

	// delta = (419 760 000 − 1 284 500 000) / 1 284 500 000 * 100 ≈ -67.3211 after Round(4)
	require.Equal(t, "-67.3211", result.DeltaPctVsLastClose.String())
}

func TestAssembleValuationAllocationSumsToWithinRounding(t *testing.T) {
	pos := newAOT()
	quote := &contract.MarketQuote{Price: d("65.25"), Currency: "THB", EffectiveAt: time.Now().UTC()}
	result := &IntradayValuationResult{
		ValuationCcy: "THB",
		CashBalance:  d("73935000"),
	}
	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)
	require.Len(t, result.Allocation, 2, "expect EQUITY + CASH buckets")

	var totalPct decimal.Decimal
	for _, b := range result.Allocation {
		totalPct = totalPct.Add(b.PctOfTotal)
	}
	diff := totalPct.Sub(decimal.NewFromInt(100)).Abs()
	require.True(t, diff.LessThanOrEqual(decimal.NewFromFloat(0.05)),
		"allocation % should sum to ~100; got %s", totalPct)
}

func alwaysFresh(time.Time) bool { return false }
