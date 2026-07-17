package exposure_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/exposure"
)

func aumParams(maxPct float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{"max_percent_aum": maxPct})
	return spi.NewParameterSet(raw)
}

func TestMaxOrderPercentAUM_WithinLimit_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("exposure.max_order_percent_aum")
	if !ok {
		t.Fatal("not registered")
	}

	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "PTT",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(100),
			Price:    decimal.NewFromInt(50_000), // 5,000,000 trade value
		},
	}
	bundle := spi.DataBundle{NAV: &spi.NAVSnapshot{NAV: decimal.NewFromInt(100_000_000)}}
	result, err := r.Evaluate(context.Background(), input, bundle, aumParams(10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMaxOrderPercentAUM_ExceedsLimit_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("exposure.max_order_percent_aum")

	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "PTT",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(300),
			Price:    decimal.NewFromInt(50_000), // 15,000,000 trade value = 15% of 100M
		},
	}
	bundle := spi.DataBundle{NAV: &spi.NAVSnapshot{NAV: decimal.NewFromInt(100_000_000)}}
	result, err := r.Evaluate(context.Background(), input, bundle, aumParams(10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMaxOrderPercentAUM_NoProposedOrder_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("exposure.max_order_percent_aum")

	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	bundle := spi.DataBundle{NAV: &spi.NAVSnapshot{NAV: decimal.NewFromInt(100_000_000)}}
	result, err := r.Evaluate(context.Background(), input, bundle, aumParams(10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS (no order), got %v", result.Verdict)
	}
}

func TestMaxOrderPercentAUM_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("exposure.max_order_percent_aum")

	input := spi.CheckInput{
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "PTT",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(1),
			Price:    decimal.NewFromInt(1),
		},
	}
	result, err := r.Evaluate(context.Background(), input, spi.DataBundle{NAV: nil}, aumParams(10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}
