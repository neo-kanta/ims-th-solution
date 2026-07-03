package query

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// applyMarketValueOverride is the mechanism guaranteeing allocation cannot
// silently disagree with the holdings mark-to-market view (issue: allocation
// used an independent, unconditional "latest ever" price lookup). These
// tests exercise it directly, without a database, mirroring how the intraday
// valuation service tests assembleValuation as a pure function.
func TestApplyMarketValueOverrideReplacesValueForKnownInstrument(t *testing.T) {
	instID := uuid.New()
	rows := []rawAllocationRow{
		{instrumentID: instID, assetClassCode: "EQUITY", marketValue: decimal.RequireFromString("100.00")},
	}
	mtm := map[uuid.UUID]decimal.Decimal{
		instID: decimal.RequireFromString("142.50"),
	}

	out := applyMarketValueOverride(rows, mtm)

	require.Len(t, out, 1)
	require.True(t, decimal.RequireFromString("142.50").Equal(out[0].marketValue),
		"allocation must use the holdings valuation's resolved market value, got %s", out[0].marketValue)
}

func TestApplyMarketValueOverrideLeavesUnknownInstrumentsUntouched(t *testing.T) {
	knownID := uuid.New()
	unknownID := uuid.New()
	rows := []rawAllocationRow{
		{instrumentID: knownID, marketValue: decimal.RequireFromString("100.00")},
		{instrumentID: unknownID, marketValue: decimal.RequireFromString("50.00")},
	}
	mtm := map[uuid.UUID]decimal.Decimal{
		knownID: decimal.RequireFromString("110.00"),
	}

	out := applyMarketValueOverride(rows, mtm)

	require.True(t, decimal.RequireFromString("110.00").Equal(out[0].marketValue))
	require.True(t, decimal.RequireFromString("50.00").Equal(out[1].marketValue),
		"an instrument the mark-to-market view didn't resolve must keep its SQL-derived (business_date-bounded) value")
}

// When the holdings valuation is entirely unavailable (e.g. market_data not
// wired, or the intraday call failed), allocation must fall back to the raw
// SQL-computed values rather than losing data — this is the "never crash,
// degrade gracefully" contract applied to allocation.
func TestApplyMarketValueOverrideNoOpWhenMTMMapEmpty(t *testing.T) {
	instID := uuid.New()
	rows := []rawAllocationRow{
		{instrumentID: instID, marketValue: decimal.RequireFromString("333900000")},
	}

	out := applyMarketValueOverride(rows, nil)

	require.True(t, decimal.RequireFromString("333900000").Equal(out[0].marketValue))
}

// allocationPctDenominator is the fix for allocation disagreeing with the
// holdings mark-to-market view: bucket values can be MTM-overridden, but
// before this fix the pct denominator stayed the official accounting NAV
// unconditionally, so percentages didn't foot against a book that had
// actually moved on live prices.
func TestAllocationPctDenominatorUsesOfficialNAVWhenNoMTMOverride(t *testing.T) {
	rows := []rawAllocationRow{{marketValue: decimal.RequireFromString("100")}}

	got := allocationPctDenominator(
		decimal.RequireFromString("1000"), decimal.Zero, rows, nil,
	)

	require.True(t, decimal.RequireFromString("1000").Equal(got),
		"with no MTM override, the denominator must stay the official NAV")
}

func TestAllocationPctDenominatorUsesMTMTotalWhenOverrideApplied(t *testing.T) {
	instID := uuid.New()
	rows := []rawAllocationRow{
		{instrumentID: instID, marketValue: decimal.RequireFromString("142.50")},
	}
	mtm := map[uuid.UUID]decimal.Decimal{instID: decimal.RequireFromString("142.50")}

	// Official NAV (1000) diverges from the MTM book (142.50 rows + 50 cash
	// = 192.50) — the denominator must follow the MTM book, not the stale
	// official figure, so the bucket's own pct matches its own market value.
	got := allocationPctDenominator(
		decimal.RequireFromString("1000"), decimal.RequireFromString("50"), rows, mtm,
	)

	require.True(t, decimal.RequireFromString("192.50").Equal(got),
		"with an MTM override applied, the denominator must be the MTM book (rows + cash), got %s", got)
}

func TestAllocationPctDenominatorFallsBackToOfficialWhenMTMTotalIsZero(t *testing.T) {
	instID := uuid.New()
	rows := []rawAllocationRow{{instrumentID: instID, marketValue: decimal.Zero}}
	mtm := map[uuid.UUID]decimal.Decimal{instID: decimal.Zero}

	got := allocationPctDenominator(decimal.RequireFromString("1000"), decimal.Zero, rows, mtm)

	require.True(t, decimal.RequireFromString("1000").Equal(got),
		"a zero/degenerate MTM total must not divide-by-zero the percentages — fall back to official NAV")
}
