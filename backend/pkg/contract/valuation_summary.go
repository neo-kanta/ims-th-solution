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

// ValuationSummaryResult is the aggregate AUM / today's P&L for the
// requested scope, expressed in Currency. DataAvailable is false when no
// fund in scope has a valuation snapshot yet — callers must render an
// explicit "not available" state rather than a misleading zero.
type ValuationSummaryResult struct {
	Currency        string
	BusinessDate    time.Time
	AsOf            time.Time
	AUM             decimal.Decimal
	TodayPnL        decimal.Decimal
	TodayPnLPercent *decimal.Decimal
	DataAvailable   bool
	FundCount       int
}

// ValuationSummaryProvider aggregates today's AUM and P&L across a scoped
// set of funds. Implemented by the investment module; consumed by the
// integration module's dashboard valuation-summary endpoint.
type ValuationSummaryProvider interface {
	GetValuationSummary(ctx context.Context, req ValuationSummaryRequest) (*ValuationSummaryResult, error)
}
