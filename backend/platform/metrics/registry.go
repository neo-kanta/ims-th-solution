// Package metrics owns the IMS-wide Prometheus registry and the metric
// handles consumed by Phase 1+ instrumentation.
//
// Design:
//   - One private *prometheus.Registry. We don't use prometheus.DefaultRegisterer
//     so test runs don't accidentally collide with another caller's metrics.
//   - Metric handles live as exported package-level vars in this package
//     (investment.go, market_data.go). Phase 1 imports them and calls
//     Inc / Observe at the right moments.
//   - Handler() returns an http.Handler that exposes the registry in the
//     Prometheus text exposition format. cmd/server mounts it at /metrics.
//   - Reset() resets all metrics to zero. Used by tests; never call from
//     production code.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// registry is the IMS-private Prometheus registry. Tests in this package
// reset it between cases so per-test state stays isolated.
var registry = newRegistry()

func newRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	r.MustRegister(collectors.NewGoCollector())
	return r
}

// Handler returns the HTTP handler that emits the IMS metrics in the
// Prometheus text exposition format. Callers must mount this at /metrics
// on a network the operator considers safe to scrape (no auth).
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Registry returns the underlying *prometheus.Registry so additional
// collectors (e.g. test fixtures) can be registered. Production code
// should not call this — declare metrics in this package's *.go files
// using the helpers below.
func Registry() *prometheus.Registry { return registry }

// MustRegister wraps registry.MustRegister so this package's *.go files
// can declare metrics tersely.
func mustRegister(c prometheus.Collector) {
	registry.MustRegister(c)
}

// resetAllForTest unregisters every collector and re-installs runtime
// collectors so each test starts from a clean slate. This MUST NOT be
// called from production code.
func resetAllForTest() {
	registry = newRegistry()
	registerInvestmentMetrics()
	registerMarketDataMetrics()
	registerChatMetrics()
}
