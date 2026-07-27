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
	ID                     uuid.UUID  `json:"id"`
	Code                   string     `json:"code"`
	Name                   string     `json:"name"`
	ShortName              string     `json:"short_name,omitempty"`
	FundCategoryID         uuid.UUID  `json:"fund_category_id"`
	BaseCurrency           string     `json:"base_currency"`
	InceptionDate          string     `json:"inception_date"`
	ManagerUserID          *uuid.UUID `json:"manager_user_id,omitempty"`
	Benchmark              string     `json:"benchmark,omitempty"`
	RiskProfile            string     `json:"risk_profile,omitempty"`
	HasUnits               bool       `json:"has_units"`
	RequirePretradePreview bool       `json:"require_pretrade_preview"`
	ExternalPAMRef         string     `json:"external_pam_ref,omitempty"`
	Status                 string     `json:"status"`
	Version                int        `json:"version"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// PortfolioResponse mirrors entity.Portfolio for HTTP transport.
type PortfolioResponse struct {
	ID uuid.UUID `json:"id"`
	// FundID is omitted for a fund-less portfolio ("Bind with Fund: N").
	FundID            *uuid.UUID `json:"fund_id,omitempty"`
	PortfolioType     string     `json:"portfolio_type"`
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
	FundID                *uuid.UUID `json:"fund_id,omitempty"`
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

// TransactionSimulationResponse previews a post without mutating investment
// ledger, position, or cash tables.
type TransactionSimulationResponse struct {
	PortfolioID     uuid.UUID                   `json:"portfolio_id"`
	TransactionType string                      `json:"transaction_type"`
	InstrumentID    *uuid.UUID                  `json:"instrument_id,omitempty"`
	GrossAmount     string                      `json:"gross_amount"`
	NetAmount       string                      `json:"net_amount"`
	Cash            CashProjectionResponse      `json:"cash"`
	Position        *PositionProjectionResponse `json:"position,omitempty"`
	Compliance      *CompliancePreviewResponse  `json:"compliance,omitempty"`
}

type CashProjectionResponse struct {
	Currency         string `json:"currency"`
	CurrentBalance   string `json:"current_balance"`
	CashImpact       string `json:"cash_impact"`
	ProjectedBalance string `json:"projected_balance"`
}

type PositionProjectionResponse struct {
	InstrumentID         uuid.UUID `json:"instrument_id"`
	CurrentQuantity      string    `json:"current_quantity"`
	CurrentAverageCost   string    `json:"current_average_cost"`
	CurrentCostBasis     string    `json:"current_cost_basis"`
	ProjectedQuantity    string    `json:"projected_quantity"`
	ProjectedAverageCost string    `json:"projected_average_cost"`
	ProjectedCostBasis   string    `json:"projected_cost_basis"`
}

type CompliancePreviewResponse struct {
	CheckGroupID   uuid.UUID                         `json:"check_group_id"`
	Verdict        string                            `json:"verdict"`
	RulesEvaluated int                               `json:"rules_evaluated"`
	Breaches       []ComplianceBreachPreviewResponse `json:"breaches,omitempty"`
}

type ComplianceBreachPreviewResponse struct {
	BreachID    uuid.UUID `json:"breach_id"`
	RuleTypeID  string    `json:"rule_type_id"`
	Verdict     string    `json:"verdict"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Overridable bool      `json:"overridable"`
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

