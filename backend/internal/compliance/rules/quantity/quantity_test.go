package quantity_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/quantity"
)

func sellInput(ticker string, qty float64) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID:  uuid.New(),
			Ticker:   ticker,
			Side:     vo.OrderSideSell,
			Quantity: decimal.NewFromFloat(qty),
			Price:    decimal.NewFromInt(10000),
		},
	}
}

func posBundle(ticker string, held, pendingSell float64) spi.DataBundle {
	holdings := []spi.Holding{{Ticker: ticker, Quantity: decimal.NewFromFloat(held)}}
	var pending []spi.PendingOrderInfo
	if pendingSell > 0 {
		pending = append(pending, spi.PendingOrderInfo{
			Ticker:   ticker,
			Side:     string(vo.OrderSideSell),
			Quantity: decimal.NewFromFloat(pendingSell),
		})
	}
	return spi.DataBundle{
		Positions: &spi.PositionSnapshot{
			PortfolioID:   uuid.New(),
			AsOf:          time.Now(),
			Holdings:      holdings,
			PendingOrders: pending,
		},
	}
}

func noAllowShort() spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{"allow_short_sell": false})
	return spi.NewParameterSet(raw)
}

// ─── sell quantity tests ──────────────────────────────────────────────────────

func TestSellQuantity_WithinHolding_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("quantity.sell_available")
	if !ok {
		t.Fatal("not registered")
	}
	result, err := r.Evaluate(context.Background(), sellInput("AAA", 500), posBundle("AAA", 1000, 0), noAllowShort())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSellQuantity_ExceedsHolding_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.sell_available")
	result, err := r.Evaluate(context.Background(), sellInput("AAA", 1001), posBundle("AAA", 1000, 0), noAllowShort())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v", result.Verdict)
	}
}

func TestSellQuantity_PendingSellsReduceAvailable(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.sell_available")
	// Held 1000, pending sell 200 → available 800. Proposed 900 → BLOCK
	result, err := r.Evaluate(context.Background(), sellInput("AAA", 900), posBundle("AAA", 1000, 200), noAllowShort())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK (pending sell reduces available), got %v", result.Verdict)
	}
}

func TestSellQuantity_BuyOrder_Skipped(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.sell_available")
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker: "AAA", Side: vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(9999),
		},
	}
	result, err := r.Evaluate(context.Background(), input, spi.DataBundle{}, noAllowShort())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS for buy, got %v", result.Verdict)
	}
}

// ─── min trading unit tests ───────────────────────────────────────────────────

func mtuParams(defaultLot int, overrides map[string]int64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{
		"default_lot_size": defaultLot,
		"overrides":        overrides,
	})
	return spi.NewParameterSet(raw)
}

func buyQty(ticker string, qty int64) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   ticker,
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(qty),
			Price:    decimal.NewFromInt(100),
		},
	}
}

func TestMinTradingUnit_ValidMultiple_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("quantity.min_trading_unit")
	if !ok {
		t.Fatal("not registered")
	}
	// 500 is a multiple of 100
	result, err := r.Evaluate(context.Background(), buyQty("AAA", 500), spi.DataBundle{}, mtuParams(100, nil))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMinTradingUnit_InvalidLot_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.min_trading_unit")
	// 501 is not a multiple of 100
	result, err := r.Evaluate(context.Background(), buyQty("AAA", 501), spi.DataBundle{}, mtuParams(100, nil))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
}

func TestMinTradingUnit_PerTickerOverride(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.min_trading_unit")
	// Default lot 100, but BBB has lot size 1000.
	// 500 is multiple of 100 (default) but NOT of 1000.
	params := mtuParams(100, map[string]int64{"BBB": 1000})
	result, err := r.Evaluate(context.Background(), buyQty("BBB", 500), spi.DataBundle{}, params)
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK (override lot 1000), got %v", result.Verdict)
	}
}

func TestMinTradingUnit_ExactLotSize_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("quantity.min_trading_unit")
	// Exactly 1 lot
	result, err := r.Evaluate(context.Background(), buyQty("AAA", 100), spi.DataBundle{}, mtuParams(100, nil))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v", result.Verdict)
	}
}
