package domain

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ValuationScope selects the aggregation breadth for a dashboard valuation
// summary request.
type ValuationScope string

const (
	ValuationScopeCompany ValuationScope = "company"
	ValuationScopeMine    ValuationScope = "mine"
)

// ValuationSummary is the dashboard read model for aggregate AUM and today's
// P&L, scoped to either "company" or "mine".
//
// DataAvailable is false when no fund in scope has a valuation snapshot yet
// (or the caller lacks dashboard permission) — the numeric fields are zero
// in that case and MUST be rendered as an explicit "not available" state,
// never as a misleading zero.
type ValuationSummary struct {
	Scope           ValuationScope
	Username        string // set only for ValuationScopeMine
	Status          contract.ValuationSummaryStatus
	BusinessDate    time.Time
	Currency        string
	AUM             decimal.Decimal
	TodayPnL        decimal.Decimal
	TodayPnLPercent *decimal.Decimal
	AsOf            time.Time
	DataAvailable   bool
	Coverage        contract.ValuationSummaryCoverage
}
