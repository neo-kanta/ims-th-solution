package service

import (
	"context"
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
		ticker:          "AOT",
		primaryExchange: "SET",
		providerYahoo:   "AOT.BK",
		providerAlpha:   "AOT.XYZ",
	}

	require.Equal(t, "AOT.XYZ", pickProviderSymbol(p, "alpha_vantage"))
	require.Equal(t, "AOT.BK", pickProviderSymbol(p, "yahoo"))
}

func TestPickProviderSymbolFallsBackToOtherMappingWhenPrimaryMissing(t *testing.T) {
	p := rawPosition{
		ticker:          "AOT",
		primaryExchange: "SET",
		providerYahoo:   "AOT.BK",
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

// -----------------------------------------------------------------------------
// assembleValuation — ROI and mark-to-market metadata (new fields)
// -----------------------------------------------------------------------------

func TestAssembleValuationComputesROIAndSourceForLiveQuote(t *testing.T) {
	pos := newAOT() // costBasis 325950000
	effectiveAt := time.Date(2026, 7, 1, 9, 30, 0, 0, time.UTC)
	fetchedAt := time.Date(2026, 7, 1, 9, 31, 0, 0, time.UTC)
	quote := &contract.MarketQuote{
		Symbol: "AOT.BK", Provider: "yahoo", Price: d("65.25"), Currency: "THB",
		EffectiveAt: effectiveAt, FetchedAt: fetchedAt, Source: contract.QuoteSourceLive,
	}
	result := &IntradayValuationResult{ValuationCcy: "THB", CashBalance: d("0")}

	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)

	row := result.Positions[0]
	require.NotNil(t, row.ROI)
	// unrealised pnl 19,875,000 / cost 325,950,000 = 0.060977...
	require.True(t, d("0.06098").Sub(*row.ROI).Abs().LessThan(d("0.00001")), "roi = %s", row.ROI)
	require.Equal(t, contract.QuoteSourceLive, row.Source)
	require.True(t, row.PriceEffectiveDate.Equal(effectiveAt))
	require.True(t, row.FetchedAt.Equal(fetchedAt))

	require.NotNil(t, result.ROI)
	require.True(t, result.CostBasis.Equal(d("325950000")))
	require.True(t, result.MarketValue.Equal(d("345825000")))
}

// Zero cost basis must never divide-by-zero: ROI (and the legacy pct field)
// stay nil rather than panicking or emitting Inf/NaN.
func TestAssembleValuationROINilWhenCostBasisZero(t *testing.T) {
	pos := rawPosition{
		instrumentID: uuid.New(),
		ticker:       "FREESHARES",
		quantity:     d("100"),
		averageCost:  d("0"),
		costBasis:    d("0"),
	}
	quote := &contract.MarketQuote{Price: d("10.00"), Currency: "THB", EffectiveAt: time.Now().UTC()}
	result := &IntradayValuationResult{ValuationCcy: "THB"}

	assembleValuation(
		result, false,
		[]rawPosition{pos},
		map[uuid.UUID]*contract.MarketQuote{pos.instrumentID: quote},
		nil, []string{"yahoo"}, true, false,
		alwaysFresh,
	)

	require.Nil(t, result.Positions[0].ROI)
	require.Nil(t, result.Positions[0].UnrealisedPnLPct)
	require.Nil(t, result.ROI, "aggregate ROI must be nil when total cost basis is zero")
	require.True(t, result.Positions[0].MarketValue.Equal(d("1000")))
}

// The official-price-snapshot fallback must surface its own source metadata
// (id, effective date, fetched-at, provider/source label) for the mark-to-
// market contract, in addition to the existing stale flag/reason.
func TestAssembleValuationOfficialPriceSnapshotCarriesMetadata(t *testing.T) {
	pos := newAOT()
	pos.lastOfficialPrice = d("63.00")
	pos.lastOfficialPriceID = "11111111-1111-1111-1111-111111111111"
	pos.lastOfficialPriceSource = "DEMO_SEED"
	snapDate := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	capturedAt := time.Date(2026, 6, 28, 7, 0, 0, 0, time.UTC)
	pos.lastOfficialPriceDate = &snapDate
	pos.lastOfficialPriceCapturedAt = &capturedAt

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
	require.Equal(t, "official_price_snapshot", row.Source)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", row.PriceSnapshotID)
	require.Equal(t, "DEMO_SEED", row.Provider)
	require.True(t, row.PriceEffectiveDate.Equal(snapDate))
	require.True(t, row.FetchedAt.Equal(capturedAt))
}

// -----------------------------------------------------------------------------
// applyHoldingsAsOfStatus — honesty flag for historical business_date
// requests: only prices are resolved as of that date, not positions/cash.
// -----------------------------------------------------------------------------

func TestApplyHoldingsAsOfStatusConfirmedForToday(t *testing.T) {
	result := &IntradayValuationResult{}
	applyHoldingsAsOfStatus(result, true, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

	require.True(t, result.HoldingsAsOfConfirmed)
	require.Empty(t, result.HoldingsAsOfNote)
	require.False(t, result.IsStale)
}

func TestApplyHoldingsAsOfStatusFlagsHistoricalDateAsUnconfirmed(t *testing.T) {
	result := &IntradayValuationResult{}
	applyHoldingsAsOfStatus(result, false, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC))

	require.False(t, result.HoldingsAsOfConfirmed)
	require.Contains(t, result.HoldingsAsOfNote, "2026-06-28")
	require.True(t, result.IsStale, "an unconfirmed historical view must surface via the existing stale signal")
	require.Contains(t, result.StaleReason, "2026-06-28")
}

func TestApplyHoldingsAsOfStatusAppendsToExistingStaleReason(t *testing.T) {
	result := &IntradayValuationResult{IsStale: true, StaleReason: "no live quote for AOT"}
	applyHoldingsAsOfStatus(result, false, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC))

	require.Contains(t, result.StaleReason, "no live quote for AOT")
	require.Contains(t, result.StaleReason, "2026-06-28")
}

