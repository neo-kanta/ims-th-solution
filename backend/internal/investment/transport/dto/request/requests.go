// Package request holds inbound HTTP DTOs for the investment module.
//
// All money / quantity fields are typed as string and parsed into
// decimal.Decimal at the handler boundary. JSON numbers MUST NOT be used —
// they round-trip through float64 and lose precision.
package request

import "github.com/google/uuid"

// ─── Fund ─────────────────────────────────────────────────────────────────────

// CreateFundRequest is the JSON body for POST /investment/funds.
type CreateFundRequest struct {
	Code           string     `json:"code"            validate:"required,max=40"`
	Name           string     `json:"name"            validate:"required,max=255"`
	ShortName      string     `json:"short_name"      validate:"max=80"`
	FundCategoryID uuid.UUID  `json:"fund_category_id" validate:"required"`
	BaseCurrency   string     `json:"base_currency"   validate:"required,len=3"`
	InceptionDate  string     `json:"inception_date"  validate:"required"`
	ManagerUserID  *uuid.UUID `json:"manager_user_id"`
	Benchmark      string     `json:"benchmark"`
	RiskProfile    string     `json:"risk_profile"`
	HasUnits       bool       `json:"has_units"`
	ExternalPAMRef string     `json:"external_pam_ref"`
}

// UpdateFundRequest is the JSON body for PUT /investment/funds/{id}.
type UpdateFundRequest struct {
	ExpectedVersion int        `json:"expected_version" validate:"required,min=1"`
	Name            *string    `json:"name"`
	ShortName       *string    `json:"short_name"`
	FundCategoryID  *uuid.UUID `json:"fund_category_id"`
	ManagerUserID   *uuid.UUID `json:"manager_user_id"`
	Benchmark       *string    `json:"benchmark"`
	RiskProfile     *string    `json:"risk_profile"`
	Status          *string    `json:"status"`
	ExternalPAMRef  *string    `json:"external_pam_ref"`
}

// DeleteFundRequest is the JSON body for DELETE /investment/funds/{id}.
type DeleteFundRequest struct {
	ExpectedVersion int `json:"expected_version" validate:"required,min=1"`
}

// ─── Portfolio ────────────────────────────────────────────────────────────────

