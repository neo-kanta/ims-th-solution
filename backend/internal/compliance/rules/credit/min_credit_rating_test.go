package credit_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/credit"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func minRatingParams(t *testing.T, m map[string]any) spi.ParameterSet {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return spi.NewParameterSet(raw)
}

func buyOrder(ticker string) spi.CheckInput {
	return spi.CheckInput{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProposedOrder: &spi.ProposedOrder{
			OrderID: uuid.New(),
			Ticker:  ticker,
			Side:    vo.OrderSideBuy,
		},
	}
}

func classificationBundle(ticker, issuer string, isGov bool) *spi.ClassificationSnapshot {
	return spi.NewClassificationSnapshot([]spi.InstrumentClassification{{
		Ticker:       ticker,
		Issuer:       issuer,
		ParentEntity: issuer,
		IsGovernment: isGov,
	}})
}

// ratingsFor builds a snapshot in an explicit form: issuer → []rating.
func ratingsFor(entries map[string][]spi.CreditRating) *spi.CreditRatingSnapshot {
	return &spi.CreditRatingSnapshot{Ratings: entries}
}

func getMinRatingRule(t *testing.T) spi.RuleEvaluator {
	t.Helper()
	r, ok := spi.GlobalRegistry().Get("credit.min_rating")
	if !ok {
		t.Fatal("credit.min_rating not registered")
	}
	return r
}

// ─── happy paths ────────────────────────────────────────────────────────────

func TestMinRating_MeetsThreshold_Passes(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("ACME", "ACME-CORP", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"ACME-CORP": {{Rating: "A-", Agency: "TRIS"}},
		}),
	}
	res, err := r.Evaluate(context.Background(), buyOrder("ACME"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("want PASS, got %s (%s)", res.Verdict, res.Message)
	}
}

func TestMinRating_BelowThreshold_Blocks(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("JUNK", "JUNK-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"JUNK-CO": {{Rating: "B", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("JUNK"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("want BLOCK, got %s", res.Verdict)
	}
	if !strings.Contains(res.Message, "fails credit.min_rating") {
		t.Fatalf("message does not mention failure: %s", res.Message)
	}
}

// ─── missing rating → BLOCK with RATING_UNAVAILABLE ─────────────────────────

func TestMinRating_NoRating_BlocksWithRatingUnavailable(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("UNKNOWN", "UNKNOWN-CO", false),
		CreditRatings:   ratingsFor(map[string][]spi.CreditRating{}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("UNKNOWN"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("missing rating must BLOCK, got %s", res.Verdict)
	}
	if got := res.Evidence.Metrics["error"]; got != "RATING_UNAVAILABLE" {
		t.Fatalf("expected RATING_UNAVAILABLE in evidence.metrics.error, got %q", got)
	}
}

func TestMinRating_NilSnapshot_FailsClosed(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("X", "X-CO", false),
		CreditRatings:   nil,
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("X"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "A"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("nil CreditRatings must fail-closed (BLOCK), got %s", res.Verdict)
	}
}

// ─── government exemption ───────────────────────────────────────────────────

func TestMinRating_GovernmentExempt_Passes(t *testing.T) {
	r := getMinRatingRule(t)
	// Government issuer with no rating at all — would BLOCK otherwise.
	bundle := spi.DataBundle{
		Classifications: classificationBundle("LB366A", "MOF-TH", true),
		CreditRatings:   ratingsFor(map[string][]spi.CreditRating{}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("LB366A"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "AAA"}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("government-exempt must pass, got %s", res.Verdict)
	}
}

func TestMinRating_GovernmentExemptDisabled_Blocks(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("LB366A", "MOF-TH", true),
		CreditRatings:   ratingsFor(map[string][]spi.CreditRating{}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("LB366A"), bundle,
		minRatingParams(t, map[string]any{
			"min_rating":        "AAA",
			"government_exempt": false,
		}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("government_exempt=false must enforce rating; got %s", res.Verdict)
	}
}

// ─── split-rating policy ────────────────────────────────────────────────────

func TestMinRating_SplitRatingWorst_Blocks(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("SPLIT", "SPLIT-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"SPLIT-CO": {
				{Rating: "A", Agency: "TRIS"},
				{Rating: "BB+", Agency: "FITCH"}, // worse side
			},
		}),
	}
	// Policy defaults to "worst" — BB+ is below BBB-, so BLOCK.
	res, _ := r.Evaluate(context.Background(), buyOrder("SPLIT"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("split-rating worst policy must BLOCK on BB+, got %s (%s)", res.Verdict, res.Message)
	}
}

func TestMinRating_SplitRatingBest_Passes(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("SPLIT", "SPLIT-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"SPLIT-CO": {
				{Rating: "A", Agency: "TRIS"},
				{Rating: "BB+", Agency: "FITCH"},
			},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("SPLIT"), bundle,
		minRatingParams(t, map[string]any{
			"min_rating":          "BBB-",
			"split_rating_policy": "best",
		}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("best policy must choose A (pass), got %s (%s)", res.Verdict, res.Message)
	}
}

// ─── agency priority ────────────────────────────────────────────────────────

func TestMinRating_AgencyPriority_Wins(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("ACME", "ACME-CORP", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"ACME-CORP": {
				{Rating: "B", Agency: "MOODY'S"}, // lowest
				{Rating: "AA-", Agency: "TRIS"},  // prioritised, above threshold
				{Rating: "BB-", Agency: "FITCH"},
			},
		}),
	}
	// Priority = TRIS → AA- selected → PASS despite worse ratings elsewhere.
	res, _ := r.Evaluate(context.Background(), buyOrder("ACME"), bundle,
		minRatingParams(t, map[string]any{
			"min_rating":             "BBB-",
			"rating_agency_priority": []string{"TRIS"},
		}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("agency priority TRIS→AA- must PASS; got %s (%s)", res.Verdict, res.Message)
	}
	if got := res.Evidence.Metrics["selected_agency"]; got != "TRIS" {
		t.Fatalf("expected selected_agency=TRIS, got %q", got)
	}
}

