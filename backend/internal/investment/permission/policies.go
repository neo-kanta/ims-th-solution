// Package permission declares the permission codes owned by the investment
// module.
//
// These constants are the single source of truth for permission names. The
// cmd/seed binary upserts each definition into
// permissions_function_definitions via contract.PermissionCatalog;
// transport middleware uses them when gating routes.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "investment"

// Permission codes owned by the investment module.
//
// Naming convention: INVESTMENT_<RESOURCE>_<ACTION>. Codes are stable
// contracts — renaming one is a breaking change for DB seed data and any UI
// that references them.
const (
	CodeFundView         = "INVESTMENT_FUND_VIEW"
	CodeFundManage       = "INVESTMENT_FUND_MANAGE"
	CodePortfolioView    = "INVESTMENT_PORTFOLIO_VIEW"
	CodePortfolioManage  = "INVESTMENT_PORTFOLIO_MANAGE"
	CodeInstrumentView   = "INVESTMENT_INSTRUMENT_VIEW"
	CodeInstrumentManage = "INVESTMENT_INSTRUMENT_MANAGE"
	CodeReferenceView    = "INVESTMENT_REFERENCE_VIEW"
	CodeLedgerView       = "INVESTMENT_LEDGER_VIEW"
	CodeLedgerPost       = "INVESTMENT_LEDGER_POST"
	CodeLedgerForcePost  = "INVESTMENT_LEDGER_FORCE_POST"
	CodeLedgerReverse    = "INVESTMENT_LEDGER_REVERSE"
	CodeValuationView    = "INVESTMENT_VALUATION_VIEW"
	CodeValuationRun     = "INVESTMENT_VALUATION_RUN"
	CodePricePost        = "INVESTMENT_PRICE_POST"
	CodeFundAUMCompute   = "INVESTMENT_FUND_AUM_COMPUTE"
)

// All returns every permission code owned by this module.
// Used by the permissions module seeder and any caller that wants only codes.
func All() []string {
	defs := Provider{}.Permissions()
	codes := make([]string, 0, len(defs))
	for _, d := range defs {
		codes = append(codes, d.Code)
	}
	return codes
}

// Provider implements contract.PermissionCatalog for the investment module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of investment permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeFundView, Name: "Investment Fund View", Description: "Read fund master data and fund-level AUM history."},
		{Code: CodeFundManage, Name: "Investment Fund Manage", Description: "Create, update, and soft-delete fund master records."},
		{Code: CodePortfolioView, Name: "Investment Portfolio View", Description: "Read portfolio master data, positions, cash, and valuations."},
		{Code: CodePortfolioManage, Name: "Investment Portfolio Manage", Description: "Create, update, and soft-delete portfolio master records."},
		{Code: CodeInstrumentView, Name: "Investment Instrument View", Description: "Read instrument master data and provider mappings."},
		{Code: CodeInstrumentManage, Name: "Investment Instrument Manage", Description: "Create, update, and soft-delete instrument master records."},
		{Code: CodeReferenceView, Name: "Investment Reference View", Description: "Read taxonomy reference data (asset classes, sectors, regions, fund categories, styles)."},
		{Code: CodeLedgerView, Name: "Investment Ledger View", Description: "Read the immutable transaction ledger and reversals."},
		{Code: CodeLedgerPost, Name: "Investment Ledger Post", Description: "Post BUY / SELL / cash transactions when the workflow day is open."},
		{Code: CodeLedgerForcePost, Name: "Investment Ledger Force Post", Description: "Post against a locked workflow day; runbook §2 reversal authorisation."},
		{Code: CodeLedgerReverse, Name: "Investment Ledger Reverse", Description: "Post a REVERSAL transaction against an existing ledger row."},
		{Code: CodeValuationView, Name: "Investment Valuation View", Description: "Read valuation snapshots and holding-line breakdowns."},
		{Code: CodeValuationRun, Name: "Investment Valuation Run", Description: "Trigger the valuation runner manually."},
		{Code: CodePricePost, Name: "Investment Price Post", Description: "Manually post a price snapshot (operator authorised, runbook §1 fallback)."},
		{Code: CodeFundAUMCompute, Name: "Investment Fund AUM Compute", Description: "Aggregate per-portfolio AUM into a fund-level AUM snapshot for a business date."},
	}
}
