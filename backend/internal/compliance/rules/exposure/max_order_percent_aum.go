// Package exposure implements per-order exposure limits relative to
// portfolio-level aggregates (AUM/NAV). Self-registers
// "exposure.max_order_percent_aum" via init().
package exposure

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&MaxOrderPercentAUMRule{})
}

// MaxOrderPercentAUMRule caps a single proposed order's trade value as a % of
// portfolio AUM. AUM is read from the same NAV snapshot every other rule in
// this module uses (investment__valuation_snapshots.aum) — Portfolio
// Compliance V2 treats AUM and NAV as the same figure at the portfolio level.
// TypeID: "exposure.max_order_percent_aum"
type MaxOrderPercentAUMRule struct{}

// Params configures the order-size ceiling.
type Params struct {
	// MaxPercentAUM is the maximum single-order trade value as % of AUM
	// (e.g. 10 = an order may not exceed 10% of AUM).
	MaxPercentAUM decimal.Decimal `json:"max_percent_aum"`
}

func (r *MaxOrderPercentAUMRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "exposure.max_order_percent_aum",
		Version:         "1.0.0",
		Category:        spi.CategoryMandate,
		DefaultSeverity: vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopeGlobal,
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Caps a single proposed order's trade value as a percentage of portfolio AUM.",
	}
}

func (r *MaxOrderPercentAUMRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["max_percent_aum"],
		"properties": {
			"max_percent_aum": {
				"type": "number",
				"minimum": 0,
				"maximum": 100,
				"description": "Maximum single-order trade value as % of AUM"
			}
		}
	}`)}
}

func (r *MaxOrderPercentAUMRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{NAV: true}
}

func (r *MaxOrderPercentAUMRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "exposure.max_order_percent_aum: no proposed order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	var p Params
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.MaxPercentAUM.IsZero() || p.MaxPercentAUM.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("max_percent_aum must be positive, got %s", p.MaxPercentAUM)
	}

	if data.NAV == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictBlock,
			Message:  "AUM data unavailable — blocking as fail-safe",
			Evidence: vo.Evidence{Metrics: map[string]string{"error": "aum_data_unavailable"}},
		}, nil
	}

	aum := data.NAV.NAV
	if aum.IsZero() || aum.IsNegative() {
		return spi.EvalResult{
			Verdict:  vo.VerdictBlock,
			Message:  "AUM is zero or negative — cannot evaluate order size limit",
			Evidence: vo.Evidence{Metrics: map[string]string{"error": "aum_zero_or_negative"}},
		}, nil
	}

	tradeValue := input.ProposedOrder.TradeValue()
	orderPct := tradeValue.Div(aum).Mul(decimal.NewFromInt(100))
	orderPctDisplay := orderPct.StringFixed(4)
	limitDisplay := p.MaxPercentAUM.StringFixed(4)

	metrics := map[string]string{
		"trade_value":      tradeValue.StringFixed(2),
		"aum":              aum.StringFixed(2),
		"order_pct_of_aum": orderPctDisplay,
		"max_percent_aum":  limitDisplay,
	}
	references := map[string]string{"ticker": input.ProposedOrder.Ticker}

	if orderPct.GreaterThan(p.MaxPercentAUM) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"order value %s is %s%% of AUM %s (limit %s%%)",
				tradeValue.StringFixed(2), orderPctDisplay, aum.StringFixed(2), limitDisplay,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "order_pct_of_aum",
					Actual:     orderPctDisplay,
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
			"order value %s (%s%% of AUM) is within limit %s%%",
			tradeValue.StringFixed(2), orderPctDisplay, limitDisplay,
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *MaxOrderPercentAUMRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p Params
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'exposure.max_order_percent_aum' caps a single order at %s%% of AUM. Verdict: %s. %s",
		p.MaxPercentAUM.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
