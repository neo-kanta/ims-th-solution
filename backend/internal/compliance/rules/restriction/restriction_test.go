package restriction_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/restriction"
)

func emptyParams() spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{})
	return spi.NewParameterSet(raw)
}

func orderFor(ticker string) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID: uuid.New(),
			Ticker:  ticker,
			Side:    vo.OrderSideBuy,
		},
	}
}

// ─── blacklist tests ─────────────────────────────────────────────────────────

func TestBlacklist_TickerOnBlacklist_Blocks(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("restriction.blacklist")
	if !ok {
		t.Fatal("restriction.blacklist not registered")
	}

	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{
				"BLOCKED_CO": {Ticker: "BLOCKED_CO", Reason: "regulatory sanction", Source: "GLOBAL"},
			},
			Whitelisted: map[string]bool{},
		},
	}
	result, err := r.Evaluate(context.Background(), orderFor("BLOCKED_CO"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v", result.Verdict)
	}
}

func TestBlacklist_CleanTicker_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("restriction.blacklist")
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{},
			Whitelisted: map[string]bool{},
		},
	}
	result, err := r.Evaluate(context.Background(), orderFor("CLEAN_CO"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v", result.Verdict)
	}
}

func TestBlacklist_NilRestrictions_FailClosed(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("restriction.blacklist")
	bundle := spi.DataBundle{Restrictions: nil}
	result, err := r.Evaluate(context.Background(), orderFor("ANY"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	// Fail-closed: no data → BLOCK
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil restrictions, got %v", result.Verdict)
	}
}

func TestBlacklist_NoProposedOrder_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("restriction.blacklist")
	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	result, err := r.Evaluate(context.Background(), input, spi.DataBundle{}, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS for no order, got %v", result.Verdict)
	}
}

// ─── whitelist tests ──────────────────────────────────────────────────────────

func TestWhitelist_NoWhitelistConfigured_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("restriction.whitelist")
	if !ok {
		t.Fatal("restriction.whitelist not registered")
	}
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{HasWhitelist: false, Whitelisted: map[string]bool{}},
	}
	result, err := r.Evaluate(context.Background(), orderFor("ANY"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS (no whitelist), got %v", result.Verdict)
	}
}

func TestWhitelist_TickerOnWhitelist_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("restriction.whitelist")
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			HasWhitelist: true,
			Whitelisted:  map[string]bool{"APPROVED_CO": true},
		},
	}
	result, err := r.Evaluate(context.Background(), orderFor("APPROVED_CO"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v", result.Verdict)
	}
}

func TestWhitelist_TickerNotOnWhitelist_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("restriction.whitelist")
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			HasWhitelist: true,
			Whitelisted:  map[string]bool{"APPROVED_CO": true},
		},
	}
	result, err := r.Evaluate(context.Background(), orderFor("NOT_APPROVED"), bundle, emptyParams())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v", result.Verdict)
	}
}
