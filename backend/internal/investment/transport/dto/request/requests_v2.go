package request

import "github.com/google/uuid"

// Portfolio V2 request DTOs (docs/api/portfolio-v2-api-ddd.md, Milestone 5 of
// docs/handoff/portfolio-v2-claude-implementation-prompt.md).
//
// These mirror the V1 request shapes in requests.go with two differences:
//   - fund_id / portfolio_id are removed. The handler resolves the portfolio
//     from the {portfolioCode} route param and derives FundID from it before
//     calling the same V1 command handlers.
//   - decision_id / execution_id are removed from the nested create
//     requests (POST .../decisions/{decisionId}/executions and
//     POST .../executions/{executionId}/confirmations) because the parent
//     ID already comes from the URL path.

// CreateDecisionV2Request is the JSON body for
// POST /api/v2/portfolios/{portfolioCode}/decisions.
type CreateDecisionV2Request struct {
	InstrumentID     *uuid.UUID `json:"instrument_id"`
	InstrumentCode   string     `json:"instrument_code"    validate:"required,max=40"`
	BusinessDate     string     `json:"business_date"      validate:"required"`
	Side             string     `json:"side"               validate:"required,oneof=BUY SELL"`
	Quantity         string     `json:"quantity"`
	Amount           string     `json:"amount"`
	LimitPrice       string     `json:"limit_price"`
	Currency         string     `json:"currency"           validate:"required,len=3"`
	Exchange         string     `json:"exchange"`
	ResearchReportID *uuid.UUID `json:"research_report_id"`
	Rationale        string     `json:"rationale"`
}

// CreateExecutionV2Request is the JSON body for
// POST /api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions.
type CreateExecutionV2Request struct {
	OrderedQuantity string `json:"ordered_quantity"`
	OrderedAmount   string `json:"ordered_amount"`
	BrokerReference string `json:"broker_reference"`
}

// RecordConfirmationV2Request is the JSON body for
// POST /api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations.
type RecordConfirmationV2Request struct {
	ConfirmedQuantity string     `json:"confirmed_quantity"`
	ConfirmedAmount   string     `json:"confirmed_amount"`
	ConfirmedPrice    string     `json:"confirmed_price"`
	BrokerReference   string     `json:"broker_reference"`
	ImportBatchID     *uuid.UUID `json:"import_batch_id"`
}

// CreatePortfolioV2Request is the JSON body for POST /api/v2/portfolios.
//
// Unlike V1's CreatePortfolioRequest, this accepts a business fund_code
// instead of a raw fund_id — the handler resolves fund_code to the internal
// fund_id server-side via the fund repository (never trusting a
// client-supplied fund_id/portfolio_id/contract_id per
// docs/MANAGER/MEMORY.md's V2 request-body rule) and verifies the caller has
// data-permission on the resolved fund before creating the portfolio.
type CreatePortfolioV2Request struct {
	FundCode          string     `json:"fund_code"          validate:"required,max=40"`
	PortfolioType     string     `json:"portfolio_type"     validate:"required,oneof=LIVE SIMULATION MODEL"`
	Code              string     `json:"code"               validate:"required,max=40"`
	Name              string     `json:"name"               validate:"required,max=255"`
	Description       string     `json:"description"`
	BaseCurrency      string     `json:"base_currency"      validate:"required,len=3"`
	ValuationCurrency string     `json:"valuation_currency" validate:"required,len=3"`
	StrategyCode      string     `json:"strategy_code"`
	StyleID           *uuid.UUID `json:"style_id"`
	ManagerUserID     *uuid.UUID `json:"manager_user_id"`
	Benchmark         string     `json:"benchmark"`
	RiskProfile       string     `json:"risk_profile"`
	InceptionDate     string     `json:"inception_date"     validate:"required"`
	HasUnits          bool       `json:"has_units"`
	TaxLotMethod      string     `json:"tax_lot_method"`
}

// PatchPortfolioV2Request is the JSON body for
// PATCH /api/v2/portfolios/{portfolioCode}.
//
// Deliberately excludes fund association, code, and any status/lifecycle
// field — this endpoint updates only already-established descriptive
// metadata. Lifecycle transitions (status changes) are out of scope; use the
// dedicated approval/onboarding flows instead. Uses the same
// expected_version optimistic-concurrency convention as
// request.UpdatePortfolioRequest/DeletePortfolioRequest.
type PatchPortfolioV2Request struct {
	ExpectedVersion int        `json:"expected_version" validate:"required,min=1"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	StrategyCode    *string    `json:"strategy_code"`
	StyleID         *uuid.UUID `json:"style_id"`
	ManagerUserID   *uuid.UUID `json:"manager_user_id"`
	Benchmark       *string    `json:"benchmark"`
	RiskProfile     *string    `json:"risk_profile"`
}
