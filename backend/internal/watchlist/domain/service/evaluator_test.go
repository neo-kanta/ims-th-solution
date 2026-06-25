package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	domainsvc "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/service"
)

func makeRule(lastState entity.RuleState, dir entity.Direction, threshold decimal.Decimal, cooldown int, lastAlertedAt *time.Time) *entity.ThresholdRule {
	return &entity.ThresholdRule{
		ID:              uuid.New(),
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       dir,
		ThresholdValue:  threshold,
		CooldownMinutes: cooldown,
		Status:          entity.RuleStatusEnabled,
		LastState:       lastState,
		LastAlertedAt:   lastAlertedAt,
	}
}

func makeQuote(price decimal.Decimal, stale bool, staleReason string, effectiveAt time.Time) *domain.QuoteInfo {
	return &domain.QuoteInfo{
		Price:       price,
		Stale:       stale,
		StaleReason: staleReason,
		EffectiveAt: effectiveAt,
		FetchedAt:   time.Now().UTC(),
	}
}

func TestEvaluate(t *testing.T) {
	now := time.Now().UTC()
	threshold := decimal.NewFromFloat(100.0)
	above := entity.DirectionAbove
	below := entity.DirectionBelow

	unknown := entity.RuleStateUnknown
	nonBreached := entity.RuleStateNonBreached
	breached := entity.RuleStateBreached

	alerted30MinAgo := now.Add(-30 * time.Minute)
	alerted90MinAgo := now.Add(-90 * time.Minute)

	cases := []struct {
		name         string
		rule         *entity.ThresholdRule
		quote        *domain.QuoteInfo
		maxStaleAge  time.Duration
		wantOutcome  domainsvc.EvalOutcome
		wantNewState entity.RuleState
		wantErr      error
		wantStale    bool
	}{
		// ── State seeding (UNKNOWN transitions) ──────────────────────────────
		{
			name:         "UNKNOWN_to_NON_BREACHED_seeds_state",
			rule:         makeRule(unknown, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(99.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeStateSeeded,
			wantNewState: nonBreached,
		},
		{
			name:         "UNKNOWN_to_BREACHED_seeds_state_no_alert",
			rule:         makeRule(unknown, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeStateSeeded,
			wantNewState: breached,
		},
		// ── Alert creation (NON_BREACHED → BREACHED) ─────────────────────────
		{
			name:         "NON_BREACHED_to_BREACHED_ABOVE_creates_alert",
			rule:         makeRule(nonBreached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
		},
		{
			name:         "NON_BREACHED_to_BREACHED_BELOW_creates_alert",
			rule:         makeRule(nonBreached, below, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
		},
		// ── No change ────────────────────────────────────────────────────────
		{
			name:         "NON_BREACHED_stays_NON_BREACHED",
			rule:         makeRule(nonBreached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(99.99), false, "", now),
			wantOutcome:  domainsvc.OutcomeNoChange,
			wantNewState: nonBreached,
		},
		{
			name:         "BREACHED_stays_BREACHED",
			rule:         makeRule(breached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(101.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeNoChange,
			wantNewState: breached,
		},
		// ── Re-arm ───────────────────────────────────────────────────────────
		{
			name:         "BREACHED_to_NON_BREACHED_rearms",
			rule:         makeRule(breached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(99.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeReArmed,
			wantNewState: nonBreached,
		},
		// ── Cooldown suppression ─────────────────────────────────────────────
		{
			name:         "NON_BREACHED_to_BREACHED_inside_cooldown_suppressed",
			rule:         makeRule(nonBreached, above, threshold, 60, &alerted30MinAgo),
			quote:        makeQuote(decimal.NewFromFloat(101.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeCooldownSuppressed,
			wantNewState: breached,
		},
		{
			name:         "NON_BREACHED_to_BREACHED_outside_cooldown_creates_alert",
			rule:         makeRule(nonBreached, above, threshold, 60, &alerted90MinAgo),
			quote:        makeQuote(decimal.NewFromFloat(101.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
		},
		// ── Stale quote handling ─────────────────────────────────────────────
		{
			name:        "stale_beyond_max_age_returns_ErrStaleQuote",
			rule:        makeRule(nonBreached, above, threshold, 60, nil),
			quote:       makeQuote(decimal.NewFromFloat(101.0), true, "no feed update", now.Add(-20*time.Minute)),
			maxStaleAge: 15 * time.Minute,
			wantErr:     domain.ErrStaleQuote,
		},
		{
			name:         "stale_within_max_age_accepted_creates_alert",
			rule:         makeRule(nonBreached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(101.0), true, "borderline stale", now.Add(-10*time.Minute)),
			maxStaleAge:  15 * time.Minute,
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
			wantStale:    true,
		},
		// ── Direction boundary conditions ────────────────────────────────────
		{
			name:         "ABOVE_exact_threshold_is_breached",
			rule:         makeRule(nonBreached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
		},
		{
			name:         "BELOW_exact_threshold_is_breached",
			rule:         makeRule(nonBreached, below, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.0), false, "", now),
			wantOutcome:  domainsvc.OutcomeAlertCreated,
			wantNewState: breached,
		},
		{
			name:         "ABOVE_one_below_threshold_not_breached",
			rule:         makeRule(nonBreached, above, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(99.99), false, "", now),
			wantOutcome:  domainsvc.OutcomeNoChange,
			wantNewState: nonBreached,
		},
		{
			name:         "BELOW_one_above_threshold_not_breached",
			rule:         makeRule(nonBreached, below, threshold, 60, nil),
			quote:        makeQuote(decimal.NewFromFloat(100.01), false, "", now),
			wantOutcome:  domainsvc.OutcomeNoChange,
			wantNewState: nonBreached,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			maxStaleAge := tc.maxStaleAge
			if maxStaleAge == 0 {
				maxStaleAge = domainsvc.DefaultMaxStaleAge
			}

			got, err := domainsvc.Evaluate(tc.rule, tc.quote, now, maxStaleAge)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got err=%v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Outcome != tc.wantOutcome {
				t.Errorf("outcome: got %v, want %v", got.Outcome, tc.wantOutcome)
			}
			if got.NewState != tc.wantNewState {
				t.Errorf("new state: got %v, want %v", got.NewState, tc.wantNewState)
			}
			if got.Stale != tc.wantStale {
				t.Errorf("stale: got %v, want %v", got.Stale, tc.wantStale)
			}
		})
	}
}

func TestAlertIdempotencyKey(t *testing.T) {
	ruleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	t0 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC) // same 60-min bucket

	k0 := domainsvc.AlertIdempotencyKey(ruleID, entity.DirectionAbove, t0, 60)
	k1 := domainsvc.AlertIdempotencyKey(ruleID, entity.DirectionAbove, t1, 60)
	k2 := domainsvc.AlertIdempotencyKey(ruleID, entity.DirectionAbove, t0, 60)

	if k0 != k1 {
		t.Errorf("same cooldown bucket should produce same key: %q vs %q", k0, k1)
	}
	if k0 != k2 {
		t.Errorf("identical inputs should produce same key: %q vs %q", k0, k2)
	}

	kDiff := domainsvc.AlertIdempotencyKey(ruleID, entity.DirectionBelow, t0, 60)
	if k0 == kDiff {
		t.Errorf("different direction should produce different key")
	}
}
