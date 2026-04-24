// Package credit implements credit-quality compliance rules.
//
// credit.min_rating enforces a minimum acceptable issuer credit rating. It
// understands Thai (TRIS / Fitch Thailand) and international agency scales,
// tolerates multi-agency "split" ratings via a configurable policy, and can
// exempt government-issued instruments.
//
// Self-registers via init() as "credit.min_rating".
package credit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func init() {
	spi.Register(&MinRatingRule{})
}

// MinRatingRule is the pre/post-trade rule that blocks trades into issuers
// whose credit rating is below the configured threshold.
//
// TypeID: "credit.min_rating"
// Severity (default): BLOCK — overridable per binding.
type MinRatingRule struct{}

// Exposed parameter names (used by admin UI, tests and JSON schema).
const (
	paramMinRating         = "min_rating"
	paramRatingAgencyPrio  = "rating_agency_priority"
	paramSplitRatingPolicy = "split_rating_policy"
	paramGovernmentExempt  = "government_exempt"
)

// SplitRatingPolicy enumerates how multi-agency ratings are reconciled.
type SplitRatingPolicy string

const (
	// SplitPolicyWorst selects the lowest (most conservative) rating across
	// all agencies that rated the issuer. This is the regulator-friendly
	// default when the parameter is omitted.
	SplitPolicyWorst SplitRatingPolicy = "worst"
	// SplitPolicyBest selects the highest rating, typically used when a
	// mandate explicitly allows "best-of" treatment.
	SplitPolicyBest SplitRatingPolicy = "best"
)

// Error code surfaced in Evidence when no rating is found for the issuer.
const evidenceCodeRatingUnavailable = "RATING_UNAVAILABLE"

// ─────────────────────────────────────────────────────────────────────────────
// Rule plumbing
// ─────────────────────────────────────────────────────────────────────────────

func (r *MinRatingRule) Metadata() spi.RuleMetadata {
	return spi.RuleMetadata{
		TypeID:          "credit.min_rating",
		Version:         "1.0.0",
		Category:        spi.CategoryMandate,
		DefaultSeverity: vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{
			vo.TimingPreTrade,
			vo.TimingPostTrade,
			vo.TimingPeriodic,
		},
		SupportedScopes: []vo.ScopeType{
			vo.ScopeGlobal,
			vo.ScopePortfolio,
			vo.ScopeContract,
		},
		Overridable: true,
		Description: "Enforces a minimum issuer credit rating. Supports agency priority ordering, split-rating policy, and government-issuer exemption.",
	}
}

func (r *MinRatingRule) ParameterSchema() spi.ParameterSchema {
	return spi.ParameterSchema{Schema: []byte(`{
		"type": "object",
		"required": ["min_rating"],
		"properties": {
			"min_rating": {
				"type": "string",
				"description": "Minimum acceptable rating in TRIS/S&P style notation (e.g. 'BBB-')."
			},
			"rating_agency_priority": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Ordered list of agencies. When the issuer has a rating from the first agency it wins; otherwise fall through to the next, then to split-rating policy."
			},
			"split_rating_policy": {
				"type": "string",
				"enum": ["worst", "best"],
				"default": "worst",
				"description": "When no priority agency matches and more than one rating exists, pick the worst (conservative) or best rating."
			},
			"government_exempt": {
				"type": "boolean",
				"default": true,
				"description": "Skip the rating check for government-issued instruments (IsGovernment classification)."
			}
		}
	}`)}
}

func (r *MinRatingRule) DataDependencies() spi.DataDependencies {
	return spi.DataDependencies{
		CreditRatings:   true,
		Classifications: true,
	}
}

