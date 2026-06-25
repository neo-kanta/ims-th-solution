package command

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

// fakePgxTx embeds pgx.Tx so we only need to override Commit and Rollback.
// All other methods panic on call; since fake repos ignore the tx, this is safe.
type fakePgxTx struct{ pgx.Tx }

func (f *fakePgxTx) Commit(ctx context.Context) error   { return nil }
func (f *fakePgxTx) Rollback(ctx context.Context) error { return nil }
func (f *fakePgxTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return f, nil
}

// fakeTxStarter satisfies the unexported txStarter interface in update_item.go.
type fakeTxStarter struct{}

func (f *fakeTxStarter) Begin(ctx context.Context) (pgx.Tx, error) {
	return &fakePgxTx{}, nil
}

type fakeUpdateItemRepo struct {
	items map[uuid.UUID]*entity.WatchlistItem
}

func (f *fakeUpdateItemRepo) Create(_ context.Context, _ pgx.Tx, item *entity.WatchlistItem) error {
	f.items[item.ID] = item
	return nil
}
func (f *fakeUpdateItemRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.WatchlistItem, error) {
	return f.items[id], nil
}
func (f *fakeUpdateItemRepo) Update(_ context.Context, _ pgx.Tx, item *entity.WatchlistItem) error {
	f.items[item.ID] = item
	return nil
}
func (f *fakeUpdateItemRepo) SoftDelete(_ context.Context, _ pgx.Tx, id uuid.UUID, _ uuid.UUID) error {
	delete(f.items, id)
	return nil
}
func (f *fakeUpdateItemRepo) List(_ context.Context, _ repository.WatchlistItemFilter) ([]*entity.WatchlistItem, int, error) {
	return nil, 0, nil
}

type fakeUpdateRuleRepo struct {
	rules    map[uuid.UUID]*entity.ThresholdRule
	disabled []uuid.UUID
	updated  []uuid.UUID
	created  []uuid.UUID
}

func (f *fakeUpdateRuleRepo) Create(_ context.Context, _ pgx.Tx, r *entity.ThresholdRule) error {
	f.rules[r.ID] = r
	f.created = append(f.created, r.ID)
	return nil
}
func (f *fakeUpdateRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.ThresholdRule, error) {
	return f.rules[id], nil
}
func (f *fakeUpdateRuleRepo) GetByIDForUpdate(_ context.Context, _ pgx.Tx, id uuid.UUID) (*entity.ThresholdRule, error) {
	return f.rules[id], nil
}
func (f *fakeUpdateRuleRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]*entity.ThresholdRule, error) {
	var out []*entity.ThresholdRule
	for _, r := range f.rules {
		if r.WatchlistItemID == itemID {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeUpdateRuleRepo) Update(_ context.Context, _ pgx.Tx, r *entity.ThresholdRule) error {
	f.rules[r.ID] = r
	f.updated = append(f.updated, r.ID)
	return nil
}
func (f *fakeUpdateRuleRepo) UpdateState(_ context.Context, _ pgx.Tx, upd repository.RuleStateUpdate) error {
	if r, ok := f.rules[upd.ID]; ok {
		r.LastState = upd.LastState
	}
	return nil
}
func (f *fakeUpdateRuleRepo) SoftDisable(_ context.Context, _ pgx.Tx, id uuid.UUID, _ uuid.UUID) error {
	if _, ok := f.rules[id]; !ok {
		return watchlistdomain.ErrRuleDisabled
	}
	delete(f.rules, id)
	f.disabled = append(f.disabled, id)
	return nil
}
func (f *fakeUpdateRuleRepo) ListForEvaluation(_ context.Context, _ repository.EvaluatorRuleFilter) ([]*entity.ThresholdRule, error) {
	return nil, nil
}

type fakeUpdatePortfolioPort struct{}

func (f *fakeUpdatePortfolioPort) GetPortfolioScope(_ context.Context, _ uuid.UUID) (*watchlistdomain.PortfolioScopeInfo, error) {
	return nil, nil
}

type fakeUpdatePermissionChecker struct{}

func (f *fakeUpdatePermissionChecker) HasFunctionPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *fakeUpdatePermissionChecker) HasDataPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *fakeUpdatePermissionChecker) GetAccessibleContracts(_ uuid.UUID) ([]string, error) {
	return nil, nil
}

