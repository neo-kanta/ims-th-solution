package policy

import (
	"strings"
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

func navSource() []ToolResultSource {
	return []ToolResultSource{{
		ToolName: "get_fund_nav",
		Raw:      `{"source":{"tool":"get_fund_nav","as_of":"2026-06-05T00:00:00Z"},"data":{"nav":"10.25","as_of_date":"2026-06-05"}}`,
	}}
}

// Allowed: a NAV the model states matches a value in the current-turn tool
// result.
func TestValidator_AllowsTracedNAV(t *testing.T) {
	res := ValidateAnswer("The latest NAV is 10.25 as reported.", navSource())
	if !res.OK() {
		t.Fatalf("expected allow, got %s with violations %#v", res.Decision, res.Violations)
	}
	if res.Checked != 1 {
		t.Fatalf("expected 1 figure checked, got %d", res.Checked)
	}
	if len(res.Bindings) != 1 || res.Bindings[0].Source != valueobject.SourceToolRaw {
		t.Fatalf("expected one tool_raw binding, got %#v", res.Bindings)
	}
}

// Blocked: a NAV that does NOT appear in the tool result.
func TestValidator_BlocksUntracedNAV(t *testing.T) {
	res := ValidateAnswer("The latest NAV is 10.99.", navSource())
	if res.OK() {
		t.Fatal("expected block for untraced NAV")
	}
	if res.Decision != valueobject.DecisionBlock {
		t.Fatalf("expected block decision, got %s", res.Decision)
	}
	if res.FinalText != SafeBlockedMessage {
		t.Fatalf("expected safe blocked message, got %q", res.FinalText)
	}
	if len(res.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %#v", res.Violations)
	}
}

// No tool data at all: any financial figure is unverified -> block.
func TestValidator_BlocksWhenNoToolData(t *testing.T) {
	res := ValidateAnswer("Your balance is 1,000.00 THB.", nil)
	if res.OK() {
		t.Fatal("expected block when no tool data exists")
	}
}

// Harmless counts must NOT be flagged.
func TestValidator_AllowsHarmlessCounts(t *testing.T) {
	for _, ans := range []string{
		"You have 3 portfolios and 2 funds.",
		"Here are the top 5 holdings.",
		"I performed 4 steps to answer this.",
	} {
		res := ValidateAnswer(ans, nil)
		if !res.OK() {
			t.Fatalf("expected allow for counts in %q, violations %#v", ans, res.Violations)
		}
		if res.Checked != 0 {
			t.Fatalf("expected 0 financial figures in %q, got %d", ans, res.Checked)
		}
	}
}

// Percentages are validated; a calc_* tool result is a valid source.
func TestValidator_AllowsPercentageFromCalcTool(t *testing.T) {
	src := []ToolResultSource{{
		ToolName: "calc_exposure_pct",
		Raw:      `{"calculation":{"result":"12.5"},"data":{"exposure_pct":"12.5"}}`,
	}}
	res := ValidateAnswer("That position is 12.5% of the portfolio.", src)
	if !res.OK() {
		t.Fatalf("expected allow, got %s %#v", res.Decision, res.Violations)
	}
	if len(res.Bindings) != 1 || res.Bindings[0].Source != valueobject.SourceCalculation {
		t.Fatalf("expected a calculation-sourced binding, got %#v", res.Bindings)
	}
}

// Money matches regardless of thousands separators / trailing zeros.
func TestValidator_MoneyNormalisation(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_portfolio_valuation", Raw: `{"data":{"market_value":"1234.5"}}`}}
	res := ValidateAnswer("Total market value is 1,234.50 THB.", src)
	if !res.OK() {
		t.Fatalf("expected allow for normalised money, got %#v", res.Violations)
	}
}

// Currency-adjacent integers ARE validated (so a bare integer next to a
// currency code is not waved through).
func TestValidator_CurrencyAdjacentIntegerBlockedWhenUntraced(t *testing.T) {
	res := ValidateAnswer("Your cash balance is USD 500.", nil)
	if res.OK() {
		t.Fatal("expected block: currency-adjacent amount with no source")
	}
}

// Financial-context dates are validated; non-financial dates are ignored.
func TestValidator_FinancialDateContext(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_portfolio_valuation", Raw: `{"data":{"business_date":"2026-06-05"}}`}}

	blocked := ValidateAnswer("The valuation date is 2026-06-30.", src)
	if blocked.OK() {
		t.Fatal("expected block: valuation date not in tool result")
	}

	allowed := ValidateAnswer("The valuation date is 2026-06-05.", src)
	if !allowed.OK() {
		t.Fatalf("expected allow: valuation date present in tool result, got %#v", allowed.Violations)
	}

	ignored := ValidateAnswer("Let's schedule a call on 2026-06-30.", src)
	if !ignored.OK() || ignored.Checked != 0 {
		t.Fatalf("expected non-financial date ignored, got decision=%s checked=%d", ignored.Decision, ignored.Checked)
	}
}

// ── Bare integers in financial context (must be validated) ──────────────────

// 1. NAV stated as a bare integer that IS in the tool result -> pass.
func TestValidator_IntegerNAVTracedPasses(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_fund_nav", Raw: `{"data":{"nav":"10"}}`}}
	res := ValidateAnswer("The NAV is 10.", src)
	if !res.OK() {
		t.Fatalf("expected allow for traced integer NAV, got %#v", res.Violations)
	}
	if res.Checked != 1 {
		t.Fatalf("expected the integer NAV to be checked, checked=%d", res.Checked)
	}
}

// 2. NAV integer NOT in the tool result -> block.
func TestValidator_IntegerNAVUntracedBlocks(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_fund_nav", Raw: `{"data":{"nav":"12"}}`}}
	res := ValidateAnswer("The NAV is 10.", src)
	if res.OK() {
		t.Fatal("expected block for untraced integer NAV")
	}
}

// 3. price integer traced -> pass.
func TestValidator_IntegerPriceTracedPasses(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_portfolio_valuation", Raw: `{"data":{"price":"50"}}`}}
	res := ValidateAnswer("Price is 50.", src)
	if !res.OK() {
		t.Fatalf("expected allow for traced integer price, got %#v", res.Violations)
	}
}

// 4. quantity integer untraced -> block.
func TestValidator_IntegerQuantityUntracedBlocks(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_portfolio_holdings", Raw: `{"data":{"quantity":"999"}}`}}
	res := ValidateAnswer("Quantity is 1000.", src)
	if res.OK() {
		t.Fatal("expected block for untraced integer quantity")
	}
}

// Unit-bearing phrase: "3000 units" must be validated (and blocked if untraced).
func TestValidator_IntegerWithUnitFollowingIsValidated(t *testing.T) {
	res := ValidateAnswer("The portfolio holds 3000 units.", nil)
	if res.OK() {
		t.Fatal("expected block: '3000 units' is a financial quantity with no source")
	}
	if res.Checked != 1 {
		t.Fatalf("expected the unit-bearing integer to be checked, checked=%d", res.Checked)
	}
}

// Larger financial integers in context.
func TestValidator_LargeFinancialIntegersValidated(t *testing.T) {
	for _, ans := range []string{
		"AUM is 50000000.",
		"Cash balance is 200000.",
		"Market value is 1000000.",
	} {
		res := ValidateAnswer(ans, nil) // no sources -> must block
		if res.OK() {
			t.Fatalf("expected block for untraced financial integer in %q", ans)
		}
	}
}

// 5 & 6. Normal counts must be ignored.
func TestValidator_CountsRemainIgnored(t *testing.T) {
	for _, ans := range []string{
		"I performed 3 steps.",
		"Here are the top 5 holdings.",
		"There are 2 funds.",
		"Show 10 rows.",
		"1 portfolio found.",
		"In total there are 2 funds.",
	} {
		res := ValidateAnswer(ans, nil)
		if !res.OK() || res.Checked != 0 {
			t.Fatalf("expected count ignored in %q (decision=%s checked=%d violations=%#v)",
				ans, res.Decision, res.Checked, res.Violations)
		}
	}
}

// UUIDs and codes in the answer must NOT be treated as financial figures.
func TestValidator_IgnoresUUIDAndCodeFragments(t *testing.T) {
	src := []ToolResultSource{{
		ToolName: "list_funds",
		Raw:      `{"data":[{"id":"00000000-0000-0000-0000-000000000004","code":"GEF-0001000","name":"Global Equity"}]}`,
	}}
	answer := "You can access: Global Equity (id 00000000-0000-0000-0000-000000000004, code GEF-0001000)."
	res := ValidateAnswer(answer, src)
	if !res.OK() {
		t.Fatalf("UUIDs/codes are not financial figures; expected allow, got violations %#v", res.Violations)
	}
	if res.Checked != 0 {
		t.Fatalf("expected no financial figures detected, got %d", res.Checked)
	}
}

// A "list funds" style answer with only identifiers/names must pass.
func TestValidator_ListFundsAnswerPasses(t *testing.T) {
	src := []ToolResultSource{{
		ToolName: "list_funds",
		Raw:      `{"data":[{"id":"00000000-0000-0000-0000-000000000004","name":"Equity"},{"id":"00000000-0000-0000-0000-000000000006","name":"Bond"}]}`,
	}}
	answer := "You can access 2 funds: Equity and Bond."
	res := ValidateAnswer(answer, src)
	if !res.OK() {
		t.Fatalf("expected allow for a plain fund listing, got %#v", res.Violations)
	}
}

// Real negative figures keep their sign and still validate against the source.
func TestValidator_NegativeFigureValidated(t *testing.T) {
	src := []ToolResultSource{{ToolName: "get_portfolio_valuation", Raw: `{"data":{"unrealised_pnl":"-1234.50"}}`}}
	if !ValidateAnswer("Unrealised P&L is -1234.50 THB.", src).OK() {
		t.Fatal("a traced negative figure should be allowed")
	}
	if ValidateAnswer("Unrealised P&L is -9999.00 THB.", src).OK() {
		t.Fatal("an untraced negative figure should be blocked")
	}
}

// Numbered-list markers ("1." … "6.") must never be treated as figures, even
// when financial words appear in the list — this was the "Couldn't verify: 5 6"
// false positive.
func TestValidator_NumberedListMarkersIgnored(t *testing.T) {
	answer := "Funds you can access:\n" +
		"1. Global Tech Fund\n" +
		"2. Bond Fund\n" +
		"5. Balanced Fund\n" +
		"6. Equity Fund"
	res := ValidateAnswer(answer, nil)
	if !res.OK() {
		t.Fatalf("numbered list markers must not be flagged, got %#v", res.Violations)
	}
	if res.Checked != 0 {
		t.Fatalf("expected 0 financial figures in a plain numbered list, got %d", res.Checked)
	}
}

// A keyword far away in the sentence must not sweep up an unrelated integer,
// but an immediate "NAV is 10" still validates.
func TestValidator_KeywordProximityIsTight(t *testing.T) {
	// "NAV" is far from "6" here (different clause) -> not flagged.
	farAnswer := "The NAV report covers several funds and lists 6 items."
	if !ValidateAnswer(farAnswer, nil).OK() {
		t.Fatal("a distant keyword must not flag an unrelated integer")
	}
	// Immediate keyword -> still flagged/blocked when untraced.
	if ValidateAnswer("The NAV is 6.", nil).OK() {
		t.Fatal("an immediate 'NAV is 6' must still be validated/blocked when untraced")
	}
}

// The RedactFigures helper is position-safe and removes the figures.
func TestRedactFigures(t *testing.T) {
	answer := "NAV is 10.99 and value is 1,234.50."
	figures := detectFigures(answer)
	if len(figures) != 2 {
		t.Fatalf("expected 2 detected figures, got %d (%#v)", len(figures), figures)
	}
	out := RedactFigures(answer, figures, "[unverified]")
	if strings.Contains(out, "10.99") || strings.Contains(out, "1,234.50") {
		t.Fatalf("figures not redacted: %q", out)
	}
	if !strings.Contains(out, "[unverified]") {
		t.Fatalf("placeholder missing: %q", out)
	}
}

func TestValidator_BenchmarkAndRatioFixes(t *testing.T) {
	// 1. Benchmark numbers in string leaves (like KTB-BALANCED benchmark "Balanced 50/40/10")
	// If the tool returns a string "Balanced 50/40/10", those numbers should be extracted
	// into the corpus. In the final text, "50", "40", and "10" (which are flagged as financial
	// figures due to proximity to "balanced") must be allowed because they exist in the corpus.
	src1 := []ToolResultSource{{
		ToolName: "list_funds",
		Raw:      `{"data":[{"id":"d0001000-0000-0000-0000-000000000004","code":"KTB-BALANCED","benchmark":"Balanced 50/40/10"}]}`,
	}}
	res1 := ValidateAnswer("Here is KTB-BALANCED (Balanced 50/40/10).", src1)
	if !res1.OK() {
		t.Fatalf("expected allow for benchmark numbers, got violations %#v", res1.Violations)
	}

	// 2. Percentage to ratio translation
	// Text has "0.448%", which matches the ratio "0.00448" in the corpus (from the tool).
	src2 := []ToolResultSource{{
		ToolName: "get_portfolio_valuation",
		Raw:      `{"data":{"roi":"0.00448"}}`,
	}}
	res2 := ValidateAnswer("The portfolio ROI is 0.448%.", src2)
	if !res2.OK() {
		t.Fatalf("expected allow for ratio-to-percentage conversion, got violations %#v", res2.Violations)
	}

	// 3. Time interval counts near financial keywords
	// "NAV is 10.25 (last updated 2 days ago)"
	// Here, "2" is followed by "days", which is in countNouns, so it should be ignored.
	src3 := []ToolResultSource{{
		ToolName: "get_fund_nav",
		Raw:      `{"data":{"nav":"10.25"}}`,
	}}
	res3 := ValidateAnswer("The NAV is 10.25 (last updated 2 days ago).", src3)
	if !res3.OK() {
		t.Fatalf("expected allow: '2 days' should be ignored as a count, got violations %#v", res3.Violations)
	}
	if res3.Checked != 1 { // Only 10.25 should be checked
		t.Fatalf("expected 1 checked figure, got %d", res3.Checked)
	}

	// 4. Thai language support
	// Test Thai count nouns like "กองทุน" (funds) and "วัน" (days)
	res4 := ValidateAnswer("คุณมีสิทธิ์เข้าถึง 10 กองทุน (ข้อมูล ณ 2 วันที่ผ่านมา)", nil)
	if !res4.OK() {
		t.Fatalf("expected allow for Thai counts, got violations %#v", res4.Violations)
	}
	if res4.Checked != 0 {
		t.Fatalf("expected 0 checked figures for Thai counts, got %d", res4.Checked)
	}

	// Test Thai financial keywords preceding validated numbers, and verified from tool
	src5 := []ToolResultSource{{
		ToolName: "get_fund_nav",
		Raw:      `{"data":{"nav":"10.25"}}`,
	}}
	res5 := ValidateAnswer("มูลค่าสินทรัพย์สุทธิคือ 10.25", src5)
	if !res5.OK() {
		t.Fatalf("expected allow for traced Thai financial figure, got violations %#v", res5.Violations)
	}
	if res5.Checked != 1 {
		t.Fatalf("expected 1 checked figure for Thai financial text, got %d", res5.Checked)
	}
}

func TestValidator_RoundingAndSuffixTolerance(t *testing.T) {
	// Corpus with raw numbers
	src := []ToolResultSource{
		{
			ToolName: "get_portfolio_valuation",
			Raw:      `{"data":{"market_value":"2511600000.00","cash_balance":"11200000.00","roi":"0.00448"}}`,
		},
	}

	// 1. English suffix: "2.51 billion" matching 2,511,600,000
	res1 := ValidateAnswer("The portfolio market value is 2.51 billion USD.", src)
	if !res1.OK() {
		t.Fatalf("expected allow for '2.51 billion' matching 2,511,600,000, got violations: %#v", res1.Violations)
	}

	// 2. English suffix exact: "11.2 million" matching 11,200,000
	res2 := ValidateAnswer("The cash balance is 11.2 million.", src)
	if !res2.OK() {
		t.Fatalf("expected allow for '11.2 million' matching 11,200,000, got violations: %#v", res2.Violations)
	}

	// 3. Thai suffix: "2.51 พันล้าน" matching 2,511,600,000
	res3 := ValidateAnswer("มูลค่าตลาดของพอร์ตคือ 2.51 พันล้าน", src)
	if !res3.OK() {
		t.Fatalf("expected allow for '2.51 พันล้าน' matching 2,511,600,000, got violations: %#v", res3.Violations)
	}

	// 4. Thai suffix with currency: "2.51 พันล้านบาท" matching 2,511,600,000
	res4 := ValidateAnswer("มูลค่าตลาดของพอร์ตคือ 2.51 พันล้านบาท", src)
	if !res4.OK() {
		t.Fatalf("expected allow for '2.51 พันล้านบาท' matching 2,511,600,000, got violations: %#v", res4.Violations)
	}

	// 5. Percentage ratio conversion and rounding: "0.45%" matching 0.00448
	res5 := ValidateAnswer("The portfolio ROI is 0.45%.", src)
	if !res5.OK() {
		t.Fatalf("expected allow for '0.45%%' matching ratio 0.00448, got violations: %#v", res5.Violations)
	}

	// 6. Out of tolerance: "2.50 billion" matching 2,511,600,000 (claimed: 2,500,000,000, tolerance: 5,000,000 -> [2.495B, 2.505B] - actual is 2.5116B, out of range -> block)
	res6 := ValidateAnswer("The portfolio market value is 2.50 billion.", src)
	if res6.OK() {
		t.Fatal("expected block for '2.50 billion' as it is out of tolerance range for 2,511,600,000")
	}

	// 7. Tighter precision block: "2.510 billion" (claimed: 2,510,000,000, tolerance: 500,000 -> [2.5095B, 2.5105B] - actual is 2.5116B, out of range -> block)
	res7 := ValidateAnswer("The portfolio market value is 2.510 billion.", src)
	if res7.OK() {
		t.Fatal("expected block for '2.510 billion' due to tighter precision limits")
	}
}
