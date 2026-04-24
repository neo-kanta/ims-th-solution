package restriction

import (
	"context"
	"fmt"
	"strings"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&ListEnforcementRule{})
}

// ListEnforcementRule is the unified restriction-list enforcer. Unlike the
// legacy blacklist/whitelist rules it understands all four list types in a
// single evaluator configured per binding:
//
//   - BLACKLIST — always BLOCK.
//   - WHITELIST — BLOCK if HasWhitelist and ticker not in list.
//   - ALERT     — WARN (overridable). Entry source surfaced in evidence.
//   - DISPOSAL  — BUY is BLOCK; SELL is PASS (wind-down).
//
// TypeID: "restriction.list_enforcement"
// Overridable: true (severity cap still applies per-binding).
type ListEnforcementRule struct{}

// Parameter names (exported for tests and admin UI).
const (
	paramEnforcedTypes = "enforced_types" // []string subset of {"BLACKLIST","WHITELIST","ALERT","DISPOSAL"}
)

func (r *ListEnforcementRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "restriction.list_enforcement",
		Version:         "1.0.0",
		Category:        spi.CategoryRestriction,
		DefaultSeverity: vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
			vo.TimingPostTrade,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopeGlobal,
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Unified restriction list enforcer: supports BLACKLIST, WHITELIST, ALERT, DISPOSAL.",
	}
}

