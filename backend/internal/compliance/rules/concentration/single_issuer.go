// Package concentration implements single-issuer concentration rules.
// Self-registers as "concentration.single_issuer" via init().
package concentration

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&SingleIssuerRule{})
}

// SingleIssuerRule enforces the single-issuer concentration limit as a % of NAV.
// It groups holdings by parent entity (e.g. conglomerate parent) before summing.
// TypeID: "concentration.single_issuer"
type SingleIssuerRule struct{}

// Params are the parameters decoded from the rule instance version.
type Params struct {
	// MaxPct is the maximum allowed exposure per parent entity as a % of NAV (e.g. 10 = 10%).
	MaxPct decimal.Decimal `json:"max_pct"`
	// ExemptGovernment skips the limit for government-issued instruments.
	ExemptGovernment bool `json:"exempt_government"`
}

func (r *SingleIssuerRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "concentration.single_issuer",
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
		Description: "Limits exposure to any single issuer (grouped by parent entity) as a percentage of NAV.",
	}
}

func (r *SingleIssuerRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["max_pct"],
		"properties": {
			"max_pct": {"type": "number", "minimum": 0, "maximum": 100,
				"description": "Maximum single-issuer exposure as % of NAV"},
			"exempt_government": {"type": "boolean",
				"description": "If true, government bonds are not subject to this limit"}
		}
	}`)}
}

func (r *SingleIssuerRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Positions:       true,
		NAV:             true,
		MarketPrices:    true,
		Classifications: true,
	}
}

func (r *SingleIssuerRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p Params
	if err := params.Decode(&p); err != nil {
		return spi.EvalResult{}, fmt.Errorf("decoding params: %w", err)
	}
	if p.MaxPct.IsZero() || p.MaxPct.IsNegative() {
		return spi.EvalResult{}, fmt.Errorf("max_pct must be positive, got %s", p.MaxPct)
	}

	if data.NAV == nil {
		return blockResult("NAV data unavailable — cannot compute concentration"), nil
	}
	nav := data.NAV.NAV
	if nav.IsZero() || nav.IsNegative() {
		return blockResult("NAV is zero or negative — cannot compute concentration"), nil
	}

	// Build parent-entity → market value map from current holdings.
	entityMV := make(map[string]decimal.Decimal)
	if data.Positions != nil {
		for _, h := range data.Positions.Holdings {
			// Skip if exempt government
			if p.ExemptGovernment && data.Classifications != nil &&
				data.Classifications.IsGovernment(h.Ticker) {
				continue
			}
			entity := resolveEntity(h.Ticker, data.Classifications)
			mv := spi.HoldingMarketValue(h, data.MarketPrices)
			entityMV[entity] = entityMV[entity].Add(mv)
		}
	}

	// Apply the proposed order if pre-trade.
	var proposedEntity string
	var proposedValue decimal.Decimal
	if input.ProposedOrder != nil {
		ord := input.ProposedOrder
		if p.ExemptGovernment && data.Classifications != nil &&
			data.Classifications.IsGovernment(ord.Ticker) {
			// exempt — skip addition
		} else {
			proposedEntity = resolveEntity(ord.Ticker, data.Classifications)
			proposedValue = ord.TradeValue()
			if ord.Side == vo.OrderSideBuy {
				entityMV[proposedEntity] = entityMV[proposedEntity].Add(proposedValue)
			} else {
				// SELL reduces concentration — still check other entities
				entityMV[proposedEntity] = entityMV[proposedEntity].Sub(proposedValue)
				if entityMV[proposedEntity].IsNegative() {
					entityMV[proposedEntity] = decimal.Zero
				}
			}
		}
	}

	// Find the worst offender.
	limit := p.MaxPct.Div(decimal.NewFromInt(100))
	var worstEntity string
	worstPct := decimal.Zero

	for entity, mv := range entityMV {
		pct := mv.Div(nav)
		if pct.GreaterThan(worstPct) {
			worstPct = pct
			worstEntity = entity
		}
	}

	worstPctDisplay := worstPct.Mul(decimal.NewFromInt(100)).StringFixed(4)
	limitDisplay := p.MaxPct.StringFixed(4)

	references := map[string]string{}
	if input.ProposedOrder != nil {
		references["proposed_ticker"] = input.ProposedOrder.Ticker
		references["proposed_entity"] = proposedEntity
	}
	if worstEntity != "" {
		references["worst_entity"] = worstEntity
	}

	if worstPct.GreaterThan(limit) {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf(
				"single-issuer concentration for '%s' would be %s%% (limit %s%%)",
				worstEntity, worstPctDisplay, limitDisplay,
			),
			Evidence: vo.Evidence{
				Metrics: map[string]string{
					"proposed_concentration_pct": worstPctDisplay,
					"nav":                        nav.StringFixed(2),
					"entity_market_value":        entityMV[worstEntity].StringFixed(2),
				},
				ThresholdBreached: &vo.ThresholdBreach{
					MetricName: "single_issuer_concentration_pct",
					Actual:     worstPctDisplay,
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
			"single-issuer concentration check passed — worst entity '%s' at %s%% (limit %s%%)",
			worstEntity, worstPctDisplay, limitDisplay,
		),
		Evidence: vo.Evidence{
			Metrics: map[string]string{
				"highest_concentration_pct": worstPctDisplay,
				"nav":                       nav.StringFixed(2),
			},
			References: references,
		},
	}, nil
}

func (r *SingleIssuerRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p Params
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'concentration.single_issuer' caps any single issuer (grouped by parent entity) at %s%% of NAV. "+
			"Verdict: %s. %s",
		p.MaxPct.StringFixed(2), result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// --- helpers ---

func resolveEntity(ticker string, cls *spi.ClassificationSnapshot) string {
	if cls == nil {
		return ticker
	}
	return cls.ParentEntity(ticker)
}

func blockResult(msg string) spi.EvalResult {
	return spi.EvalResult{
		Verdict: vo.VerdictBlock,
		Message: msg,
		Evidence: vo.Evidence{
			Metrics: map[string]string{"error": msg},
		},
	}
}
