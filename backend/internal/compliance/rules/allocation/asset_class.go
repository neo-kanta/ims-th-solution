// Package allocation implements portfolio asset-allocation rules.
// Self-registers "allocation.asset_class_max" and "allocation.asset_class_min"
// via init().
package allocation

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&AssetClassMaxRule{})
	spi.Register(&AssetClassMinRule{})
}

// ─────────────────────────────────────────────────────────────────────────────
// allocation.asset_class_max
// ─────────────────────────────────────────────────────────────────────────────

// AssetClassMaxRule caps portfolio exposure to a named asset class as a % of NAV.
// TypeID: "allocation.asset_class_max"
type AssetClassMaxRule struct{}

// MaxParams configures the asset class and its ceiling.
type MaxParams struct {
	// AssetClass is the asset class code to monitor (must match
	// ClassificationSnapshot / InstrumentClassification.AssetClass values,
	// e.g. EQUITY, FIXED_INCOME, ALTERNATIVE).
	AssetClass string `json:"asset_class"`
	// MaxPercentNAV is the maximum allowed exposure as % of NAV (e.g. 60 = 60%).
	MaxPercentNAV decimal.Decimal `json:"max_percent_nav"`
}

func (r *AssetClassMaxRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "allocation.asset_class_max",
		Version:         "1.0.0",
		Category:        spi.CategoryMandate,
		DefaultSeverity: vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
			vo.TimingPeriodic,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopeGlobal,
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Caps exposure to a named asset class (e.g. stock, bond, alternative) as a percentage of NAV.",
	}
}

func (r *AssetClassMaxRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["asset_class", "max_percent_nav"],
		"properties": {
			"asset_class": {
				"type": "string",
				"minLength": 1,
				"description": "Asset class code to cap (e.g. EQUITY, FIXED_INCOME, ALTERNATIVE)"
			},
			"max_percent_nav": {
				"type": "number",
				"minimum": 0,
				"maximum": 100,
				"description": "Maximum asset-class exposure as % of NAV"
			}
		}
	}`)}
}

func (r *AssetClassMaxRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Positions:       true,
		NAV:             true,
		MarketPrices:    true,
		Classifications: true,
		PortfolioMeta:   true,
	}
}

func (r *AssetClassMaxRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p MaxParams
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.AssetClass == "" {
		return spi.EvalResult{}, fmt.Errorf("asset_class parameter is required")
	}
	if p.MaxPercentNAV.IsZero() || p.MaxPercentNAV.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("max_percent_nav must be positive, got %s", p.MaxPercentNAV)
	}
	if missing := missingAssetClassifications(input, data); len(missing) > 0 && !isNonLivePortfolio(data.PortfolioMeta) {
		return assetClassUnavailable(p.AssetClass, missing), nil
	}

	pct, classMV, nav, ok := assetClassExposure(input, data, p.AssetClass)
	if !ok {
		return assetClassBlock("NAV data unavailable — cannot compute asset class exposure"), nil
	}

	limit := p.MaxPercentNAV
	pctDisplay := pct.Mul(decimal.NewFromInt(100)).StringFixed(4)
	limitDisplay := limit.StringFixed(4)

	metrics := map[string]string{
		"asset_class_exposure_pct": pctDisplay,
		"asset_class_market_value": classMV.StringFixed(2),
		"nav":                      nav.StringFixed(2),
	}
	references := map[string]string{"asset_class": p.AssetClass}
	if input.ProposedOrder != nil {
		references["proposed_ticker"] = input.ProposedOrder.Ticker
	}

	if pct.Mul(decimal.NewFromInt(100)).GreaterThan(limit) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"asset class '%s' exposure would be %s%% (max %s%%)",
				p.AssetClass, pctDisplay, limitDisplay,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "asset_class_exposure_pct",
					Actual:     pctDisplay,
					Limit:      limitDisplay,
					Operator:   "<=",
					Unit:       "percent",
				},
				References: references,
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"asset class '%s' exposure %s%% is within max %s%%",
			p.AssetClass, pctDisplay, limitDisplay,
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *AssetClassMaxRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p MaxParams
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'allocation.asset_class_max' caps '%s' exposure at %s%% of NAV. Verdict: %s. %s",
		p.AssetClass, p.MaxPercentNAV.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// ─────────────────────────────────────────────────────────────────────────────
// allocation.asset_class_min
// ─────────────────────────────────────────────────────────────────────────────

// AssetClassMinRule enforces a minimum portfolio exposure to a named asset
// class as a % of NAV — e.g. a mandate that must keep at least 10% in bonds.
// TypeID: "allocation.asset_class_min"
type AssetClassMinRule struct{}

// MinParams configures the asset class and its floor.
type MinParams struct {
	AssetClass string `json:"asset_class"`
	// MinPercentNAV is the minimum required exposure as % of NAV.
	MinPercentNAV decimal.Decimal `json:"min_percent_nav"`
}

func (r *AssetClassMinRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "allocation.asset_class_min",
		Version:         "1.0.0",
		Category:        spi.CategoryMandate,
		DefaultSeverity: vo.SeverityWarn,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
			vo.TimingPeriodic,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopeGlobal,
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Enforces a minimum exposure to a named asset class (e.g. stock, bond, alternative) as a percentage of NAV.",
	}
}

func (r *AssetClassMinRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["asset_class", "min_percent_nav"],
		"properties": {
			"asset_class": {
				"type": "string",
				"minLength": 1,
				"description": "Asset class code to floor (e.g. EQUITY, FIXED_INCOME, ALTERNATIVE)"
			},
			"min_percent_nav": {
				"type": "number",
				"minimum": 0,
				"maximum": 100,
				"description": "Minimum required asset-class exposure as % of NAV"
			}
		}
	}`)}
}