// CashRequestResponse mirrors entity.PortfolioCashRequest — the mutable staging
// row for a LIVE-portfolio cash movement awaiting approval. Returned by the
// pending-list and cancel endpoints, and by POST .../transactions (HTTP 202)
// when a LIVE cash movement is staged instead of posted.
type CashRequestResponse struct {
	ID                uuid.UUID  `json:"id"`
	PortfolioID       uuid.UUID  `json:"portfolio_id"`
	FundID            *uuid.UUID `json:"fund_id,omitempty"`
	TransactionType   string     `json:"transaction_type"`
	Amount            string     `json:"amount"`
	Currency          string     `json:"currency"`
	Fees              string     `json:"fees"`
	ValueDate         string     `json:"value_date"`
	Memo              string     `json:"memo,omitempty"`
	Status            string     `json:"status"`
	ApprovalRequestID *uuid.UUID `json:"approval_request_id,omitempty"`
	ResultingTxnID    *uuid.UUID `json:"resulting_txn_id,omitempty"`
	SubmittedBy       uuid.UUID  `json:"submitted_by"`
	SubmittedAt       time.Time  `json:"submitted_at"`
	DecidedBy         *uuid.UUID `json:"decided_by,omitempty"`
	DecidedAt         *time.Time `json:"decided_at,omitempty"`
	Version           int        `json:"version"`
}