// Evaluate implements the RuleEvaluator contract.
func (r *MinRatingRule) Evaluate(
	_ context.Context,
	input spi.CheckInput,
	data spi.DataBundle,
	params spi.ParameterSet,
) (spi.EvalResult, error) {
	if input.ProposedOrder == nil {
		return spi.EvalResult{
			Verdict:  vo.VerdictPass,
			Message:  "credit.min_rating: no proposed order to evaluate",
			Evidence: vo.Evidence{Metrics: map[string]string{"check": "skipped_no_order"}},
		}, nil
	}

	p, perr := parseMinRatingParams(params)
	if perr != nil {
		// Misconfigured rule → fail-closed so operators see the misconfig.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("credit.min_rating: invalid parameters (%s)", perr.Error()),
			Evidence: vo.Evidence{
				Metrics:    map[string]string{"error": "INVALID_PARAMETERS"},
				References: map[string]string{"detail": perr.Error()},
			},
		}, nil
	}

	ticker := input.ProposedOrder.Ticker

	// Government exemption (default-on) — check classification first so we can
	// skip issuers that the mandate explicitly carves out.
	if p.GovernmentExempt {
		if data.Classifications != nil && data.Classifications.IsGovernment(ticker) {
			return spi.EvalResult{
				Verdict: vo.VerdictPass,
				Message: fmt.Sprintf("'%s' is government-issued; credit.min_rating is exempt.", ticker),
				Evidence: vo.Evidence{
					Metrics:    map[string]string{"exempt": "government"},
					References: map[string]string{"ticker": ticker, "min_rating": p.MinRating.Code},
				},
			}, nil
		}
	}

	// Resolve issuer key. We prefer the direct issuer (legal entity) over the
	// ticker so issuers with multiple listed instruments share one rating.
	issuer := ticker
	if data.Classifications != nil {
		issuer = data.Classifications.DirectIssuer(ticker)
	}

	ratings := issuerRatings(data.CreditRatings, issuer, ticker)
	if len(ratings) == 0 {
		// Missing rating → BLOCK (fail-closed) with a specific evidence code so
		// the UI can render a meaningful error.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("'%s': no credit rating available for issuer '%s'.", ticker, issuer),
			Evidence: vo.Evidence{
				Metrics: map[string]string{
					"error":      evidenceCodeRatingUnavailable,
					"min_rating": p.MinRating.Code,
				},
				References: map[string]string{
					"ticker":   ticker,
					"issuer":   issuer,
					"agencies": "none",
				},
			},
		}, nil
	}

	selected, agenciesConsidered := selectRating(ratings, p.AgencyPriority, p.SplitPolicy)
	if !selected.Ordinal.Valid {
		// Rating present but unrecognised notation — fail-closed rather than
		// silently passing something that may be junk.
		return spi.EvalResult{
			Verdict: vo.VerdictBlock,
			Message: fmt.Sprintf("'%s': issuer '%s' has rating '%s' in unrecognised notation.", ticker, issuer, selected.Rating),
			Evidence: vo.Evidence{
				Metrics: map[string]string{
					"error":      "UNRECOGNISED_RATING",
					"rating":     selected.Rating,
					"agency":     selected.Agency,
					"min_rating": p.MinRating.Code,
					"policy":     string(p.SplitPolicy),
				},
				References: map[string]string{"ticker": ticker, "issuer": issuer},
			},
		}, nil
	}

	meets := selected.Ordinal.Value >= p.MinRating.Value
	verdict := vo.VerdictPass
	message := fmt.Sprintf("'%s' meets credit.min_rating: %s/%s ≥ %s",
		ticker, selected.Rating, selected.Agency, p.MinRating.Code)
	if !meets {
		verdict = vo.VerdictBlock
		message = fmt.Sprintf("'%s' fails credit.min_rating: %s/%s < %s",
			ticker, selected.Rating, selected.Agency, p.MinRating.Code)
	}

	return spi.EvalResult{
		Verdict: verdict,
		Message: message,
		Evidence: vo.Evidence{
			Metrics: map[string]string{
				"min_rating":          p.MinRating.Code,
				"selected_rating":     selected.Rating,
				"selected_agency":     selected.Agency,
				"policy":              string(p.SplitPolicy),
				"agencies_considered": agenciesConsidered,
				"government_exempt":   fmt.Sprintf("%v", p.GovernmentExempt),
			},
			References: map[string]string{"ticker": ticker, "issuer": issuer},
		},
	}, nil
}

func (r *MinRatingRule) Explain(
	input spi.CheckInput, result spi.EvalResult, _ spi.ParameterSet,
) spi.Explanation {
	ticker := ""
	if input.ProposedOrder != nil {
		ticker = input.ProposedOrder.Ticker
	}
	text := fmt.Sprintf(
		"Rule 'credit.min_rating' checks issuer-level rating for '%s' against the configured floor. "+
			"Verdict: %s. %s",
		ticker, result.Verdict, result.Message,
	)
	return spi.Explanation{PlainText: text, Structured: result.Evidence}
}

// ─────────────────────────────────────────────────────────────────────────────
// Parameter parsing
// ─────────────────────────────────────────────────────────────────────────────

type minRatingParamsRaw struct {
	MinRating            string   `json:"min_rating"`
	RatingAgencyPriority []string `json:"rating_agency_priority"`
	SplitRatingPolicy    string   `json:"split_rating_policy"`
	GovernmentExempt     *bool    `json:"government_exempt"`
}

