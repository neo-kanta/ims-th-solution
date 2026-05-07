// Package health implements the IMS liveness and readiness HTTP probes.
//
// Liveness ("/health/live") asserts only that the process is alive — it
// performs no I/O and always returns 200 unless the process is in the
// middle of shutting down. Kubernetes / load balancers use this to decide
// whether to restart the pod.
//
// Readiness ("/health/ready") asserts that the process is willing to
// accept traffic. It runs a DB ping and asks every registered Probe for
// its current state. A single Probe failure flips the response to 503.
//
// Both endpoints emit the canonical error envelope on failure so callers
// can switch on `error_code` consistently with every other API surface.
package health

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Probe is implemented by any subsystem that wants to gate readiness on
// its state — e.g. a workflow scheduler that needs Postgres warm before
// announcing ready, or a market-data ingester that needs an API key.
//
// Implementations must respect the context deadline; readiness should
// be fast (sub-second) so probes do not stall the load balancer.
type Probe interface {
	Name() string
	Ready(ctx context.Context) error
}

// Service collects every readiness probe and the DB pool used for the
// baseline ping. Add/remove probes via Register / Deregister at startup.
type Service struct {
	pool      *pgxpool.Pool
	probes    []Probe
	shuttingDown atomic.Bool
	probeTimeout time.Duration
}

// NewService returns a Service bound to the given pool. probeTimeout
// caps how long a single probe is allowed to run; the readiness handler
// short-circuits the request if it's exceeded. Pass zero to use the
// 2-second default.
func NewService(pool *pgxpool.Pool, probeTimeout time.Duration) *Service {
	if probeTimeout <= 0 {
		probeTimeout = 2 * time.Second
	}
	return &Service{pool: pool, probeTimeout: probeTimeout}
}

// Register adds a probe to the readiness list. Safe to call only at
// boot; the probe slice is not mutex-guarded for runtime mutation.
func (s *Service) Register(p Probe) {
	if s == nil || p == nil {
		return
	}
	s.probes = append(s.probes, p)
}

// MarkShuttingDown flips liveness to "draining" so the load balancer
// stops sending new traffic during a graceful shutdown window.
func (s *Service) MarkShuttingDown() {
	if s == nil {
		return
	}
	s.shuttingDown.Store(true)
}

// Live is the /health/live handler.
func (s *Service) Live(w http.ResponseWriter, r *http.Request) {
	if s != nil && s.shuttingDown.Load() {
		httputil.WriteEnvelope(
			w, r,
			http.StatusServiceUnavailable,
			errcode.CodeInternal,
			"shutting down",
			nil,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"alive"}`))
}

// Ready is the /health/ready handler. Returns 200 with `{"status":"ready"}`
// when DB ping succeeds and every probe reports nil; otherwise emits the
// envelope with status 503 and a `details` map of the failing probes.
func (s *Service) Ready(w http.ResponseWriter, r *http.Request) {
	if s == nil {
		httputil.WriteEnvelope(w, r, http.StatusServiceUnavailable, errcode.CodeInternal, "health service not initialised", nil)
		return
	}
	if s.shuttingDown.Load() {
		httputil.WriteEnvelope(w, r, http.StatusServiceUnavailable, errcode.CodeInternal, "shutting down", nil)
		return
	}

	failures := map[string]any{}

	if s.pool != nil {
		ctx, cancel := context.WithTimeout(r.Context(), s.probeTimeout)
		if err := s.pool.Ping(ctx); err != nil {
			failures["database"] = err.Error()
		}
		cancel()
	}

	for _, p := range s.probes {
		ctx, cancel := context.WithTimeout(r.Context(), s.probeTimeout)
		if err := p.Ready(ctx); err != nil {
			failures[p.Name()] = err.Error()
		}
		cancel()
	}

	if len(failures) > 0 {
		httputil.WriteEnvelope(w, r, http.StatusServiceUnavailable, errcode.CodeInternal, "not ready", failures)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}

// ProbeFunc adapts a function into the Probe interface.
type ProbeFunc struct {
	NameValue string
	ReadyFunc func(ctx context.Context) error
}

// Name implements Probe.
func (p ProbeFunc) Name() string { return p.NameValue }

// Ready implements Probe.
func (p ProbeFunc) Ready(ctx context.Context) error {
	if p.ReadyFunc == nil {
		return nil
	}
	return p.ReadyFunc(ctx)
}
