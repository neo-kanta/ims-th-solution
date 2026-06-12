package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// authMetaKey is the _meta key the chat backend uses to forward the caller's
// JWT. It is NOT part of any tool's input schema, so the model can neither
// see nor set it — identity is injected server-side by the chat agent loop.
const authMetaKey = "ims_auth_token"

// ── Tool input types ────────────────────────────────────────────────────────
// Field descriptions (jsonschema tags) become the schema the LLM sees.

type noArgs struct{}

type portfolioArgs struct {
	PortfolioID string `json:"portfolio_id" jsonschema:"UUID of the portfolio"`
}

type fundArgs struct {
	FundID string `json:"fund_id" jsonschema:"UUID of the fund"`
}

type allocationArgs struct {
	PortfolioID string `json:"portfolio_id" jsonschema:"UUID of the portfolio"`
	By          string `json:"by" jsonschema:"group allocation by 'instrument' (default) or 'currency'"`
}

type exposureArgs struct {
	PortfolioID  string `json:"portfolio_id" jsonschema:"UUID of the portfolio"`
	InstrumentID string `json:"instrument_id" jsonschema:"UUID of the instrument to measure exposure for"`
}

type calcSumArgs struct {
	Values []string `json:"values" jsonschema:"Array of decimal number strings to sum"`
}

type calcPnlArgs struct {
	MarketValue string `json:"market_value" jsonschema:"Total current market value string"`
	CostBasis   string `json:"cost_basis" jsonschema:"Total cost basis/investment string"`
}

type calcPercentageArgs struct {
	Value string `json:"value" jsonschema:"The numerator decimal number string"`
	Total string `json:"total" jsonschema:"The denominator decimal number string"`
}

// ── Result envelope ─────────────────────────────────────────────────────────

// sourceInfo is the provenance stamp attached to every tool result so both
// the model and the audit trail can trace a figure to its origin.
type sourceInfo struct {
	Tool       string `json:"tool"`
	Endpoint   string `json:"endpoint"`
	HTTPStatus int    `json:"http_status"`
	AsOf       string `json:"as_of"` // RFC3339 UTC, when the read was performed
}

// resultEnvelope is what the tool returns. `data` is the verbatim upstream
// JSON — financial numbers are never reshaped, recomputed, or rounded by the
// passthrough tools. For calc_* tools, `calculation` carries the deterministic,
// decimal-safe result plus its method, inputs and as-of stamp.
type resultEnvelope struct {
	Source      sourceInfo      `json:"source"`
	Data        json.RawMessage `json:"data,omitempty"`
	Calculation *calcInfo       `json:"calculation,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// calcInfo is the provenance for a deterministic calculation result. Result is
// the computed value as a string; inputs reference the authoritative source.
type calcInfo struct {
	Method   string      `json:"method"`
	Result   string      `json:"result"`
	Currency string      `json:"currency,omitempty"`
	Inputs   []calcInput `json:"inputs"`
	AsOf     string      `json:"as_of"`
}

type calcInput struct {
	Ref   string `json:"ref"`
	Value string `json:"value,omitempty"`
}

// toolDeps carries the IMS client into each handler.
type toolDeps struct {
	ims *imsClient
}

// callerToken extracts the forwarded JWT from request _meta.
func callerToken(req *mcp.CallToolRequest) string {
	if req == nil || req.Params == nil {
		return ""
	}
	meta := req.Params.GetMeta()
	if meta == nil {
		return ""
	}
	if v, ok := meta[authMetaKey].(string); ok {
		return v
	}
	return ""
}

func (d *toolDeps) resolveFundID(ctx context.Context, req *mcp.CallToolRequest, input string) (string, error) {
	if _, err := uuid.Parse(input); err == nil {
		return input, nil
	}
	token := callerToken(req)
	if token == "" {
		return "", fmt.Errorf("no caller identity was supplied to the tool")
	}
	resp, err := d.ims.get(ctx, token, "/investment/funds?limit=200")
	if err != nil {
		return "", fmt.Errorf("could not reach the IMS service: %w", err)
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return "", fmt.Errorf("failed to lookup fund: status %d", resp.Status)
	}
	var pageRes struct {
		Items []struct {
			ID   string `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &pageRes); err != nil {
		return "", fmt.Errorf("invalid response from IMS: %w", err)
	}
	
	// 1. Exact match on Code or Name (case-insensitive)
	for _, f := range pageRes.Items {
		if strings.EqualFold(f.Code, input) || strings.EqualFold(f.Name, input) {
			return f.ID, nil
		}
	}
	
	// 2. Substring matching (input contains code/name, or name contains input)
	cleanInput := strings.ToLower(strings.TrimSpace(input))
	for _, f := range pageRes.Items {
		c := strings.ToLower(f.Code)
		n := strings.ToLower(f.Name)
		if c != "" && strings.Contains(cleanInput, c) {
			return f.ID, nil
		}
		if n != "" && strings.Contains(cleanInput, n) {
			return f.ID, nil
		}
	}
	
	for _, f := range pageRes.Items {
		n := strings.ToLower(f.Name)
		if n != "" && strings.Contains(n, cleanInput) {
			return f.ID, nil
		}
	}
	
	return "", fmt.Errorf("fund identifier %q not found or not accessible", input)
}

