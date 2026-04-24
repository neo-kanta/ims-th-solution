package cash_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/cash"
)

func getRule() spi.RuleEvaluator {
	r, ok := spi.GlobalRegistry().Get("cash.availability")
	if !ok {
		panic("cash.availability not registered")
	}
	return r
}

func makeParams(bufferPct float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{"min_cash_buffer_pct": bufferPct})
	return spi.NewParameterSet(raw)
}

func navBundle(cashBalance, reservedCash float64) spi.DataBundle {
	return spi.DataBundle{
		NAV: &spi.NAVSnapshot{
			NAV:          decimal.NewFromFloat(100_000_000),
			CashBalance:  decimal.NewFromFloat(cashBalance),
			ReservedCash: decimal.NewFromFloat(reservedCash),
		},
	}
}

func buyInput(qty, price float64) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID:  uuid.New(),
			Ticker:   "AAA",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromFloat(qty),
			Price:    decimal.NewFromFloat(price),
		},
	}
}

func TestCashAvailability_SufficientCash_Passes(t *testing.T) {
	t.Parallel()
	r := getRule()

	// Trade = 1M, available = 10M, buffer 5% of 100M = 5M. After trade: 9M > 5M → PASS.
	bundle := navBundle(10_000_000, 0)
	result, err := r.Evaluate(context.Background(), buyInput(100, 10000), bundle, makeParams(5))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestCashAvailability_InsufficientCash_Blocks(t *testing.T) {
	t.Parallel()
	r := getRule()

	// Trade = 6M, available = 5M → shortfall → BLOCK
	bundle := navBundle(5_000_000, 0)
	result, err := r.Evaluate(context.Background(), buyInput(600, 10000), bundle, makeParams(0))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
	if result.Evidence.ThresholdBreached == nil {
		t.Error("expected ThresholdBreached in evidence")
	}
}

func TestCashAvailability_BufferBreached_Warns(t *testing.T) {
	t.Parallel()
	r := getRule()

	// Trade = 4M, available = 5M, remaining = 1M.
	// Buffer 5% of 100M = 5M. 1M < 5M → WARN (but trade is covered).
	bundle := navBundle(5_000_000, 0)
	result, err := r.Evaluate(context.Background(), buyInput(400, 10000), bundle, makeParams(5))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictWarn {
		t.Errorf("expected WARN for buffer breach, got %v: %s", result.Verdict, result.Message)
	}
}

func TestCashAvailability_SellOrder_Skipped(t *testing.T) {
	t.Parallel()
	r := getRule()

	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "AAA",
			Side:     vo.OrderSideSell,
			Quantity: decimal.NewFromInt(100),
			Price:    decimal.NewFromInt(10000),
		},
	}
	result, err := r.Evaluate(context.Background(), input, spi.DataBundle{}, makeParams(5))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS for sell, got %v", result.Verdict)
	}
}

func TestCashAvailability_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r := getRule()
	result, err := r.Evaluate(context.Background(), buyInput(1, 1), spi.DataBundle{NAV: nil}, makeParams(0))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}

func TestCashAvailability_ReservedCashReducesAvailable(t *testing.T) {
	t.Parallel()
	r := getRule()

	// Balance = 10M, reserved = 8M → available = 2M.
	// Trade = 3M → shortfall → BLOCK
	bundle := navBundle(10_000_000, 8_000_000)
	result, err := r.Evaluate(context.Background(), buyInput(300, 10000), bundle, makeParams(0))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK (reserved reduces available), got %v: %s", result.Verdict, result.Message)
	}
}