// -----------------------------------------------------------------------------
// resolveQuote — live vs as-of tier selection, provider-down never crashes
// -----------------------------------------------------------------------------

type fakeQuoteProvider struct {
	latestQuote *contract.MarketQuote
	latestErr   error
	asOfQuote   *contract.MarketQuote
	asOfErr     error
}

func (f *fakeQuoteProvider) GetLatestQuote(context.Context, string) (*contract.MarketQuote, error) {
	return f.latestQuote, f.latestErr
}
func (f *fakeQuoteProvider) GetQuoteAsOf(context.Context, string, time.Time) (*contract.MarketQuote, error) {
	return f.asOfQuote, f.asOfErr
}
func (f *fakeQuoteProvider) PrimaryProviderName() string    { return "fake" }
func (f *fakeQuoteProvider) ProviderConfigured(string) bool { return true }

func newTestIntradayService(quotes contract.MarketQuoteProvider) *IntradayValuationService {
	return NewIntradayValuationService(nil, quotes, nil, nil, nil, nil, IntradayConfig{})
}

func TestResolveQuotePrefersLiveWhenTodayAndFresh(t *testing.T) {
	live := &contract.MarketQuote{Price: d("65.25"), Source: contract.QuoteSourceLive}
	asOf := &contract.MarketQuote{Price: d("60.00"), Source: contract.QuoteSourceMarketDataSnapshot}
	svc := newTestIntradayService(&fakeQuoteProvider{latestQuote: live, asOfQuote: asOf})

	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Same(t, live, got)
}

func TestResolveQuoteFallsBackToAsOfWhenLiveStale(t *testing.T) {
	live := &contract.MarketQuote{Price: d("65.25"), Stale: true, StaleReason: "rate limited"}
	asOf := &contract.MarketQuote{Price: d("64.00"), Stale: false}
	svc := newTestIntradayService(&fakeQuoteProvider{latestQuote: live, asOfQuote: asOf})

	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Same(t, asOf, got, "a fresh as-of snapshot must win over a stale live quote")
}

func TestResolveQuoteKeepsStaleLiveWhenAsOfAlsoUnavailable(t *testing.T) {
	live := &contract.MarketQuote{Price: d("65.25"), Stale: true, StaleReason: "rate limited"}
	svc := newTestIntradayService(&fakeQuoteProvider{latestQuote: live, asOfErr: contract.ErrMarketDataUnavailable})

	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Same(t, live, got, "a stale live quote is still better than nothing")
}

