package quantity

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&MinTradingUnitRule{})
}

// MinTradingUnitRule verifies the proposed quantity is a multiple of the
// instrument's board lot (trading unit). The Thai stock exchange uses 100-share
// lots for most equities. Odd-lot orders are routed separately; this rule
// enforces that board-market orders land on valid lot boundaries.
// TypeID: "quantity.min_trading_unit"
type MinTradingUnitRule struct{}

// MTUParams defines the default lot size and per-ticker overrides.
type MTUParams struct {
	DefaultLotSize int64            `json:"default_lot_size"`
	Overrides      map[string]int64 `json:"overrides,omitempty"`
}

func (r *MinTradingUnitRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "quantity.min_trading_unit",
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
		Description: "Verifies that the order quantity is a multiple of the instrument's board lot size.",
	}
}

func (r *MinTradingUnitRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["default_lot_size"],
		"properties": {
			"default_lot_size": {
				"type": "integer",
				"minimum": 1,
				"description": "Default board lot size (e.g. 100 for SET equities)"
			},
			"overrides": {
				"type": "object",
				"additionalProperties": {"type": "integer", "minimum": 1},
				"description": "Per-ticker lot size overrides"
			}
		}
	}`)}
}

func (r *MinTradingUnitRule) DataDependencies() spi.DataDependencies {
	// No external data fetch needed — lot size comes from params.
	return spi.DataDependencies{}
}

func (r *MinTradingUnitRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "quantity.min_trading_unit: no proposed order",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	var p MTUParams
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.DefaultLotSize <= 0 {
		return spi.EvalResult{}, fmt.Errorf("default_lot_size must be >= 1, got %d", p.DefaultLotSize)
	}

	ticker := input.ProposedOrder.Ticker
	proposedQty := input.ProposedOrder.Quantity

	lotSize := p.DefaultLotSize
	if override, ok := p.Overrides[ticker]; ok && override > 0 {
		lotSize = override
	}

	lotDecimal := decimal.NewFromInt(lotSize)
	remainder := proposedQty.Mod(lotDecimal)

	metrics := map[string]string{
		"proposed_quantity": proposedQty.String(),
		"lot_size":          lotDecimal.String(),
		"remainder":         remainder.String(),
	}
	references := map[string]string{"ticker": ticker}

	if !remainder.IsZero() {
		nearestValidLots := proposedQty.Div(lotDecimal).Floor().Mul(lotDecimal)
		metrics["nearest_valid_quantity"] = nearestValidLots.String()
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"quantity %s for '%s' is not a multiple of board lot size %d (remainder: %s, nearest valid: %s)",
				proposedQty, ticker, lotSize, remainder, nearestValidLots,
			),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "quantity_modulo_lot_size",
					Actual:     remainder.String(),
					Limit:      "0",
					Operator:   "==",
					Unit:       "lots",
				},
				References: references,
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"quantity %s for '%s' is a valid multiple of lot size %d (%s lots)",
			proposedQty, ticker, lotSize,
			proposedQty.Div(lotDecimal).StringFixed(0),
		),
		Evidence: vo.Evidence{Metrics: metrics, References: references},
	}, nil
}

func (r *MinTradingUnitRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p MTUParams
	_ = params.Decode(&p)
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'quantity.min_trading_unit' verifies that the order quantity for '%s' is a multiple "+
			"of the board lot size (default: %d). Verdict: %s. %s",
		ticker, p.DefaultLotSize, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
