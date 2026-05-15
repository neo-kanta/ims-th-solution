// Package response holds outbound HTTP DTOs for the investment module.
//
// All money / quantity values are emitted as JSON strings to preserve decimal
// precision through the wire — clients MUST parse them with a decimal library.
package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// FundResponse mirrors entity.Fund for HTTP transport.
type FundResponse struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	ShortName      string     `json:"short_name,omitempty"`
	FundCategoryID uuid.UUID  `json:"fund_category_id"`
	BaseCurrency   string     `json:"base_currency"`
	InceptionDate  string     `json:"inception_date"`
	ManagerUserID  *uuid.UUID `json:"manager_user_id,omitempty"`
	Benchmark      string     `json:"benchmark,omitempty"`
	RiskProfile    string     `json:"risk_profile,omitempty"`
	HasUnits       bool       `json:"has_units"`
	ExternalPAMRef string     `json:"external_pam_ref,omitempty"`
	Status         string     `json:"status"`
	Version        int        `json:"version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PortfolioResponse mirrors entity.Portfolio for HTTP transport.
type PortfolioResponse struct {
	ID                uuid.UUID  `json:"id"`
	FundID            uuid.UUID  `json:"fund_id"`
	Code              string     `json:"code"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	BaseCurrency      string     `json:"base_currency"`
	ValuationCurrency string     `json:"valuation_currency"`
	StrategyCode      string     `json:"strategy_code,omitempty"`
	StyleID           *uuid.UUID `json:"style_id,omitempty"`
	ManagerUserID     *uuid.UUID `json:"manager_user_id,omitempty"`
	Benchmark         string     `json:"benchmark,omitempty"`
	RiskProfile       string     `json:"risk_profile,omitempty"`
	InceptionDate     string     `json:"inception_date"`
	Status            string     `json:"status"`
	HasUnits          bool       `json:"has_units"`
	TaxLotMethod      string     `json:"tax_lot_method"`
	Version           int        `json:"version"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// InstrumentResponse mirrors entity.Instrument for HTTP transport.
type InstrumentResponse struct {
	ID              uuid.UUID      `json:"id"`
	PrimaryTicker   string         `json:"primary_ticker"`
	Name            string         `json:"name"`
	AssetClassID    uuid.UUID      `json:"asset_class_id"`
	AssetSubtypeID  uuid.UUID      `json:"asset_subtype_id"`
	Currency        string         `json:"currency"`
	CountryID       uuid.UUID      `json:"country_id"`
	RegionID        *uuid.UUID     `json:"region_id,omitempty"`
	PrimaryExchange string         `json:"primary_exchange,omitempty"`
	SectorID        *uuid.UUID     `json:"sector_id,omitempty"`
	FundCategoryID  *uuid.UUID     `json:"fund_category_id,omitempty"`
	LotSize         int            `json:"lot_size"`
	TickSize        string         `json:"tick_size,omitempty"`
	IsTradable      bool           `json:"is_tradable"`
	Status          string         `json:"status"`
	Attributes      map[string]any `json:"attributes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// HoldingResponse mirrors entity.PortfolioPosition.
type HoldingResponse struct {
	InstrumentID      uuid.UUID  `json:"instrument_id"`
	Quantity          string     `json:"quantity"`
	AverageCost       string     `json:"average_cost"`
	CostBasis         string     `json:"cost_basis"`
	LastBusinessDate  *string    `json:"last_business_date,omitempty"`
	LastTransactionID *uuid.UUID `json:"last_transaction_id,omitempty"`
	Version           int        `json:"version"`
}

// TransactionResponse mirrors entity.PortfolioTransaction.
type TransactionResponse struct {
	ID                    uuid.UUID  `json:"id"`
	PortfolioID           uuid.UUID  `json:"portfolio_id"`
	FundID                uuid.UUID  `json:"fund_id"`
	InstrumentID          *uuid.UUID `json:"instrument_id,omitempty"`
	TransactionType       string     `json:"transaction_type"`
	Side                  string     `json:"side,omitempty"`
	Quantity              string     `json:"quantity,omitempty"`
	Price                 string     `json:"price,omitempty"`
	Currency              string     `json:"currency"`
	GrossAmount           string     `json:"gross_amount"`
	Fees                  string     `json:"fees"`
	NetAmount             string     `json:"net_amount"`
	FxRateToBase          string     `json:"fx_rate_to_base,omitempty"`
	BusinessDate          string     `json:"business_date"`
	SettlementDate        string     `json:"settlement_date,omitempty"`
	SourceDecisionID      *uuid.UUID `json:"source_decision_id,omitempty"`
	ReversesTransactionID *uuid.UUID `json:"reverses_transaction_id,omitempty"`
	ExternalRef           string     `json:"external_ref,omitempty"`
	Reason                string     `json:"reason,omitempty"`
	Status                string     `json:"status"`
	CreatedAt             time.Time  `json:"created_at"`
	CreatedBy             uuid.UUID  `json:"created_by"`
}

// CashBalanceResponse mirrors entity.CashBalance.
type CashBalanceResponse struct {
	Currency         string  `json:"currency"`
	Balance          string  `json:"balance"`
	LastBusinessDate *string `json:"last_business_date,omitempty"`
	Version          int     `json:"version"`
}

