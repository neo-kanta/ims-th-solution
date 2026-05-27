// Package amount implements trade amount compliance rules.
package amount

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&MinimumTradeRule{})
}

// MinimumTradeRule enforces a minimum notional trade value.
// TypeID: "amount.minimum_trade"
type MinimumTradeRule struct{}

type Params struct {
	MinAmount decimal.Decimal            `json:"min_amount"`
	Currency  string                     `json:"currency,omitempty"`
	PerSide   map[string]decimal.Decimal `json:"per_side,omitempty"`
}

func (r *MinimumTradeRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "amount.minimum_trade",
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
		Overridable: false,
		Description: "Verifies that the proposed order value is at least the configured minimum trade amount.",
	}
}

func (r *MinimumTradeRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["min_amount"],
		"properties": {
			"min_amount": {
				"type": "number",
				"exclusiveMinimum": 0,
				"description": "Minimum trade value in the order currency unless currency is set."
			},
			"currency": {
				"type": "string",
				"minLength": 3,
				"maxLength": 3,
				"description": "Optional currency this threshold applies to."
			},
			"per_side": {
				"type": "object",
				"additionalProperties": {"type": "number", "exclusiveMinimum": 0},
				"description": "Optional BUY/SELL minimum overrides."
			}
		}
	}`)}
}

func (r *MinimumTradeRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{}
}

func (r *MinimumTradeRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "amount.minimum_trade: no proposed order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	var p Params
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.MinAmount.Sign() <= 0 {
		return spi.EvalResult{}, fmt.Errorf("min_amount must be positive, got %s", p.MinAmount)
	}

	order := input.ProposedOrder
	minAmount := p.MinAmount
	if p.PerSide != nil {
		if sideMin, ok := p.PerSide[string(order.Side)]; ok {
			if sideMin.Sign() <= 0 {
				return spi.EvalResult{}, fmt.Errorf("per_side.%s must be positive, got %s", order.Side, sideMin)
			}
			minAmount = sideMin
		}
	}

	if p.Currency != "" && order.Currency != "" && p.Currency != order.Currency {
		return spi.EvalResult{}, fmt.Errorf("configured currency %s does not match order currency %s", p.Currency, order.Currency)
	}

	tradeValue := order.TradeValue()
	metrics := map[string]string{
		"trade_value": tradeValue.StringFixed(2),
		"min_amount":  minAmount.StringFixed(2),
		"side":        string(order.Side),
	}
	if order.Currency != "" {
		metrics["currency"] = order.Currency
	}
	references := map[string]string{"ticker": order.Ticker}

	if tradeValue.LessThan(minAmount) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"trade value %s is below minimum amount %s",
				tradeValue.StringFixed(2), minAmount.StringFixed(2),
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "trade_value",
					Actual:     tradeValue.StringFixed(2),
					Limit:      minAmount.StringFixed(2),
					Operator:   ">=",
					Unit:       "amount",
				},
				References: references,
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict:  vo.VerdictPass,
		Message:  fmt.Sprintf("trade value %s meets minimum amount %s", tradeValue.StringFixed(2), minAmount.StringFixed(2)),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *MinimumTradeRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p Params
	_ = params.Decode(&p)
	return spi.Explanation{
		PlainText:  fmt.Sprintf("Rule 'amount.minimum_trade' requires order value to be at least %s. Verdict: %s. %s", p.MinAmount.StringFixed(2), result.Verdict, result.Message),
		Structured: result.Evidence,
	}
}