func (d *toolDeps) resolvePortfolioID(ctx context.Context, req *mcp.CallToolRequest, input string) (string, error) {
	if _, err := uuid.Parse(input); err == nil {
		return input, nil
	}
	token := callerToken(req)
	if token == "" {
		return "", fmt.Errorf("no caller identity was supplied to the tool")
	}
	resp, err := d.ims.get(ctx, token, "/investment/portfolios?limit=200")
	if err != nil {
		return "", fmt.Errorf("could not reach the IMS service: %w", err)
	}
	if resp.Status < 200 || resp.Status >= 300 {
		return "", fmt.Errorf("failed to lookup portfolio: status %d", resp.Status)
	}
	var pageRes struct {
		Items []struct {
			ID   string `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(resp.Body, &pageRes); err != nil {
		return "", fmt.Errorf("invalid response from IMS: %w", err)
	}
	
	// 1. Exact match on Code or Name (case-insensitive)
	for _, p := range pageRes.Items {
		if strings.EqualFold(p.Code, input) || strings.EqualFold(p.Name, input) {
			return p.ID, nil
		}
	}
	
	// 2. Substring matching (input contains code/name, or name contains input)
	cleanInput := strings.ToLower(strings.TrimSpace(input))
	for _, p := range pageRes.Items {
		c := strings.ToLower(p.Code)
		n := strings.ToLower(p.Name)
		if c != "" && strings.Contains(cleanInput, c) {
			return p.ID, nil
		}
		if n != "" && strings.Contains(cleanInput, n) {
			return p.ID, nil
		}
	}
	
	for _, p := range pageRes.Items {
		n := strings.ToLower(p.Name)
		if n != "" && strings.Contains(n, cleanInput) {
			return p.ID, nil
		}
	}
	
	return "", fmt.Errorf("portfolio identifier %q not found or not accessible", input)
}

// fetch runs an authenticated GET and wraps the outcome into the standard
// envelope, covering the three adversarial cases by construction:
//   - upstream error (non-2xx)         → IsError result with a safe message
//   - empty / no-data (2xx empty body) → envelope with whatever upstream sent
//   - success                          → verbatim data + as-of provenance
func (d *toolDeps) fetch(ctx context.Context, req *mcp.CallToolRequest, toolName, endpoint string) (*mcp.CallToolResult, any, error) {
	body, src, errRes := d.authedGet(ctx, req, toolName, endpoint)
	if errRes != nil {
		return errRes, nil, nil
	}
	env := resultEnvelope{Source: src, Data: json.RawMessage(body)}
	return jsonResult(env, false), nil, nil
}

// authedGet performs the authenticated GET + status handling shared by the
// passthrough and calc tools. On any failure it returns a ready-to-send error
// result; on success it returns the verbatim body and the populated source
// stamp. Identity (JWT) is taken from request _meta — never from tool args.
func (d *toolDeps) authedGet(ctx context.Context, req *mcp.CallToolRequest, toolName, endpoint string) ([]byte, sourceInfo, *mcp.CallToolResult) {
	src := sourceInfo{Tool: toolName, Endpoint: endpoint, AsOf: nowRFC3339()}

	token := callerToken(req)
	if token == "" {
		return nil, src, errResult(src, "no caller identity was supplied to the tool")
	}
	resp, err := d.ims.get(ctx, token, endpoint)
	if err != nil {
		return nil, src, errResult(src, "could not reach the IMS service")
	}
	src.HTTPStatus = resp.Status

	switch {
	case resp.Status == 401 || resp.Status == 403:
		return nil, src, errResult(src, "you do not have permission to read this resource")
	case resp.Status == 404:
		return nil, src, errResult(src, "the requested resource was not found")
	case resp.Status < 200 || resp.Status >= 300:
		return nil, src, errResult(src, fmt.Sprintf("the IMS service returned status %d", resp.Status))
	}
	return resp.Body, src, nil
}

// ── Authoritative valuation decoding ────────────────────────────────────────

type valuationEnvelope struct {
	Data valuationDTO `json:"data"`
}

type valuationDTO struct {
	PortfolioID  string             `json:"portfolio_id"`
	ValuationCcy string             `json:"valuation_ccy"`
	MarketValue  string             `json:"market_value"`
	BusinessDate string             `json:"business_date"`
	HoldingLines []valuationLineDTO `json:"holding_lines"`
}

type valuationLineDTO struct {
	InstrumentID    string `json:"instrument_id"`
	Quantity        string `json:"quantity"`
	PriceInQuoteCcy string `json:"price_in_quote_ccy"`
	QuoteCurrency   string `json:"quote_currency"`
	MarketValue     string `json:"market_value"`
}

// decodeValuation parses the IMS REST valuation payload (wrapped as {"data":{…}}
// by httputil.OK, or bare) into the calc Valuation shape.
func decodeValuation(body []byte) (Valuation, error) {
	var env valuationEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Valuation{}, err
	}
	d := env.Data
	if d.MarketValue == "" && len(d.HoldingLines) == 0 {
		// Not enveloped — try decoding the body directly.
		var direct valuationDTO
		if err := json.Unmarshal(body, &direct); err == nil {
			d = direct
		}
	}
	v := Valuation{
		PortfolioID:  d.PortfolioID,
		ValuationCcy: d.ValuationCcy,
		MarketValue:  d.MarketValue,
		BusinessDate: d.BusinessDate,
	}
	for _, ln := range d.HoldingLines {
		v.Lines = append(v.Lines, ValuationLine{
			InstrumentID:    ln.InstrumentID,
			Quantity:        ln.Quantity,
			PriceInQuoteCcy: ln.PriceInQuoteCcy,
			QuoteCurrency:   ln.QuoteCurrency,
			MarketValue:     ln.MarketValue,
		})
	}
	return v, nil
}

func rawJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

// ── Tool handlers ───────────────────────────────────────────────────────────

func (d *toolDeps) listFunds(ctx context.Context, req *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, any, error) {
	return d.fetch(ctx, req, "list_funds", "/investment/funds")
}

func (d *toolDeps) listPortfolios(ctx context.Context, req *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, any, error) {
	return d.fetch(ctx, req, "list_portfolios", "/investment/portfolios")
}

func (d *toolDeps) getPortfolioHoldings(ctx context.Context, req *mcp.CallToolRequest, in portfolioArgs) (*mcp.CallToolResult, any, error) {
	if in.PortfolioID == "" {
		return errResult(sourceInfo{Tool: "get_portfolio_holdings", AsOf: nowRFC3339()}, "portfolio_id is required"), nil, nil
	}
	pid, err := d.resolvePortfolioID(ctx, req, in.PortfolioID)
	if err != nil {
		return errResult(sourceInfo{Tool: "get_portfolio_holdings", AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	endpoint := "/investment/portfolios/" + url.PathEscape(pid) + "/holdings"
	return d.fetch(ctx, req, "get_portfolio_holdings", endpoint)
}

func (d *toolDeps) getFundNAV(ctx context.Context, req *mcp.CallToolRequest, in fundArgs) (*mcp.CallToolResult, any, error) {
	if in.FundID == "" {
		return errResult(sourceInfo{Tool: "get_fund_nav", AsOf: nowRFC3339()}, "fund_id is required"), nil, nil
	}
	fid, err := d.resolveFundID(ctx, req, in.FundID)
	if err != nil {
		return errResult(sourceInfo{Tool: "get_fund_nav", AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	endpoint := "/investment/funds/" + url.PathEscape(fid) + "/nav/latest"
	return d.fetch(ctx, req, "get_fund_nav", endpoint)
}

// valuationEndpoint builds the authoritative latest-valuation endpoint for a
// portfolio.
func valuationEndpoint(portfolioID string) string {
	return "/investment/portfolios/" + url.PathEscape(portfolioID) + "/valuations/latest"
}

// getPortfolioValuation returns the AUTHORITATIVE latest valuation snapshot
// verbatim (market value, cost basis, unrealised P&L, ROI, AUM, holding lines).
// These figures are computed by the IMS valuation engine — never recomputed
// here — so the assistant can report totals/P&L/ROI without doing arithmetic.
func (d *toolDeps) getPortfolioValuation(ctx context.Context, req *mcp.CallToolRequest, in portfolioArgs) (*mcp.CallToolResult, any, error) {
	if in.PortfolioID == "" {
		return errResult(sourceInfo{Tool: "get_portfolio_valuation", AsOf: nowRFC3339()}, "portfolio_id is required"), nil, nil
	}
	pid, err := d.resolvePortfolioID(ctx, req, in.PortfolioID)
	if err != nil {
		return errResult(sourceInfo{Tool: "get_portfolio_valuation", AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	return d.fetch(ctx, req, "get_portfolio_valuation", valuationEndpoint(pid))
}

// calcWeightedAveragePrice computes the quantity-weighted average price from the
// authoritative valuation holding lines (explicit price_in_quote_ccy only).
func (d *toolDeps) calcWeightedAveragePrice(ctx context.Context, req *mcp.CallToolRequest, in portfolioArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_weighted_average_price"
	if in.PortfolioID == "" {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "portfolio_id is required"), nil, nil
	}
	pid, err := d.resolvePortfolioID(ctx, req, in.PortfolioID)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	endpoint := valuationEndpoint(pid)
	body, src, errRes := d.authedGet(ctx, req, tool, endpoint)
	if errRes != nil {
		return errRes, nil, nil
	}
	val, err := decodeValuation(body)
	if err != nil {
		return errResult(src, "the valuation data could not be read"), nil, nil
	}
	result, currency, cErr := WeightedAveragePrice(val)
	if cErr != nil {
		return errResult(src, cErr.Error()), nil, nil
	}
	env := resultEnvelope{
		Source: src,
		Calculation: &calcInfo{
			Method:   "weighted_average_price = sum(quantity * price_in_quote_ccy) / sum(quantity); single quote currency required",
			Result:   result,
			Currency: currency,
			AsOf:     src.AsOf,
			Inputs:   []calcInput{{Ref: endpoint + " -> holding_lines[].quantity, holding_lines[].price_in_quote_ccy"}},
		},
		Data: rawJSON(map[string]any{
			"portfolio_id":           val.PortfolioID,
			"weighted_average_price": result,
			"currency":               currency,
			"business_date":          val.BusinessDate,
		}),
	}
	return jsonResult(env, false), nil, nil
}

// calcAllocation computes each group's share of the AUTHORITATIVE portfolio
// market value (denominator never recomputed). by = instrument|currency.
func (d *toolDeps) calcAllocation(ctx context.Context, req *mcp.CallToolRequest, in allocationArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_allocation"
	if in.PortfolioID == "" {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "portfolio_id is required"), nil, nil
	}
	pid, err := d.resolvePortfolioID(ctx, req, in.PortfolioID)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	by := in.By
	if by == "" {
		by = "instrument"
	}
	if by != "instrument" && by != "currency" {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "by must be 'instrument' or 'currency'"), nil, nil
	}
	endpoint := valuationEndpoint(pid)
	body, src, errRes := d.authedGet(ctx, req, tool, endpoint)
	if errRes != nil {
		return errRes, nil, nil
	}
	val, err := decodeValuation(body)
	if err != nil {
		return errResult(src, "the valuation data could not be read"), nil, nil
	}
	entries, cErr := Allocation(val, by)
	if cErr != nil {
		return errResult(src, cErr.Error()), nil, nil
	}
	env := resultEnvelope{
		Source: src,
		Calculation: &calcInfo{
			Method: "allocation_pct = group_market_value / authoritative_portfolio_market_value * 100",
			Result: "see data.allocations",
			AsOf:   src.AsOf,
			Inputs: []calcInput{{Ref: endpoint + " -> market_value (denominator), holding_lines[].market_value (per group)", Value: val.MarketValue}},
		},
		Data: rawJSON(map[string]any{
			"portfolio_id":       val.PortfolioID,
			"by":                 by,
			"total_market_value": val.MarketValue,
			"valuation_ccy":      val.ValuationCcy,
			"allocations":        entries,
		}),
	}
	return jsonResult(env, false), nil, nil
}

// calcExposurePct computes one instrument's market value as a percentage of the
// AUTHORITATIVE portfolio market value.
func (d *toolDeps) calcExposurePct(ctx context.Context, req *mcp.CallToolRequest, in exposureArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_exposure_pct"
	if in.PortfolioID == "" || in.InstrumentID == "" {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "portfolio_id and instrument_id are required"), nil, nil
	}
	pid, err := d.resolvePortfolioID(ctx, req, in.PortfolioID)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, err.Error()), nil, nil
	}
	endpoint := valuationEndpoint(pid)
	body, src, errRes := d.authedGet(ctx, req, tool, endpoint)
	if errRes != nil {
		return errRes, nil, nil
	}
	val, err := decodeValuation(body)
	if err != nil {
		return errResult(src, "the valuation data could not be read"), nil, nil
	}
	pct, mv, cErr := ExposurePct(val, in.InstrumentID)
	if cErr != nil {
		return errResult(src, cErr.Error()), nil, nil
	}
	env := resultEnvelope{
		Source: src,
		Calculation: &calcInfo{
			Method:   "exposure_pct = instrument_market_value / authoritative_portfolio_market_value * 100",
			Result:   pct,
			Currency: val.ValuationCcy,
			AsOf:     src.AsOf,
			Inputs:   []calcInput{{Ref: endpoint + " -> market_value (denominator), holding_lines[].market_value", Value: val.MarketValue}},
		},
		Data: rawJSON(map[string]any{
			"portfolio_id":       val.PortfolioID,
			"instrument_id":      in.InstrumentID,
			"market_value":       mv,
			"total_market_value": val.MarketValue,
			"exposure_pct":       pct,
			"valuation_ccy":      val.ValuationCcy,
		}),
	}
	return jsonResult(env, false), nil, nil
}

// ── Result builders ─────────────────────────────────────────────────────────

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

func jsonResult(env resultEnvelope, isErr bool) *mcp.CallToolResult {
	b, err := json.Marshal(env)
	if err != nil {
		b = []byte(`{"error":"failed to encode tool result"}`)
		isErr = true
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
		IsError: isErr,
	}
}

func errResult(src sourceInfo, msg string) *mcp.CallToolResult {
	return jsonResult(resultEnvelope{Source: src, Error: msg}, true)
}

// calcSum sums a list of decimal strings deterministically.
func (d *toolDeps) calcSum(ctx context.Context, req *mcp.CallToolRequest, in calcSumArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_sum"
	if len(in.Values) == 0 {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "values array is empty"), nil, nil
	}
	sum := new(big.Rat)
	var inputs []calcInput
	for i, v := range in.Values {
		rat, err := ratFromString(v)
		if err != nil {
			return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, fmt.Sprintf("invalid value at index %d: %q", i, v)), nil, nil
		}
		sum.Add(sum, rat)
		inputs = append(inputs, calcInput{Ref: fmt.Sprintf("value[%d]", i), Value: v})
	}
	resStr := sum.FloatString(pricePrecision)
	env := resultEnvelope{
		Source: sourceInfo{Tool: tool, AsOf: nowRFC3339()},
		Calculation: &calcInfo{
			Method: "sum = value[0] + value[1] + ...",
			Result: resStr,
			AsOf:   nowRFC3339(),
			Inputs: inputs,
		},
		Data: rawJSON(map[string]any{
			"sum":   resStr,
			"count": len(in.Values),
		}),
	}
	return jsonResult(env, false), nil, nil
}

// calcPnl computes Profit & Loss.
func (d *toolDeps) calcPnl(ctx context.Context, req *mcp.CallToolRequest, in calcPnlArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_pnl"
	mv, err := ratFromString(in.MarketValue)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "invalid market_value"), nil, nil
	}
	cb, err := ratFromString(in.CostBasis)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "invalid cost_basis"), nil, nil
	}
	absPnl := new(big.Rat).Sub(mv, cb)
	pctPnl := new(big.Rat)
	if cb.Sign() != 0 {
		pctPnl.Quo(absPnl, cb).Mul(pctPnl, big.NewRat(100, 1))
	}
	absStr := absPnl.FloatString(pricePrecision)
	pctStr := pctPnl.FloatString(pctPrecision)
	env := resultEnvelope{
		Source: sourceInfo{Tool: tool, AsOf: nowRFC3339()},
		Calculation: &calcInfo{
			Method: "abs_pnl = market_value - cost_basis; pct_pnl = abs_pnl / cost_basis * 100",
			Result: absStr + " (" + pctStr + "%)",
			AsOf:   nowRFC3339(),
			Inputs: []calcInput{
				{Ref: "market_value", Value: in.MarketValue},
				{Ref: "cost_basis", Value: in.CostBasis},
			},
		},
		Data: rawJSON(map[string]any{
			"market_value": in.MarketValue,
			"cost_basis":   in.CostBasis,
			"abs_pnl":      absStr,
			"pct_pnl":      pctStr,
		}),
	}
	return jsonResult(env, false), nil, nil
}

// calcPercentage computes percentage.
func (d *toolDeps) calcPercentage(ctx context.Context, req *mcp.CallToolRequest, in calcPercentageArgs) (*mcp.CallToolResult, any, error) {
	const tool = "calc_percentage"
	val, err := ratFromString(in.Value)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "invalid value"), nil, nil
	}
	tot, err := ratFromString(in.Total)
	if err != nil {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "invalid total"), nil, nil
	}
	if tot.Sign() == 0 {
		return errResult(sourceInfo{Tool: tool, AsOf: nowRFC3339()}, "total is zero, percentage undefined"), nil, nil
	}
	pct := new(big.Rat).Quo(val, tot)
	pct.Mul(pct, big.NewRat(100, 1))
	pctStr := pct.FloatString(pctPrecision)
	env := resultEnvelope{
		Source: sourceInfo{Tool: tool, AsOf: nowRFC3339()},
		Calculation: &calcInfo{
			Method: "percentage = value / total * 100",
			Result: pctStr,
			AsOf:   nowRFC3339(),
			Inputs: []calcInput{
				{Ref: "value", Value: in.Value},
				{Ref: "total", Value: in.Total},
			},
		},
		Data: rawJSON(map[string]any{
			"value":      in.Value,
			"total":      in.Total,
			"percentage": pctStr,
		}),
	}
	return jsonResult(env, false), nil, nil
}
