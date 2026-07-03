package main

import (
	"errors"
	"testing"
)

func lines() []ValuationLine {
	return []ValuationLine{
		{InstrumentID: "i1", Quantity: "100", PriceInQuoteCcy: "10.00", QuoteCurrency: "THB", MarketValue: "1000.00"},
		{InstrumentID: "i2", Quantity: "300", PriceInQuoteCcy: "30.00", QuoteCurrency: "THB", MarketValue: "9000.00"},
	}
}

// Weighted average price: (100*10 + 300*30) / (100+300) = 10000/400 = 25.
func TestWeightedAveragePrice_Exact(t *testing.T) {
	v := Valuation{MarketValue: "10000.00", Lines: lines()}
	got, cur, err := WeightedAveragePrice(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cur != "THB" {
		t.Fatalf("expected THB, got %q", cur)
	}
	if got != "25.000000" {
		t.Fatalf("expected 25.000000, got %q", got)
	}
}

// Decimal-safe: 1/3 weighting must not drift like float64 would.
func TestWeightedAveragePrice_DecimalSafe(t *testing.T) {
	v := Valuation{Lines: []ValuationLine{
		{Quantity: "1", PriceInQuoteCcy: "0.10", QuoteCurrency: "THB", MarketValue: "0.10"},
		{Quantity: "2", PriceInQuoteCcy: "0.20", QuoteCurrency: "THB", MarketValue: "0.40"},
	}}
	got, _, err := WeightedAveragePrice(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (1*0.10 + 2*0.20) / 3 = 0.50/3 = 0.166666...
	if got != "0.166667" {
		t.Fatalf("expected 0.166667, got %q", got)
	}
}

func TestWeightedAveragePrice_CurrencyMismatch(t *testing.T) {
	v := Valuation{Lines: []ValuationLine{
		{Quantity: "1", PriceInQuoteCcy: "10", QuoteCurrency: "THB", MarketValue: "10"},
		{Quantity: "1", PriceInQuoteCcy: "10", QuoteCurrency: "USD", MarketValue: "10"},
	}}
	if _, _, err := WeightedAveragePrice(v); !errors.Is(err, errCurrencyMix) {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}
}

func TestWeightedAveragePrice_MissingPrice(t *testing.T) {
	v := Valuation{Lines: []ValuationLine{{Quantity: "1", PriceInQuoteCcy: "", QuoteCurrency: "THB", MarketValue: "0"}}}
	if _, _, err := WeightedAveragePrice(v); !errors.Is(err, errMissingPrice) {
		t.Fatalf("expected missing price error, got %v", err)
	}
}

func TestWeightedAveragePrice_EmptyLines(t *testing.T) {
	if _, _, err := WeightedAveragePrice(Valuation{}); !errors.Is(err, errNoLines) {
		t.Fatalf("expected no-lines error, got %v", err)
	}
}

func TestWeightedAveragePrice_ZeroQuantity(t *testing.T) {
	v := Valuation{Lines: []ValuationLine{{Quantity: "0", PriceInQuoteCcy: "10", QuoteCurrency: "THB", MarketValue: "0"}}}
	if _, _, err := WeightedAveragePrice(v); !errors.Is(err, errZeroQuantity) {
		t.Fatalf("expected zero-quantity error, got %v", err)
	}
}

// Allocation uses the AUTHORITATIVE total as denominator (not a re-sum of lines).
// Authoritative total = 10000; i1=1000 -> 10%, i2=9000 -> 90%.
func TestAllocation_ByInstrument_UsesAuthoritativeTotal(t *testing.T) {
	v := Valuation{MarketValue: "10000.00", Lines: lines()}
	entries, err := Allocation(v, "instrument")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "i1" || entries[0].Pct != "10.0000" {
		t.Fatalf("expected i1=10.0000, got %+v", entries[0])
	}
	if entries[1].Key != "i2" || entries[1].Pct != "90.0000" {
		t.Fatalf("expected i2=90.0000, got %+v", entries[1])
	}
}

// If the authoritative total differs from the sum of lines, the percentages
// must still use the authoritative total (proving we don't recompute it).
func TestAllocation_DenominatorIsAuthoritativeNotLineSum(t *testing.T) {
	v := Valuation{MarketValue: "20000.00", Lines: lines()} // lines sum to 10000
	entries, err := Allocation(v, "instrument")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries[0].Pct != "5.0000" { // 1000 / 20000 = 5%
		t.Fatalf("expected 5.0000 against authoritative total, got %q", entries[0].Pct)
	}
}

func TestAllocation_ByCurrency(t *testing.T) {
	v := Valuation{MarketValue: "10000.00", Lines: lines()}
	entries, err := Allocation(v, "currency")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].Key != "THB" || entries[0].Pct != "100.0000" {
		t.Fatalf("expected single THB=100.0000 group, got %+v", entries)
	}
}

func TestAllocation_ZeroTotal(t *testing.T) {
	v := Valuation{MarketValue: "0", Lines: lines()}
	if _, err := Allocation(v, "instrument"); !errors.Is(err, errZeroTotalValue) {
		t.Fatalf("expected zero-total error, got %v", err)
	}
}

func TestExposurePct(t *testing.T) {
	v := Valuation{MarketValue: "10000.00", Lines: lines()}
	pct, mv, err := ExposurePct(v, "i2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pct != "90.0000" {
		t.Fatalf("expected 90.0000, got %q", pct)
	}
	if mv != "9000.000000" {
		t.Fatalf("expected market value 9000.000000, got %q", mv)
	}
}

func TestExposurePct_InstrumentNotFound(t *testing.T) {
	v := Valuation{MarketValue: "10000.00", Lines: lines()}
	if _, _, err := ExposurePct(v, "nope"); !errors.Is(err, errInstrNotFound) {
		t.Fatalf("expected instrument-not-found error, got %v", err)
	}
}

func TestDecodeValuation_Enveloped(t *testing.T) {
	body := []byte(`{"data":{"portfolio_id":"p1","valuation_ccy":"THB","market_value":"10000.00","business_date":"2026-06-05","holding_lines":[{"instrument_id":"i1","quantity":"100","price_in_quote_ccy":"10.00","quote_currency":"THB","market_value":"1000.00"}]}}`)
	v, err := decodeValuation(body)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if v.MarketValue != "10000.00" || len(v.Lines) != 1 || v.Lines[0].InstrumentID != "i1" {
		t.Fatalf("unexpected decode: %+v", v)
	}
}
