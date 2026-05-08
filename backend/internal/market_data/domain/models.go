package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type AssetType string

const (
	AssetTypeUnknown    AssetType = "UNKNOWN"
	AssetTypeEquity     AssetType = "EQUITY"
	AssetTypeETF        AssetType = "ETF"
	AssetTypeMutualFund AssetType = "MUTUAL_FUND"
)

const (
	ProviderAlphaVantage = "alpha_vantage"
	ProviderYahoo        = "yahoo"
)

type Quote struct {
	Symbol        string          `json:"symbol"`
	Provider      string          `json:"provider"`
	AssetType     AssetType       `json:"asset_type,omitempty"`
	Price         decimal.Decimal `json:"price"`
	Open          decimal.Decimal `json:"open,omitempty"`
	High          decimal.Decimal `json:"high,omitempty"`
	Low           decimal.Decimal `json:"low,omitempty"`
	PreviousClose decimal.Decimal `json:"previous_close,omitempty"`
	Change        decimal.Decimal `json:"change,omitempty"`
	ChangePercent decimal.Decimal `json:"change_percent,omitempty"`
	Volume        int64           `json:"volume,omitempty"`
	Currency      string          `json:"currency,omitempty"`
	AsOf          time.Time       `json:"as_of"`
	CapturedAt    time.Time       `json:"captured_at"`
	Stale         bool            `json:"stale"`
	StaleReason   string          `json:"stale_reason,omitempty"`
	Cached        bool            `json:"cached,omitempty"`
}

type PriceBar struct {
	Symbol        string           `json:"symbol"`
	Provider      string           `json:"provider"`
	Date          time.Time        `json:"date"`
	Open          decimal.Decimal  `json:"open"`
	High          decimal.Decimal  `json:"high"`
	Low           decimal.Decimal  `json:"low"`
	Close         decimal.Decimal  `json:"close"`
	AdjustedClose *decimal.Decimal `json:"adjusted_close,omitempty"`
	Volume        int64            `json:"volume,omitempty"`
	Currency      string           `json:"currency,omitempty"`
	Stale         bool             `json:"stale"`
	StaleReason   string           `json:"stale_reason,omitempty"`
}

type ProviderError struct {
	Provider   string `json:"provider"`
	Operation  string `json:"operation,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Retriable  bool   `json:"retriable"`
	Cause      error  `json:"-"`
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return fmt.Sprintf("%s %s failed [%s]: %s", e.Provider, e.Operation, e.Code, e.Message)
	}
	return fmt.Sprintf("%s %s failed: %s", e.Provider, e.Operation, e.Message)
}

func (e *ProviderError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type SymbolMapping struct {
	Symbol              string    `json:"symbol"`
	AssetType           AssetType `json:"asset_type"`
	Name                string    `json:"name,omitempty"`
	Currency            string    `json:"currency,omitempty"`
	AlphaVantageSymbol  string    `json:"alpha_vantage_symbol,omitempty"`
	YahooFinanceSymbol  string    `json:"yahoo_finance_symbol,omitempty"`
	ProviderDescription string    `json:"provider_description,omitempty"`
}

type MarketDataProvider interface {
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetDailyPrices(ctx context.Context, symbol string) ([]PriceBar, error)
	ProviderName() string
}

type SnapshotRepository interface {
	UpsertSymbol(ctx context.Context, mapping SymbolMapping) error
	GetSymbolMapping(ctx context.Context, symbol string) (*SymbolMapping, error)
	SaveQuote(ctx context.Context, quote Quote) error
	SaveDailyPrices(ctx context.Context, symbol string, provider string, bars []PriceBar) error
	GetLatestQuote(ctx context.Context, symbol string) (*Quote, error)
	ListDailyPrices(ctx context.Context, symbol string, limit int) ([]PriceBar, error)
}

type ProviderRequestLogger interface {
	LogProviderRequest(ctx context.Context, log ProviderRequestLog) error
	ListLatestProviderRequests(ctx context.Context, limit int) ([]ProviderRequestLog, error)
}

type QuoteCache interface {
	GetQuote(ctx context.Context, symbol string) (*Quote, bool, error)
	SetQuote(ctx context.Context, quote Quote, ttl time.Duration) error
}

type ProviderRequestLog struct {
	ProviderName string
	Symbol       string
	Operation    string
	Status       string
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
	Duration     time.Duration
	RequestedAt  time.Time
}

type ProviderHealth struct {
	ProviderName string    `json:"provider_name"`
	Role         string    `json:"role"`
	Configured   bool      `json:"configured"`
	Official     bool      `json:"official"`
	Healthy      bool      `json:"healthy"`
	LastStatus   string    `json:"last_status,omitempty"`
	LastError    string    `json:"last_error,omitempty"`
	LastChecked  time.Time `json:"last_checked,omitempty"`
}

var ErrNotConfigured = errors.New("market data provider is not configured")
