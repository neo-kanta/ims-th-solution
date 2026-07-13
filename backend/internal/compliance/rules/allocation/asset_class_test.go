package allocation_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/allocation"
)

func maxParams(assetClass string, maxPct float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{
		"asset_class":     assetClass,
		"max_percent_nav": maxPct,
	})
	return spi.NewParameterSet(raw)
}

func minParams(assetClass string, minPct float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{
		"asset_class":     assetClass,
		"min_percent_nav": minPct,
	})
	return spi.NewParameterSet(raw)
}

func allocationBundle(
	holdings []spi.Holding,
	nav float64,
	prices map[string]decimal.Decimal,
	cls []spi.InstrumentClassification,
) spi.DataBundle {
	return spi.DataBundle{
		Positions: &spi.PositionSnapshot{
			PortfolioID: uuid.New(),
			AsOf:        time.Now(),
			Holdings:    holdings,
		},
		NAV:             &spi.NAVSnapshot{NAV: decimal.NewFromFloat(nav)},
		MarketPrices:    &spi.MarketPriceSnapshot{Prices: prices, AsOf: time.Now()},
		Classifications: spi.NewClassificationSnapshot(cls),
	}
}

func TestAssetClassMax_WithinLimit_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("allocation.asset_class_max")
	if !ok {
		t.Fatal("not registered")
	}

	// EQUITY at 40M / 100M NAV = 40% -> limit 60% -> PASS
	bundle := allocationBundle(
		[]spi.Holding{{Ticker: "PTT", Quantity: decimal.NewFromInt(400), MarketValue: decimal.NewFromInt(40_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"PTT": decimal.NewFromInt(100000)},
		[]spi.InstrumentClassification{{Ticker: "PTT", AssetClass: "EQUITY"}},
	)
	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	result, err := r.Evaluate(context.Background(), input, bundle, maxParams("EQUITY", 60))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestAssetClassMax_BuyExceedsLimit_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("allocation.asset_class_max")

	// EQUITY at 55M, proposed buy 10M more -> 65M/100M = 65% > 60% -> BLOCK
	bundle := allocationBundle(
		[]spi.Holding{{Ticker: "PTT", Quantity: decimal.NewFromInt(550), MarketValue: decimal.NewFromInt(55_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"PTT": decimal.NewFromInt(100000), "ADVANC": decimal.NewFromInt(200000)},
		[]spi.InstrumentClassification{
			{Ticker: "PTT", AssetClass: "EQUITY"},
			{Ticker: "ADVANC", AssetClass: "EQUITY"},
		},
	)
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "ADVANC",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(50),
			Price:    decimal.NewFromInt(200000),
		},
	}
	result, err := r.Evaluate(context.Background(), input, bundle, maxParams("EQUITY", 60))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
}

func TestAssetClassMax_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("allocation.asset_class_max")
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, spi.DataBundle{NAV: nil}, maxParams("EQUITY", 60))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}

func TestAssetClassMin_BelowFloor_Warns(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("allocation.asset_class_min")
	if !ok {
		t.Fatal("not registered")
	}

	// FIXED_INCOME at 5M / 100M NAV = 5% -> min 10% -> WARN
	bundle := allocationBundle(
		[]spi.Holding{{Ticker: "TGOV", Quantity: decimal.NewFromInt(50), MarketValue: decimal.NewFromInt(5_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"TGOV": decimal.NewFromInt(100000)},
		[]spi.InstrumentClassification{{Ticker: "TGOV", AssetClass: "FIXED_INCOME"}},
	)
	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	result, err := r.Evaluate(context.Background(), input, bundle, minParams("FIXED_INCOME", 10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictWarn {
		t.Errorf("expected WARN, got %v: %s", result.Verdict, result.Message)
	}
}

func TestAssetClassMin_SellBelowFloor_Warns(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("allocation.asset_class_min")

	// FIXED_INCOME at 12M, proposed sell 5M -> 7M/100M = 7% < 10% -> WARN
	bundle := allocationBundle(
		[]spi.Holding{{Ticker: "TGOV", Quantity: decimal.NewFromInt(120), MarketValue: decimal.NewFromInt(12_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"TGOV": decimal.NewFromInt(100000)},
		[]spi.InstrumentClassification{{Ticker: "TGOV", AssetClass: "FIXED_INCOME"}},
	)
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "TGOV",
			Side:     vo.OrderSideSell,
			Quantity: decimal.NewFromInt(50),
			Price:    decimal.NewFromInt(100000),
		},
	}
	result, err := r.Evaluate(context.Background(), input, bundle, minParams("FIXED_INCOME", 10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictWarn {
		t.Errorf("expected WARN, got %v: %s", result.Verdict, result.Message)
	}
}

func TestAssetClassMin_AtOrAboveFloor_Passes(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("allocation.asset_class_min")

	bundle := allocationBundle(
		[]spi.Holding{{Ticker: "TGOV", Quantity: decimal.NewFromInt(150), MarketValue: decimal.NewFromInt(15_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"TGOV": decimal.NewFromInt(100000)},
		[]spi.InstrumentClassification{{Ticker: "TGOV", AssetClass: "FIXED_INCOME"}},
	)
	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	result, err := r.Evaluate(context.Background(), input, bundle, minParams("FIXED_INCOME", 10))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}
