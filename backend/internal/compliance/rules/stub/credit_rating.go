// Package stub contains Phase 2 placeholder rule implementations that always
// return PASS. They self-register so the engine recognises their TypeIDs and
// can persist check records; operators can bind them to portfolios today and
// the real logic will slot in without any binding-layer changes.
package stub

import (
	"context"
	"fmt"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&CreditRatingRule{})
}

// CreditRatingRule (stub) will enforce minimum credit rating requirements in Phase 2.
// Currently returns PASS with a NOT_IMPLEMENTED marker in the evidence.
// TypeID: "credit_rating.minimum"
type CreditRatingRule struct{}

// CreditRatingParams defines the minimum acceptable rating (Phase 2).
type CreditRatingParams struct {
	// MinRating is the minimum rating string (e.g. "BBB-", "A").
	MinRating string `json:"min_rating"`
	// Agency is the rating agency to use (e.g. "TRIS", "MOODY'S", "S&P").
	Agency string `json:"agency"`
}

func (r *CreditRatingRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "credit_rating.minimum",
		Version:         "0.1.0-stub",
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
		Description: "[STUB — Phase 2] Enforces a minimum credit rating for debt instruments. Not yet active.",
	}
}

func (r *CreditRatingRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["min_rating"],
		"properties": {
			"min_rating": {"type": "string", "description": "Minimum acceptable credit rating"},
			"agency": {"type": "string", "description": "Rating agency (TRIS, MOODY, SP, FITCH)"}
		}
	}`)}
}

func (r *CreditRatingRule) DataDependencies() spi.DataDependencies {
	// Phase 2 will add: CreditRatings: true, Classifications: true
	return spi.DataDependencies{}
}

func (r *CreditRatingRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p CreditRatingParams
	_ = params.Decode(&p)

	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf(
			"credit_rating.minimum: NOT IMPLEMENTED in PoC — passing '%s' by default (min_rating=%s, agency=%s)",
			ticker, p.MinRating, p.Agency,
		),
		Evidence: vo.Evidence{
			Metrics: map[string]string{
				"status":     "NOT_IMPLEMENTED",
				"min_rating": p.MinRating,
				"agency":     p.Agency,
			},
			References: map[string]string{"ticker": ticker},
		},
	}, nil
}

func (r *CreditRatingRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p CreditRatingParams
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'credit_rating.minimum' (STUB) will enforce a minimum credit rating of '%s' from '%s'. "+
			"Implementation deferred to Phase 2. Verdict: %s.",
		p.MinRating, p.Agency, result.Verdict,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
