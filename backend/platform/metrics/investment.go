package metrics

import "github.com/prometheus/client_golang/prometheus"

// Outcome label values for investment_post_total.
const (
	OutcomeSuccess  = "success"
	OutcomeRejected = "rejected"
	OutcomeConflict = "conflict"
)

// ForcePost type label values for investment_force_post_total.
const (
	ForcePostNormal   = "normal"
	ForcePostReversal = "reversal"
)

// Investment metric handles. These are the call sites Phase 1+ will use:
//
//	metrics.InvestmentPostTotal.WithLabelValues(metrics.OutcomeSuccess).Inc()
//	metrics.InvestmentValuationDurationSeconds.Observe(elapsed.Seconds())
//
// Initial values are zero for every label combination; Prometheus exposes
// them automatically once Inc/Observe is called for that combination.
var (
	InvestmentPostTotal              *prometheus.CounterVec
	InvestmentPostViolationTotal     *prometheus.CounterVec
	InvestmentProjectorRetryTotal    prometheus.Counter
	InvestmentValuationDurationSecs  prometheus.Histogram
	InvestmentValuationStaleInputs   prometheus.Counter
	InvestmentForcePostTotal         *prometheus.CounterVec
)

func init() { registerInvestmentMetrics() }

// registerInvestmentMetrics constructs every investment metric and registers
// it with the IMS registry. Called from init() and from resetAllForTest()
// after the registry has been re-created for a clean test slate.
func registerInvestmentMetrics() {
	InvestmentPostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "investment_post_total",
			Help: "Number of investment ledger post attempts by outcome (success|rejected|conflict).",
		},
		[]string{"outcome"},
	)
	InvestmentPostViolationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "investment_post_violation_total",
			Help: "Number of investment post precondition violations by violation type.",
		},
		[]string{"violation"},
	)
	InvestmentProjectorRetryTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "investment_projector_retry_total",
			Help: "Number of position projector retries triggered by optimistic-locking conflicts.",
		},
	)
	InvestmentValuationDurationSecs = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "investment_valuation_duration_seconds",
			Help:    "Wall-clock duration of one ValuationRunner pass per portfolio.",
			Buckets: prometheus.ExponentialBuckets(0.005, 2, 12),
		},
	)
	InvestmentValuationStaleInputs = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "investment_valuation_stale_inputs_total",
			Help: "Number of valuation inputs (price or FX) that exceeded the configured staleness threshold.",
		},
	)
	InvestmentForcePostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "investment_force_post_total",
			Help: "Number of force-post operations by type (normal|reversal).",
		},
		[]string{"type"},
	)

	mustRegister(InvestmentPostTotal)
	mustRegister(InvestmentPostViolationTotal)
	mustRegister(InvestmentProjectorRetryTotal)
	mustRegister(InvestmentValuationDurationSecs)
	mustRegister(InvestmentValuationStaleInputs)
	mustRegister(InvestmentForcePostTotal)
}
