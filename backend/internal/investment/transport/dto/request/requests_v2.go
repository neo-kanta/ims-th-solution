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
