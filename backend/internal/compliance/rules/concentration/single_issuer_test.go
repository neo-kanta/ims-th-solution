package concentration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	// trigger init() self-registration
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/concentration"
)

func newRule() spi.RuleEvaluator {
	e, ok := spi.GlobalRegistry().Get("concentration.single_issuer")
	if !ok {
		panic("concentration.single_issuer not registered")
	}
	return e
}

func params(maxPct float64, exemptGov bool) spi.ParameterSet {
	raw, _ := json.Marshal(map[string]interface{}{
		"max_pct":           maxPct,
		"exempt_government": exemptGov,
	})
	return spi.NewParameterSet(raw)
}

func makeBundle(
	holdings []spi.Holding,
	nav decimal.Decimal,
	prices map[string]decimal.Decimal,
	cls []spi.InstrumentClassification,
) spi.DataBundle {
	snap := &spi.PositionSnapshot{Holdings: holdings, PortfolioID: uuid.New(), AsOf: time.Now()}
	navSnap := &spi.NAVSnapshot{NAV: nav}
	priceSnap := &spi.MarketPriceSnapshot{Prices: prices}
	clsSnap := spi.NewClassificationSnapshot(cls)
	return spi.DataBundle{
		Positions:       snap,
		NAV:             navSnap,
		MarketPrices:    priceSnap,
		Classifications: clsSnap,
	}
}

func buyOrder(ticker string, qty, price float64) *spi.ProposedOrder {
	return &spi.ProposedOrder{
		OrderID:  uuid.New(),
		Ticker:   ticker,
		Side:     vo.OrderSideBuy,
		Quantity: decimal.NewFromFloat(qty),
		Price:    decimal.NewFromFloat(price),
	}
}

func TestSingleIssuer_NoBreach(t *testing.T) {
	t.Parallel()
	r := newRule()

	// 5% concentration, limit 10% → PASS
	bundle := makeBundle(
		[]spi.Holding{{Ticker: "AAA", Quantity: decimal.NewFromInt(1000), MarketValue: decimal.NewFromInt(5_000_000)}},
		decimal.NewFromInt(100_000_000),
		map[string]decimal.Decimal{"AAA": decimal.NewFromInt(5000)},
		[]spi.InstrumentClassification{{Ticker: "AAA", Issuer: "CORP_A", ParentEntity: "CORP_A"}},
	)
	input := spi.CheckInput{
		PortfolioID: uuid.New(),
		ContractID:  uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: buyOrder("AAA", 100, 5000), // adds 500_000 → 5.5M / 100M = 5.5%
	}

	result, err := r.Evaluate(context.Background(), input, bundle, params(10, false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSingleIssuer_Breach_BuyWouldExceed(t *testing.T) {
	t.Parallel()
	r := newRule()

	// Currently at 8% (8M/100M). Proposed buy adds 3M → 11% > 10% limit.
	bundle := makeBundle(
		[]spi.Holding{{Ticker: "BBB", Quantity: decimal.NewFromInt(800), MarketValue: decimal.NewFromInt(8_000_000)}},
		decimal.NewFromInt(100_000_000),
		map[string]decimal.Decimal{"BBB": decimal.NewFromInt(10000)},
		[]spi.InstrumentClassification{{Ticker: "BBB", Issuer: "CORP_B", ParentEntity: "CORP_B"}},
	)
	input := spi.CheckInput{
		PortfolioID:   uuid.New(),
		ContractID:    uuid.New(),
		BusinessDate:  time.Now(),
		ProposedOrder: buyOrder("BBB", 300, 10000), // 3M additional → 11M/100M = 11%
	}

	result, err := r.Evaluate(context.Background(), input, bundle, params(10, false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK, got %v: %s", result.Verdict, result.Message)
	}
	if result.Evidence.ThresholdBreached == nil {
		t.Error("expected ThresholdBreached in evidence")
	}
}

func TestSingleIssuer_ParentEntityGrouping(t *testing.T) {
	t.Parallel()
	r := newRule()

	// Two tickers with same parent entity → grouped together.
	// CORP_X: 6M (BBB) + proposed buy 5M (CCC) = 11M / 100M = 11% > 10%
	bundle := makeBundle(
		[]spi.Holding{
			{Ticker: "BBB", Quantity: decimal.NewFromInt(600), MarketValue: decimal.NewFromInt(6_000_000)},
		},
		decimal.NewFromInt(100_000_000),
		map[string]decimal.Decimal{"BBB": decimal.NewFromInt(10000), "CCC": decimal.NewFromInt(10000)},
		[]spi.InstrumentClassification{
			{Ticker: "BBB", Issuer: "BBB_INC", ParentEntity: "CORP_X"},
			{Ticker: "CCC", Issuer: "CCC_INC", ParentEntity: "CORP_X"}, // same parent
		},
	)
	input := spi.CheckInput{
		PortfolioID:   uuid.New(),
		ContractID:    uuid.New(),
		BusinessDate:  time.Now(),
		ProposedOrder: buyOrder("CCC", 500, 10000), // 5M additional → CORP_X = 11M/100M = 11%
	}

	result, err := r.Evaluate(context.Background(), input, bundle, params(10, false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK due to parent entity grouping, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSingleIssuer_ExemptGovernment(t *testing.T) {
	t.Parallel()
	r := newRule()

	// Government bond at 30% — exempt, so should PASS.
	bundle := makeBundle(
		[]spi.Holding{{Ticker: "LB32DA", Quantity: decimal.NewFromInt(3000), MarketValue: decimal.NewFromInt(30_000_000)}},
		decimal.NewFromInt(100_000_000),
		map[string]decimal.Decimal{"LB32DA": decimal.NewFromInt(10000)},
		[]spi.InstrumentClassification{{Ticker: "LB32DA", Issuer: "MOF", ParentEntity: "MOF", IsGovernment: true}},
	)
	input := spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
	}

	result, err := r.Evaluate(context.Background(), input, bundle, params(10, true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != vo.VerdictPass {
		t.Errorf("expected PASS for exempt government bond, got %v: %s", result.Verdict, result.Message)
	}
}

func TestSingleIssuer_NilNAV_FailClosed(t *testing.T) {
	t.Parallel()
	r := newRule()

	bundle := spi.DataBundle{Positions: &spi.PositionSnapshot{}}
	// NAV is nil → fail-closed → BLOCK
	result, err := r.Evaluate(context.Background(), spi.CheckInput{}, bundle, params(10, false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != vo.VerdictBlock {
		t.Errorf("expected BLOCK on nil NAV, got %v", result.Verdict)
	}
}
