package logging

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Standard log field names used across modules. Centralising these lets
// log aggregation pipelines pivot on the same key regardless of which
// module produced the line.
const (
	FieldRequestID     = "request_id"
	FieldActorID       = "actor_id"
	FieldPortfolioID   = "portfolio_id"
	FieldFundID        = "fund_id"
	FieldBusinessDate  = "business_date"
	FieldTransactionID = "transaction_id"
	FieldInstrumentID  = "instrument_id"
	FieldContractID    = "contract_id"
)

// LogContext is a fluent builder for log attribute slices. Modules build a
// LogContext once at the start of an operation, derive sub-contexts as
// they discover more identifiers, and pass the slice into slog.Log /
// slog.WithGroup helpers.
type LogContext struct {
	attrs []slog.Attr
}

// FromRequest seeds a context with the chi-injected request ID so every
// log line emitted during the request lifecycle carries the same key.
func FromRequest(ctx context.Context) *LogContext {
	lc := &LogContext{}
	if id := middleware.GetReqID(ctx); id != "" {
		lc.attrs = append(lc.attrs, slog.String(FieldRequestID, id))
	}
	return lc
}

// New returns an empty LogContext.
func New() *LogContext { return &LogContext{} }

// WithRequestID returns a context with an explicit request ID attached.
// Used by the scheduler binary, which has no inbound HTTP request.
func WithRequestID(id string) *LogContext {
	lc := &LogContext{}
	if id != "" {
		lc.attrs = append(lc.attrs, slog.String(FieldRequestID, id))
	}
	return lc
}

// Actor adds the acting user ID. uuid.Nil is silently ignored so callers
// don't have to guard.
func (lc *LogContext) Actor(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldActorID, id.String()))
	return lc
}

// Portfolio adds the portfolio ID.
func (lc *LogContext) Portfolio(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldPortfolioID, id.String()))
	return lc
}

// Fund adds the fund ID.
func (lc *LogContext) Fund(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldFundID, id.String()))
	return lc
}

// Contract adds the contract ID.
func (lc *LogContext) Contract(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldContractID, id.String()))
	return lc
}

// Instrument adds the instrument ID.
func (lc *LogContext) Instrument(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldInstrumentID, id.String()))
	return lc
}

// Transaction adds the ledger transaction ID.
func (lc *LogContext) Transaction(id uuid.UUID) *LogContext {
	if lc == nil || id == uuid.Nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldTransactionID, id.String()))
	return lc
}

// BusinessDate adds the trading business date in YYYY-MM-DD form.
func (lc *LogContext) BusinessDate(t time.Time) *LogContext {
	if lc == nil || t.IsZero() {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(FieldBusinessDate, t.Format("2006-01-02")))
	return lc
}

// String adds an arbitrary string attribute. Prefer the typed builders
// above for canonical fields so log shape stays consistent.
func (lc *LogContext) String(key, value string) *LogContext {
	if lc == nil {
		return lc
	}
	lc.attrs = append(lc.attrs, slog.String(key, value))
	return lc
}

// Args returns the accumulated attributes as a []any suitable for the
// slog.Log/Info/Warn/Error variadic argument list.
func (lc *LogContext) Args() []any {
	if lc == nil || len(lc.attrs) == 0 {
		return nil
	}
	out := make([]any, 0, len(lc.attrs)*2)
	for _, a := range lc.attrs {
		out = append(out, a.Key, a.Value.Any())
	}
	return out
}

// Logger returns a slog.Logger pre-bound to the accumulated attributes.
func (lc *LogContext) Logger() *slog.Logger {
	if lc == nil {
		return slog.Default()
	}
	args := lc.Args()
	if len(args) == 0 {
		return slog.Default()
	}
	return slog.Default().With(args...)
}
