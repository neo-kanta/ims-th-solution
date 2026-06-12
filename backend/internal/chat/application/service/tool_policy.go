package service

import "strings"

// RequiredPermission maps a tool to the IAM function permission the user must
// hold to invoke it. It mirrors the permission the underlying REST endpoint
// enforces, so the in-process gate and the downstream call agree. An empty
// string means no specific function permission is required at the gate (the
// downstream REST call still applies its own checks and data scoping).
func RequiredPermission(toolName string) string {
	switch toolName {
	case "list_funds", "get_fund_nav":
		return "INVESTMENT_FUND_VIEW"
	case "list_portfolios", "get_portfolio_holdings":
		return "INVESTMENT_PORTFOLIO_VIEW"
	case "get_portfolio_valuation",
		"calc_weighted_average_price",
		"calc_allocation",
		"calc_exposure_pct":
		// These read the authoritative valuation snapshot (and its holding
		// lines), which the IMS REST API gates behind INVESTMENT_VALUATION_VIEW.
		// The in-process gate mirrors that; the downstream REST call enforces it
		// again under the user's own token (defense-in-depth).
		return "INVESTMENT_VALUATION_VIEW"
	default:
		return ""
	}
}

// mutatingPatterns flags tools that could change state. Read-only tools never
// match. A mutating tool may execute only when CHAT_WRITE_ENABLED is set.
var mutatingPatterns = []string{
	"create", "update", "delete", "post", "submit",
	"approve", "reverse", "cancel", "transfer", "write", "set_",
}

// IsMutatingTool reports whether a tool name looks state-changing. This is a
// conservative, defense-in-depth classifier — the MCP server also only
// exposes read tools in this slice.
func IsMutatingTool(toolName string) bool {
	n := strings.ToLower(toolName)
	for _, p := range mutatingPatterns {
		if strings.Contains(n, p) {
			return true
		}
	}
	return false
}
