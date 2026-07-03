package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrInvalidSecurity          = errors.New("watchlist: invalid or inactive security")
	ErrDuplicateWatchlistItem   = errors.New("watchlist: duplicate active item for scope and security")
	ErrInvalidQuery             = errors.New("watchlist: invalid query")
	ErrInvalidThreshold         = errors.New("watchlist: invalid threshold rule")
	ErrForbiddenScope           = errors.New("watchlist: forbidden scope")
	ErrItemNotFound             = errors.New("watchlist: item not found")
	ErrAlertNotFound            = errors.New("watchlist: alert event not found")
	ErrStaleQuote               = errors.New("watchlist: quote is stale beyond max age")
	ErrProviderUnavailable      = errors.New("watchlist: market data provider unavailable")
	ErrAlertAlreadyAcknowledged = errors.New("watchlist: alert already acknowledged")
	ErrRuleDisabled             = errors.New("watchlist: rule is disabled")
	ErrAlertIdempotencyConflict = errors.New("watchlist: alert idempotency key conflict")
	ErrDuplicateThresholdRule   = errors.New("watchlist: duplicate active threshold rule")
)

// ─── Security port ────────────────────────────────────────────────────────────

// SecurityInfo is a narrow snapshot of a canonical security.
type SecurityInfo struct {
	ID            string
	IMSSymbol     string
	DisplaySymbol string
	Name          string
	AssetType     string
	Currency      string
	ExchangeMIC   string
	Status        string // "ACTIVE" when available for use
}

// ProviderMapping is a security → provider symbol mapping.
type ProviderMapping struct {
	ProviderCode   string
	ProviderSymbol string
}

// SecurityPort is the narrow lookup dependency injected by the reference_data module.
type SecurityPort interface {
	GetSecurityByID(ctx context.Context, id string) (*SecurityInfo, error)
	ResolveProviderSymbol(ctx context.Context, securityID, providerCode string) (*ProviderMapping, error)
}

// ─── Quote port ───────────────────────────────────────────────────────────────

// QuoteInfo is the market quote used by the evaluator.
type QuoteInfo struct {
	Symbol        string
	Provider      string
	Currency      string
	MarketStatus  string
	Price         decimal.Decimal
	PreviousClose decimal.Decimal
	ChangePercent decimal.Decimal
	EffectiveAt   time.Time
	FetchedAt     time.Time
	Stale         bool
	StaleReason   string
}

// QuotePort is the narrow market-data dependency injected by the market_data module.
type QuotePort interface {
	GetLatestQuote(ctx context.Context, providerSymbol string) (*QuoteInfo, error)
	PrimaryProviderName() string
}

// ─── Portfolio scope port ─────────────────────────────────────────────────────

// PortfolioScopeInfo carries the fields needed for data-permission checks and
// descriptor hydration.
type PortfolioScopeInfo struct {
	PortfolioID   uuid.UUID
	PortfolioCode string
	PortfolioName string
	FundID        uuid.UUID
	FundCode      string
	FundName      string
}

// PortfolioScopePort resolves a portfolio ID to its IAM data scope and
// descriptor fields.
type PortfolioScopePort interface {
	GetPortfolioScope(ctx context.Context, portfolioID uuid.UUID) (*PortfolioScopeInfo, error)
}

// ─── Notification port ────────────────────────────────────────────────────────

// WatchlistAlertNotificationInput contains the data passed to the notifier.
type WatchlistAlertNotificationInput struct {
	RecipientUserID uuid.UUID
	AlertEventID    uuid.UUID
	WatchlistItemID uuid.UUID
	ThresholdRuleID uuid.UUID
	SecurityID      uuid.UUID
	SecurityName    string
	SecuritySymbol  string
	PortfolioID     *uuid.UUID
	ScopeType       string
	Direction       string
	ObservedPrice   decimal.Decimal
	ThresholdValue  decimal.Decimal
	Currency        string
	Stale           bool
	StaleReason     string
	IdempotencyKey  string
}

// WatchlistAlertNotifier delivers breach notifications.
type WatchlistAlertNotifier interface {
	NotifyThresholdBreached(ctx context.Context, input WatchlistAlertNotificationInput) error
}

// ─── User lookup port ─────────────────────────────────────────────────────────

// UserInfo carries display information for a user.
type UserInfo struct {
	UserID      uuid.UUID
	DisplayName string
}

// UserLookupPort resolves user IDs to display information.
// Implemented by querying the iam_users table directly (read-only projection).
type UserLookupPort interface {
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*UserInfo, error)
}

// ─── Audit port ───────────────────────────────────────────────────────────────

// WatchlistAuditRecorder records watchlist audit events.
type WatchlistAuditRecorder interface {
	Record(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID, ipAddress, userAgent string, metadata map[string]interface{})
}
