// Package quantity implements quantity-level pre-trade rules.
// Self-registers sell quantity and min-trading-unit evaluators via init().
package quantity

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&SellQuantityRule{})
}

// SellQuantityRule prevents selling more shares than the portfolio actually holds
// (net of pending sells already in flight). Does NOT allow short selling by default.
// TypeID: "quantity.sell_available"
type SellQuantityRule struct{}

// SellParams controls short-selling allowance.
type SellParams struct {
	// AllowShortSell permits short positions when true. Default false.
	AllowShortSell bool `json:"allow_short_sell"`
}

func (r *SellQuantityRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "quantity.sell_available",
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
		Description: "Prevents selling more shares than the portfolio holds net of pending sells. Short selling is disabled by default.",
	}
}

func (r *SellQuantityRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"properties": {
			"allow_short_sell": {
				"type": "boolean",
				"description": "Allow short selling. Defaults to false."
			}
		}
	}`)}
}

func (r *SellQuantityRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Positions: true,
	}
}

func (r *SellQuantityRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "quantity.sell_available: no proposed order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}
	if input.ProposedOrder.Side != vo.OrderSideSell {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "quantity.sell_available: not a sell order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_buy"}},
		}, nil
	}

	var p SellParams
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}

	// Short selling explicitly permitted — skip check.
	if p.AllowShortSell {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "short selling is enabled for this portfolio — quantity check skipped",
			Evidence: vo.Evidence{Metrics: map[string]string{"allow_short_sell": "true"}},
		}, nil
	}

	if data.Positions == nil {
		// Fail-closed.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: "position data unavailable — blocking sell as fail-safe",
			Evidence: vo.Evidence{Metrics: map[string]string{"error": "positions_unavailable"}},
		}, nil
	}

	ticker := input.ProposedOrder.Ticker
	proposedQty := input.ProposedOrder.Quantity

	// Find current long position.
	heldQty := decimal.Zero
	for _, h := range data.Positions.Holdings {
		if h.Ticker == ticker {
			heldQty = h.Quantity
			break
		}
	}

	// Subtract pending sells already in flight.
	pendingSells := decimal.Zero
	for _, po := range data.Positions.PendingOrders {
		if po.Ticker == ticker && po.Side == string(vo.OrderSideSell) {
			pendingSells = pendingSells.Add(po.Quantity)
		}
	}

	availableQty := heldQty.Sub(pendingSells)
	if availableQty.IsNegative() {
		availableQty = decimal.Zero
	}

	metrics := map[string]string{
		"held_quantity":      heldQty.String(),
		"pending_sells":      pendingSells.String(),
		"available_quantity": availableQty.String(),
		"proposed_quantity":  proposedQty.String(),
	}
	references := map[string]string{"ticker": ticker}

	if proposedQty.GreaterThan(availableQty) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"sell quantity %s exceeds available position %s for '%s' (held: %s, pending sells: %s)",
				proposedQty, availableQty, ticker, heldQty, pendingSells,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "sell_quantity",
					Actual:     proposedQty.String(),
					Limit:      availableQty.String(),
					Operator:   "<=",
					Unit:       "count",
				},
				References: references,
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"sell quantity %s is within available position %s for '%s'",
			proposedQty, availableQty, ticker,
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *SellQuantityRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'quantity.sell_available' ensures the portfolio holds sufficient shares of '%s' to fulfill "+
			"the sell order net of any pending sells. Short selling is not permitted. "+
			"Verdict: %s. %s",
		ticker, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
