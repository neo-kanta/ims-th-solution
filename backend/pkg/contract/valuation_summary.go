package contract

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValuationSummaryScope selects the aggregation breadth for a dashboard
// valuation summary request.
type ValuationSummaryScope string

const (
	// ValuationSummaryScopeCompany aggregates every fund the caller's data
	// scope allows (i.e. all data the authenticated user is authorized to
	// see — not necessarily literally every fund in the system).
	ValuationSummaryScopeCompany ValuationSummaryScope = "company"
	// ValuationSummaryScopeMine restricts the aggregate to funds the caller
	// manages (Fund.ManagerUserID == UserID), still bounded by data scope.
	ValuationSummaryScopeMine ValuationSummaryScope = "mine"
)

// ValuationSummaryRequest is the input to ValuationSummaryProvider.
//
// AccessibleFundIDs mirrors the accessibleFundIDs() convention used by the
// investment module's own handlers: nil means "no filter" (the caller's
// data scope is the "*" wildcard), a non-nil slice (possibly empty)
// restricts the aggregate to exactly those fund IDs.
type ValuationSummaryRequest struct {
	Scope             ValuationSummaryScope
	UserID            uuid.UUID
	AccessibleFundIDs []uuid.UUID
}

// ValuationSummaryStatus states whether an authoritative reporting-currency
// total can be presented. INCOMPLETE deliberately carries no usable total:
// callers may inspect Coverage for audit/operational diagnostics, but must not
// present a partial subtotal as company or manager AUM.
type ValuationSummaryStatus string

const (
	ValuationSummaryStatusAvailable  ValuationSummaryStatus = "AVAILABLE"
	ValuationSummaryStatusNoData     ValuationSummaryStatus = "NO_DATA"
	ValuationSummaryStatusIncomplete ValuationSummaryStatus = "INCOMPLETE"
)

// ValuationSummaryExclusionReason is a stable, machine-readable explanation
// for why a scoped fund or portfolio did not contribute to the official total.
type ValuationSummaryExclusionReason string

const (
	ValuationSummaryExclusionNoActivePortfolio ValuationSummaryExclusionReason = "NO_ACTIVE_PORTFOLIO"
	ValuationSummaryExclusionNonOfficial       ValuationSummaryExclusionReason = "NON_OFFICIAL_PORTFOLIO"
	ValuationSummaryExclusionMissingValuation  ValuationSummaryExclusionReason = "MISSING_VALUATION"
	ValuationSummaryExclusionStaleValuation    ValuationSummaryExclusionReason = "STALE_VALUATION"
	ValuationSummaryExclusionValuationDate     ValuationSummaryExclusionReason = "VALUATION_DATE_MISMATCH"
	ValuationSummaryExclusionValuationCurrency ValuationSummaryExclusionReason = "VALUATION_CURRENCY_MISMATCH"
	ValuationSummaryExclusionMissingFX         ValuationSummaryExclusionReason = "MISSING_FX_RATE"
	ValuationSummaryExclusionStaleFX           ValuationSummaryExclusionReason = "STALE_FX_RATE"
	ValuationSummaryExclusionWrongDateFX       ValuationSummaryExclusionReason = "WRONG_BUSINESS_DATE_FX_RATE"
	ValuationSummaryExclusionInvalidFX         ValuationSummaryExclusionReason = "INVALID_FX_RATE"
	ValuationSummaryExclusionFXCurrency        ValuationSummaryExclusionReason = "FX_QUOTE_CURRENCY_MISMATCH"
	ValuationSummaryExclusionFXSymbol          ValuationSummaryExclusionReason = "FX_SYMBOL_MISMATCH"
)

// ValuationSummaryExclusion identifies one excluded scoped item without
// exposing internal UUIDs. Business codes are suitable for operator-facing
// diagnostics and preserve the portfolio-first identity rule.
type ValuationSummaryExclusion struct {
	FundCode             string
	PortfolioCode        string
	Currency             string
	BusinessDate         time.Time
	RequiredBusinessDate time.Time
	Reason               ValuationSummaryExclusionReason
}

// ValuationSummaryCoverage reports exactly how much of the authenticated
// scope was eligible for the official reporting-currency total.
//
// TotalPortfolioCount includes every active portfolio. SIMULATION and MODEL
// portfolios appear as NON_OFFICIAL_PORTFOLIO exclusions but do not by
// themselves make a LIVE-only official total incomplete.
type ValuationSummaryCoverage struct {
	TotalFundCount         int
	IncludedFundCount      int
	ExcludedFundCount      int
	TotalPortfolioCount    int
	IncludedPortfolioCount int
	ExcludedPortfolioCount int
	ExcludedCurrencies     []string
	ExcludedBusinessDates  []time.Time
	ExclusionReasons       []ValuationSummaryExclusionReason
	Exclusions             []ValuationSummaryExclusion
}

// ValuationSummaryResult is the official aggregate AUM / today's P&L for the
// requested scope, expressed in the configured reporting Currency.
// DataAvailable is true only for AVAILABLE. NO_DATA and INCOMPLETE must be
// rendered as unavailable; INCOMPLETE exposes Coverage diagnostics but no
// usable partial numeric total.
type ValuationSummaryResult struct {
	Currency        string
	BusinessDate    time.Time
	AsOf            time.Time
	AUM             decimal.Decimal
	TodayPnL        decimal.Decimal
	TodayPnLPercent *decimal.Decimal
	DataAvailable   bool
	Status          ValuationSummaryStatus
	Coverage        ValuationSummaryCoverage
}

// ValuationSummaryProvider aggregates today's AUM and P&L across a scoped
// set of funds. Implemented by the investment module; consumed by the
// integration module's dashboard valuation-summary endpoint.
type ValuationSummaryProvider interface {
	GetValuationSummary(ctx context.Context, req ValuationSummaryRequest) (*ValuationSummaryResult, error)
}
