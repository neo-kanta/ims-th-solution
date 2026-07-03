package stub

import (
	"context"
	"fmt"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&ThaiSECRule{})
}

// ThaiSECRule (stub) will enforce Thai SEC / BOT investment regulations in Phase 2.
// This includes foreign ownership limits, specific instrument restrictions for
// provident funds, and any ORPP/CAAT mandate constraints.
// TypeID: "regulatory.thai_sec"
type ThaiSECRule struct{}

// ThaiSECParams will hold regulatory-specific parameters in Phase 2.
type ThaiSECParams struct {
	// Regulation is the specific SEC regulation code being checked (e.g. "SRO01", "PVD_LIMIT").
	Regulation string `json:"regulation"`
	// MaxForeignOwnershipPct caps foreign ownership of a Thai-listed stock.
	MaxForeignOwnershipPct string `json:"max_foreign_ownership_pct,omitempty"`
}

func (r *ThaiSECRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "regulatory.thai_sec",
		Version:         "0.1.0-stub",
		Category:        spi.CategoryRegulatory,
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
		Overridable: false,
		Description: "[STUB — Phase 2] Thai SEC / BOT regulatory compliance checks. Not overridable.",
	}
}

func (r *ThaiSECRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["regulation"],
		"properties": {
			"regulation": {"type": "string", "description": "SEC regulation code"},
			"max_foreign_ownership_pct": {"type": "string", "description": "Foreign ownership cap (%)"}
		}
	}`)}
}

func (r *ThaiSECRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{}
}

func (r *ThaiSECRule) Evaluate(
	ctx context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	var p ThaiSECParams
	_ = params.Decode(&p)

	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}

	// Return WARN (not PASS) so the IRG pipeline surfaces this rule as unimplemented
	// rather than silently allowing the trade. A PASS from an unimplemented stub is a
	// false negative; WARN flags the gap without hard-blocking trading. When this rule
	// is fully implemented in Phase 2, replace the body of this method.
	return spi.EvalResult{
		Verdict: vo.VerdictWarn,
		Message: fmt.Sprintf(
			"regulatory.thai_sec: STUB — NOT_CONFIGURED; emitting WARN to flag non-enforcement (ticker=%s, regulation=%s)",
			ticker, p.Regulation,
		),
		Evidence: vo.Evidence{
			Metrics: map[string]string{
				"status":     "NOT_CONFIGURED",
				"stub":       "true",
				"regulation": p.Regulation,
			},
			References: map[string]string{"ticker": ticker},
		},
	}, nil
}

func (r *ThaiSECRule) Explain(input spi.CheckInput, result spi.EvalResult, params spi.ParameterSet) spi.Explanation {
	var p ThaiSECParams
	_ = params.Decode(&p)
	text := fmt.Sprintf(
		"Rule 'regulatory.thai_sec' (STUB) will enforce Thai SEC regulation '%s'. "+
			"Implementation deferred to Phase 2. This rule is NOT overridable. Verdict: %s.",
		p.Regulation, result.Verdict,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}