func (r *AssetClassMinRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Positions:       true,
		NAV:             true,
		MarketPrices:    true,
		Classifications: true,
		PortfolioMeta:   true,
	}
}

func (r *AssetClassMinRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p MinParams
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.AssetClass == "" {
		return spi.EvalResult{}, fmt.Errorf("asset_class parameter is required")
	}
	if p.MinPercentNAV.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("min_percent_nav must be non-negative, got %s", p.MinPercentNAV)
	}
	if missing := missingAssetClassifications(input, data); len(missing) > 0 && !isNonLivePortfolio(data.PortfolioMeta) {
		return assetClassUnavailable(p.AssetClass, missing), nil
	}

	pct, classMV, nav, ok := assetClassExposure(input, data, p.AssetClass)
	if !ok {
		return assetClassBlock("NAV data unavailable — cannot compute asset class exposure"), nil
	}

	floor := p.MinPercentNAV
	pctDisplay := pct.Mul(decimal.NewFromInt(100)).StringFixed(4)
	floorDisplay := floor.StringFixed(4)

	metrics := map[string]string{
		"asset_class_exposure_pct": pctDisplay,
		"asset_class_market_value": classMV.StringFixed(2),
		"nav":                      nav.StringFixed(2),
	}
	references := map[string]string{"asset_class": p.AssetClass}
	if input.ProposedOrder != nil {
		references["proposed_ticker"] = input.ProposedOrder.Ticker
	}

	if pct.Mul(decimal.NewFromInt(100)).LessThan(floor) {
		return spi.EvalResult{
			Verdict: vo.VerdictWarn,
			Message: fmt.Sprintf(
				"asset class '%s' exposure would be %s%% (min %s%%)",
				p.AssetClass, pctDisplay, floorDisplay,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "asset_class_exposure_pct",
					Actual:     pctDisplay,
					Limit:      floorDisplay,
					Operator:   ">=",
					Unit:       "percent",
				},
				References: references,
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"asset class '%s' exposure %s%% meets min %s%%",
			p.AssetClass, pctDisplay, floorDisplay,
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *AssetClassMinRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p MinParams
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'allocation.asset_class_min' requires at least %s%% of NAV in '%s'. Verdict: %s. %s",
		p.MinPercentNAV.StringFixed(2), p.AssetClass, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// ─────────────────────────────────────────────────────────────────────────────
// shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// assetClassExposure sums the market value of all holdings in assetClass,
// applies the proposed order's delta if it falls in that asset class, and
// returns the resulting exposure as a fraction of NAV (0.30 = 30%).
func assetClassExposure(
	input spi.CheckInput,
	data spi.DataBundle,
	assetClass string,
) (pct decimal.Decimal, classMV decimal.Decimal, nav decimal.Decimal, ok bool) {
	if data.NAV == nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, false
	}
	nav = data.NAV.NAV
	if nav.IsZero() || nav.IsNegative() {
		return decimal.Zero, decimal.Zero, nav, false
	}

	classMV = decimal.Zero
	if data.Positions != nil {
		for _, h := range data.Positions.Holdings {
			ac := ""
			if data.Classifications != nil {
				if c, found := data.Classifications.Get(h.Ticker); found {
					ac = c.AssetClass
				}
			}
			if ac != assetClass {
				continue
			}
			classMV = classMV.Add(assetClassHoldingMV(h, data.MarketPrices))
		}
	}

	if input.ProposedOrder != nil {
		ord := input.ProposedOrder
		orderClass := ""
		if data.Classifications != nil {
			if c, found := data.Classifications.Get(ord.Ticker); found {
				orderClass = c.AssetClass
			}
		}
		if orderClass == assetClass {
			if ord.Side == vo.OrderSideBuy {
				classMV = classMV.Add(ord.TradeValue())
			} else {
				classMV = classMV.Sub(ord.TradeValue())
				if classMV.IsNegative() {
					classMV = decimal.Zero
				}
			}
		}
	}

	pct = classMV.Div(nav)
	return pct, classMV, nav, true
}

func assetClassHoldingMV(h spi.Holding, prices *spi.MarketPriceSnapshot) decimal.Decimal {
	if prices != nil {
		if price, ok := prices.Prices[h.Ticker]; ok && price.IsPositive() {
			return h.Quantity.Mul(price)
		}
	}
	return h.MarketValue
}

func assetClassBlock(msg string) spi.EvalResult {
	return spi.EvalResult{
		Verdict:  vo.VerdictBlock,
		Message:  msg,
		Evidence: vo.Evidence{Metrics: map[string]string{"error": msg}},
	}
}

func missingAssetClassifications(input spi.CheckInput, data spi.DataBundle) []string {
	missing := map[string]struct{}{}
	check := func(ticker string) {
		ticker = strings.TrimSpace(ticker)
		if ticker == "" {
			return
		}
		classification, ok := data.Classifications.Get(ticker)
		if !ok || strings.TrimSpace(classification.AssetClass) == "" {
			missing[ticker] = struct{}{}
		}
	}

	if data.Positions != nil {
		for _, holding := range data.Positions.Holdings {
			check(holding.Ticker)
		}
	}
	if input.ProposedOrder != nil {
		check(input.ProposedOrder.Ticker)
	}

	tickers := make([]string, 0, len(missing))
	for ticker := range missing {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)
	return tickers
}

func assetClassUnavailable(assetClass string, missing []string) spi.EvalResult {
	missingList := strings.Join(missing, ", ")
	return spi.EvalResult{
		Status:  vo.ComplianceStatusUnavailable,
		Verdict: vo.VerdictBlock,
		Message: fmt.Sprintf("asset class classification unavailable for instruments: %s", missingList),
		Evidence: vo.Evidence{
			Metrics: map[string]string{
				"compliance_status":        string(vo.ComplianceStatusUnavailable),
				"reason":                   "MISSING_ASSET_CLASSIFICATION",
				"missing_instrument_count": fmt.Sprintf("%d", len(missing)),
			},
			References: map[string]string{
				"asset_class":         assetClass,
				"missing_instruments": missingList,
			},
		},
	}
}

func isNonLivePortfolio(meta *spi.PortfolioMetadata) bool {
	if meta == nil {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(meta.PortfolioType)) {
	case "SIMULATION", "MODEL":
		return true
	default:
		return false
	}
}
