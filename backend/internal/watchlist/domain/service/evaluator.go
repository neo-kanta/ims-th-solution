package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
)

// EvalOutcome classifies a single rule evaluation result.
type EvalOutcome int

const (
	OutcomeStateSeeded        EvalOutcome = iota // UNKNOWN → any state; no alert
	OutcomeNoChange                              // state unchanged; no alert
	OutcomeReArmed                               // BREACHED → NON_BREACHED; no alert
	OutcomeAlertCreated                          // NON_BREACHED → BREACHED; alert emitted
	OutcomeCooldownSuppressed                    // NON_BREACHED → BREACHED inside cooldown; no notification
)

// EvalResult is the output of evaluating one threshold rule against a quote.
type EvalResult struct {
	Outcome       EvalOutcome
	PreviousState entity.RuleState
	NewState      entity.RuleState
	ObservedPrice decimal.Decimal
	Stale         bool
	StaleReason   string
	AlertKey      string // idempotency key; populated only when an alert is expected
}

// DefaultMaxStaleAge is the maximum quote age allowed for alert evaluation.
const DefaultMaxStaleAge = 15 * time.Minute

// isBreached returns true when the observed price satisfies the crossing predicate.
//
// ABOVE: breached when observed_price >= threshold_value
// BELOW: breached when observed_price <= threshold_value
func isBreached(direction entity.Direction, price, threshold decimal.Decimal) bool {
	switch direction {
	case entity.DirectionAbove:
		return price.GreaterThanOrEqual(threshold)
	case entity.DirectionBelow:
		return price.LessThanOrEqual(threshold)
	}
	return false
}

// AlertIdempotencyKey generates a deterministic deduplication key for one
// crossing event. The timestamp is truncated to the rule cooldown window so
// that evaluator re-runs within the same window produce the same key.
func AlertIdempotencyKey(ruleID uuid.UUID, direction entity.Direction, observedAt time.Time, cooldownMinutes int) string {
	var bucket time.Time
	if cooldownMinutes > 0 {
		bucket = observedAt.Truncate(time.Duration(cooldownMinutes) * time.Minute)
	} else {
		bucket = observedAt.UTC()
	}
	return fmt.Sprintf("watchlist:rule:%s:crossing:%s:bucket:%s",
		ruleID.String(),
		string(direction),
		bucket.UTC().Format(time.RFC3339),
	)
}

// Evaluate applies the state machine to one rule and quote.
//
// It returns ErrStaleQuote when the quote is stale beyond maxStaleAge.
// It does NOT persist anything; callers must commit the returned state.
func Evaluate(
	rule *entity.ThresholdRule,
	quote *domain.QuoteInfo,
	now time.Time,
	maxStaleAge time.Duration,
) (EvalResult, error) {
	if maxStaleAge == 0 {
		maxStaleAge = DefaultMaxStaleAge
	}

	// Reject quotes stale beyond the accepted max age.
	if quote.Stale && now.Sub(quote.EffectiveAt) > maxStaleAge {
		return EvalResult{}, domain.ErrStaleQuote
	}

	prevState := rule.LastState
	breached := isBreached(rule.Direction, quote.Price, rule.ThresholdValue)

	var newState entity.RuleState
	if breached {
		newState = entity.RuleStateBreached
	} else {
		newState = entity.RuleStateNonBreached
	}

	// UNKNOWN → any: seed state, no alert.
	if prevState == entity.RuleStateUnknown {
		return EvalResult{
			Outcome:       OutcomeStateSeeded,
			PreviousState: prevState,
			NewState:      newState,
			ObservedPrice: quote.Price,
			Stale:         quote.Stale,
			StaleReason:   quote.StaleReason,
		}, nil
	}

	// BREACHED → NON_BREACHED: re-arm, no alert.
	if prevState == entity.RuleStateBreached && newState == entity.RuleStateNonBreached {
		return EvalResult{
			Outcome:       OutcomeReArmed,
			PreviousState: prevState,
			NewState:      newState,
			ObservedPrice: quote.Price,
			Stale:         quote.Stale,
			StaleReason:   quote.StaleReason,
		}, nil
	}

	// BREACHED → BREACHED: keep state, no repeated alert.
	if prevState == entity.RuleStateBreached && newState == entity.RuleStateBreached {
		return EvalResult{
			Outcome:       OutcomeNoChange,
			PreviousState: prevState,
			NewState:      newState,
			ObservedPrice: quote.Price,
			Stale:         quote.Stale,
			StaleReason:   quote.StaleReason,
		}, nil
	}

	// NON_BREACHED → NON_BREACHED: no alert.
	if prevState == entity.RuleStateNonBreached && newState == entity.RuleStateNonBreached {
		return EvalResult{
			Outcome:       OutcomeNoChange,
			PreviousState: prevState,
			NewState:      newState,
			ObservedPrice: quote.Price,
			Stale:         quote.Stale,
			StaleReason:   quote.StaleReason,
		}, nil
	}

	// NON_BREACHED → BREACHED: crossing candidate.
	key := AlertIdempotencyKey(rule.ID, rule.Direction, quote.EffectiveAt, rule.CooldownMinutes)

	// Cooldown suppression: if last alert was within the cooldown window, suppress notification.
	if rule.LastAlertedAt != nil {
		cooldown := time.Duration(rule.CooldownMinutes) * time.Minute
		if now.Before(rule.LastAlertedAt.Add(cooldown)) {
			return EvalResult{
				Outcome:       OutcomeCooldownSuppressed,
				PreviousState: prevState,
				NewState:      newState,
				ObservedPrice: quote.Price,
				Stale:         quote.Stale,
				StaleReason:   quote.StaleReason,
				AlertKey:      key,
			}, nil
		}
	}

	return EvalResult{
		Outcome:       OutcomeAlertCreated,
		PreviousState: prevState,
		NewState:      newState,
		ObservedPrice: quote.Price,
		Stale:         quote.Stale,
		StaleReason:   quote.StaleReason,
		AlertKey:      key,
	}, nil
}
