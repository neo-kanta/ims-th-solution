// Package ratio implements portfolio ratio / exposure rules.
// Self-registers as "ratio.sector_exposure" via init().
package ratio

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&SectorExposureRule{})
}

// SectorExposureRule caps portfolio exposure to a named sector as a % of NAV.
// TypeID: "ratio.sector_exposure"
type SectorExposureRule struct{}

// SectorParams configures the sector and its limit.
type SectorParams struct {
	// Sector is the sector name to monitor (must match ClassificationSnapshot.Sector() values).
	Sector string `json:"sector"`
	// MaxPct is the maximum allowed sector exposure as % of NAV (e.g. 30 = 30%).
	MaxPct decimal.Decimal `json:"max_pct"`
}

func (r *SectorExposureRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "ratio.sector_exposure",
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
		Description: "Limits exposure to a single GICS/SET sector as a percentage of NAV.",
	}
}

func (r *SectorExposureRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["sector", "max_pct"],
		"properties": {
			"sector": {
				"type": "string",
				"minLength": 1,
				"description": "Sector name to cap (e.g. ENERGY, FINANCIALS, TECHNOLOGY)"
			},
			"max_pct": {
				"type": "number",
				"minimum": 0,
				"maximum": 100,
				"description": "Maximum sector exposure as % of NAV"
			}
		}
	}`)}
}

func (r *SectorExposureRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Positions:       true,
		NAV:             true,
		MarketPrices:    true,
		Classifications: true,
	}
}

func (r *SectorExposureRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p SectorParams
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.Sector == "" {
		return spi.EvalResult{}, fmt.Errorf("sector parameter is required")
	}
	if p.MaxPct.IsZero() || p.MaxPct.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("max_pct must be positive, got %s", p.MaxPct)
	}

	if data.NAV == nil {
		return sectorBlock("NAV data unavailable — cannot compute sector exposure"), nil
	}
	nav := data.NAV.NAV
	if nav.IsZero() || nav.IsNegative() {
		return sectorBlock("NAV is zero or negative — cannot compute sector exposure"), nil
	}

	// Sum market value of all holdings in the target sector.
	sectorMV := decimal.Zero
	proposedInSector := false

	if data.Positions != nil {
		for _, h := range data.Positions.Holdings {
			sector := ""
			if data.Classifications != nil {
				sector = data.Classifications.Sector(h.Ticker)
			}
			if sector != p.Sector {
				continue
			}
			sectorMV = sectorMV.Add(holdingMV(h, data.MarketPrices))
		}
	}

	// Include proposed BUY if it lands in this sector.
	if input.ProposedOrder != nil {
		ord := input.ProposedOrder
		orderSector := ""
		if data.Classifications != nil {
			orderSector = data.Classifications.Sector(ord.Ticker)
		}
		if orderSector == p.Sector {
			proposedInSector = true
			if ord.Side == vo.OrderSideBuy {
				sectorMV = sectorMV.Add(ord.TradeValue())
			} else {
				delta := ord.TradeValue()
				sectorMV = sectorMV.Sub(delta)
				if sectorMV.IsNegative() {
					sectorMV = decimal.Zero
				}
			}
		}
	}

	limit := p.MaxPct.Div(decimal.NewFromInt(100))
	sectorPct := sectorMV.Div(nav)
	sectorPctDisplay := sectorPct.Mul(decimal.NewFromInt(100)).StringFixed(4)
	limitDisplay := p.MaxPct.StringFixed(4)

	metrics := map[string]string{
		"sector_exposure_pct": sectorPctDisplay,
		"sector_market_value": sectorMV.StringFixed(2),
		"nav":                 nav.StringFixed(2),
		"proposed_in_sector":  fmt.Sprintf("%t", proposedInSector),
	}
	references := map[string]string{"sector": p.Sector}
	if input.ProposedOrder != nil {
		references["proposed_ticker"] = input.ProposedOrder.Ticker
	}

	if sectorPct.GreaterThan(limit) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"sector '%s' exposure would be %s%% (limit %s%%)",
				p.Sector, sectorPctDisplay, limitDisplay,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "sector_exposure_pct",
					Actual:     sectorPctDisplay,
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
			"sector '%s' exposure %s%% is within limit %s%%",
			p.Sector, sectorPctDisplay, limitDisplay,
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *SectorExposureRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p SectorParams
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'ratio.sector_exposure' caps exposure to the '%s' sector at %s%% of NAV. "+
			"Verdict: %s. %s",
		p.Sector, p.MaxPct.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// --- helpers ---

func holdingMV(h spi.Holding, prices *spi.MarketPriceSnapshot) decimal.Decimal {
	if prices != nil {
		if price, ok := prices.Prices[h.Ticker]; ok && price.IsPositive() {
			return h.Quantity.Mul(price)
		}
	}
	return h.MarketValue
}

func sectorBlock(msg string) spi.EvalResult {
	return spi.EvalResult{
		Verdict: vo.VerdictBlock,
		Message: msg,
		Evidence: vo.Evidence{
			Metrics: map[string]string{"error": msg},
		},
	}
}