type minRatingParams struct {
	MinRating        ratingOrdinal
	AgencyPriority   []string // normalised upper-case
	SplitPolicy      SplitRatingPolicy
	GovernmentExempt bool
}

func parseMinRatingParams(params spi.ParameterSet) (minRatingParams, error) {
	var raw minRatingParamsRaw
	_ = params.Decode(&raw) // tolerant; missing min_rating caught below

	if strings.TrimSpace(raw.MinRating) == "" {
		return minRatingParams{}, fmt.Errorf("%s is required", paramMinRating)
	}

	ord, ok := ParseRating(raw.MinRating)
	if !ok {
		return minRatingParams{}, fmt.Errorf("unrecognised %s=%q", paramMinRating, raw.MinRating)
	}

	policy := SplitPolicyWorst // conservative default
	switch strings.ToLower(strings.TrimSpace(raw.SplitRatingPolicy)) {
	case "", "worst":
		policy = SplitPolicyWorst
	case "best":
		policy = SplitPolicyBest
	default:
		return minRatingParams{}, fmt.Errorf("%s must be 'worst' or 'best', got %q",
			paramSplitRatingPolicy, raw.SplitRatingPolicy)
	}

	// Government exemption defaults to TRUE (regulatory convention — sovereigns
	// are treated as risk-free). Explicitly set to false to enforce ratings on
	// government bonds as well.
	govExempt := true
	if raw.GovernmentExempt != nil {
		govExempt = *raw.GovernmentExempt
	}

	prio := make([]string, 0, len(raw.RatingAgencyPriority))
	for _, a := range raw.RatingAgencyPriority {
		if clean := strings.ToUpper(strings.TrimSpace(a)); clean != "" {
			prio = append(prio, clean)
		}
	}

	return minRatingParams{
		MinRating:        ord,
		AgencyPriority:   prio,
		SplitPolicy:      policy,
		GovernmentExempt: govExempt,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Rating selection
// ─────────────────────────────────────────────────────────────────────────────

// issuerRatings resolves the slice of agency ratings to consider. It looks up
// by the direct issuer first and falls back to the ticker so adapters that
// key by ticker (not issuer code) still work.
func issuerRatings(snap *spi.CreditRatingSnapshot, issuer, ticker string) []spi.CreditRating {
	if snap == nil {
		return nil
	}
	if r := snap.ForIssuer(issuer); len(r) > 0 {
		return r
	}
	if issuer != ticker {
		return snap.ForIssuer(ticker)
	}
	return nil
}

// selectedRating is the chosen rating plus its ordinal value.
type selectedRating struct {
	Rating  string
	Agency  string
	Ordinal ratingOrdinal
}

// selectRating applies agency_priority first; if no priority agency is
// present, falls back to split_rating_policy across all parsed ratings.
// Returns the chosen rating and a comma-separated summary of the agencies
// considered (stable order for audit reproducibility).
func selectRating(
	ratings []spi.CreditRating,
	priority []string,
	policy SplitRatingPolicy,
) (selectedRating, string) {
	// Index parsed ratings by agency (upper-cased).
	parsed := make(map[string]selectedRating, len(ratings))
	agencies := make([]string, 0, len(ratings))
	for _, r := range ratings {
		agency := strings.ToUpper(strings.TrimSpace(r.Agency))
		if agency == "" {
			agency = "UNKNOWN"
		}
		ord, _ := ParseRating(r.Rating) // preserves Valid=false when unparseable
		parsed[agency] = selectedRating{
			Rating:  r.Rating,
			Agency:  agency,
			Ordinal: ord,
		}
		agencies = append(agencies, agency)
	}
	sort.Strings(agencies)
	considered := strings.Join(agencies, ",")

	// 1) Agency priority: first hit wins, even if its ordinal is invalid. A
	//    recognised-but-unparseable rating from the prioritised agency still
	//    fails closed (see caller).
	for _, a := range priority {
		if r, ok := parsed[a]; ok {
			return r, considered
		}
	}

	// 2) Split policy: ignore unparseable entries; pick worst or best of the
	//    remainder. If everything is unparseable, return the first we saw so
	//    the caller can emit a helpful UNRECOGNISED_RATING verdict.
	var valid []selectedRating
	for _, r := range parsed {
		if r.Ordinal.Valid {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		return parsed[agencies[0]], considered
	}

	sort.Slice(valid, func(i, j int) bool { return valid[i].Ordinal.Value < valid[j].Ordinal.Value })
	switch policy {
	case SplitPolicyBest:
		return valid[len(valid)-1], considered
	default:
		return valid[0], considered
	}
}
