package command_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	// Registers "amount.minimum_trade" in the global SPI registry via init().
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/amount"
)

// ─── fakes ───────────────────────────────────────────────────────────────────

// pretradeBindingRepo returns a fixed binding slice and counts ResolveApplicable calls.
type pretradeBindingRepo struct {
	bindings []domain.ResolvedBinding
}

func (r *pretradeBindingRepo) ResolveApplicable(_ context.Context, _ []vo.Scope, _ time.Time) ([]domain.ResolvedBinding, error) {
	return r.bindings, nil
}
func (r *pretradeBindingRepo) Create(_ context.Context, _ *entity.RuleBinding) error {
	panic("pretradeBindingRepo.Create not expected in pre-trade pipeline tests")
}
func (r *pretradeBindingRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.RuleBinding, error) {
	panic("pretradeBindingRepo.GetByID not expected")
}
func (r *pretradeBindingRepo) List(_ context.Context, _ domain.BindingFilter) ([]entity.RuleBinding, int64, error) {
	panic("pretradeBindingRepo.List not expected")
}
func (r *pretradeBindingRepo) Deactivate(_ context.Context, _ uuid.UUID) error {
	panic("pretradeBindingRepo.Deactivate not expected")
}

// pretradeCheckRepo records CreateBatch calls for dry-run assertions.
type pretradeCheckRepo struct {
	batchCalls int32
}

