package allocation_test

import (
	"context"
	"encoding/json"
	"strings"
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

func TestAssetClassRules_MissingClassification_IsUnavailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ruleTypeID string
		params     spi.ParameterSet
		holdings   []spi.Holding
		order      *spi.ProposedOrder
		classes    []spi.InstrumentClassification
		missing    []string
	}{
		{
			name:       "max unclassified existing holding",
			ruleTypeID: "allocation.asset_class_max",
			params:     maxParams("EQUITY", 60),
			holdings: []spi.Holding{
				{Ticker: "UNKNOWN-HOLDING", MarketValue: decimal.NewFromInt(20_000_000)},
			},
			missing: []string{"UNKNOWN-HOLDING"},
		},
		{
			name:       "min unclassified existing holding",
			ruleTypeID: "allocation.asset_class_min",
			params:     minParams("FIXED_INCOME", 10),
			holdings: []spi.Holding{
				{Ticker: "UNCLASSIFIED-BOND", MarketValue: decimal.NewFromInt(20_000_000)},
			},
			missing: []string{"UNCLASSIFIED-BOND"},
		},
		{
			name:       "max unclassified proposed instrument",
			ruleTypeID: "allocation.asset_class_max",
			params:     maxParams("EQUITY", 60),
			order: &spi.ProposedOrder{
				Ticker: "NEW-EQUITY", Side: vo.OrderSideBuy,
				Quantity: decimal.NewFromInt(100), Price: decimal.NewFromInt(10),
			},
			missing: []string{"NEW-EQUITY"},
		},
		{
			name:       "min unclassified proposed instrument",
			ruleTypeID: "allocation.asset_class_min",
			params:     minParams("FIXED_INCOME", 10),
			order: &spi.ProposedOrder{
				Ticker: "NEW-BOND", Side: vo.OrderSideBuy,
				Quantity: decimal.NewFromInt(100), Price: decimal.NewFromInt(10),
			},
			missing: []string{"NEW-BOND"},
		},
		{
			name:       "max partial classification snapshot",
			ruleTypeID: "allocation.asset_class_max",
			params:     maxParams("EQUITY", 60),
			holdings: []spi.Holding{
				{Ticker: "CLASSIFIED", MarketValue: decimal.NewFromInt(10_000_000)},
				{Ticker: "MISSING", MarketValue: decimal.NewFromInt(10_000_000)},
			},
			classes: []spi.InstrumentClassification{{Ticker: "CLASSIFIED", AssetClass: "EQUITY"}},
			missing: []string{"MISSING"},
		},
		{
			name:       "min partial classification snapshot",
			ruleTypeID: "allocation.asset_class_min",
			params:     minParams("FIXED_INCOME", 10),
			holdings: []spi.Holding{
				{Ticker: "CLASSIFIED", MarketValue: decimal.NewFromInt(10_000_000)},
				{Ticker: "BLANK", MarketValue: decimal.NewFromInt(10_000_000)},
			},
			classes: []spi.InstrumentClassification{
				{Ticker: "CLASSIFIED", AssetClass: "FIXED_INCOME"},
				{Ticker: "BLANK", AssetClass: ""},
			},
			missing: []string{"BLANK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rule, ok := spi.GlobalRegistry().Get(tt.ruleTypeID)
			if !ok {
				t.Fatalf("rule %s not registered", tt.ruleTypeID)
			}
			bundle := allocationBundle(tt.holdings, 100_000_000, nil, tt.classes)
			bundle.PortfolioMeta = &spi.PortfolioMetadata{PortfolioType: "LIVE"}
			result, err := rule.Evaluate(context.Background(), spi.CheckInput{
				PortfolioID:   uuid.New(),
				BusinessDate:  time.Now(),
				ProposedOrder: tt.order,
			}, bundle, tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Status != vo.ComplianceStatusUnavailable {
				t.Fatalf("status = %q, want %q", result.Status, vo.ComplianceStatusUnavailable)
			}
			if result.Verdict != vo.VerdictBlock {
				t.Fatalf("verdict = %q, want BLOCK", result.Verdict)
			}
			if result.Evidence.Metrics["reason"] != "MISSING_ASSET_CLASSIFICATION" {
				t.Fatalf("reason = %q", result.Evidence.Metrics["reason"])
			}
			for _, ticker := range tt.missing {
				if !strings.Contains(result.Evidence.References["missing_instruments"], ticker) {
					t.Fatalf("missing_instruments %q does not identify %q", result.Evidence.References["missing_instruments"], ticker)
				}
				if !strings.Contains(result.Message, ticker) {
					t.Fatalf("message %q does not identify %q", result.Message, ticker)
				}
			}
		})
	}
}

func TestAssetClassRules_NonLiveMissingClassificationPreservesPriorEvaluation(t *testing.T) {
	t.Parallel()
	for _, portfolioType := range []string{"SIMULATION", "MODEL"} {
		portfolioType := portfolioType
		t.Run(portfolioType, func(t *testing.T) {
			t.Parallel()
			tests := []struct {
				name       string
				ruleTypeID string
				params     spi.ParameterSet
				holdings   []spi.Holding
				classes    []spi.InstrumentClassification
				want       vo.Verdict
			}{
				{
					name:       "max ignores missing classification as before",
					ruleTypeID: "allocation.asset_class_max",
					params:     maxParams("EQUITY", 60),
					holdings:   []spi.Holding{{Ticker: "UNKNOWN", MarketValue: decimal.NewFromInt(80_000_000)}},
					want:       vo.VerdictPass,
				},
				{
					name:       "min retains prior warning math",
					ruleTypeID: "allocation.asset_class_min",
					params:     minParams("FIXED_INCOME", 10),
					holdings:   []spi.Holding{{Ticker: "UNKNOWN", MarketValue: decimal.NewFromInt(20_000_000)}},
					want:       vo.VerdictWarn,
				},
				{
					name:       "valid classified breach still blocks",
					ruleTypeID: "allocation.asset_class_max",
					params:     maxParams("EQUITY", 60),
					holdings:   []spi.Holding{{Ticker: "PTT", MarketValue: decimal.NewFromInt(80_000_000)}},
					classes:    []spi.InstrumentClassification{{Ticker: "PTT", AssetClass: "EQUITY"}},
					want:       vo.VerdictBlock,
				},
			}
			for _, tc := range tests {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					rule, ok := spi.GlobalRegistry().Get(tc.ruleTypeID)
					if !ok {
						t.Fatalf("rule %s not registered", tc.ruleTypeID)
					}
					bundle := allocationBundle(tc.holdings, 100_000_000, nil, tc.classes)
					bundle.PortfolioMeta = &spi.PortfolioMetadata{PortfolioType: portfolioType}
					result, err := rule.Evaluate(context.Background(), spi.CheckInput{
						PortfolioID:  uuid.New(),
						BusinessDate: time.Now(),
					}, bundle, tc.params)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if result.Status == vo.ComplianceStatusUnavailable {
						t.Fatal("non-LIVE missing classification must not introduce the LIVE-only unavailable status")
					}
					if result.Verdict != tc.want {
						t.Fatalf("verdict = %q, want prior policy %q", result.Verdict, tc.want)
					}
				})
			}
		})
	}
}