type fakeCreateSecurityPort struct{}

func (f *fakeCreateSecurityPort) GetSecurityByID(_ context.Context, id string) (*watchlistdomain.SecurityInfo, error) {
	return &watchlistdomain.SecurityInfo{ID: id, Status: "ACTIVE"}, nil
}
func (f *fakeCreateSecurityPort) ResolveProviderSymbol(_ context.Context, _, _ string) (*watchlistdomain.ProviderMapping, error) {
	return nil, nil
}

type auditCapture struct {
	events []string
}

func (a *auditCapture) Record(_ context.Context, _ *uuid.UUID, eventType, _, _, _, _ string, _ map[string]interface{}) {
	a.events = append(a.events, eventType)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func buildUpdateHandler(items *fakeUpdateItemRepo, rules *fakeUpdateRuleRepo, audit *auditCapture) *UpdateItemHandler {
	return &UpdateItemHandler{
		pool:          &fakeTxStarter{},
		items:         items,
		rules:         rules,
		portfolioPort: &fakeUpdatePortfolioPort{},
		iamChecker:    &fakeUpdatePermissionChecker{},
		auditRecorder: audit,
	}
}

func seedPersonalItem(items *fakeUpdateItemRepo, actorID uuid.UUID) *entity.WatchlistItem {
	item := &entity.WatchlistItem{
		ID:          uuid.New(),
		ScopeType:   entity.ScopePersonal,
		OwnerUserID: &actorID,
		SecurityID:  uuid.New(),
		Status:      entity.ItemStatusActive,
		CreatedBy:   actorID,
	}
	items.items[item.ID] = item
	return item
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// Test 4 (P1-4): PATCH with id references existing rule and updates it in-place.
func TestUpdateItem_UpdateExistingRuleByID(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	audit := &auditCapture{}

	item := seedPersonalItem(items, actorID)

	existingRuleID := uuid.New()
	rules.rules[existingRuleID] = &entity.ThresholdRule{
		ID:              existingRuleID,
		WatchlistItemID: item.ID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       entity.RuleStateBreached,
		CreatedBy:       actorID,
	}

	newThreshold := decimal.NewFromFloat(120)
	rulesInput := []ThresholdRuleInput{
		{
			ID:              &existingRuleID,
			Direction:       entity.DirectionAbove,
			ThresholdValue:  newThreshold,
			CooldownMinutes: 60,
			Status:          entity.RuleStatusEnabled,
		},
	}

	h := buildUpdateHandler(items, rules, audit)
	out, err := h.Handle(context.Background(), UpdateItemInput{
		ItemID:         item.ID,
		ThresholdRules: &rulesInput,
		ActorID:        actorID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules.updated) != 1 || rules.updated[0] != existingRuleID {
		t.Errorf("expected rule %s to be updated; updated=%v", existingRuleID, rules.updated)
	}
	if len(rules.created) != 0 {
		t.Errorf("expected no new rules created, got %v", rules.created)
	}
	if len(rules.disabled) != 0 {
		t.Errorf("expected no rules disabled, got %v", rules.disabled)
	}

	// Threshold changed → last_state must reset to UNKNOWN.
	updated := rules.rules[existingRuleID]
	if updated.LastState != entity.RuleStateUnknown {
		t.Errorf("LastState = %v after threshold change, want UNKNOWN", updated.LastState)
	}
	if !updated.ThresholdValue.Equal(newThreshold) {
		t.Errorf("ThresholdValue = %v, want %v", updated.ThresholdValue, newThreshold)
	}
	if len(out.Rules) != 1 {
		t.Errorf("output Rules count = %d, want 1", len(out.Rules))
	}
}

// Test 5 (P1-4): PATCH without matching ID soft-disables the omitted rule.
func TestUpdateItem_OmittedRuleGetsSoftDisabled(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	audit := &auditCapture{}

	item := seedPersonalItem(items, actorID)
	existingRuleID := uuid.New()
	rules.rules[existingRuleID] = &entity.ThresholdRule{
		ID:              existingRuleID,
		WatchlistItemID: item.ID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       entity.RuleStateNonBreached,
		CreatedBy:       actorID,
	}

	// PATCH with a new rule (no id) — existing rule not referenced → soft-disable.
	rulesInput := []ThresholdRuleInput{
		{
			Direction:       entity.DirectionBelow,
			ThresholdValue:  decimal.NewFromFloat(80),
			CooldownMinutes: 60,
			Status:          entity.RuleStatusEnabled,
		},
	}

	h := buildUpdateHandler(items, rules, audit)
	if _, err := h.Handle(context.Background(), UpdateItemInput{
		ItemID:         item.ID,
		ThresholdRules: &rulesInput,
		ActorID:        actorID,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules.disabled) != 1 || rules.disabled[0] != existingRuleID {
		t.Errorf("expected existing rule %s to be disabled; disabled=%v", existingRuleID, rules.disabled)
	}
	if len(rules.created) != 1 {
		t.Errorf("expected 1 new rule created, got %d", len(rules.created))
	}

	wantEvents := map[string]bool{"WATCHLIST_THRESHOLD_DISABLED": false, "WATCHLIST_THRESHOLD_CREATED": false}
	for _, ev := range audit.events {
		wantEvents[ev] = true
	}
	for ev, seen := range wantEvents {
		if !seen {
			t.Errorf("expected audit event %q not emitted; events=%v", ev, audit.events)
		}
	}
}

// Test 6 (P1-4): WATCHLIST_THRESHOLD_UPDATED audit is emitted on in-place update.
func TestUpdateItem_ThresholdUpdatedAuditEvent(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	audit := &auditCapture{}

	item := seedPersonalItem(items, actorID)
	ruleID := uuid.New()
	rules.rules[ruleID] = &entity.ThresholdRule{
		ID:              ruleID,
		WatchlistItemID: item.ID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       entity.RuleStateNonBreached,
		CreatedBy:       actorID,
	}

	rulesInput := []ThresholdRuleInput{
		{ID: &ruleID, Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(105), CooldownMinutes: 60, Status: entity.RuleStatusEnabled},
	}
	h := buildUpdateHandler(items, rules, audit)
	if _, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedCount := 0
	for _, ev := range audit.events {
		if ev == "WATCHLIST_THRESHOLD_UPDATED" {
			updatedCount++
		}
	}
	if updatedCount != 1 {
		t.Errorf("WATCHLIST_THRESHOLD_UPDATED emitted %d times, want 1; events=%v", updatedCount, audit.events)
	}
}

// Bonus: unchanged direction+threshold preserves LastState.
func TestUpdateItem_UnchangedParams_LastStatePreserved(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	item := seedPersonalItem(items, actorID)
	ruleID := uuid.New()
	rules.rules[ruleID] = &entity.ThresholdRule{
		ID:              ruleID,
		WatchlistItemID: item.ID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       entity.RuleStateBreached,
		CreatedBy:       actorID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	rulesInput := []ThresholdRuleInput{
		{ID: &ruleID, Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(100), CooldownMinutes: 30, Status: entity.RuleStatusEnabled},
	}
	h := buildUpdateHandler(items, rules, &auditCapture{})
	if _, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rules.rules[ruleID].LastState != entity.RuleStateBreached {
		t.Errorf("LastState changed to %v but direction+threshold unchanged", rules.rules[ruleID].LastState)
	}
}

func TestUpdateItem_ForeignRuleID_ReturnsInvalidThresholdWithoutCreate(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	item := seedPersonalItem(items, actorID)
	foreignItemID := uuid.New()
	foreignRuleID := uuid.New()
	rules.rules[foreignRuleID] = &entity.ThresholdRule{
		ID:              foreignRuleID,
		WatchlistItemID: foreignItemID,
		MetricType:      entity.MetricTypeMarketPrice,
		Direction:       entity.DirectionAbove,
		ThresholdValue:  decimal.NewFromFloat(100),
		CooldownMinutes: 60,
		Status:          entity.RuleStatusEnabled,
		LastState:       entity.RuleStateNonBreached,
		CreatedBy:       actorID,
	}

	rulesInput := []ThresholdRuleInput{
		{ID: &foreignRuleID, Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(110), CooldownMinutes: 60, Status: entity.RuleStatusEnabled},
	}
	h := buildUpdateHandler(items, rules, &auditCapture{})
	_, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(rules.created) != 0 {
		t.Errorf("foreign rule id should not create a new rule; created=%v", rules.created)
	}
}

func TestUpdateItem_UnknownRuleID_ReturnsInvalidThresholdWithoutCreate(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	item := seedPersonalItem(items, actorID)
	unknownRuleID := uuid.New()
	rulesInput := []ThresholdRuleInput{
		{ID: &unknownRuleID, Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(110), CooldownMinutes: 60, Status: entity.RuleStatusEnabled},
	}
	h := buildUpdateHandler(items, rules, &auditCapture{})
	_, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(rules.created) != 0 {
		t.Errorf("unknown rule id should not create a new rule; created=%v", rules.created)
	}
}

func TestUpdateItem_InvalidRuleStatus_ReturnsInvalidThreshold(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	item := seedPersonalItem(items, actorID)
	rulesInput := []ThresholdRuleInput{
		{Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(100), CooldownMinutes: 60, Status: entity.RuleStatus("PAUSED")},
	}
	h := buildUpdateHandler(items, rules, &auditCapture{})
	_, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(rules.created) != 0 {
		t.Errorf("invalid status should not create a new rule; created=%v", rules.created)
	}
}

func TestUpdateItem_NegativeCooldown_ReturnsInvalidThreshold(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}

	item := seedPersonalItem(items, actorID)
	rulesInput := []ThresholdRuleInput{
		{Direction: entity.DirectionAbove, ThresholdValue: decimal.NewFromFloat(100), CooldownMinutes: -1, Status: entity.RuleStatusEnabled},
	}
	h := buildUpdateHandler(items, rules, &auditCapture{})
	_, err := h.Handle(context.Background(), UpdateItemInput{ItemID: item.ID, ThresholdRules: &rulesInput, ActorID: actorID})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(rules.created) != 0 {
		t.Errorf("negative cooldown should not create a new rule; created=%v", rules.created)
	}
}

func TestCreateItem_InvalidRuleStatus_ReturnsInvalidThresholdBeforePersistence(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	h := NewCreateItemHandler(nil, items, rules, &fakeCreateSecurityPort{}, nil, nil, &auditCapture{})

	_, err := h.Handle(context.Background(), CreateItemInput{
		ScopeType:  entity.ScopePersonal,
		SecurityID: uuid.New(),
		ThresholdRules: []ThresholdRuleInput{{
			Direction:       entity.DirectionAbove,
			ThresholdValue:  decimal.NewFromFloat(100),
			CooldownMinutes: 60,
			Status:          entity.RuleStatus("PAUSED"),
		}},
		ActorID: actorID,
	})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(items.items) != 0 || len(rules.created) != 0 {
		t.Errorf("invalid status should not persist item/rules; items=%d createdRules=%v", len(items.items), rules.created)
	}
}

func TestCreateItem_NegativeCooldown_ReturnsInvalidThresholdBeforePersistence(t *testing.T) {
	actorID := uuid.New()
	items := &fakeUpdateItemRepo{items: make(map[uuid.UUID]*entity.WatchlistItem)}
	rules := &fakeUpdateRuleRepo{rules: make(map[uuid.UUID]*entity.ThresholdRule)}
	h := NewCreateItemHandler(nil, items, rules, &fakeCreateSecurityPort{}, nil, nil, &auditCapture{})

	_, err := h.Handle(context.Background(), CreateItemInput{
		ScopeType:  entity.ScopePersonal,
		SecurityID: uuid.New(),
		ThresholdRules: []ThresholdRuleInput{{
			Direction:       entity.DirectionAbove,
			ThresholdValue:  decimal.NewFromFloat(100),
			CooldownMinutes: -1,
			Status:          entity.RuleStatusEnabled,
		}},
		ActorID: actorID,
	})
	if !errors.Is(err, watchlistdomain.ErrInvalidThreshold) {
		t.Fatalf("err = %v, want ErrInvalidThreshold", err)
	}
	if len(items.items) != 0 || len(rules.created) != 0 {
		t.Errorf("negative cooldown should not persist item/rules; items=%d createdRules=%v", len(items.items), rules.created)
	}
}
