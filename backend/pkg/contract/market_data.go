package contract

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// Quote source tags surfaced on MarketQuote.Source. Callers (e.g. the
// investment holdings valuation) use these to label the "source" field of a
// mark-to-market position without needing to know how market_data resolved
// the price.
const (
	QuoteSourceLive               = "live_quote"
	QuoteSourceMarketDataSnapshot = "market_data_snapshot"
)

// MarketQuote is the minimal cross-module shape of a live (or last-known)
// market data observation. It is intentionally small: investment never
// imports market_data internal types — it talks to a port that returns this.
//
// Stale=true means the provider chain failed and the value is a cached
// snapshot. StaleReason carries a short, displayable explanation.
type MarketQuote struct {
	Symbol         string          // input provider symbol
	Provider       string          // provider that produced the value (e.g. "yahoo")
	Price          decimal.Decimal // last price in Currency
	Currency       string          // ISO 4217
	PreviousClose  decimal.Decimal // optional; zero when not provided
	ChangePercent  decimal.Decimal // signed, percentage points (e.g. 1.25 = +1.25%)
	EffectiveAt    time.Time       // when the price was observed by the provider
	FetchedAt      time.Time       // when the platform pulled the value
	MarketStatus   string          // optional ("OPEN","CLOSED","UNKNOWN")
	RawPayloadHash string          // optional fingerprint for audit/dedup
	Stale          bool
	StaleReason    string

	// Source is one of the QuoteSource* constants above. Empty for callers
	// that predate this field (treat as QuoteSourceLive).
	Source string
	// SnapshotID is the market_data_snapshots row id this quote was read
	// from, when known. Empty when the quote was not (yet) persisted with a
	// retrievable id (e.g. a freshly fetched live quote).
	SnapshotID string
}

// MarketQuoteProvider is the read port the investment module consumes for
// intraday quotes. Implementations live in the market_data module / adapters.
//
// GetLatestQuote MUST NOT crash the caller: on total provider failure it
// either returns the latest cached snapshot with Stale=true OR returns
// ErrMarketDataUnavailable (when no cached snapshot exists). Typed errors
// (ErrSymbolNotMapped, ErrRateLimited, ErrUnauthorized, ErrProviderTimeout,
// ErrMalformedResponse) are also valid returns.
type MarketQuoteProvider interface {
	GetLatestQuote(ctx context.Context, providerSymbol string) (*MarketQuote, error)

	// GetQuoteAsOf resolves the latest normalized market-data snapshot for
	// providerSymbol dated on or before businessDate. Unlike GetLatestQuote
	// it never performs a live provider HTTP call — it only reads persisted
	// snapshots, so it is safe to call for historical business dates.
	//
	// Returns ErrMarketDataUnavailable (not a panic/500) when no snapshot
	// exists on or before businessDate for the symbol.
	GetQuoteAsOf(ctx context.Context, providerSymbol string, businessDate time.Time) (*MarketQuote, error)

	// PrimaryProviderName returns the configured primary provider tag
	// (e.g. "alpha_vantage", "yahoo", "mock"). Used by the status endpoint.
	PrimaryProviderName() string

	// ProviderConfigured reports whether the named provider is wired and has
	// credentials. The status endpoint surfaces this in its "Feeds OK" badge.
	ProviderConfigured(name string) bool
}

// Common typed errors a MarketQuoteProvider may return. Callers compare via
// errors.Is so adapters are free to wrap with extra context.
var (
	ErrMarketDataUnavailable = errors.New("market data unavailable")
	ErrSymbolNotMapped       = errors.New("symbol not mapped for any provider")
	ErrRateLimited           = errors.New("provider rate limited")
	ErrProviderUnauthorized  = errors.New("provider unauthorized")
	ErrProviderTimeout       = errors.New("provider timeout")
	ErrMalformedResponse     = errors.New("provider malformed response")
)
