// Package restriction implements instrument restriction list rules.
// Self-registers blacklist and whitelist evaluators via init().
package restriction

import (
	"context"
	"fmt"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&BlacklistRule{})
}

// BlacklistRule blocks any order for an instrument on the restriction blacklist.
// TypeID: "restriction.blacklist"
// NOT overridable — regulatory/internal compliance requirement.
type BlacklistRule struct{}

func (r *BlacklistRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "restriction.blacklist",
		Version:         "1.0.0",
		Category:        spi.CategoryRestriction,
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
		Description: "Blocks trading in any instrument listed on the restriction blacklist. Not overridable.",
	}
}

func (r *BlacklistRule) ParameterSchema() spi.ParameterSchema {
	// No parameters — the blacklist itself is managed through the restriction port.
	return spi.ParameterSchema{Schema: []byte(`{"type": "object", "properties": {}}`)}
}

func (r *BlacklistRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Restrictions: true,
	}
}

func (r *BlacklistRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		// No order to check — periodic scans are not meaningful for this rule.
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "restriction.blacklist: no proposed order to evaluate",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	ticker := input.ProposedOrder.Ticker

	if data.Restrictions == nil {
		// Fail-closed: if restriction data is unavailable, block the trade.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("restriction list unavailable — blocking '%s' as fail-safe", ticker),
			Evidence: vo.Evidence{
				Metrics:    map[string]string{"error": "restriction_data_unavailable"},
				References: map[string]string{"ticker": ticker},
			},
		}, nil
	}

	if entry, ok := data.Restrictions.Blacklisted[ticker]; ok {
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("'%s' is on the restriction blacklist: %s (source: %s)",
				ticker, entry.Reason, entry.Source),
			Evidence: vo.Evidence{
				Metrics: map[string]string{
					"list_type": entry.ListType,
					"source":    entry.Source,
				},
				References: map[string]string{
					"ticker": ticker,
					"reason": entry.Reason,
				},
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf("'%s' is not on the restriction blacklist", ticker),
		Evidence: vo.Evidence{
			References: map[string]string{"ticker": ticker},
		},
	}, nil
}

func (r *BlacklistRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'restriction.blacklist' checks whether the proposed instrument ('%s') appears on any "+
			"active restriction blacklist. This rule is NOT overridable. Verdict: %s. %s",
		ticker, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
