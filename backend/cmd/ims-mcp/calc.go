package main

// calc.go holds the deterministic, decimal-safe financial calculations exposed
// as calc_* MCP tools. All arithmetic uses math/big.Rat (arbitrary precision) —
// never float64 — so money math is exact. These functions are PURE (no HTTP,
// no MCP) so they are fully unit-testable.
//
// Design rules enforced here:
//   - calc_weighted_average_price uses ONLY the explicit authoritative line
//     price (price_in_quote_ccy). If a line is missing that price, or lines span
//     multiple quote currencies, it returns a safe error rather than guessing.
//   - Allocation and exposure use the AUTHORITATIVE portfolio market value
//     (Valuation.MarketValue from /valuations/latest) as the denominator. We do
//     NOT recompute the authoritative total by re-summing lines.

import (
	"errors"
	"math/big"
	"strings"
)

// ValuationLine is the subset of an authoritative valuation holding line the
// calculations consume.
type ValuationLine struct {
	InstrumentID    string
	Quantity        string
	PriceInQuoteCcy string
	QuoteCurrency   string
	MarketValue     string
}

// Valuation is the subset of the authoritative /valuations/latest snapshot the
// calculations consume. MarketValue is the authoritative portfolio total.
type Valuation struct {
	PortfolioID  string
	ValuationCcy string
	MarketValue  string
	BusinessDate string
	Lines        []ValuationLine
}

// AllocationEntry is one group's share of the authoritative portfolio market
// value. Pct is a percentage (0..100) formatted to 4 dp.
type AllocationEntry struct {
	Key         string `json:"key"`
	MarketValue string `json:"market_value"`
	Pct         string `json:"pct"`
}

// pricePrecision / pctPrecision are the fixed scales used to format derived
// values. Inputs remain exact; only the final derived figure is rounded.
const (
	pricePrecision = 6
	pctPrecision   = 4
)

var (
	errNoLines        = errors.New("valuation has no holding lines; cannot compute")
	errCurrencyMix    = errors.New("holding lines span multiple currencies; cannot aggregate into a single figure")
	errZeroQuantity   = errors.New("total quantity is zero; weighted average price is undefined")
	errZeroTotalValue = errors.New("portfolio market value is zero; percentage is undefined")
	errMissingPrice   = errors.New("a holding line is missing an explicit price; cannot compute")
	errInstrNotFound  = errors.New("instrument not found in the portfolio valuation")
)

// ratFromString parses a decimal string into an exact rational. Empty/invalid
// input is an error (callers must not fabricate).
func ratFromString(s string) (*big.Rat, error) {
	clean := strings.TrimSpace(s)
	if clean == "" {
		return nil, errors.New("empty numeric value")
	}
	r := new(big.Rat)
	if _, ok := r.SetString(clean); !ok {
		return nil, errors.New("invalid numeric value: " + s)
	}
	return r, nil
}

// WeightedAveragePrice computes Sum(qty*price) / Sum(qty) across lines using the
// explicit authoritative line price only. Requires a single quote currency.
func WeightedAveragePrice(v Valuation) (result string, currency string, err error) {
	if len(v.Lines) == 0 {
		return "", "", errNoLines
	}
	num := new(big.Rat)
	den := new(big.Rat)
	cur := ""
	for _, ln := range v.Lines {
		if strings.TrimSpace(ln.PriceInQuoteCcy) == "" {
			return "", "", errMissingPrice
		}
		lineCur := strings.TrimSpace(ln.QuoteCurrency)
		if cur == "" {
			cur = lineCur
		} else if lineCur != cur {
			return "", "", errCurrencyMix
		}
		qty, qErr := ratFromString(ln.Quantity)
		if qErr != nil {
			return "", "", qErr
		}
		price, pErr := ratFromString(ln.PriceInQuoteCcy)
		if pErr != nil {
			return "", "", pErr
		}
		num.Add(num, new(big.Rat).Mul(qty, price))
		den.Add(den, qty)
	}
	if den.Sign() == 0 {
		return "", "", errZeroQuantity
	}
	avg := new(big.Rat).Quo(num, den)
	return avg.FloatString(pricePrecision), cur, nil
}

// Allocation computes each group's percentage of the AUTHORITATIVE portfolio
// market value. by is "instrument" or "currency". Group market value is the sum
// of that group's line market values (the numerator); the denominator is the
// authoritative total, never recomputed from lines.
func Allocation(v Valuation, by string) ([]AllocationEntry, error) {
	if len(v.Lines) == 0 {
		return nil, errNoLines
	}
	total, err := ratFromString(v.MarketValue)
	if err != nil {
		return nil, err
	}
	if total.Sign() == 0 {
		return nil, errZeroTotalValue
	}

	order := make([]string, 0)
	sums := map[string]*big.Rat{}
	for _, ln := range v.Lines {
		var key string
		switch by {
		case "currency":
			key = strings.TrimSpace(ln.QuoteCurrency)
		default: // "instrument"
			key = ln.InstrumentID
		}
		mv, mErr := ratFromString(ln.MarketValue)
		if mErr != nil {
			return nil, mErr
		}
		if _, ok := sums[key]; !ok {
			sums[key] = new(big.Rat)
			order = append(order, key)
		}
		sums[key].Add(sums[key], mv)
	}

	hundred := big.NewRat(100, 1)
	out := make([]AllocationEntry, 0, len(order))
	for _, key := range order {
		mv := sums[key]
		pct := new(big.Rat).Quo(mv, total)
		pct.Mul(pct, hundred)
		out = append(out, AllocationEntry{
			Key:         key,
			MarketValue: mv.FloatString(pricePrecision),
			Pct:         pct.FloatString(pctPrecision),
		})
	}
	return out, nil
}

// ExposurePct computes one instrument's market value as a percentage of the
// authoritative portfolio market value. Returns the instrument market value and
// the percentage.
func ExposurePct(v Valuation, instrumentID string) (pct string, marketValue string, err error) {
	if len(v.Lines) == 0 {
		return "", "", errNoLines
	}
	total, tErr := ratFromString(v.MarketValue)
	if tErr != nil {
		return "", "", tErr
	}
	if total.Sign() == 0 {
		return "", "", errZeroTotalValue
	}
	mv := new(big.Rat)
	found := false
	for _, ln := range v.Lines {
		if ln.InstrumentID != instrumentID {
			continue
		}
		found = true
		lineMV, mErr := ratFromString(ln.MarketValue)
		if mErr != nil {
			return "", "", mErr
		}
		mv.Add(mv, lineMV)
	}
	if !found {
		return "", "", errInstrNotFound
	}
	pctRat := new(big.Rat).Quo(mv, total)
	pctRat.Mul(pctRat, big.NewRat(100, 1))
	return pctRat.FloatString(pctPrecision), mv.FloatString(pricePrecision), nil
}
