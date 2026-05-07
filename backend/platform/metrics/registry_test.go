package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHandler_ExposesPrometheusFormat(t *testing.T) {
	resetAllForTest()

	// Touch every metric so the exposition emits HELP/TYPE lines (Prometheus
	// only renders a label-less counter or histogram once it has at least
	// one observation; a vec with no observed label combinations stays
	// silent in the text format).
	InvestmentPostTotal.WithLabelValues(OutcomeSuccess).Add(0)
	InvestmentPostViolationTotal.WithLabelValues("warmup").Add(0)
	InvestmentProjectorRetryTotal.Add(0)
	InvestmentValuationDurationSecs.Observe(0)
	InvestmentValuationStaleInputs.Add(0)
	InvestmentForcePostTotal.WithLabelValues(ForcePostNormal).Add(0)
	MarketDataIngestTotal.WithLabelValues(ProviderAlphaVantage, IngestOutcomeSuccess).Add(0)
	MarketDataProviderLatencySecs.WithLabelValues(ProviderAlphaVantage).Observe(0)

	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()

	wantSubstrings := []string{
		"# HELP investment_post_total",
		"# TYPE investment_post_total counter",
		"# HELP investment_valuation_duration_seconds",
		"# HELP investment_force_post_total",
		"# HELP market_data_ingest_total",
		"# HELP market_data_provider_latency_seconds",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics body missing %q", want)
		}
	}
}

func TestInvestmentPostTotal_LabelsAndIncrements(t *testing.T) {
	resetAllForTest()

	InvestmentPostTotal.WithLabelValues(OutcomeSuccess).Inc()
	InvestmentPostTotal.WithLabelValues(OutcomeSuccess).Inc()
	InvestmentPostTotal.WithLabelValues(OutcomeRejected).Inc()

	if got := testutil.ToFloat64(InvestmentPostTotal.WithLabelValues(OutcomeSuccess)); got != 2 {
		t.Errorf("success counter = %v, want 2", got)
	}
	if got := testutil.ToFloat64(InvestmentPostTotal.WithLabelValues(OutcomeRejected)); got != 1 {
		t.Errorf("rejected counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(InvestmentPostTotal.WithLabelValues(OutcomeConflict)); got != 0 {
		t.Errorf("unused conflict counter = %v, want 0", got)
	}
}

func TestInvestmentValuationDuration_HistogramObservation(t *testing.T) {
	resetAllForTest()
	InvestmentValuationDurationSecs.Observe(0.123)
	InvestmentValuationDurationSecs.Observe(0.456)

	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := w.Body.String()

	if !strings.Contains(body, "investment_valuation_duration_seconds_count 2") {
		t.Errorf("histogram count missing: body=\n%s", body)
	}
}

func TestForcePostTotal_NormalAndReversalLabels(t *testing.T) {
	resetAllForTest()
	InvestmentForcePostTotal.WithLabelValues(ForcePostNormal).Inc()
	InvestmentForcePostTotal.WithLabelValues(ForcePostReversal).Add(3)

	if got := testutil.ToFloat64(InvestmentForcePostTotal.WithLabelValues(ForcePostNormal)); got != 1 {
		t.Errorf("normal force-post = %v, want 1", got)
	}
	if got := testutil.ToFloat64(InvestmentForcePostTotal.WithLabelValues(ForcePostReversal)); got != 3 {
		t.Errorf("reversal force-post = %v, want 3", got)
	}
}

func TestMarketDataIngest_LabelMatrix(t *testing.T) {
	resetAllForTest()
	MarketDataIngestTotal.WithLabelValues(ProviderAlphaVantage, IngestOutcomeSuccess).Inc()
	MarketDataIngestTotal.WithLabelValues(ProviderAlphaVantage, IngestOutcomeRateLimited).Inc()
	MarketDataIngestTotal.WithLabelValues(ProviderManual, IngestOutcomeSuccess).Add(2)

	if got := testutil.ToFloat64(MarketDataIngestTotal.WithLabelValues(ProviderAlphaVantage, IngestOutcomeSuccess)); got != 1 {
		t.Errorf("alpha-vantage success = %v, want 1", got)
	}
	if got := testutil.ToFloat64(MarketDataIngestTotal.WithLabelValues(ProviderManual, IngestOutcomeSuccess)); got != 2 {
		t.Errorf("manual success = %v, want 2", got)
	}
}