type CashRequestListResponse struct {
	Items []CashRequestResponse `json:"items"`
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

// ExecutionListResponse documents the shape of
// PaginatedResponse[ExecutionResponse] for Swagger — swag cannot parse Go
// generic instantiations directly in @Success annotations, so Portfolio V2's
// GET /portfolios/{portfolioCode}/executions handler documents this
// concrete type instead (see PortfolioListResponse/TransactionListResponse
// above for the same pattern on other V2 paginated list endpoints).
type ExecutionListResponse struct {
	Items []ExecutionResponse `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// TradeConfirmationListResponse documents the shape of
// PaginatedResponse[TradeConfirmationResponse] for Swagger. See
// ExecutionListResponse's doc comment for why this concrete type exists
// alongside the generic runtime type.
type TradeConfirmationListResponse struct {
	Items []TradeConfirmationResponse `json:"items"`
	Total int                         `json:"total"`
	Page  int                         `json:"page"`
	Limit int                         `json:"limit"`
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

// FundNAVResponse is the aggregated fund-level latest valuation, served by
// GET /investment/funds/{id}/nav/latest. Values mirror entity.ValuationSnapshot
// for one or more portfolios under the same fund. Decimal fields are emitted
// as strings to preserve precision.
type FundNAVResponse struct {
	FundID         string `json:"fund_id"`
	BusinessDate   string `json:"business_date"`
	ValuationCcy   string `json:"valuation_ccy"`
	MarketValue    string `json:"market_value"`
	AUM            string `json:"aum"`
	CashBalance    string `json:"cash_balance"`
	UnrealisedPnL  string `json:"unrealised_pnl"`
	RealisedPnL    string `json:"realised_pnl"`
	ROI            string `json:"roi,omitempty"`
	TotalUnits     string `json:"total_units,omitempty"`
	NAVPerUnit     string `json:"nav_per_unit,omitempty"`
	HasStaleInputs bool   `json:"has_stale_inputs"`
	IsIndicative   bool   `json:"is_indicative"`
	PortfolioCount int    `json:"portfolio_count"`
}

// AllocationBucketResponse is one entry in a breakdown — e.g. an
// asset-class, sector, country, or currency contribution to the fund's NAV.
type AllocationBucketResponse struct {
	Key         string `json:"key"`          // stable machine key
	Label       string `json:"label"`        // human label
	MarketValue string `json:"market_value"` // decimal string in valuation_ccy
	PctOfNAV    string `json:"pct_of_nav"`   // 0..100, two-decimal precision
}

// FundAllocationResponse breaks a fund's market value across four orthogonal
// dimensions: asset class, sector, country, and currency. Served by
// GET /investment/funds/{id}/allocation.
type FundAllocationResponse struct {
	FundID         string                     `json:"fund_id"`
	AsOf           string                     `json:"as_of"`
	ValuationCcy   string                     `json:"valuation_ccy"`
	TotalNAV       string                     `json:"total_nav"`
	TotalCash      string                     `json:"total_cash"`
	PortfolioCount int                        `json:"portfolio_count"`
	ByAssetClass   []AllocationBucketResponse `json:"by_asset_class"`
	BySector       []AllocationBucketResponse `json:"by_sector"`
	ByCountry      []AllocationBucketResponse `json:"by_country"`
	ByCurrency     []AllocationBucketResponse `json:"by_currency"`
}

// NAVHistoryPointResponse is one daily observation in a fund's NAV/AUM trail.
// NAVPerUnit is empty for non-unitised funds; AUM is always populated.
type NAVHistoryPointResponse struct {
	BusinessDate string `json:"business_date"`
	NAVPerUnit   string `json:"nav_per_unit,omitempty"`
	AUM          string `json:"aum"`
}

// FundNAVHistoryResponse carries the time series plus convenience aggregates
// (high/low/latest/delta) so the UI doesn't redo arithmetic.
type FundNAVHistoryResponse struct {
	FundID   string                    `json:"fund_id"`
	Range    string                    `json:"range"`
	HasUnits bool                      `json:"has_units"`
	From     string                    `json:"from"`
	To       string                    `json:"to"`
	Series   []NAVHistoryPointResponse `json:"series"`
	High     string                    `json:"high"`
	Low      string                    `json:"low"`
	Latest   string                    `json:"latest"`
	DeltaPct string                    `json:"delta_pct"`
	IsEmpty  bool                      `json:"is_empty"`
}

// ResearchReportResponse mirrors entity.ResearchReport for HTTP transport.
type ResearchReportResponse struct {
	ID                   uuid.UUID  `json:"id"`
	ReportNo             string     `json:"report_no"`
	ReportDate           string     `json:"report_date"`
	EffectiveDate        *string    `json:"effective_date,omitempty"`
	OwnerUserID          uuid.UUID  `json:"owner_user_id"`
	AuthorUserID         uuid.UUID  `json:"author_user_id"`
	ApplicableContractID *uuid.UUID `json:"applicable_contract_id,omitempty"`

	InstrumentType string `json:"instrument_type,omitempty"`
	InstrumentCode string `json:"instrument_code"`
	InstrumentName string `json:"instrument_name,omitempty"`
	Market         string `json:"market,omitempty"`
	Currency       string `json:"currency,omitempty"`

	Recommendation string `json:"recommendation"`
	ReportTitle    string `json:"report_title,omitempty"`

	CompanyOverview    string `json:"company_overview,omitempty"`
	CompanyOutlook     string `json:"company_outlook,omitempty"`
	ESGComment         string `json:"esg_comment,omitempty"`
	FinancialStatus    string `json:"financial_status,omitempty"`
	InvestmentAnalysis string `json:"investment_analysis"`

	RejectionReason    string `json:"rejection_reason,omitempty"`
	PostSubmissionNote string `json:"post_submission_note,omitempty"`

	ReportStatus string `json:"report_status"`
	ReviewStatus string `json:"review_status"`

	// Multi-level review badge derived from the approval engine. Empty when
	// the report has not been submitted; "PENDING_LEVEL_<N>" while the
	// approval is in progress; "REVIEW_COMPLETED" / "REJECTED" / "INVALIDATED"
	// at terminal states. Backend remains the source of truth; UI uses this
	// field for display only.
	DerivedReviewStage string `json:"derived_review_stage,omitempty"`

	// Invalidation metadata. Populated together when the report is in
	// INVALIDATED state; omitted otherwise.
	InvalidatedAt      *time.Time `json:"invalidated_at,omitempty"`
	InvalidatedBy      *uuid.UUID `json:"invalidated_by,omitempty"`
	InvalidationReason string     `json:"invalidation_reason,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// ResearchReportListResponse is the wire shape for paginated research-report
// list calls. Mirrors the fund/portfolio list pattern.
type ResearchReportListResponse struct {
	Items []ResearchReportResponse `json:"items"`
	Total int                      `json:"total"`
	Page  int                      `json:"page"`
	Limit int                      `json:"limit"`
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

// DecisionResponse mirrors entity.Decision for HTTP transport.
type DecisionResponse struct {
	ID                                 uuid.UUID  `json:"id"`
	DecisionNumber                     string     `json:"decision_number"`
	FundID                             *uuid.UUID `json:"fund_id,omitempty"`
	PortfolioID                        uuid.UUID  `json:"portfolio_id"`
	InstrumentID                       *uuid.UUID `json:"instrument_id,omitempty"`
	InstrumentCode                     string     `json:"instrument_code,omitempty"`
	BusinessDate                       string     `json:"business_date"`
	Side                               string     `json:"side,omitempty"`
	Quantity                           string     `json:"quantity,omitempty"`
	Amount                             string     `json:"amount,omitempty"`
	LimitPrice                         string     `json:"limit_price,omitempty"`
	Currency                           string     `json:"currency"`
	Exchange                           string     `json:"exchange,omitempty"`
	DecisionType                       string     `json:"decision_type"`
	ProcessType                        string     `json:"process_type"`
	ProductType                        string     `json:"product_type"`
	StrategyCode                       string     `json:"strategy_code,omitempty"`
	AmendmentNo                        int        `json:"amendment_no"`
	ResearchReportID                   *uuid.UUID `json:"research_report_id,omitempty"`
	ResearchReportNo                   string     `json:"research_report_no,omitempty"`
	Rationale                          string     `json:"rationale,omitempty"`
	Status                             string     `json:"status"`
	ApprovalRequestID                  *uuid.UUID `json:"approval_request_id,omitempty"`
	ComplianceReleaseApprovalRequestID *uuid.UUID `json:"compliance_release_approval_request_id,omitempty"`
	ApprovalStatus                     string     `json:"approval_status,omitempty"`
	ComplianceCheckGroupID             *uuid.UUID `json:"compliance_check_group_id,omitempty"`
	SubmitterUserID                    uuid.UUID  `json:"submitter_user_id"`
	SubmittedAt                        *time.Time `json:"submitted_at,omitempty"`
	CancelledAt                        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason                 string     `json:"cancellation_reason,omitempty"`
	ReadyForExecutionAt                *time.Time `json:"ready_for_execution_at,omitempty"`
	CreatedAt                          time.Time  `json:"created_at"`
	UpdatedAt                          time.Time  `json:"updated_at"`

	// Lines are included when the decision has basket/rebalance/switch lines.
	Lines []DecisionLineResponse `json:"lines,omitempty"`

	// Approval enrichment — populated by the approval-items endpoint.
	ApprovalStage       int      `json:"approval_stage,omitempty"`
	ApprovalTotalStages int      `json:"approval_total_stages,omitempty"`
	CurrentApprovers    []string `json:"current_approvers,omitempty"`
	PreviousApprovers   []string `json:"previous_approvers,omitempty"`
}

// DecisionLineResponse mirrors entity.DecisionLine for HTTP transport.
type DecisionLineResponse struct {
	ID             uuid.UUID  `json:"id"`
	LineNumber     int        `json:"line_number"`
	InstrumentID   *uuid.UUID `json:"instrument_id,omitempty"`
	InstrumentCode string     `json:"instrument_code"`
	ProductType    string     `json:"product_type"`
	Side           string     `json:"side"`
	Quantity       string     `json:"quantity,omitempty"`
	Amount         string     `json:"amount,omitempty"`
	TargetWeight   string     `json:"target_weight,omitempty"`
	LimitPrice     string     `json:"limit_price,omitempty"`
	Currency       string     `json:"currency"`
	Notes          string     `json:"notes,omitempty"`
}

// ApproverInfo is a minimal approver identity for the batch-approval screen.
type ApproverInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

// BatchApprovalResultResponse reports the per-decision outcome of a batch action.
type BatchApprovalResultResponse struct {
	DecisionNo string `json:"decision_no"`
	OK         bool   `json:"ok"`
	Error      string `json:"error,omitempty"`
}

// BatchApprovalResponse is the envelope for batch approve/reject.
type BatchApprovalResponse struct {
	Results   []BatchApprovalResultResponse `json:"results"`
	Succeeded int                           `json:"succeeded"`
	Failed    int                           `json:"failed"`
}

type DecisionListResponse struct {
	Items []DecisionResponse `json:"items"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// ExecutionResponse mirrors entity.Execution.
type ExecutionResponse struct {
	ID                 uuid.UUID  `json:"id"`
	DecisionID         uuid.UUID  `json:"decision_id"`
	FundID             *uuid.UUID `json:"fund_id,omitempty"`
	PortfolioID        uuid.UUID  `json:"portfolio_id"`
	InstrumentID       *uuid.UUID `json:"instrument_id,omitempty"`
	InstrumentCode     string     `json:"instrument_code"`
	BusinessDate       string     `json:"business_date"`
	Side               string     `json:"side"`
	OrderedQuantity    string     `json:"ordered_quantity,omitempty"`
	OrderedAmount      string     `json:"ordered_amount,omitempty"`
	ExecutedQuantity   string     `json:"executed_quantity,omitempty"`
	ExecutedAmount     string     `json:"executed_amount,omitempty"`
	ExecutionPrice     string     `json:"execution_price,omitempty"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	TraderUserID       *uuid.UUID `json:"trader_user_id,omitempty"`
	BrokerReference    string     `json:"broker_reference,omitempty"`
	ExecutedAt         *time.Time `json:"executed_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ConfirmationBatchImportRowResponse summarises one row outcome of a batch
// import. Accepted rows expose the new confirmation ID; rejected rows expose
// the validation message so the caller does not need to query the batch
// items table to understand the result.
type ConfirmationBatchImportRowResponse struct {
	RowIndex       int        `json:"row_index"`
	Accepted       bool       `json:"accepted"`
	ConfirmationID *uuid.UUID `json:"confirmation_id,omitempty"`
	Error          string     `json:"error,omitempty"`
}

// ConfirmationBatchImportResponse is the wire shape returned by
// POST /investment/trade-confirmations/batch. Rejected rows are reported in
// the same payload so the importer never has to "hide" failures behind a
// second request — every row's fate is right there.
type ConfirmationBatchImportResponse struct {
	BatchID         uuid.UUID                            `json:"batch_id"`
	Status          string                               `json:"status"`
	TotalRecords    int                                  `json:"total_records"`
	AcceptedRecords int                                  `json:"accepted_records"`
	RejectedRecords int                                  `json:"rejected_records"`
	Rows            []ConfirmationBatchImportRowResponse `json:"rows"`
}

// TradeConfirmationResponse mirrors entity.TradeConfirmation.
type TradeConfirmationResponse struct {
	ID                uuid.UUID  `json:"id"`
	ExecutionID       uuid.UUID  `json:"execution_id"`
	DecisionID        uuid.UUID  `json:"decision_id"`
	FundID            *uuid.UUID `json:"fund_id,omitempty"`
	PortfolioID       uuid.UUID  `json:"portfolio_id"`
	BusinessDate      string     `json:"business_date"`
	ConfirmedQuantity string     `json:"confirmed_quantity,omitempty"`
	ConfirmedAmount   string     `json:"confirmed_amount,omitempty"`
	ConfirmedPrice    string     `json:"confirmed_price,omitempty"`
	Currency          string     `json:"currency"`
	BrokerReference   string     `json:"broker_reference,omitempty"`
	ImportBatchID     *uuid.UUID `json:"import_batch_id,omitempty"`
	Status            string     `json:"status"`
	DiscrepancyReason string     `json:"discrepancy_reason,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy        *uuid.UUID `json:"reviewed_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
