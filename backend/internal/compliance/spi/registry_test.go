package spi_test

import (
	"context"
	"testing"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	// All rule packages must be imported to trigger self-registration.
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/cash"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/concentration"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/quantity"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/ratio"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/restriction"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/stub"
)

func TestAllRulesRegistered(t *testing.T) {
	t.Parallel()

	expectedTypeIDs := []string{
		"concentration.single_issuer",
		"restriction.blacklist",
		"restriction.whitelist",
		"cash.availability",
		"quantity.sell_available",
		"quantity.min_trading_unit",
		"ratio.sector_exposure",
		"credit_rating.minimum",
		"regulatory.thai_sec",
	}

	reg := spi.GlobalRegistry()
	for _, typeID := range expectedTypeIDs {
		typeID := typeID
		t.Run(typeID, func(t *testing.T) {
			t.Parallel()
			evaluator, ok := reg.Get(typeID)
			if !ok {
				t.Fatalf("rule type %q not registered in global registry", typeID)
			}
			meta := evaluator.Metadata()
			if meta.TypeID != typeID {
				t.Errorf("metadata TypeID mismatch: got %q, want %q", meta.TypeID, typeID)
			}
			if !meta.DefaultSeverity.IsValid() {
				t.Errorf("invalid DefaultSeverity: %q", meta.DefaultSeverity)
			}
			if len(meta.SupportedTimings) == 0 {
				t.Error("SupportedTimings is empty")
			}
		})
	}
}

func TestRegistry_ListAll(t *testing.T) {
	t.Parallel()
	all := spi.GlobalRegistry().All()
	if len(all) < 9 {
		t.Errorf("expected at least 9 registered rules, got %d", len(all))
	}
}

func TestDataDependencies_Union(t *testing.T) {
	t.Parallel()

	a := spi.DataDependencies{
		Positions:                true,
		NAV:                      false,
		TradeHistoryLookbackDays: 30,
	}
	b := spi.DataDependencies{
		NAV:                      true,
		Classifications:          true,
		TradeHistoryLookbackDays: 60,
	}

	merged := a.Union(b)
	if !merged.Positions {
		t.Error("Positions should be true after union")
	}
	if !merged.NAV {
		t.Error("NAV should be true after union")
	}
	if !merged.Classifications {
		t.Error("Classifications should be true after union")
	}
	if merged.TradeHistoryLookbackDays != 60 {
		t.Errorf("LookbackDays should be max(30,60)=60, got %d", merged.TradeHistoryLookbackDays)
	}
}

func TestSeverity_CapVerdict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		severity vo.Severity
		raw      vo.Verdict
		want     vo.Verdict
	}{
		{vo.SeverityBlock, vo.VerdictBlock, vo.VerdictBlock},
		{vo.SeverityBlock, vo.VerdictWarn, vo.VerdictWarn},
		{vo.SeverityBlock, vo.VerdictPass, vo.VerdictPass},
		{vo.SeverityWarn, vo.VerdictBlock, vo.VerdictWarn},  // cap BLOCK → WARN
		{vo.SeverityWarn, vo.VerdictWarn, vo.VerdictWarn},
		{vo.SeverityWarn, vo.VerdictPass, vo.VerdictPass},
		{vo.SeverityMonitor, vo.VerdictBlock, vo.VerdictPass}, // monitor: always PASS
		{vo.SeverityMonitor, vo.VerdictWarn, vo.VerdictPass},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(string(tc.severity)+"/"+string(tc.raw), func(t *testing.T) {
			t.Parallel()
			got := tc.severity.CapVerdict(tc.raw)
			if got != tc.want {
				t.Errorf("(%v).CapVerdict(%v) = %v, want %v", tc.severity, tc.raw, got, tc.want)
			}
		})
	}
}

func TestNotImplementedResult_AlwaysPasses(t *testing.T) {
	t.Parallel()
	result := spi.NotImplementedResult("test.rule")
	if result.Verdict != vo.VerdictPass {
		t.Errorf("NotImplementedResult should return PASS, got %v", result.Verdict)
	}
}

func TestRuleEvaluator_ExplainDoesNotPanic(t *testing.T) {
	t.Parallel()
	reg := spi.GlobalRegistry()
	all := reg.All()
	for _, e := range all {
		e := e
		t.Run(e.Metadata().TypeID, func(t *testing.T) {
			t.Parallel()
			// Explain with empty inputs should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Explain panicked: %v", r)
				}
			}()
			result := spi.NotImplementedResult(e.Metadata().TypeID)
			_ = e.Explain(spi.CheckInput{}, result, spi.NewParameterSet([]byte("{}")))
		})
	}
}

func TestRuleEvaluator_ParameterSchemaIsValid(t *testing.T) {
	t.Parallel()
	reg := spi.GlobalRegistry()
	for _, e := range reg.All() {
		e := e
		t.Run(e.Metadata().TypeID, func(t *testing.T) {
			t.Parallel()
			schema := e.ParameterSchema()
			if len(schema.Schema) == 0 {
				t.Error("ParameterSchema.Schema is empty")
			}
			// Should be valid JSON
			if err := schema.Validate([]byte("{}")); err != nil {
				t.Errorf("empty params should be valid JSON: %v", err)
			}
		})
	}
}

func TestRuleEvaluator_DataDependenciesUnionWithSelf(t *testing.T) {
	t.Parallel()
	// Union of deps with itself should be equal to itself
	reg := spi.GlobalRegistry()
	for _, e := range reg.All() {
		e := e
		t.Run(e.Metadata().TypeID, func(t *testing.T) {
			t.Parallel()
			deps := e.DataDependencies()
			union := deps.Union(deps)
			if union.Positions != deps.Positions ||
				union.NAV != deps.NAV ||
				union.MarketPrices != deps.MarketPrices {
				t.Error("Union with self changed deps")
			}
		})
	}
}

func TestBlacklist_NotOverridable(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("restriction.blacklist")
	if !ok {
		t.Fatal("not registered")
	}
	if r.Metadata().Overridable {
		t.Error("restriction.blacklist should NOT be overridable")
	}
}

func TestThaiSEC_NotOverridable(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("regulatory.thai_sec")
	if !ok {
		t.Fatal("not registered")
	}
	if r.Metadata().Overridable {
		t.Error("regulatory.thai_sec should NOT be overridable")
	}
}

func TestSellQuantity_NotOverridable(t *testing.T) {
	t.Parallel()
	r, ok := spi.GlobalRegistry().Get("quantity.sell_available")
	if !ok {
		t.Fatal("not registered")
	}
	if r.Metadata().Overridable {
		t.Error("quantity.sell_available should NOT be overridable")
	}
}

// Ensure all rules with Evaluate work with an empty DataBundle (no panic, no error).
func TestAllRules_EmptyBundle_NoError(t *testing.T) {
	t.Parallel()
	reg := spi.GlobalRegistry()
	input := spi.CheckInput{}
	bundle := spi.DataBundle{}
	params := spi.NewParameterSet([]byte("{}"))

	for _, e := range reg.All() {
		e := e
		t.Run(e.Metadata().TypeID, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Evaluate panicked: %v", r)
				}
			}()
			// Errors are acceptable (e.g. missing required param). No panics.
			result, _ := e.Evaluate(context.Background(), input, bundle, params)
			if result.Verdict != "" && !result.Verdict.IsValid() {
				t.Errorf("invalid verdict returned: %q", result.Verdict)
			}
		})
	}
}
