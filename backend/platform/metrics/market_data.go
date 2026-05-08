package metrics

import "github.com/prometheus/client_golang/prometheus"

// Provider label values used across market_data metrics.
const (
	ProviderAlphaVantage = "alpha_vantage"
	ProviderManual       = "manual"
)

// Outcome label values for market_data_ingest_total.
const (
	IngestOutcomeSuccess     = "success"
	IngestOutcomeRateLimited = "rate_limited"
	IngestOutcomeFailed      = "failed"
)

// Market-data metric handles. Call sites:
//
//	metrics.MarketDataIngestTotal.WithLabelValues(metrics.ProviderAlphaVantage, metrics.IngestOutcomeSuccess).Inc()
//	metrics.MarketDataProviderLatencySecs.WithLabelValues(metrics.ProviderAlphaVantage).Observe(elapsed.Seconds())
var (
	MarketDataIngestTotal         *prometheus.CounterVec
	MarketDataProviderLatencySecs *prometheus.HistogramVec
)

func init() { registerMarketDataMetrics() }

// registerMarketDataMetrics constructs every market_data metric and
// registers it with the IMS registry.
func registerMarketDataMetrics() {
	MarketDataIngestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "market_data_ingest_total",
			Help: "Number of market-data ingestion attempts by provider and outcome.",
		},
		[]string{"provider", "outcome"},
	)
	MarketDataProviderLatencySecs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "market_data_provider_latency_seconds",
			Help:    "Latency of upstream market-data provider calls.",
			Buckets: prometheus.ExponentialBuckets(0.05, 2, 10),
		},
		[]string{"provider"},
	)

	mustRegister(MarketDataIngestTotal)
	mustRegister(MarketDataProviderLatencySecs)
}
