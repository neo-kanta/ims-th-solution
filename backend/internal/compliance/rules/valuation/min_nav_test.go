package valuation_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/valuation"
)

func navParams(minNAV float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{"min_nav": minNAV})
	return spi.NewParameterSet(raw)
}

func TestMinNAV_AboveFloor_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("valuation.min_nav")
	if !ok {
		t.Fatal("not registered")
	}
	bundle := spi.DataBundle{NAV: &spi.NAVSnapshot{NAV: decimal.NewFromInt(50_000_000)}}
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, bundle, navParams(10_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMinNAV_BelowFloor_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("valuation.min_nav")
	bundle := spi.DataBundle{NAV: &spi.NAVSnapshot{NAV: decimal.NewFromInt(5_000_000)}}
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, bundle, navParams(10_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMinNAV_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("valuation.min_nav")
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, spi.DataBundle{NAV: nil}, navParams(10_000_000))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}
