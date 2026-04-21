package restriction

import (
	"context"
	"fmt"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&WhitelistRule{})
}

// WhitelistRule blocks trades in instruments NOT on the approved list,
// but only when a whitelist is actually configured for the portfolio.
// TypeID: "restriction.whitelist"
type WhitelistRule struct{}

func (r *WhitelistRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "restriction.whitelist",
		Version:         "1.0.0",
		Category:        spi.CategoryRestriction,
		DefaultSeverity: vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Blocks trading in instruments not on the approved (white) list. Only active when a whitelist is configured.",
	}
}

func (r *WhitelistRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{"type": "object", "properties": {}}`)}
}

func (r *WhitelistRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		Restrictions: true,
	}
}

func (r *WhitelistRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "restriction.whitelist: no proposed order to evaluate",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	ticker := input.ProposedOrder.Ticker

	if data.Restrictions == nil {
		// Fail-closed.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("restriction list unavailable — blocking '%s' as fail-safe", ticker),
			Evidence: vo.Evidence{
				Metrics:    map[string]string{"error": "restriction_data_unavailable"},
				References: map[string]string{"ticker": ticker},
			},
		}, nil
	}

	// If no whitelist is configured for this portfolio, skip the check.
	if !data.Restrictions.HasWhitelist {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: "no whitelist configured for this portfolio — check skipped",
			Evidence: vo.Evidence{
				Metrics:    map[string]string{"has_whitelist": "false"},
				References: map[string]string{"ticker": ticker},
			},
		}, nil
	}

	if data.Restrictions.Whitelisted[ticker] {
		return spi.EvalResult{
			Verdict: vo.VerdictPass,
			Message: fmt.Sprintf("'%s' is on the approved whitelist", ticker),
			Evidence: vo.Evidence{
				Metrics:    map[string]string{"whitelisted": "true"},
				References: map[string]string{"ticker": ticker},
			},
		}, nil
	}

	return spi.EvalResult{
		Verdict: vo.VerdictBlock,
		Message: fmt.Sprintf("'%s' is NOT on the approved whitelist for this portfolio", ticker),
		Evidence: vo.Evidence{
			Metrics:    map[string]string{"whitelisted": "false"},
			References: map[string]string{"ticker": ticker},
		},
	}, nil
}

func (r *WhitelistRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'restriction.whitelist' verifies that the proposed instrument ('%s') appears on the "+
			"portfolio's approved instrument list. The rule is inactive when no whitelist is configured. "+
			"Verdict: %s. %s",
		ticker, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
