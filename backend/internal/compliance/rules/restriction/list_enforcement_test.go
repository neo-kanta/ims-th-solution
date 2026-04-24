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

// helpers

func listEnfParams(types ...string) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]any{"enforced_types": types})
	return spi.NewParameterSet(raw)
}

func orderWithSide(ticker string, side vo.OrderSide) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID: uuid.New(),
			Ticker:  ticker,
			Side:    side,
		},
	}
}

func getListEnforcementRule(t *testing.T) spi.RuleEvaluator {
	t.Helper()
	r, ok := spi.GlobalRegistry().Get("restriction.list_enforcement")
	if !ok {
		t.Fatal("restriction.list_enforcement not registered")
	}
	return r
}

// ─── BLACKLIST ───────────────────────────────────────────────────────────────

func TestListEnforcement_Blacklist_Blocks(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{
				"BLK": {Ticker: "BLK", Reason: "sanction", Source: "GLOBAL", ListType: "BLACKLIST"},
			},
		},
	}
	res, err := r.Evaluate(context.Background(), orderWithSide("BLK", vo.OrderSideBuy), bundle, listEnfParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("want BLOCK, got %s", res.Verdict)
	}
}

// ─── WHITELIST ───────────────────────────────────────────────────────────────

func TestListEnforcement_Whitelist_Miss_Blocks(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			HasWhitelist: true,
			Whitelisted:  map[string]bool{"ONLY_THIS": true},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("NOT_LISTED", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("whitelist miss must BLOCK, got %s", res.Verdict)
	}
}

func TestListEnforcement_Whitelist_Hit_Passes(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			HasWhitelist: true,
			Whitelisted:  map[string]bool{"OK": true},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("OK", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("whitelisted must pass, got %s", res.Verdict)
	}
}

func TestListEnforcement_Whitelist_NotActive_Passes(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			HasWhitelist: false,
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("ANY", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("no active whitelist → pass, got %s", res.Verdict)
	}
}

// ─── ALERT ───────────────────────────────────────────────────────────────────

func TestListEnforcement_Alert_Warns(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Alerted: map[string]spi.RestrictionEntry{
				"ALRT": {Ticker: "ALRT", Reason: "under investigation", Source: "CONTRACT", ListType: "ALERT"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("ALRT", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictWarn {
		t.Fatalf("alert list must WARN, got %s", res.Verdict)
	}
}

func TestListEnforcement_LegacyGrayListMapsToAlert(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			GrayListed: map[string]spi.RestrictionEntry{
				"GRY": {Ticker: "GRY", Reason: "legacy gray", Source: "GLOBAL", ListType: "GRAYLIST"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("GRY", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictWarn {
		t.Fatalf("legacy gray list must WARN via ALERT path, got %s", res.Verdict)
	}
}

// ─── DISPOSAL ────────────────────────────────────────────────────────────────

func TestListEnforcement_Disposal_BuyBlocked(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Disposal: map[string]spi.RestrictionEntry{
				"DSP": {Ticker: "DSP", Reason: "wind-down", Source: "CONTRACT", ListType: "DISPOSAL"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("DSP", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("disposal BUY must BLOCK, got %s", res.Verdict)
	}
}

func TestListEnforcement_Disposal_SellPasses(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Disposal: map[string]spi.RestrictionEntry{
				"DSP": {Ticker: "DSP", Reason: "wind-down", Source: "CONTRACT", ListType: "DISPOSAL"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("DSP", vo.OrderSideSell), bundle, listEnfParams())
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("disposal SELL must PASS, got %s", res.Verdict)
	}
}

// ─── scoping & fail-closed ──────────────────────────────────────────────────

func TestListEnforcement_NoRestrictionData_FailsClosed(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{Restrictions: nil}
	res, _ := r.Evaluate(context.Background(), orderWithSide("ANY", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("nil restriction snapshot must fail-closed, got %s", res.Verdict)
	}
}

func TestListEnforcement_NoProposedOrder_Passes(t *testing.T) {
	r := getListEnforcementRule(t)
	input := spi.CheckInput{ProposedOrder: nil}
	res, _ := r.Evaluate(context.Background(), input, spi.DataBundle{Restrictions: &spi.RestrictionSnapshot{}}, listEnfParams())
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("no order → pass, got %s", res.Verdict)
	}
}

func TestListEnforcement_EnforcedTypesParameter_Restricts(t *testing.T) {
	r := getListEnforcementRule(t)

	// Ticker is blacklisted AND alerted, but we only enforce ALERT.
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{
				"TK": {Ticker: "TK", Reason: "sanction", Source: "GLOBAL"},
			},
			Alerted: map[string]spi.RestrictionEntry{
				"TK": {Ticker: "TK", Reason: "under review", Source: "GLOBAL"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("TK", vo.OrderSideBuy), bundle, listEnfParams("ALERT"))
	if res.Verdict != vo.VerdictWarn {
		t.Fatalf("when only ALERT enforced, blacklist hit must be ignored; got %s", res.Verdict)
	}

	// Now enforce BLACKLIST only — should block.
	res2, _ := r.Evaluate(context.Background(), orderWithSide("TK", vo.OrderSideBuy), bundle, listEnfParams("BLACKLIST"))
	if res2.Verdict != vo.VerdictBlock {
		t.Fatalf("BLACKLIST enforcement must BLOCK, got %s", res2.Verdict)
	}
}

func TestListEnforcement_PrecedenceBlacklistOverAlert(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{
				"X": {Ticker: "X", Reason: "sanction", Source: "GLOBAL"},
			},
			Alerted: map[string]spi.RestrictionEntry{
				"X": {Ticker: "X", Reason: "also suspect", Source: "CONTRACT"},
			},
		},
	}
	res, _ := r.Evaluate(context.Background(), orderWithSide("X", vo.OrderSideBuy), bundle, listEnfParams())
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("BLACKLIST must take precedence, got %s", res.Verdict)
	}
}

func TestListEnforcement_UnknownParamTypeIgnored(t *testing.T) {
	r := getListEnforcementRule(t)
	bundle := spi.DataBundle{
		Restrictions: &spi.RestrictionSnapshot{
			Blacklisted: map[string]spi.RestrictionEntry{
				"BLK": {Ticker: "BLK", Reason: "sanction", Source: "GLOBAL"},
			},
		},
	}
	// Unknown type → no enforced list types → default (= enforce all) should NOT
	// be re-triggered; instead, unknown yields an EMPTY enforced set and so the
	// blacklist hit is not evaluated.
	res, _ := r.Evaluate(context.Background(), orderWithSide("BLK", vo.OrderSideBuy), bundle, listEnfParams("NOT_A_LIST_TYPE"))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("unknown param filters out all; want PASS, got %s", res.Verdict)
	}
}