// Regression: a provider outage must not be silently erased. GetLatestQuote
// falls back to the last cached snapshot row and marks it Stale=true.
// GetQuoteAsOf resolves independently and its staleness is a pure
// calendar-date match against businessDate — if it happens to land on the
// exact same underlying snapshot row (same SnapshotID), it would report that
// row as fresh even though it is the same outage-era data. Without the
// same-row guard, resolveQuote would "upgrade" to it and the frontend would
// see a fresh valuation during a live outage.
func TestResolveQuoteDoesNotUnstaleSameCachedRowDuringOutage(t *testing.T) {
	live := &contract.MarketQuote{
		Price: d("65.25"), Stale: true, StaleReason: "provider unavailable — cached",
		Source: contract.QuoteSourceMarketDataSnapshot, SnapshotID: "snap-123",
	}
	asOf := &contract.MarketQuote{
		Price: d("65.25"), Stale: false, // pure date-match sees "today" and calls it fresh
		Source: contract.QuoteSourceMarketDataSnapshot, SnapshotID: "snap-123", // same row
	}
	svc := newTestIntradayService(&fakeQuoteProvider{latestQuote: live, asOfQuote: asOf})

	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Same(t, live, got, "must not un-stale a cached row via a same-row as-of lookup")
	require.True(t, got.Stale)
}

// When the as-of tier resolves a genuinely different (newer) row — e.g. a
// DAILY close posted after the cached row was captured — it is a legitimate
// upgrade and must still replace the stale cached quote.
func TestResolveQuoteUpgradesToDifferentFreshRowDuringOutage(t *testing.T) {
	live := &contract.MarketQuote{
		Price: d("65.25"), Stale: true, StaleReason: "provider unavailable — cached",
		SnapshotID: "snap-123",
	}
	asOf := &contract.MarketQuote{
		Price: d("66.10"), Stale: false,
		SnapshotID: "snap-456", // a different, newer row
	}
	svc := newTestIntradayService(&fakeQuoteProvider{latestQuote: live, asOfQuote: asOf})

	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Same(t, asOf, got, "a genuinely different fresh row must still win over a stale cached quote")
}

func TestResolveQuoteSkipsLiveForPastBusinessDate(t *testing.T) {
	live := &contract.MarketQuote{Price: d("65.25")}
	asOf := &contract.MarketQuote{Price: d("63.00")}
	provider := &fakeQuoteProvider{latestQuote: live, asOfQuote: asOf}
	svc := newTestIntradayService(provider)

	past := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	got := svc.resolveQuote(context.Background(), "AOT.BK", past, false)
	require.Same(t, asOf, got, "a historical business_date must resolve from the as-of tier only")
}

// Provider chain fully down (both live and as-of error) must degrade to nil,
// not panic — the caller (assembleValuation) already proves nil quotes
// produce a stale-but-valid valuation, never a 500.
func TestResolveQuoteReturnsNilWhenProviderFullyUnavailable(t *testing.T) {
	svc := newTestIntradayService(&fakeQuoteProvider{
		latestErr: contract.ErrProviderTimeout,
		asOfErr:   contract.ErrMarketDataUnavailable,
	})
	require.NotPanics(t, func() {
		got := svc.resolveQuote(context.Background(), "TH-LB30DA", time.Now().UTC(), true)
		require.Nil(t, got)
	})
}

func TestResolveQuoteNilWhenNoProviderConfigured(t *testing.T) {
	svc := newTestIntradayService(nil)
	got := svc.resolveQuote(context.Background(), "AOT.BK", time.Now().UTC(), true)
	require.Nil(t, got)
}

// -----------------------------------------------------------------------------
// normalizeBusinessDate
// -----------------------------------------------------------------------------

func TestNormalizeBusinessDateDefaultsToTodayWhenZero(t *testing.T) {
	now := time.Date(2026, 7, 1, 14, 22, 0, 0, time.UTC)
	bd, isToday := normalizeBusinessDate(time.Time{}, now)
	require.True(t, isToday)
	require.Equal(t, "2026-07-01", bd.Format("2006-01-02"))
}

func TestNormalizeBusinessDateTruncatesToUTCCalendarDate(t *testing.T) {
	now := time.Date(2026, 7, 1, 14, 22, 0, 0, time.UTC)
	requested := time.Date(2026, 6, 28, 23, 59, 0, 0, time.UTC)
	bd, isToday := normalizeBusinessDate(requested, now)
	require.False(t, isToday)
	require.Equal(t, "2026-06-28", bd.Format("2006-01-02"))
	require.True(t, bd.Equal(time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)))
}

func TestNormalizeBusinessDateEqualsTodayWhenSameCalendarDate(t *testing.T) {
	now := time.Date(2026, 7, 1, 23, 0, 0, 0, time.UTC)
	requested := time.Date(2026, 7, 1, 1, 0, 0, 0, time.UTC)
	_, isToday := normalizeBusinessDate(requested, now)
	require.True(t, isToday)
}
