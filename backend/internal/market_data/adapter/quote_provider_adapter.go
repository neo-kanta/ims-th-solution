// Package adapter exposes the market_data module behind cross-module ports
// declared in backend/pkg/contract so consumers (investment, etc.) never have
// to import internal/market_data application/domain types directly.
package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mdapp "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	mddomain "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// QuoteProviderAdapter implements contract.MarketQuoteProvider by delegating
// to the market_data application service. Wired into the investment module
// from cmd/server/main.go so the dependency arrow points
// investment → contract ← market_data/adapter.
type QuoteProviderAdapter struct {
	service *mdapp.Service
}

// NewQuoteProviderAdapter wires the adapter around the running service.
func NewQuoteProviderAdapter(service *mdapp.Service) *QuoteProviderAdapter {
	return &QuoteProviderAdapter{service: service}
}

// GetLatestQuote fetches the latest quote for the given provider symbol. On
// total provider failure the service returns the latest cached snapshot with
// Stale=true; we propagate that as a successful response so callers do not
// need to distinguish "stale" from "live" at the call site.
func (a *QuoteProviderAdapter) GetLatestQuote(ctx context.Context, providerSymbol string) (*contract.MarketQuote, error) {
	if a == nil || a.service == nil {
		return nil, contract.ErrMarketDataUnavailable
	}
	providerSymbol = strings.TrimSpace(providerSymbol)
	if providerSymbol == "" {
		return nil, contract.ErrSymbolNotMapped
	}

	q, err := a.service.GetQuote(ctx, providerSymbol)
	if err != nil {
		return nil, translateError(err)
	}
	if q == nil {
		return nil, contract.ErrMarketDataUnavailable
	}
	// A stale result here means the live provider call failed and the service
	// fell back to the last cached snapshot row — that is NOT a live quote,
	// so it must be labeled as a snapshot source or callers (and the frontend
	// stale badge) will show a provider outage as fresh, live data.
	source := contract.QuoteSourceLive
	if q.Stale {
		source = contract.QuoteSourceMarketDataSnapshot
	}
	return &contract.MarketQuote{
		Symbol:        q.Symbol,
		Provider:      q.Provider,
		Price:         q.Price,
		Currency:      q.Currency,
		PreviousClose: q.PreviousClose,
		ChangePercent: q.ChangePercent,
		EffectiveAt:   q.AsOf,
		FetchedAt:     timeOr(q.CapturedAt, time.Now().UTC()),
		MarketStatus:  marketStatusFromQuote(q),
		Stale:         q.Stale,
		StaleReason:   q.StaleReason,
		Source:        source,
		SnapshotID:    q.SnapshotID,
	}, nil
}

// GetQuoteAsOf resolves the latest normalized market-data snapshot dated on
// or before businessDate. It never performs a live provider call, so a
// historical business_date never triggers a provider request. Staleness is
// derived by comparing the snapshot's own date to the requested businessDate
// — a snapshot dated earlier than requested is still a valid "latest known
// as of" mark, but is flagged stale so callers can label it as carried
// forward rather than an exact match.
func (a *QuoteProviderAdapter) GetQuoteAsOf(ctx context.Context, providerSymbol string, businessDate time.Time) (*contract.MarketQuote, error) {
	if a == nil || a.service == nil {
		return nil, contract.ErrMarketDataUnavailable
	}
	providerSymbol = strings.TrimSpace(providerSymbol)
	if providerSymbol == "" {
		return nil, contract.ErrSymbolNotMapped
	}

	q, err := a.service.GetQuoteAsOf(ctx, providerSymbol, businessDate)
	if err != nil {
		return nil, translateError(err)
	}
	if q == nil {
		return nil, contract.ErrMarketDataUnavailable
	}

	stale := q.AsOf.IsZero() || !sameCalendarDate(q.AsOf, businessDate)
	reason := ""
	if stale {
		reason = fmt.Sprintf(
			"no market-data snapshot dated %s; carried forward from %s",
			businessDate.Format("2006-01-02"), q.AsOf.Format("2006-01-02"),
		)
	}

	return &contract.MarketQuote{
		Symbol:        q.Symbol,
		Provider:      q.Provider,
		Price:         q.Price,
		Currency:      q.Currency,
		PreviousClose: q.PreviousClose,
		ChangePercent: q.ChangePercent,
		EffectiveAt:   q.AsOf,
		FetchedAt:     timeOr(q.CapturedAt, q.AsOf),
		MarketStatus:  "SNAPSHOT",
		Stale:         stale,
		StaleReason:   reason,
		Source:        contract.QuoteSourceMarketDataSnapshot,
		SnapshotID:    q.SnapshotID,
	}, nil
}

// PrimaryProviderName returns the configured primary provider tag.
func (a *QuoteProviderAdapter) PrimaryProviderName() string {
	if a == nil || a.service == nil {
		return ""
	}
	health, err := a.service.ProviderHealth(context.Background())
	if err != nil {
		return ""
	}
	for _, h := range health {
		if h.Role == "primary" {
			return h.ProviderName
		}
	}
	return ""
}

// ProviderConfigured reports whether the named provider has credentials wired.
func (a *QuoteProviderAdapter) ProviderConfigured(name string) bool {
	if a == nil || a.service == nil || strings.TrimSpace(name) == "" {
		return false
	}
	health, err := a.service.ProviderHealth(context.Background())
	if err != nil {
		return false
	}
	want := strings.ToLower(strings.TrimSpace(name))
	for _, h := range health {
		if strings.ToLower(h.ProviderName) == want {
			return h.Configured
		}
	}
	return false
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %s", contract.ErrProviderTimeout, err.Error())
	}
	var perr *mddomain.ProviderError
	if errors.As(err, &perr) && perr != nil {
		switch strings.ToLower(perr.Code) {
		case "rate_limited", "rate_limit", "throttled":
			return fmt.Errorf("%w: %s", contract.ErrRateLimited, perr.Message)
		case "unauthorized", "forbidden", "missing_api_key":
			return fmt.Errorf("%w: %s", contract.ErrProviderUnauthorized, perr.Message)
		case "timeout":
			return fmt.Errorf("%w: %s", contract.ErrProviderTimeout, perr.Message)
		case "symbol_not_found", "no_data":
			return fmt.Errorf("%w: %s", contract.ErrSymbolNotMapped, perr.Message)
		case "malformed", "malformed_response", "parse_error":
			return fmt.Errorf("%w: %s", contract.ErrMalformedResponse, perr.Message)
		}
	}
	return fmt.Errorf("%w: %s", contract.ErrMarketDataUnavailable, err.Error())
}

func marketStatusFromQuote(q *mddomain.Quote) string {
	if q == nil {
		return "UNKNOWN"
	}
	if q.Cached || q.Stale {
		return "CACHED"
	}
	return "LIVE"
}

func timeOr(t, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t
}

func sameCalendarDate(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

var _ contract.MarketQuoteProvider = (*QuoteProviderAdapter)(nil)