func (r *ListEnforcementRule) ParameterSchema() spi.ParameterSchema {
	// enforced_types is an optional array of list-type names. When absent or
	// empty the rule enforces all four list types.
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"properties": {
			"enforced_types": {
				"type": "array",
				"items": {
					"type": "string",
					"enum": ["BLACKLIST","WHITELIST","ALERT","DISPOSAL"]
				},
				"uniqueItems": true
			}
		}
	}`)}
}

func (r *ListEnforcementRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{Restrictions: true}
}

func (r *ListEnforcementRule) Evaluate(
	_ context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		// Post-trade / periodic sweeps: nothing to evaluate at order-scope.
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "restriction.list_enforcement: no proposed order to evaluate",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	ticker := input.ProposedOrder.Ticker
	side := input.ProposedOrder.Side

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

	enforced := parseEnforcedTypes(params)

	// BLACKLIST — hard block.
	if enforced[spi.RestrictionListTypeBlacklist] {
		if entry, ok := data.Restrictions.Blacklisted[ticker]; ok {
			return blockResult(ticker, entry, "blacklist"), nil
		}
	}

	// DISPOSAL — BUY blocked, SELL allowed.
	if enforced[spi.RestrictionListTypeDisposal] && data.Restrictions.Disposal != nil {
		if entry, ok := data.Restrictions.Disposal[ticker]; ok {
			if side == vo.OrderSideSell {
				return spi.EvalResult{
					Verdict: vo.VerdictPass,
					Message: fmt.Sprintf("'%s' is on the disposal list; SELL permitted for wind-down (%s)", ticker, entry.Reason),
					Evidence: vo.Evidence{
						Metrics:    map[string]string{"list_type": string(spi.RestrictionListTypeDisposal), "side": string(side)},
						References: map[string]string{"ticker": ticker, "reason": entry.Reason, "source": entry.Source},
					},
				}, nil
			}
			return spi.EvalResult{
				Verdict: vo.VerdictBlock,
				Message: fmt.Sprintf("'%s' is on the disposal list; BUY blocked during wind-down: %s (source: %s)",
					ticker, entry.Reason, entry.Source),
				Evidence: vo.Evidence{
					Metrics:    map[string]string{"list_type": string(spi.RestrictionListTypeDisposal), "side": string(side)},
					References: map[string]string{"ticker": ticker, "reason": entry.Reason, "source": entry.Source},
				},
			}, nil
		}
	}

	// WHITELIST — BLOCK when an active whitelist exists and the ticker is absent.
	if enforced[spi.RestrictionListTypeWhitelist] && data.Restrictions.HasWhitelist {
		if _, ok := data.Restrictions.Whitelisted[ticker]; !ok {
			return spi.EvalResult{
				Verdict: vo.VerdictBlock,
				Message: fmt.Sprintf("'%s' is not on the active whitelist", ticker),
				Evidence: vo.Evidence{
					Metrics:    map[string]string{"list_type": string(spi.RestrictionListTypeWhitelist), "has_whitelist": "true"},
					References: map[string]string{"ticker": ticker},
				},
			}, nil
		}
	}

	// ALERT — WARN-level breach (legacy GrayListed is merged for backwards compat).
	if enforced[spi.RestrictionListTypeAlert] {
		if entry, ok := lookupAlert(data.Restrictions, ticker); ok {
			return spi.EvalResult{
				Verdict: vo.VerdictWarn,
				Message: fmt.Sprintf("'%s' is on the alert list: %s (source: %s)", ticker, entry.Reason, entry.Source),
				Evidence: vo.Evidence{
					Metrics:    map[string]string{"list_type": string(spi.RestrictionListTypeAlert), "source": entry.Source},
					References: map[string]string{"ticker": ticker, "reason": entry.Reason},
				},
			}, nil
		}
	}

	return spi.EvalResult{
		Verdict: vo.VerdictPass,
		Message: fmt.Sprintf("'%s' passes all enforced restriction lists (%s)", ticker, joinTypes(enforced)),
		Evidence: vo.Evidence{
			Metrics:    map[string]string{"enforced": joinTypes(enforced)},
			References: map[string]string{"ticker": ticker},
		},
	}, nil
}

func (r *ListEnforcementRule) Explain(
	input spi.CheckInput, result spi.EvalResult, _ spi.ParameterSet,
) spi.Explanation {
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'restriction.list_enforcement' enforces BLACKLIST / WHITELIST / ALERT / DISPOSAL "+
			"list types for instrument '%s'. Verdict: %s. %s",
		ticker, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

type listEnforcementParams struct {
	EnforcedTypes []string `json:"enforced_types"`
}

func parseEnforcedTypes(params spi.ParameterSet) map[spi.RestrictionListType]bool {
	var p listEnforcementParams
	_ = params.Decode(&p) // tolerant: empty params → enforce all

	// Empty / unset → enforce all four types (most conservative default).
	if len(p.EnforcedTypes) == 0 {
		return map[spi.RestrictionListType]bool{
			spi.RestrictionListTypeBlacklist: true,
			spi.RestrictionListTypeWhitelist: true,
			spi.RestrictionListTypeAlert:     true,
			spi.RestrictionListTypeDisposal:  true,
		}
	}
	out := make(map[spi.RestrictionListType]bool, len(p.EnforcedTypes))
	for _, v := range p.EnforcedTypes {
		t := spi.RestrictionListType(strings.ToUpper(strings.TrimSpace(v)))
		if t.IsValid() {
			if t == spi.RestrictionListTypeGray {
				t = spi.RestrictionListTypeAlert
			}
			out[t] = true
		}
	}
	return out
}

func lookupAlert(r *spi.RestrictionSnapshot, ticker string) (spi.RestrictionEntry, bool) {
	if r == nil {
		return spi.RestrictionEntry{}, false
	}
	if entry, ok := r.Alerted[ticker]; ok {
		return entry, true
	}
	// Legacy GrayListed maps to ALERT.
	if entry, ok := r.GrayListed[ticker]; ok {
		return entry, true
	}
	return spi.RestrictionEntry{}, false
}

func blockResult(ticker string, entry spi.RestrictionEntry, kind string) spi.EvalResult {
	return spi.EvalResult{
		Verdict: vo.VerdictBlock,
		Message: fmt.Sprintf("'%s' is on the %s: %s (source: %s)", ticker, kind, entry.Reason, entry.Source),
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
	}
}

func joinTypes(enforced map[spi.RestrictionListType]bool) string {
	keys := make([]string, 0, len(enforced))
	for k := range enforced {
		keys = append(keys, string(k))
	}
	// Stable order for determinism.
	sortedStrings(keys)
	return strings.Join(keys, ",")
}

// sortedStrings sorts in-place (tiny local helper to avoid sort import sprawl).
func sortedStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