// ValuationLineResponse mirrors entity.ValuationHoldingLine.
type ValuationLineResponse struct {
	InstrumentID    uuid.UUID  `json:"instrument_id"`
	PriceSnapshotID *uuid.UUID `json:"price_snapshot_id,omitempty"`
	Quantity        string     `json:"quantity"`
	PriceInQuoteCcy string     `json:"price_in_quote_ccy"`
	QuoteCurrency   string     `json:"quote_currency"`
	FxRate          string     `json:"fx_rate_to_valuation_ccy"`
	MarketValue     string     `json:"market_value"`
	CostBasis       string     `json:"cost_basis"`
	UnrealisedPnL   string     `json:"unrealised_pnl"`
	IsStale         bool       `json:"is_stale"`
}

// ValuationResponse mirrors entity.ValuationSnapshot (with optional lines).
type ValuationResponse struct {
	ID             uuid.UUID               `json:"id"`
	PortfolioID    uuid.UUID               `json:"portfolio_id"`
	BusinessDate   string                  `json:"business_date"`
	ValuationCcy   string                  `json:"valuation_ccy"`
	MarketValue    string                  `json:"market_value"`
	CostBasis      string                  `json:"cost_basis"`
	UnrealisedPnL  string                  `json:"unrealised_pnl"`
	RealisedPnL    string                  `json:"realised_pnl"`
	ROI            string                  `json:"roi,omitempty"`
	AUM            string                  `json:"aum"`
	CashBalance    string                  `json:"cash_balance"`
	PriceSetHash   string                  `json:"price_set_hash"`
	HasStaleInputs bool                    `json:"has_stale_inputs"`
	IsIndicative   bool                    `json:"is_indicative"`
	Source         string                  `json:"source"`
	CreatedAt      time.Time               `json:"created_at"`
	HoldingLines   []ValuationLineResponse `json:"holding_lines,omitempty"`
}

// NAVResponse mirrors entity.NAVSnapshot.
type NAVResponse struct {
	ID                  uuid.UUID `json:"id"`
	PortfolioID         uuid.UUID `json:"portfolio_id"`
	BusinessDate        string    `json:"business_date"`
	TotalUnits          string    `json:"total_units"`
	NAVPerUnit          string    `json:"nav_per_unit"`
	ValuationSnapshotID uuid.UUID `json:"valuation_snapshot_id"`
	IsIndicative        bool      `json:"is_indicative"`
	CreatedAt           time.Time `json:"created_at"`
}

// AUMResponse mirrors entity.AUMSnapshot.
type AUMResponse struct {
	ID           uuid.UUID `json:"id"`
	ScopeType    string    `json:"scope_type"`
	ScopeID      uuid.UUID `json:"scope_id"`
	BusinessDate string    `json:"business_date"`
	AUM          string    `json:"aum"`
	ValuationCcy string    `json:"valuation_ccy"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}

// PriceResponse mirrors entity.PriceSnapshot.
type PriceResponse struct {
	ID           uuid.UUID `json:"id"`
	InstrumentID uuid.UUID `json:"instrument_id"`
	BusinessDate string    `json:"business_date"`
	Price        string    `json:"price"`
	Currency     string    `json:"currency"`
	PriceSource  string    `json:"price_source"`
	ProviderRef  string    `json:"provider_ref,omitempty"`
	IsStale      bool      `json:"is_stale"`
	CapturedAt   time.Time `json:"captured_at"`
}

// PortfolioSummaryResponse aggregates the most useful headline metrics for a
// portfolio detail view.
type PortfolioSummaryResponse struct {
	PortfolioID     uuid.UUID             `json:"portfolio_id"`
	HoldingCount    int                   `json:"holding_count"`
	NonZeroHoldings int                   `json:"non_zero_holdings"`
	CashBalances    []CashBalanceResponse `json:"cash_balances"`
	LatestValuation *ValuationResponse    `json:"latest_valuation,omitempty"`
}

// PaginatedResponse is a generic envelope for list endpoints.
type PaginatedResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type FundListResponse struct {
	Items []FundResponse `json:"items"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type PortfolioListResponse struct {
	Items []PortfolioResponse `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

type InstrumentListResponse struct {
	Items []InstrumentResponse `json:"items"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

type TransactionListResponse struct {
	Items []TransactionResponse `json:"items"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

type ValuationListResponse struct {
	Items []ValuationResponse `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// ComputeFundAUMResponse is the body returned by
// POST /investment/funds/{id}/aum/compute. `idempotent` is true when a
// snapshot for (fund, date, source=INTERNAL) already existed and the
// returned snapshot is the pre-existing row rather than a fresh insert.
type ComputeFundAUMResponse struct {
	Snapshot       AUMResponse `json:"snapshot"`
	PortfolioCount int         `json:"portfolio_count"`
	Idempotent     bool        `json:"idempotent"`
}

// FormatDecimal returns "" for nil pointers, otherwise a decimal string.
// Centralised here so handlers all serialise the same way.
func FormatDecimal(d *decimal.Decimal) string {
	if d == nil {
		return ""
	}
	return d.String()
}

// FormatDate returns YYYY-MM-DD.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDatePtr returns a pointer to YYYY-MM-DD or nil.
func FormatDatePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
