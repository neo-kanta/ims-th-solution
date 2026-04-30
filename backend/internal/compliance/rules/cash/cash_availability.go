// Package cash implements cash availability pre-trade checks.
// Self-registers as "cash.availability" via init().
package cash

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&AvailabilityRule{})
}

// AvailabilityRule ensures a BUY order does not exceed available cash minus any
// mandatory buffer. Also warns when available cash would drop below a soft floor.
// TypeID: "cash.availability"
type AvailabilityRule struct{}

// Params are the parameters for this rule instance.
type Params struct {
	// MinCashBufferPct is the minimum cash to retain after the trade, expressed
	// as a percentage of NAV (e.g. 5 = keep at least 5% of NAV in cash).
	// Zero means no buffer requirement beyond covering the trade cost.
	MinCashBufferPct decimal.Decimal `json:"min_cash_buffer_pct"`
}

func (r *AvailabilityRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "cash.availability",
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
		Description: "Checks that sufficient cash is available to settle a buy order, optionally enforcing a minimum cash buffer as % of NAV.",
	}
}

func (r *AvailabilityRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"properties": {
			"min_cash_buffer_pct": {
				"type": "number",
				"minimum": 0,
				"maximum": 100,
				"description": "Minimum cash to retain after the trade as % of NAV. 0 = no buffer."
			}
		}
	}`)}
}

func (r *AvailabilityRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		NAV: true,
	}
}

func (r *AvailabilityRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	// Only meaningful for BUY orders.
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "cash.availability: no proposed order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}
	if input.ProposedOrder.Side != vo.OrderSideBuy {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "cash.availability: sell orders do not require cash",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_sell"}},
		}, nil
	}

	var p Params
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}

	if data.NAV == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictBlock,
			Message:  "NAV / cash data unavailable — blocking as fail-safe",
			Evidence: vo.Evidence{Metrics: map[string]string{"error": "nav_data_unavailable"}},
		}, nil
	}

	nav := data.NAV.NAV
	availableCash := data.NAV.CashBalance.Sub(data.NAV.ReservedCash)
	tradeValue := input.ProposedOrder.TradeValue()
	remainingCash := availableCash.Sub(tradeValue)

	metrics := map[string]string{
		"available_cash":      availableCash.StringFixed(2),
		"reserved_cash":       data.NAV.ReservedCash.StringFixed(2),
		"trade_value":         tradeValue.StringFixed(2),
		"cash_after_trade":    remainingCash.StringFixed(2),
		"nav":                 nav.StringFixed(2),
		"min_cash_buffer_pct": p.MinCashBufferPct.StringFixed(4),
	}
	references := map[string]string{
		"ticker": input.ProposedOrder.Ticker,
	}

	// Hard check: available cash must cover the trade.
	if remainingCash.IsNegative() {
		shortfall := remainingCash.Neg()
		metrics["shortfall"] = shortfall.StringFixed(2)
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"insufficient cash: trade value %s exceeds available cash %s (shortfall %s)",
				tradeValue.StringFixed(2), availableCash.StringFixed(2), shortfall.StringFixed(2),
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "available_cash",
					Actual:     availableCash.StringFixed(2),
					Limit:      tradeValue.StringFixed(2),
					Operator:   ">=",
					Unit:       "amount",
				},
				References: references,
			},
		}, nil
	}

	// Soft check: buffer requirement.
	if !p.MinCashBufferPct.IsZero() && !nav.IsZero() {
		bufferRequired := nav.Mul(p.MinCashBufferPct).Div(decimal.NewFromInt(100))
		metrics["buffer_required"] = bufferRequired.StringFixed(2)
		if remainingCash.LessThan(bufferRequired) {
			return spi.EvalResult{
				Verdict: vo.VerdictWarn,
				Message: fmt.Sprintf(
					"cash after trade (%s) would fall below minimum buffer of %s%% of NAV (%s)",
					remainingCash.StringFixed(2), p.MinCashBufferPct.StringFixed(2), bufferRequired.StringFixed(2),
				),
				Evidence: vo.Evidence{
					Metrics: metrics,
					ThresholdBreached: &vo.ThresholdBreach{
						MetricName: "cash_after_trade_pct_of_nav",
						Actual:     remainingCash.Div(nav).Mul(decimal.NewFromInt(100)).StringFixed(4),
						Limit:      p.MinCashBufferPct.StringFixed(4),
						Operator:   ">=",
						Unit:       "percent",
					},
					References: references,
				},
			}, nil
		}
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"sufficient cash available: trade value %s, cash remaining after trade %s",
			tradeValue.StringFixed(2), remainingCash.StringFixed(2),
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *AvailabilityRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p Params
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'cash.availability' checks that available cash covers the BUY order value and retains "+
			"a minimum buffer of %s%% of NAV. Verdict: %s. %s",
		p.MinCashBufferPct.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
