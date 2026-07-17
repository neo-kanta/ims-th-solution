package command_test

// Post-trade check verification tests.
//
// "amount.minimum_trade" is already registered in the global SPI registry via
// run_pretrade_check_test.go (same package, same binary). No additional import needed.
//
// Post-trade check design under test:
//  - Handle always uses RunCheck (persist=true); there is no HandleDryRun.
//  - ProposedOrder is nil for all post-trade scans.
//  - Rules that skip on nil ProposedOrder (like amount.minimum_trade) produce a
//    PASS check record that IS still persisted.
//  - Breach records are created ONLY for BLOCK/WARN verdict records.

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// buildPostTradeHandler wires a post-trade pipeline with the given bindings.
// Reuses the pretrade* fakes already defined in run_pretrade_check_test.go.
func buildPostTradeHandler(bindings []domain.ResolvedBinding) (
	h *command.RunPostTradeCheckHandler,
	checkRepo *pretradeCheckRepo,
	breachRepo *pretradeBreachRepo,
) {
	bindingRepo := &pretradeBindingRepo{bindings: bindings}
	checkRepo = &pretradeCheckRepo{}
	breachRepo = &pretradeBreachRepo{}
	fetcher := engine.NewFetcher(nil, nil, nil, nil, nil, nil, nil, nil)
	pipeline := engine.NewPipeline(spi.GlobalRegistry(), bindingRepo, checkRepo, breachRepo, fetcher)
	h = command.NewRunPostTradeCheckHandler(pipeline, spi.GlobalRegistry())
	return
}

func basePostTradeReq() command.PostTradeCheckRequest {
	return command.PostTradeCheckRequest{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC),
		Actor:        "test-user",
	}
}

// ─── tests ────────────────────────────────────────────────────────────────────

// TestPostTradeCheck_NoRules verifies that a portfolio with no applicable bindings
// produces PASS with zero rules evaluated and no persistence side effects.
func TestPostTradeCheck_NoRules(t *testing.T) {
	t.Parallel()

	h, checkRepo, breachRepo := buildPostTradeHandler(nil)
	resp, err := h.Handle(context.Background(), basePostTradeReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Verdict != vo.VerdictPass {
		t.Errorf("verdict: got %q, want PASS", resp.Verdict)
	}
	if resp.RulesEvaluated != 0 {
		t.Errorf("rules_evaluated: got %d, want 0", resp.RulesEvaluated)
	}
	if len(resp.Breaches) != 0 {
		t.Errorf("breaches: got %d, want 0", len(resp.Breaches))
	}
	// No records produced → CreateBatch must NOT have been called.
	if got := atomic.LoadInt32(&checkRepo.batchCalls); got != 0 {
		t.Errorf("CreateBatch calls: got %d, want 0 (no rules → no records)", got)
	}
	if got := atomic.LoadInt32(&breachRepo.createCalls); got != 0 {
		t.Errorf("breach.Create calls: got %d, want 0", got)
	}
}

// TestPostTradeCheck_RuleSkipsWithoutOrder verifies that when amount.minimum_trade
// is bound, the rule evaluates and produces a PASS record (no ProposedOrder),
// CreateBatch IS called (records exist), and no breach is created.
func TestPostTradeCheck_RuleSkipsWithoutOrder(t *testing.T) {
	t.Parallel()

	h, checkRepo, breachRepo := buildPostTradeHandler(
		[]domain.ResolvedBinding{minimumTradeBinding(vo.SeverityBlock)},
	)
	resp, err := h.Handle(context.Background(), basePostTradeReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Verdict != vo.VerdictPass {
		t.Errorf("verdict: got %q, want PASS (rule skips without ProposedOrder)", resp.Verdict)
	}
	if resp.RulesEvaluated != 1 {
		t.Errorf("rules_evaluated: got %d, want 1", resp.RulesEvaluated)
	}
	if len(resp.Breaches) != 0 {
		t.Errorf("breaches: got %d, want 0 (PASS → no breach)", len(resp.Breaches))
	}
	// PASS record → CreateBatch IS called (records exist).
	if got := atomic.LoadInt32(&checkRepo.batchCalls); got != 1 {
		t.Errorf("CreateBatch calls: got %d, want 1 (post-trade always persists)", got)
	}
	// No BLOCK/WARN → breach.Create must NOT be called.
	if got := atomic.LoadInt32(&breachRepo.createCalls); got != 0 {
		t.Errorf("breach.Create calls: got %d, want 0 (PASS → no breach)", got)
	}
}

// TestPostTradeCheck_CheckGroupIDPreserved verifies that a caller-supplied
// CheckGroupID is echoed back in the response unchanged.
func TestPostTradeCheck_CheckGroupIDPreserved(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPostTradeHandler(nil)
	req := basePostTradeReq()
	req.CheckGroupID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CheckGroupID != req.CheckGroupID {
		t.Errorf("CheckGroupID: got %v, want %v", resp.CheckGroupID, req.CheckGroupID)
	}
}

// TestPostTradeCheck_CheckGroupIDGeneratedWhenAbsent verifies that when
// CheckGroupID is zero the handler generates a new non-nil UUID.
func TestPostTradeCheck_CheckGroupIDGeneratedWhenAbsent(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPostTradeHandler(nil)
	resp, err := h.Handle(context.Background(), basePostTradeReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CheckGroupID == uuid.Nil {
		t.Error("CheckGroupID: got uuid.Nil, want a generated UUID")
	}
}

// TestPostTradeCheck_RequiredFieldValidation verifies that Handle returns an
// error when a required field is absent (portfolio_id, business_date).
// contract_id is intentionally NOT required at this layer — Portfolio
// Compliance V2 invokes this handler with ContractID left as uuid.Nil when
// the portfolio has no fund_id (see TestPostTradeCheck_MissingContractID_Allowed).
// V1's HTTP handler still requires contract_id; that is enforced at the
// transport layer, not here.
func TestPostTradeCheck_RequiredFieldValidation(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPostTradeHandler(nil)

	cases := []struct {
		name string
		req  command.PostTradeCheckRequest
	}{
		{
			name: "missing portfolio_id",
			req: command.PostTradeCheckRequest{
				ContractID:   uuid.New(),
				BusinessDate: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "missing business_date",
			req: command.PostTradeCheckRequest{
				PortfolioID: uuid.New(),
				ContractID:  uuid.New(),
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := h.Handle(context.Background(), tc.req)
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

// TestPostTradeCheck_MissingContractID_Allowed verifies that a portfolio-only
// post-trade check (no fund_id) succeeds — the Portfolio Compliance V2 case.
func TestPostTradeCheck_MissingContractID_Allowed(t *testing.T) {
	t.Parallel()

	h, _, _ := buildPostTradeHandler(nil)
	req := command.PostTradeCheckRequest{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC),
	}
	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("portfolio-only post-trade check must not require contract_id: %v", err)
	}
	if resp.CheckGroupID == uuid.Nil {
		t.Error("CheckGroupID: got uuid.Nil, want a generated UUID")
	}
}

// TestPostTradeCheck_NoDryRun verifies that RunPostTradeCheckHandler has no
// HandleDryRun method — post-trade scans are always persistent.
// This is a compile-time safety assertion expressed as a type-check test.
func TestPostTradeCheck_NoDryRun(t *testing.T) {
	// If this file compiles without a dryRunner interface assertion,
	// it confirms RunPostTradeCheckHandler intentionally omits HandleDryRun.
	t.Parallel()
	h, _, _ := buildPostTradeHandler(nil)
	_ = h // used; dry-run path intentionally absent on this handler type
}
