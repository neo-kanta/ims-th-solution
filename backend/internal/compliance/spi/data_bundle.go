package spi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DataBundle holds all pre-fetched data for rule evaluation.
// Fields are nil if not requested in DataDependencies.
type DataBundle struct {
	Positions       *PositionSnapshot       `json:"positions,omitempty"`
	NAV             *NAVSnapshot            `json:"nav,omitempty"`
	MarketPrices    *MarketPriceSnapshot    `json:"market_prices,omitempty"`
	FXRates         *FXRateSnapshot         `json:"fx_rates,omitempty"`
	Classifications *ClassificationSnapshot `json:"classifications,omitempty"`
	CreditRatings   *CreditRatingSnapshot   `json:"credit_ratings,omitempty"`
	Restrictions    *RestrictionSnapshot    `json:"restrictions,omitempty"`
	TradeHistory    *TradeHistorySnapshot   `json:"trade_history,omitempty"`
	Calendar        *CalendarSnapshot       `json:"calendar,omitempty"`
	PortfolioMeta   *PortfolioMetadata      `json:"portfolio_meta,omitempty"`
}

// ComputeHash returns a SHA-256 hex digest of the bundle for audit reproducibility.
func (b *DataBundle) ComputeHash() string {
	raw, _ := json.Marshal(b)
	h := sha256.Sum256(raw)
	return fmt.Sprintf("%x", h)
}

// --- Position data ---

// PositionSnapshot holds current holdings and pending orders for a portfolio.
type PositionSnapshot struct {
	AsOf          time.Time          `json:"as_of"`
	PortfolioID   uuid.UUID          `json:"portfolio_id"`
	Holdings      []Holding          `json:"holdings"`
	PendingOrders []PendingOrderInfo `json:"pending_orders"`
}

// Holding is a single position in a portfolio.
type Holding struct {
	Ticker      string          `json:"ticker"`
	ISIN        string          `json:"isin"`
	Quantity    decimal.Decimal `json:"quantity"`
	CostBasis   decimal.Decimal `json:"cost_basis"`
	MarketValue decimal.Decimal `json:"market_value"`
}

// PendingOrderInfo is an unconfirmed order affecting available quantity/cash.
type PendingOrderInfo struct {
	OrderID  uuid.UUID       `json:"order_id"`
	Ticker   string          `json:"ticker"`
	Side     string          `json:"side"`
	Quantity decimal.Decimal `json:"quantity"`
	Price    decimal.Decimal `json:"price"`
}

// --- NAV data ---

// NAVSnapshot holds NAV and cash data for a portfolio.
type NAVSnapshot struct {
	PortfolioID  uuid.UUID       `json:"portfolio_id"`
	AsOf         time.Time       `json:"as_of"`
	NAV          decimal.Decimal `json:"nav"`
	CashBalance  decimal.Decimal `json:"cash_balance"`
	ReservedCash decimal.Decimal `json:"reserved_cash"`
}

// --- Market prices ---

// MarketPriceSnapshot holds current market prices keyed by ticker.
type MarketPriceSnapshot struct {
	Prices map[string]decimal.Decimal `json:"prices"`
	AsOf   time.Time                  `json:"as_of"`
}

// --- FX rates ---

// FXRateSnapshot holds FX rates for a base currency.
type FXRateSnapshot struct {
	BaseCurrency string                     `json:"base_currency"`
	Rates        map[string]decimal.Decimal `json:"rates"`
	AsOf         time.Time                  `json:"as_of"`
}

// --- Classifications ---

// ClassificationSnapshot holds instrument classification data.
type ClassificationSnapshot struct {
	data map[string]InstrumentClassification
}

// InstrumentClassification holds classification attributes for one instrument.
type InstrumentClassification struct {
	Ticker       string `json:"ticker"`
	ISIN         string `json:"isin"`
	Issuer       string `json:"issuer"`
	ParentEntity string `json:"parent_entity"`
	Sector       string `json:"sector"`
	AssetClass   string `json:"asset_class"`
	Country      string `json:"country"`
	Exchange     string `json:"exchange"`
	IsGovernment bool   `json:"is_government"`
}

// NewClassificationSnapshot creates a snapshot from a slice.
func NewClassificationSnapshot(items []InstrumentClassification) *ClassificationSnapshot {
	m := make(map[string]InstrumentClassification, len(items))
	for _, item := range items {
		m[item.Ticker] = item
	}
	return &ClassificationSnapshot{data: m}
}

