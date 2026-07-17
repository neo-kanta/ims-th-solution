// Package valuation implements portfolio-level valuation floor rules.
// Self-registers "valuation.min_nav" via init().
package valuation

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&MinNAVRule{})
}

// MinNAVRule blocks further trading once portfolio NAV falls below a
// configured floor — e.g. a mandate that must wind down once assets drop
// below a regulatory or contractual minimum.
// TypeID: "valuation.min_nav"
type MinNAVRule struct{}

// Params configures the NAV floor.
type Params struct {
	// MinNAV is the minimum required NAV, in the portfolio's base currency.
	MinNAV decimal.Decimal `json:"min_nav"`
}

func (r *MinNAVRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "valuation.min_nav",
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
		Description: "Blocks when portfolio NAV falls below a configured floor.",
	}
}

func (r *MinNAVRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["min_nav"],
		"properties": {
			"min_nav": {
				"type": "number",
				"minimum": 0,
				"description": "Minimum required NAV in the portfolio's base currency"
			}
		}
	}`)}
}

func (r *MinNAVRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{NAV: true}
}

func (r *MinNAVRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p Params
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.MinNAV.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("min_nav must be non-negative, got %s", p.MinNAV)
	}

	if data.NAV == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictBlock,
			Message:  "NAV data unavailable — blocking as fail-safe",
			Evidence: vo.Evidence{Metrics: map[string]string{"error": "nav_data_unavailable"}},
		}, nil
	}

	nav := data.NAV.NAV
	metrics := map[string]string{
		"nav":     nav.StringFixed(2),
		"min_nav": p.MinNAV.StringFixed(2),
	}

	if nav.LessThan(p.MinNAV) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("NAV %s is below the minimum %s", nav.StringFixed(2), p.MinNAV.StringFixed(2)),
			Evidence: vo.Evidence{
				Metrics: metrics,
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "nav",
					Actual:     nav.StringFixed(2),
					Limit:      p.MinNAV.StringFixed(2),
					Operator:   ">=",
					Unit:       "amount",
				},
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict:  vo.VerdictPass,
		Message:  fmt.Sprintf("NAV %s meets the minimum %s", nav.StringFixed(2), p.MinNAV.StringFixed(2)),
		Evidence: vo.Evidence{Metrics: metrics},
	}, nil
}

func (r *MinNAVRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p Params
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'valuation.min_nav' requires NAV to stay at or above %s. Verdict: %s. %s",
		p.MinNAV.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
