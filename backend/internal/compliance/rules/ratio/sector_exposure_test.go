package ratio_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/ratio"
)

func sectorParams(sector string, maxPct float64) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{
		"sector":  sector,
		"max_pct": maxPct,
	})
	return spi.NewParameterSet(raw)
}

func sectorBundle(
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
		NAV: &spi.NAVSnapshot{NAV: decimal.NewFromFloat(nav)},
		MarketPrices: &spi.MarketPriceSnapshot{
			Prices: prices,
			AsOf:   time.Now(),
		},
		Classifications: spi.NewClassificationSnapshot(cls),
	}
}

func TestSectorExposure_WithinLimit_Passes(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("ratio.sector_exposure")
	if !ok {
		t.Fatal("not registered")
	}

	// FINANCIALS sector at 20% → limit 30% → PASS
	bundle := sectorBundle(
		[]spi.Holding{{Ticker: "BBL", Quantity: decimal.NewFromInt(200), MarketValue: decimal.NewFromInt(20_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"BBL": decimal.NewFromInt(100000)},
		[]spi.InstrumentClassification{{Ticker: "BBL", Sector: "FINANCIALS"}},
	)
	input := spi.CheckInput{PortfolioID: uuid.New(), BusinessDate: time.Now()}
	result, err := r.Evaluate(context.Background(), input, bundle, sectorParams("FINANCIALS", 30))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSectorExposure_BuyExceedsLimit_Blocks(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("ratio.sector_exposure")

	// ENERGY at 20M. Proposed buy 15M more → 35M/100M = 35% > 30% → BLOCK
	bundle := sectorBundle(
		[]spi.Holding{{Ticker: "PTT", Quantity: decimal.NewFromInt(200), MarketValue: decimal.NewFromInt(20_000_000)}},
		100_000_000,
		map[string]decimal.Decimal{"PTT": decimal.NewFromInt(100000), "TOP": decimal.NewFromInt(10000)},
		[]spi.InstrumentClassification{
			{Ticker: "PTT", Sector: "ENERGY"},
			{Ticker: "TOP", Sector: "ENERGY"},
		},
	)
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "TOP",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(1500),
			Price:    decimal.NewFromInt(10000),
		},
	}
	result, err := r.Evaluate(context.Background(), input, bundle, sectorParams("ENERGY", 30))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSectorExposure_WrongSector_Ignored(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("ratio.sector_exposure")

	// Proposed buy is in TECHNOLOGY, but this rule checks ENERGY → PASS
	bundle := sectorBundle(nil, 100_000_000, map[string]decimal.Decimal{}, nil)
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			Ticker:   "ADVANC",
			Side:     vo.OrderSideBuy,
			Quantity: decimal.NewFromInt(5000),
			Price:    decimal.NewFromInt(10000),
		},
	}
	bundle.Classifications = spi.NewClassificationSnapshot([]spi.InstrumentClassification{
		{Ticker: "ADVANC", Sector: "TECHNOLOGY"},
	})
	result, err := r.Evaluate(context.Background(), input, bundle, sectorParams("ENERGY", 30))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS (wrong sector), got %v", result.Verdict)
	}
}

func TestSectorExposure_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r, _ := spi.GlobalRegistry().Get("ratio.sector_exposure")
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, spi.DataBundle{NAV: nil}, sectorParams("ENERGY", 30))
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}