// Get returns classification for a ticker.
func (s *ClassificationSnapshot) Get(ticker string) (InstrumentClassification, bool) {
	if s == nil || s.data == nil {
		return InstrumentClassification{}, false
	}
	c, ok := s.data[ticker]
	return c, ok
}

// DirectIssuer returns the direct issuer or falls back to ticker.
func (s *ClassificationSnapshot) DirectIssuer(ticker string) string {
	if c, ok := s.Get(ticker); ok && c.Issuer != "" {
		return c.Issuer
	}
	return ticker
}

// ParentEntity returns the parent entity or falls back to direct issuer.
func (s *ClassificationSnapshot) ParentEntity(ticker string) string {
	if c, ok := s.Get(ticker); ok && c.ParentEntity != "" {
		return c.ParentEntity
	}
	return s.DirectIssuer(ticker)
}

// IsGovernment returns whether the instrument is government-issued.
func (s *ClassificationSnapshot) IsGovernment(ticker string) bool {
	if c, ok := s.Get(ticker); ok {
		return c.IsGovernment
	}
	return false
}

// Sector returns the sector for a ticker.
func (s *ClassificationSnapshot) Sector(ticker string) string {
	if c, ok := s.Get(ticker); ok {
		return c.Sector
	}
	return ""
}

// MarshalJSON implements json.Marshaler for audit serialization.
func (s *ClassificationSnapshot) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	return json.Marshal(s.data)
}

// --- Credit ratings ---

// CreditRatingSnapshot holds credit ratings keyed by issuer.
type CreditRatingSnapshot struct {
	Ratings map[string]CreditRating `json:"ratings"`
}

// CreditRating holds a single issuer's credit rating.
type CreditRating struct {
	Rating string    `json:"rating"`
	Agency string    `json:"agency"`
	AsOf   time.Time `json:"as_of"`
}

// --- Restriction lists ---

// RestrictionSnapshot holds blacklist, whitelist, and gray list data.
type RestrictionSnapshot struct {
	Blacklisted  map[string]RestrictionEntry `json:"blacklisted"`
	Whitelisted  map[string]bool             `json:"whitelisted"`
	HasWhitelist bool                        `json:"has_whitelist"`
	GrayListed   map[string]RestrictionEntry `json:"gray_listed"`
}

// RestrictionEntry is a single restriction list entry.
type RestrictionEntry struct {
	Ticker   string `json:"ticker"`
	Reason   string `json:"reason"`
	ListType string `json:"list_type"` // BLACKLIST, GRAYLIST
	Source   string `json:"source"`    // GLOBAL, CONTRACT, REGULATORY
}

// --- Trade history ---

// TradeHistorySnapshot holds historical trades for look-back rules.
type TradeHistorySnapshot struct {
	PortfolioID  uuid.UUID         `json:"portfolio_id"`
	LookbackDays int               `json:"lookback_days"`
	Trades       []HistoricalTrade `json:"trades"`
}

// HistoricalTrade is a single historical trade.
type HistoricalTrade struct {
	TradeDate time.Time       `json:"trade_date"`
	Ticker    string          `json:"ticker"`
	Side      string          `json:"side"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	TradeID   uuid.UUID       `json:"trade_id"`
}

// --- Calendar ---

// CalendarSnapshot holds business day information.
type CalendarSnapshot struct {
	BusinessDays map[string]bool `json:"business_days"` // "2006-01-02" -> true
	Holidays     []time.Time     `json:"holidays"`
}

// IsBusinessDay returns true if the date is a business day.
func (c *CalendarSnapshot) IsBusinessDay(date time.Time) bool {
	if c == nil || c.BusinessDays == nil {
		return true // default assumption
	}
	return c.BusinessDays[date.Format("2006-01-02")]
}

// --- Portfolio metadata ---

// PortfolioMetadata holds mandate-level information.
type PortfolioMetadata struct {
	PortfolioID   uuid.UUID `json:"portfolio_id"`
	ContractID    uuid.UUID `json:"contract_id"`
	MandateType   string    `json:"mandate_type"`
	BaseCurrency  string    `json:"base_currency"`
	Jurisdiction  string    `json:"jurisdiction"`
	InceptionDate time.Time `json:"inception_date"`
}