func TestMinRating_AgencyPriority_FallsThroughWhenMissing(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("ACME", "ACME-CORP", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"ACME-CORP": {
				{Rating: "A-", Agency: "S&P"}, // priority agency absent → fall through
			},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("ACME"), bundle,
		minRatingParams(t, map[string]any{
			"min_rating":             "BBB-",
			"rating_agency_priority": []string{"TRIS"}, // not present
		}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("fallthrough to S&P A- must PASS; got %s (%s)", res.Verdict, res.Message)
	}
}

// ─── Moody's notation & normalisation ───────────────────────────────────────

func TestMinRating_MoodysNotation_Parses(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("ACME", "ACME-CORP", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"ACME-CORP": {
				{Rating: "Baa3", Agency: "MOODY'S"}, // == BBB-
			},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("ACME"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("Baa3 should map to BBB- and PASS ≥ BBB-, got %s (%s)", res.Verdict, res.Message)
	}
}

// ─── parameter validation ───────────────────────────────────────────────────

func TestMinRating_MissingMinRating_FailsClosed(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("X", "X-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"X-CO": {{Rating: "AAA", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("X"), bundle,
		minRatingParams(t, map[string]any{})) // no min_rating
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("missing min_rating must fail-closed; got %s", res.Verdict)
	}
	if got := res.Evidence.Metrics["error"]; got != "INVALID_PARAMETERS" {
		t.Fatalf("expected INVALID_PARAMETERS, got %q", got)
	}
}

func TestMinRating_InvalidMinRating_FailsClosed(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("X", "X-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"X-CO": {{Rating: "AAA", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("X"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "NOT_A_RATING"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("invalid min_rating must fail-closed; got %s", res.Verdict)
	}
}

func TestMinRating_InvalidSplitPolicy_FailsClosed(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("X", "X-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"X-CO": {{Rating: "AAA", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("X"), bundle,
		minRatingParams(t, map[string]any{
			"min_rating":          "BBB-",
			"split_rating_policy": "random",
		}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("invalid split_rating_policy must fail-closed; got %s", res.Verdict)
	}
}

// ─── misc ───────────────────────────────────────────────────────────────────

func TestMinRating_NoProposedOrder_Passes(t *testing.T) {
	r := getMinRatingRule(t)
	res, _ := r.Evaluate(context.Background(), spi.CheckInput{},
		spi.DataBundle{}, minRatingParams(t, map[string]any{"min_rating": "AAA"}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("no order → PASS; got %s", res.Verdict)
	}
}

func TestMinRating_UnrecognisedRating_Blocks(t *testing.T) {
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: classificationBundle("X", "X-CO", false),
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"X-CO": {{Rating: "??", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("X"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictBlock {
		t.Fatalf("unrecognised notation must fail-closed; got %s", res.Verdict)
	}
	if got := res.Evidence.Metrics["error"]; got != "UNRECOGNISED_RATING" {
		t.Fatalf("expected UNRECOGNISED_RATING evidence, got %q", got)
	}
}

func TestMinRating_FallbackToTickerLookup(t *testing.T) {
	// No Classifications → issuer key falls back to ticker; snapshot keyed by ticker.
	r := getMinRatingRule(t)
	bundle := spi.DataBundle{
		Classifications: nil,
		CreditRatings: ratingsFor(map[string][]spi.CreditRating{
			"ACME": {{Rating: "A", Agency: "TRIS"}},
		}),
	}
	res, _ := r.Evaluate(context.Background(), buyOrder("ACME"), bundle,
		minRatingParams(t, map[string]any{"min_rating": "BBB-"}))
	if res.Verdict != vo.VerdictPass {
		t.Fatalf("ticker-keyed lookup must PASS; got %s (%s)", res.Verdict, res.Message)
	}
}
