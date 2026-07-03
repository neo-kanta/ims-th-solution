package amount_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/amount"
)

func amountRule(t *testing.T) spi.RuleEvaluator {
	t.Helper()
	r, ok := spi.GlobalRegistry().Get("amount.minimum_trade")
	if !ok {
		t.Fatal("amount.minimum_trade not registered")
	}
	return r
}

func minAmountParams(t *testing.T, body map[string]any) spi.ParameterSet {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return spi.NewParameterSet(raw)
}

func order(qty, price string) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID:  uuid.New(),
			Ticker:   "ABC",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.RequireFromString(qty),
			Price:    decimal.RequireFromString(price),
			Currency: "THB",
		},
	}
}

func TestMinimumTrade_AboveMinimumPasses(t *testing.T) {
	t.Parallel()
	result, err := amountRule(t).Evaluate(
		context.Background(),
		order("100", "20"),
		spi.DataBundle{},
		minAmountParams(t, map[string]any{"min_amount": "1000", "currency": "THB"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Fatalf("expected PASS, got %s: %s", result.Verdict, result.Message)
	}
}

func TestMinimumTrade_EqualMinimumPasses(t *testing.T) {
	t.Parallel()
	result, err := amountRule(t).Evaluate(
		context.Background(),
		order("100", "10"),
		spi.DataBundle{},
		minAmountParams(t, map[string]any{"min_amount": "1000"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Fatalf("expected boundary PASS, got %s", result.Verdict)
	}
}

func TestMinimumTrade_BelowMinimumBlocks(t *testing.T) {
	t.Parallel()
	result, err := amountRule(t).Evaluate(
		context.Background(),
		order("99", "10"),
		spi.DataBundle{},
		minAmountParams(t, map[string]any{"min_amount": "1000"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Fatalf("expected BLOCK, got %s", result.Verdict)
	}
	if result.Evidence.ThresholdBreached == nil {
		t.Fatal("expected threshold evidence")
	}
}

func TestMinimumTrade_InvalidParamsFailClosedThroughError(t *testing.T) {
	t.Parallel()
	_, err := amountRule(t).Evaluate(
		context.Background(),
		order("100", "10"),
		spi.DataBundle{},
		minAmountParams(t, map[string]any{"min_amount": "0"}),
	)
	if err == nil {
		t.Fatal("expected invalid min_amount error")
	}
}

func TestMinimumTrade_PerSideOverride(t *testing.T) {
	t.Parallel()
	input := order("100", "15")
	input.ProposedOrder.Side = vo.OrderSideSell
	result, err := amountRule(t).Evaluate(
		context.Background(),
		input,
		spi.DataBundle{},
		minAmountParams(t, map[string]any{
			"min_amount": "1000",
			"per_side": map[string]any{
				"SELL": "2000",
			},
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Fatalf("expected SELL override to block, got %s", result.Verdict)
	}
}