// CreatePortfolioRequest is the JSON body for POST /investment/portfolios.
type CreatePortfolioRequest struct {
	FundID            uuid.UUID  `json:"fund_id"            validate:"required"`
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

// UpdatePortfolioRequest is the JSON body for PUT /investment/portfolios/{id}.
type UpdatePortfolioRequest struct {
	ExpectedVersion int        `json:"expected_version" validate:"required,min=1"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	StrategyCode    *string    `json:"strategy_code"`
	StyleID         *uuid.UUID `json:"style_id"`
	ManagerUserID   *uuid.UUID `json:"manager_user_id"`
	Benchmark       *string    `json:"benchmark"`
	RiskProfile     *string    `json:"risk_profile"`
	Status          *string    `json:"status"`
}

// DeletePortfolioRequest is the JSON body for DELETE /investment/portfolios/{id}.
type DeletePortfolioRequest struct {
	ExpectedVersion int `json:"expected_version" validate:"required,min=1"`
}

// ─── Instrument ───────────────────────────────────────────────────────────────

// CreateInstrumentRequest is the JSON body for POST /investment/instruments.
type CreateInstrumentRequest struct {
	PrimaryTicker   string         `json:"primary_ticker"   validate:"required,max=40"`
	Name            string         `json:"name"             validate:"required,max=255"`
	AssetClassID    uuid.UUID      `json:"asset_class_id"   validate:"required"`
	AssetSubtypeID  uuid.UUID      `json:"asset_subtype_id" validate:"required"`
	Currency        string         `json:"currency"         validate:"required,len=3"`
	CountryID       uuid.UUID      `json:"country_id"       validate:"required"`
	RegionID        *uuid.UUID     `json:"region_id"`
	PrimaryExchange string         `json:"primary_exchange"`
	SectorID        *uuid.UUID     `json:"sector_id"`
	FundCategoryID  *uuid.UUID     `json:"fund_category_id"`
	LotSize         int            `json:"lot_size"`
	TickSize        string         `json:"tick_size"`
	Attributes      map[string]any `json:"attributes"`
}

// UpdateInstrumentRequest is the JSON body for PUT /investment/instruments/{id}.
type UpdateInstrumentRequest struct {
	Name            *string        `json:"name"`
	PrimaryExchange *string        `json:"primary_exchange"`
	SectorID        *uuid.UUID     `json:"sector_id"`
	FundCategoryID  *uuid.UUID     `json:"fund_category_id"`
	LotSize         *int           `json:"lot_size"`
	TickSize        *string        `json:"tick_size"`
	IsTradable      *bool          `json:"is_tradable"`
	Status          *string        `json:"status"`
	Attributes      map[string]any `json:"attributes"`
}

// ─── Ledger ───────────────────────────────────────────────────────────────────

// PostTransactionRequest is the JSON body for POST /investment/portfolios/{id}/transactions.
//
// All decimal amount fields are JSON strings to preserve precision.
type PostTransactionRequest struct {
	TransactionType   string     `json:"transaction_type" validate:"required"`
	Side              string     `json:"side"`
	InstrumentID      *uuid.UUID `json:"instrument_id"`
	Quantity          string     `json:"quantity"`
	Price             string     `json:"price"`
	Currency          string     `json:"currency"         validate:"required,len=3"`
	Fees              string     `json:"fees"`
	GrossAmount       string     `json:"gross_amount"`
	NetAmount         string     `json:"net_amount"`
	FxRateToBase      string     `json:"fx_rate_to_base"`
	BusinessDate      string     `json:"business_date"    validate:"required"`
	SettlementDate    string     `json:"settlement_date"`
	SourceDecisionID  *uuid.UUID `json:"source_decision_id"`
	SourceExecutionID *uuid.UUID `json:"source_execution_id"`
	ExternalRef       string     `json:"external_ref"`
	Reason            string     `json:"reason"`
	ForcePost         bool       `json:"force_post"`
}

// ReverseTransactionRequest is the JSON body for POST .../transactions/{txnId}/reverse.
type ReverseTransactionRequest struct {
	BusinessDate string `json:"business_date" validate:"required"`
	Reason       string `json:"reason"        validate:"required"`
	ForcePost    bool   `json:"force_post"`
}

// ─── Pricing & Valuation ──────────────────────────────────────────────────────

// PostPriceSnapshotRequest is the JSON body for POST /investment/instruments/{id}/prices.
type PostPriceSnapshotRequest struct {
	BusinessDate string `json:"business_date" validate:"required"`
	Price        string `json:"price"         validate:"required"`
	Currency     string `json:"currency"      validate:"required,len=3"`
	PriceSource  string `json:"price_source"  validate:"required"`
	ProviderRef  string `json:"provider_ref"`
	IsStale      bool   `json:"is_stale"`
	StaleReason  string `json:"stale_reason"`
}

// RunValuationRequest is the JSON body for POST /investment/portfolios/{id}/valuations/run.
type RunValuationRequest struct {
	BusinessDate string            `json:"business_date" validate:"required"`
	FxRates      map[string]string `json:"fx_rates"`
	TotalUnits   string            `json:"total_units"`
}

// ComputeFundAUMRequest is the JSON body for
// POST /investment/funds/{id}/aum/compute. Aggregates all portfolio AUMs
// under the fund into a single fund-scoped snapshot for the business date.
type ComputeFundAUMRequest struct {
	BusinessDate string `json:"business_date" validate:"required"`
}

// ─── Research Reports ─────────────────────────────────────────────────────────

// CreateResearchReportRequest is the JSON body for POST /investment/research-reports.
type CreateResearchReportRequest struct {
	ReportNo             string     `json:"report_no"            validate:"required,max=60"`
	ReportDate           string     `json:"report_date"          validate:"required"`
	EffectiveDate        string     `json:"effective_date"`
	OwnerUserID          uuid.UUID  `json:"owner_user_id"        validate:"required"`
	AuthorUserID         uuid.UUID  `json:"author_user_id"       validate:"required"`
	ApplicableContractID *uuid.UUID `json:"applicable_contract_id"`

	InstrumentType string `json:"instrument_type"`
	InstrumentCode string `json:"instrument_code" validate:"required,max=40"`
	InstrumentName string `json:"instrument_name"`
	Market         string `json:"market"`
	Currency       string `json:"currency"`

	Recommendation string `json:"recommendation" validate:"required,oneof=BUY SELL HOLD"`
	ReportTitle    string `json:"report_title"`

	CompanyOverview    string `json:"company_overview"`
	CompanyOutlook     string `json:"company_outlook"`
	ESGComment         string `json:"esg_comment"`
	FinancialStatus    string `json:"financial_status"`
	InvestmentAnalysis string `json:"investment_analysis" validate:"required,min=25"`
}

// UpdateResearchReportRequest is the JSON body for PUT /investment/research-reports/{id}.
// Only fields set on the wire are applied (pointer-based partial update).
type UpdateResearchReportRequest struct {
	ReportDate           *string    `json:"report_date"`
	EffectiveDate        *string    `json:"effective_date"`
	OwnerUserID          *uuid.UUID `json:"owner_user_id"`
	AuthorUserID         *uuid.UUID `json:"author_user_id"`
	ApplicableContractID *uuid.UUID `json:"applicable_contract_id"`

	InstrumentType *string `json:"instrument_type"`
	InstrumentCode *string `json:"instrument_code"`
	InstrumentName *string `json:"instrument_name"`
	Market         *string `json:"market"`
	Currency       *string `json:"currency"`

	Recommendation *string `json:"recommendation"`
	ReportTitle    *string `json:"report_title"`

	CompanyOverview    *string `json:"company_overview"`
	CompanyOutlook     *string `json:"company_outlook"`
	ESGComment         *string `json:"esg_comment"`
	FinancialStatus    *string `json:"financial_status"`
	InvestmentAnalysis *string `json:"investment_analysis"`

	RejectionReason    *string `json:"rejection_reason"`
	PostSubmissionNote *string `json:"post_submission_note"`

	ReportStatus *string `json:"report_status"`
}
