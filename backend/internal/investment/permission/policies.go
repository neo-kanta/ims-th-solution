// Package permission declares the permission codes owned by the investment
// module.
//
// These constants are the single source of truth for permission names. The
// permissions module seeds these codes into permissions_function_rights;
// transport middleware uses them when gating routes.
package permission

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
)

// All returns every permission code owned by this module.
// Used by the permissions module seeder.
func All() []string {
	return []string{
		CodeFundView,
		CodeFundManage,
		CodePortfolioView,
		CodePortfolioManage,
		CodeInstrumentView,
		CodeInstrumentManage,
		CodeReferenceView,
		CodeLedgerView,
		CodeLedgerPost,
		CodeLedgerForcePost,
		CodeLedgerReverse,
		CodeValuationView,
		CodeValuationRun,
		CodePricePost,
	}
}
