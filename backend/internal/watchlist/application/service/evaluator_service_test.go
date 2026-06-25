package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
)

// ── Fakes ────────────────────────────────────────────────────────────────────

type fakeItemRepo struct {
	items map[uuid.UUID]*entity.WatchlistItem
}

func (f *fakeItemRepo) Create(_ context.Context, _ pgx.Tx, item *entity.WatchlistItem) error {
	f.items[item.ID] = item
	return nil
}
func (f *fakeItemRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.WatchlistItem, error) {
	return f.items[id], nil
}
func (f *fakeItemRepo) Update(_ context.Context, _ pgx.Tx, item *entity.WatchlistItem) error {
	f.items[item.ID] = item
	return nil
}
func (f *fakeItemRepo) SoftDelete(_ context.Context, _ pgx.Tx, id uuid.UUID, _ uuid.UUID) error {
	delete(f.items, id)
	return nil
}
func (f *fakeItemRepo) List(_ context.Context, _ repository.WatchlistItemFilter) ([]*entity.WatchlistItem, int, error) {
	return nil, 0, nil
}

type fakeRuleRepo struct {
	rules map[uuid.UUID]*entity.ThresholdRule
}

func (f *fakeRuleRepo) Create(_ context.Context, _ pgx.Tx, r *entity.ThresholdRule) error {
	f.rules[r.ID] = r
	return nil
}
func (f *fakeRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.ThresholdRule, error) {
	return f.rules[id], nil
}
func (f *fakeRuleRepo) GetByIDForUpdate(_ context.Context, _ pgx.Tx, id uuid.UUID) (*entity.ThresholdRule, error) {
	return f.rules[id], nil
}
func (f *fakeRuleRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]*entity.ThresholdRule, error) {
	var out []*entity.ThresholdRule
	for _, r := range f.rules {
		if r.WatchlistItemID == itemID {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeRuleRepo) Update(_ context.Context, _ pgx.Tx, r *entity.ThresholdRule) error {
	f.rules[r.ID] = r
	return nil
}
func (f *fakeRuleRepo) UpdateState(_ context.Context, _ pgx.Tx, upd repository.RuleStateUpdate) error {
	if r, ok := f.rules[upd.ID]; ok {
		r.LastState = upd.LastState
	}
	return nil
}
func (f *fakeRuleRepo) SoftDisable(_ context.Context, _ pgx.Tx, id uuid.UUID, _ uuid.UUID) error {
	delete(f.rules, id)
	return nil
}
func (f *fakeRuleRepo) ListForEvaluation(_ context.Context, filter repository.EvaluatorRuleFilter) ([]*entity.ThresholdRule, error) {
	var out []*entity.ThresholdRule
	for _, r := range f.rules {
		if filter.RuleID != nil && r.ID != *filter.RuleID {
			continue
		}
		if filter.ItemID != nil && r.WatchlistItemID != *filter.ItemID {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

type fakeAlertRepo struct{}

func (f *fakeAlertRepo) Insert(_ context.Context, _ pgx.Tx, _ *entity.AlertEvent) error { return nil }
func (f *fakeAlertRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.AlertEvent, error) {
	return nil, nil
}
func (f *fakeAlertRepo) UpdateNotificationStatus(_ context.Context, _ uuid.UUID, _ entity.NotificationStatus, _ *string) error {
	return nil
}
func (f *fakeAlertRepo) Acknowledge(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ uuid.UUID, _ *string) error {
	return nil
}
func (f *fakeAlertRepo) List(_ context.Context, _ repository.AlertEventFilter) ([]*entity.AlertEvent, int, error) {
	return nil, 0, nil
}

type fakeSecurityPort struct {
	symbol     string
	resolveErr error
	resolveNil bool
}

func (f *fakeSecurityPort) GetSecurityByID(_ context.Context, _ string) (*watchlistdomain.SecurityInfo, error) {
	return &watchlistdomain.SecurityInfo{Status: "ACTIVE"}, nil
}
func (f *fakeSecurityPort) ResolveProviderSymbol(_ context.Context, _, _ string) (*watchlistdomain.ProviderMapping, error) {
	if f.resolveErr != nil {
		return nil, f.resolveErr
	}
	if f.resolveNil {
		return nil, nil
	}
	return &watchlistdomain.ProviderMapping{ProviderCode: "TEST", ProviderSymbol: f.symbol}, nil
}

type fakeQuotePort struct {
	quote *watchlistdomain.QuoteInfo
	err   error
}

func (f *fakeQuotePort) GetLatestQuote(_ context.Context, _ string) (*watchlistdomain.QuoteInfo, error) {
	return f.quote, f.err
}
func (f *fakeQuotePort) PrimaryProviderName() string { return "TEST" }

type fakePortfolioPort struct {
	info *watchlistdomain.PortfolioScopeInfo
}

func (f *fakePortfolioPort) GetPortfolioScope(_ context.Context, _ uuid.UUID) (*watchlistdomain.PortfolioScopeInfo, error) {
	return f.info, nil
}

type fakeIAMChecker struct {
	hasData bool
	err     error
}

func (f *fakeIAMChecker) HasFunctionPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *fakeIAMChecker) HasDataPermission(_ uuid.UUID, _ string) (bool, error) {
	return f.hasData, f.err
}
func (f *fakeIAMChecker) GetAccessibleContracts(_ uuid.UUID) ([]string, error) {
	return nil, nil
}

type fakeAuditEvent struct {
	eventType string
	targetID  string
	metadata  map[string]interface{}
}

type fakeAuditRecorder struct {
	events []fakeAuditEvent
}

func (f *fakeAuditRecorder) Record(_ context.Context, _ *uuid.UUID, eventType, _, targetID, _, _ string, metadata map[string]interface{}) {
	f.events = append(f.events, fakeAuditEvent{eventType: eventType, targetID: targetID, metadata: metadata})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func newEvaluatorForTest(items repository.WatchlistItemRepository, rules repository.ThresholdRuleRepository, quotePort watchlistdomain.QuotePort, portfolioPort watchlistdomain.PortfolioScopePort, iamChecker interface {
	HasFunctionPermission(uuid.UUID, string) (bool, error)
	HasDataPermission(uuid.UUID, string) (bool, error)
	GetAccessibleContracts(uuid.UUID) ([]string, error)
}) *EvaluatorService {
	svc := &EvaluatorService{
		pool:          nil, // dry-run tests never use pool
		rules:         rules,
		items:         items,
		alerts:        &fakeAlertRepo{},
		securityPort:  &fakeSecurityPort{symbol: "TEST.TH"},
		quotePort:     quotePort,
		portfolioPort: portfolioPort,
		notifier:      nil,
		auditRecorder: &fakeAuditRecorder{},
		iamChecker:    iamChecker,
		maxStaleAge:   15 * time.Minute,
	}
	return svc
}

func auditEvents(svc *EvaluatorService, eventType string) []fakeAuditEvent {
	rec, ok := svc.auditRecorder.(*fakeAuditRecorder)
	if !ok || rec == nil {
		return nil
	}
	var out []fakeAuditEvent
	for _, ev := range rec.events {
		if ev.eventType == eventType {
			out = append(out, ev)
		}
	}
	return out
}

func seedRule(t *testing.T, rules *fakeRuleRepo, items *fakeItemRepo, scope entity.ScopeType, portfolioID *uuid.UUID, lastState entity.RuleState, lastAlertedAt *time.Time) (*entity.ThresholdRule, *entity.WatchlistItem) {
	t.Helper()
	itemID := uuid.New()
	ruleID := uuid.New()

	ownerID := uuid.New()
	item := &entity.WatchlistItem{
		ID:        itemID,
		ScopeType: scope,
		OwnerUserID: func() *uuid.UUID {
			if scope == entity.ScopePersonal {
				return &ownerID
			}
			return nil
		}(),
		PortfolioID: portfolioID,
		SecurityID:  uuid.New(),
		Status:      entity.ItemStatusActive,
		CreatedBy:   ownerID,
	}
	items.items[itemID] = item

	rule := &entity.ThresholdRule{
		ID:              ruleID,
		WatchlistItemID: itemID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       lastState,
		LastAlertedAt:   lastAlertedAt,
	}
	rules.rules[ruleID] = rule
	return rule, item
}

// ── Tests ────────────────────────────────────────────────────────────────────

// P1-2: Manual evaluation rejects portfolio-scoped rule when actor lacks data permission.
func TestEvaluatorService_PortfolioScopeBypass_Rejected(t *testing.T) {
	fundID := uuid.New()
	portfolioID := uuid.New()

	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	rule, _ := seedRule(t, rules, items, entity.ScopePortfolio, &portfolioID, entity.RuleStateNonBreached, nil)

	portfolioPort := &fakePortfolioPort{info: &watchlistdomain.PortfolioScopeInfo{
		PortfolioID: portfolioID,
		FundID:      fundID,
	}}
	iamChecker := &fakeIAMChecker{hasData: false}
	quotePort := &fakeQuotePort{quote: &watchlistdomain.QuoteInfo{
		Price:       decimal.NewFromFloat(110),
		EffectiveAt: time.Now().UTC(),
		FetchedAt:   time.Now().UTC(),
	}}

	svc := newEvaluatorForTest(items, rules, quotePort, portfolioPort, iamChecker)

	actorID := uuid.New()
	_, err := svc.Evaluate(context.Background(), EvaluateFilter{RuleID: &rule.ID, DryRun: true}, &actorID)
	if !errors.Is(err, watchlistdomain.ErrForbiddenScope) {
		t.Errorf("expected ErrForbiddenScope, got %v", err)
	}
}

// P1-5: Targeted evaluation returns ErrStaleQuote when quote is stale beyond max age.
func TestEvaluatorService_TargetedStaleQuote_ReturnsError(t *testing.T) {
	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	rule, _ := seedRule(t, rules, items, entity.ScopePersonal, nil, entity.RuleStateNonBreached, nil)

	staleQuote := &watchlistdomain.QuoteInfo{
		Price:       decimal.NewFromFloat(110),
		Stale:       true,
		StaleReason: "no feed update for 20 min",
		EffectiveAt: time.Now().UTC().Add(-20 * time.Minute), // beyond 15-min max
		FetchedAt:   time.Now().UTC(),
	}
	quotePort := &fakeQuotePort{quote: staleQuote}
	svc := newEvaluatorForTest(items, rules, quotePort, &fakePortfolioPort{}, nil)

	actorID := uuid.New()
	_, err := svc.Evaluate(context.Background(), EvaluateFilter{RuleID: &rule.ID, DryRun: true}, &actorID)
	if !errors.Is(err, watchlistdomain.ErrStaleQuote) {
		t.Errorf("expected ErrStaleQuote, got %v", err)
	}
	events := auditEvents(svc, "WATCHLIST_EVALUATION_FAILED")
	if len(events) != 1 {
		t.Fatalf("WATCHLIST_EVALUATION_FAILED emitted %d times, want 1", len(events))
	}
	if got := events[0].metadata["error_code"]; got != "WATCHLIST_STALE_QUOTE" {
		t.Errorf("audit error_code = %v, want WATCHLIST_STALE_QUOTE", got)
	}
}

func TestEvaluatorService_TargetedProviderUnavailable_EmitsAuditBeforeReturn(t *testing.T) {
	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	rule, _ := seedRule(t, rules, items, entity.ScopePersonal, nil, entity.RuleStateNonBreached, nil)

	quotePort := &fakeQuotePort{err: watchlistdomain.ErrProviderUnavailable}
	svc := newEvaluatorForTest(items, rules, quotePort, &fakePortfolioPort{}, nil)

	actorID := uuid.New()
	_, err := svc.Evaluate(context.Background(), EvaluateFilter{RuleID: &rule.ID, DryRun: true}, &actorID)
	if !errors.Is(err, watchlistdomain.ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got %v", err)
	}
	events := auditEvents(svc, "WATCHLIST_EVALUATION_FAILED")
	if len(events) != 1 {
		t.Fatalf("WATCHLIST_EVALUATION_FAILED emitted %d times, want 1", len(events))
	}
	if got := events[0].metadata["error_code"]; got != "WATCHLIST_PROVIDER_UNAVAILABLE" {
		t.Errorf("audit error_code = %v, want WATCHLIST_PROVIDER_UNAVAILABLE", got)
	}
}

func TestEvaluatorService_TargetedProviderMappingFailure_EmitsAuditBeforeReturn(t *testing.T) {
	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	rule, _ := seedRule(t, rules, items, entity.ScopePersonal, nil, entity.RuleStateNonBreached, nil)

	quotePort := &fakeQuotePort{quote: &watchlistdomain.QuoteInfo{
		Price:       decimal.NewFromFloat(110),
		EffectiveAt: time.Now().UTC(),
		FetchedAt:   time.Now().UTC(),
	}}
	svc := newEvaluatorForTest(items, rules, quotePort, &fakePortfolioPort{}, nil)
	svc.securityPort = &fakeSecurityPort{resolveErr: errors.New("mapping lookup failed")}

	actorID := uuid.New()
	_, err := svc.Evaluate(context.Background(), EvaluateFilter{RuleID: &rule.ID, DryRun: true}, &actorID)
	if err == nil {
		t.Fatal("expected provider mapping error, got nil")
	}
	events := auditEvents(svc, "WATCHLIST_EVALUATION_FAILED")
	if len(events) != 1 {
		t.Fatalf("WATCHLIST_EVALUATION_FAILED emitted %d times, want 1", len(events))
	}
	if got := events[0].metadata["error_code"]; got != "WATCHLIST_EVALUATION_FAILED" {
		t.Errorf("audit error_code = %v, want WATCHLIST_EVALUATION_FAILED", got)
	}
}

// P2-3: Cooldown suppression increments AlertsSuppressed (not RulesSkipped).
func TestEvaluatorService_CooldownSuppression_CountsAsSuppressed(t *testing.T) {
	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	// Rule alerted 30 minutes ago (inside 60-minute cooldown).
	recent := time.Now().UTC().Add(-30 * time.Minute)
	rule, _ := seedRule(t, rules, items, entity.ScopePersonal, nil, entity.RuleStateNonBreached, &recent)

	// Quote above threshold — would create alert but cooldown suppresses it.
	quotePort := &fakeQuotePort{quote: &watchlistdomain.QuoteInfo{
		Price:       decimal.NewFromFloat(110),
		EffectiveAt: time.Now().UTC(),
		FetchedAt:   time.Now().UTC(),
	}}
	svc := newEvaluatorForTest(items, rules, quotePort, &fakePortfolioPort{}, nil)

	actorID := uuid.New()
	out, err := svc.Evaluate(context.Background(), EvaluateFilter{RuleID: &rule.ID, DryRun: true}, &actorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AlertsSuppressed != 1 {
		t.Errorf("AlertsSuppressed = %d, want 1", out.AlertsSuppressed)
	}
	if out.AlertsCreated != 0 {
		t.Errorf("AlertsCreated = %d, want 0", out.AlertsCreated)
	}
}

// P1-5 batch: stale quote in batch mode counts as RulesSkipped, not ProviderFailures.
func TestEvaluatorService_BatchStaleQuote_CountsAsSkipped(t *testing.T) {
	items := &fakeItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	seedRule(t, rules, items, entity.ScopePersonal, nil, entity.RuleStateNonBreached, nil)

	staleQuote := &watchlistdomain.QuoteInfo{
		Price:       decimal.NewFromFloat(110),
		Stale:       true,
		StaleReason: "no feed update",
		EffectiveAt: time.Now().UTC().Add(-20 * time.Minute),
		FetchedAt:   time.Now().UTC(),
	}
	quotePort := &fakeQuotePort{quote: staleQuote}
	svc := newEvaluatorForTest(items, rules, quotePort, &fakePortfolioPort{}, nil)

	out, err := svc.Evaluate(context.Background(), EvaluateFilter{DryRun: true}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ProviderFailures != 0 {
		t.Errorf("ProviderFailures = %d, want 0 (stale is not a provider failure)", out.ProviderFailures)
	}
	if out.RulesSkipped != 1 {
		t.Errorf("RulesSkipped = %d, want 1", out.RulesSkipped)
	}
	events := auditEvents(svc, "WATCHLIST_EVALUATION_FAILED")
	if len(events) != 1 {
		t.Fatalf("WATCHLIST_EVALUATION_FAILED emitted %d times, want 1", len(events))
	}
}
