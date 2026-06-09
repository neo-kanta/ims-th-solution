// Command ims-mcp is a read-only Model Context Protocol server that exposes
// IMS investment data to the chat assistant.
//
// Design (confirmed with the user):
//   - It is a STDIO server, spawned as a child process by the chat module.
//   - Every tool is a thin, read-only wrapper over the EXISTING IMS REST API.
//     It calls those endpoints with the end user's own JWT (forwarded via MCP
//     _meta), so all permission and data-scoping middleware applies and the
//     numbers are byte-for-byte identical to the dashboard. The server never
//     queries the database directly and never reshapes a financial figure.
//   - Read-only by construction: only GET wrappers are registered here. There
//     is no path to a mutating endpoint, regardless of CHAT_WRITE_ENABLED.
//
// Configuration:
//   - IMS_API_BASE_URL (default http://localhost:8080/api/v1) — where to reach
//     the IMS REST API. Inside the backend container this is loopback.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	baseURL := os.Getenv("IMS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1"
	}

	deps := &toolDeps{ims: newIMSClient(baseURL)}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ims-mcp",
		Version: "0.1.0",
	}, nil)

	// Read-only tool catalog. Descriptions are written for the LLM: they make
	// clear the tool returns authoritative data and how to chain calls.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_funds",
		Description: "List the investment funds the current user can access. Returns fund ids and names. Use a fund id with get_fund_nav.",
	}, deps.listFunds)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_portfolios",
		Description: "List the portfolios the current user can access. Returns portfolio ids and names. Use a portfolio id with get_portfolio_holdings.",
	}, deps.listPortfolios)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_portfolio_holdings",
		Description: "Get the current holdings (positions) of one portfolio by its UUID. Returns authoritative quantities and valuations. All figures come directly from the IMS system of record — never estimate or compute them yourself.",
	}, deps.getPortfolioHoldings)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_fund_nav",
		Description: "Get the latest official Net Asset Value (NAV) for one fund by its UUID. Returns the authoritative NAV with its as-of date. Never estimate NAV; if this tool has no data, say so.",
	}, deps.getFundNAV)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_portfolio_valuation",
		Description: "Get the latest AUTHORITATIVE valuation snapshot for one portfolio by its UUID: total market value, cost basis, unrealised P&L, ROI, AUM and per-instrument holding lines. These figures are computed by the IMS valuation engine — report them verbatim; never recompute totals, P&L or ROI yourself.",
	}, deps.getPortfolioValuation)

	// Deterministic calculation tools. The model MUST use these instead of doing
	// arithmetic in prose. Each is decimal-safe and returns provenance.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_weighted_average_price",
		Description: "Compute the quantity-weighted average price of a portfolio's holdings from the authoritative valuation lines. Returns one value and its currency. Errors safely if holdings span multiple currencies or data is missing.",
	}, deps.calcWeightedAveragePrice)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_allocation",
		Description: "Compute portfolio allocation as each group's percentage of the authoritative total market value. Argument 'by' is 'instrument' (default) or 'currency'. Percentages use the IMS authoritative portfolio market value as the denominator.",
	}, deps.calcAllocation)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_exposure_pct",
		Description: "Compute one instrument's exposure as a percentage of the authoritative total portfolio market value, given portfolio_id and instrument_id.",
	}, deps.calcExposurePct)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_sum",
		Description: "Sum an array of decimal number strings. Always use this instead of manual LLM arithmetic.",
	}, deps.calcSum)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_pnl",
		Description: "Calculate absolute and percentage Profit and Loss (PnL) between a market value and a cost basis.",
	}, deps.calcPnl)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calc_percentage",
		Description: "Calculate the percentage of a value relative to a total denominator.",
	}, deps.calcPercentage)

	// Run over stdio until the parent closes the connection.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("ims-mcp server exited: %v", err)
	}
}