func (r *pretradeCheckRepo) CreateBatch(_ context.Context, _ []entity.CheckRecord) error {
	atomic.AddInt32(&r.batchCalls, 1)
	return nil
}
func (r *pretradeCheckRepo) Create(_ context.Context, _ *entity.CheckRecord) error {
	panic("pretradeCheckRepo.Create not expected; pipeline uses CreateBatch")
}
func (r *pretradeCheckRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.CheckRecord, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *pretradeCheckRepo) GetByGroupID(_ context.Context, _ uuid.UUID) ([]entity.CheckRecord, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *pretradeCheckRepo) List(_ context.Context, _ domain.CheckRecordFilter) ([]entity.CheckRecord, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

// pretradeBreachRepo counts Create calls.
type pretradeBreachRepo struct {
	createCalls int32
	breaches    []entity.Breach
}

func (r *pretradeBreachRepo) Create(_ context.Context, b *entity.Breach) error {
	atomic.AddInt32(&r.createCalls, 1)
	r.breaches = append(r.breaches, *b)
	return nil
}
func (r *pretradeBreachRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Breach, error) {
	return nil, fmt.Errorf("not implemented")
}
func (r *pretradeBreachRepo) List(_ context.Context, _ domain.BreachFilter) ([]entity.Breach, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
func (r *pretradeBreachRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ entity.BreachStatus, _ *uuid.UUID, _ *time.Time) error {
	return fmt.Errorf("not implemented")
}

// ─── builders ────────────────────────────────────────────────────────────────

// minimumTradeBinding returns a ResolvedBinding for amount.minimum_trade
// with min_amount=500 and the supplied binding severity.
func minimumTradeBinding(severity vo.Severity) domain.ResolvedBinding {
	now := time.Now().UTC()
	instanceID := uuid.New()
	return domain.ResolvedBinding{
		Binding: entity.RuleBinding{
			ID:             uuid.New(),
			RuleInstanceID: instanceID,
			Scope:          vo.Scope{Type: vo.ScopeGlobal},
			Severity:       severity,
			Priority:       0,
			EffectiveWindow: vo.EffectiveWindow{
				ValidFrom: now.AddDate(-1, 0, 0), // valid since 1 year ago
			},
			IsActive:  true,
			CreatedBy: uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		RuleInstance: entity.RuleInstance{
			ID:         instanceID,
			RuleTypeID: "amount.minimum_trade",
			Name:       "test-min-trade",
			IsActive:   true,
			EffectiveWindow: vo.EffectiveWindow{
				ValidFrom: now.AddDate(-1, 0, 0),
			},
			CreatedBy: uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		CurrentVersion: entity.RuleInstanceVersion{
			ID:             uuid.New(),
			RuleInstanceID: instanceID,
			VersionNumber:  1,
			Parameters:     json.RawMessage(`{"min_amount": 500}`),
			ChangeReason:   "test",
			CreatedBy:      uuid.New(),
			CreatedAt:      now,
		},
	}
}

// buildPreTradeHandler wires a full pre-trade pipeline with the given bindings
// and returns the handler plus the underlying check/breach repos for assertions.
func buildPreTradeHandler(bindings []domain.ResolvedBinding) (
	h *command.RunPreTradeCheckHandler,
	checkRepo *pretradeCheckRepo,
	breachRepo *pretradeBreachRepo,
) {
	bindingRepo := &pretradeBindingRepo{bindings: bindings}
	checkRepo = &pretradeCheckRepo{}
	breachRepo = &pretradeBreachRepo{}

	// Nil-port Fetcher: safe because amount.minimum_trade declares no DataDependencies.
	fetcher := engine.NewFetcher(nil, nil, nil, nil, nil, nil, nil, nil)
	pipeline := engine.NewPipeline(spi.GlobalRegistry(), bindingRepo, checkRepo, breachRepo, fetcher)
	h = command.NewRunPreTradeCheckHandler(pipeline, spi.GlobalRegistry())
	return
}

// basePreTradeReq returns a valid request with the given quantity and price.
func basePreTradeReq(qty, price float64) command.PreTradeCheckRequest {
	return command.PreTradeCheckRequest{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC),
		Actor:        "test-user",
		Ticker:       "PTT",
		Side:         vo.OrderSideBuy,
		Quantity:     decimal.NewFromFloat(qty),
		Price:        decimal.NewFromFloat(price),
		Fees:         decimal.Zero,
		Currency:     "THB",
	}
}

// ─── verdict scenarios ────────────────────────────────────────────────────────

func TestPreTradeCheck_VerdictScenarios(t *testing.T) {
	t.Parallel()

	// qty=10, price=60 → trade value=600 >= min_amount=500
	aboveMin := basePreTradeReq(10, 60)
	// qty=1, price=1 → trade value=1 < min_amount=500
	belowMin := basePreTradeReq(1, 1)

	tests := []struct {
		name            string
		req             command.PreTradeCheckRequest
		severity        vo.Severity
		wantVerdict     vo.Verdict
		wantBreachCount int
	}{
		{
			name:            "PASS: order value above minimum",
			req:             aboveMin,
			severity:        vo.SeverityBlock,
			wantVerdict:     vo.VerdictPass,
			wantBreachCount: 0,
		},
		{
			name:            "BLOCK: order below minimum, severity=BLOCK",
			req:             belowMin,
			severity:        vo.SeverityBlock,
			wantVerdict:     vo.VerdictBlock,
			wantBreachCount: 1,
		},
		{
			name:            "WARN: order below minimum, severity=WARN caps BLOCK → WARN",
			req:             belowMin,
			severity:        vo.SeverityWarn,
			wantVerdict:     vo.VerdictWarn,
			wantBreachCount: 1,
		},
		{
			name:            "PASS: order below minimum but severity=MONITOR caps BLOCK → PASS",
			req:             belowMin,
			severity:        vo.SeverityMonitor,
			wantVerdict:     vo.VerdictPass,
			wantBreachCount: 0, // MONITOR → PASS → no breach created
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h, _, _ := buildPreTradeHandler([]domain.ResolvedBinding{minimumTradeBinding(tc.severity)})
			resp, err := h.HandleDryRun(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Verdict != tc.wantVerdict {
				t.Errorf("verdict: got %q, want %q", resp.Verdict, tc.wantVerdict)
			}
			if resp.RulesEvaluated != 1 {
				t.Errorf("rules_evaluated: got %d, want 1", resp.RulesEvaluated)
			}
			if len(resp.Breaches) != tc.wantBreachCount {
				t.Errorf("breaches count: got %d, want %d", len(resp.Breaches), tc.wantBreachCount)
			}
		})
	}
}

// ─── breach output detail ─────────────────────────────────────────────────────

func TestPreTradeCheck_BreachDetails(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPreTradeHandler([]domain.ResolvedBinding{minimumTradeBinding(vo.SeverityBlock)})
	resp, err := h.HandleDryRun(context.Background(), basePreTradeReq(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Breaches) != 1 {
		t.Fatalf("expected 1 breach, got %d", len(resp.Breaches))
	}
	b := resp.Breaches[0]

	if b.RuleTypeID != "amount.minimum_trade" {
		t.Errorf("RuleTypeID: got %q, want %q", b.RuleTypeID, "amount.minimum_trade")
	}
	if b.Severity != vo.SeverityBlock {
		t.Errorf("Severity: got %q, want %q", b.Severity, vo.SeverityBlock)
	}
	if b.Verdict != vo.VerdictBlock {
		t.Errorf("Verdict: got %q, want %q", b.Verdict, vo.VerdictBlock)
	}
	if !strings.Contains(b.Message, "below minimum") {
		t.Errorf("Message %q does not mention 'below minimum'", b.Message)
	}
	// MinimumTradeRule.Metadata().Overridable == false — breach cannot be overridden.
	if b.Overridable {
		t.Errorf("Overridable: got true, want false (amount.minimum_trade is non-overridable)")
	}
}

// TestPreTradeCheck_WarnCapBreachDetail verifies that when severity=WARN, the
// breach record carries the effective (capped) verdict and binding severity.
func TestPreTradeCheck_WarnCapBreachDetail(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPreTradeHandler([]domain.ResolvedBinding{minimumTradeBinding(vo.SeverityWarn)})
	resp, err := h.HandleDryRun(context.Background(), basePreTradeReq(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Breaches) != 1 {
		t.Fatalf("expected 1 breach, got %d", len(resp.Breaches))
	}
	b := resp.Breaches[0]

	if b.Severity != vo.SeverityWarn {
		t.Errorf("Severity: got %q, want WARN (binding severity overrides rule default)", b.Severity)
	}
	if b.Verdict != vo.VerdictWarn {
		t.Errorf("Verdict: got %q, want WARN (BLOCK capped by WARN binding)", b.Verdict)
	}
}

// ─── persistence behaviour ────────────────────────────────────────────────────

// TestPreTradeCheck_HandleWritesRecordsAndBreaches verifies that Handle (persist=true)
// writes exactly one check-record batch and one breach when the order blocks.
func TestPreTradeCheck_HandleWritesRecordsAndBreaches(t *testing.T) {
	t.Parallel()

	h, checkRepo, breachRepo := buildPreTradeHandler([]domain.ResolvedBinding{minimumTradeBinding(vo.SeverityBlock)})
	_, err := h.Handle(context.Background(), basePreTradeReq(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&checkRepo.batchCalls); got != 1 {
		t.Errorf("check CreateBatch calls: got %d, want 1", got)
	}
	if got := atomic.LoadInt32(&breachRepo.createCalls); got != 1 {
		t.Errorf("breach Create calls: got %d, want 1", got)
	}
}

// TestPreTradeCheck_DryRunSkipsPersistence verifies that HandleDryRun does NOT
// call CreateBatch or breach.Create even when the order would be blocked.
func TestPreTradeCheck_DryRunSkipsPersistence(t *testing.T) {
	t.Parallel()

	h, checkRepo, breachRepo := buildPreTradeHandler([]domain.ResolvedBinding{minimumTradeBinding(vo.SeverityBlock)})
	_, err := h.HandleDryRun(context.Background(), basePreTradeReq(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&checkRepo.batchCalls); got != 0 {
		t.Errorf("dry-run must not call CreateBatch; got %d calls", got)
	}
	if got := atomic.LoadInt32(&breachRepo.createCalls); got != 0 {
		t.Errorf("dry-run must not call breach.Create; got %d calls", got)
	}
}

// TestPreTradeCheck_NoRules verifies that an order with no applicable bindings
// yields PASS with zero rules evaluated and no breaches.
func TestPreTradeCheck_NoRules(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPreTradeHandler(nil) // no bindings
	resp, err := h.HandleDryRun(context.Background(), basePreTradeReq(1, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Verdict != vo.VerdictPass {
		t.Errorf("verdict: got %q, want PASS when no rules are bound", resp.Verdict)
	}
	if resp.RulesEvaluated != 0 {
		t.Errorf("rules_evaluated: got %d, want 0", resp.RulesEvaluated)
	}
	if len(resp.Breaches) != 0 {
		t.Errorf("breaches: got %d, want 0", len(resp.Breaches))
	}
}
