package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// ─── GET /integration/dashboard/valuation-summary ──────────────────────────

func TestGetValuationSummary_RequiresAuthentication(t *testing.T) {
	t.Parallel()
	h := NewDashboardHandler(nil, nil, query.NewGetValuationSummaryHandler(&fakeHandlerIAMPort{hasPerm: true}, &fakeHandlerValuationProvider{}))

	req := httptest.NewRequest(http.MethodGet, "/integration/dashboard/valuation-summary?scope=company", nil)
	rec := httptest.NewRecorder()

	h.GetValuationSummary(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetValuationSummary_InvalidScopeReturns400(t *testing.T) {
	t.Parallel()
	h := NewDashboardHandler(nil, nil, query.NewGetValuationSummaryHandler(&fakeHandlerIAMPort{hasPerm: true}, &fakeHandlerValuationProvider{}))

	req := withHandlerUserClaims(httptest.NewRequest(http.MethodGet, "/integration/dashboard/valuation-summary?scope=bogus", nil))
	rec := httptest.NewRecorder()

	h.GetValuationSummary(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetValuationSummary_SuccessEnvelope(t *testing.T) {
	t.Parallel()
	pct := decimal.RequireFromString("1.01")
	prov := &fakeHandlerValuationProvider{result: &contract.ValuationSummaryResult{
		DataAvailable:   true,
		Status:          contract.ValuationSummaryStatusAvailable,
		Currency:        "THB",
		AUM:             decimal.NewFromInt(1234567),
		TodayPnL:        decimal.NewFromInt(12345),
		TodayPnLPercent: &pct,
	}}
	h := NewDashboardHandler(nil, nil, query.NewGetValuationSummaryHandler(&fakeHandlerIAMPort{hasPerm: true, contracts: []string{"*"}}, prov))

	req := withHandlerUserClaims(httptest.NewRequest(http.MethodGet, "/integration/dashboard/valuation-summary?scope=company", nil))
	rec := httptest.NewRecorder()

	h.GetValuationSummary(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data struct {
			Scope           string  `json:"scope"`
			Status          string  `json:"status"`
			Currency        string  `json:"currency"`
			AUMToday        string  `json:"aum_today"`
			TodayPnL        string  `json:"today_pnl"`
			TodayPnLPercent *string `json:"today_pnl_percent"`
			DataAvailable   bool    `json:"data_available"`
			Coverage        struct {
				TotalFundCount int        `json:"total_fund_count"`
				Exclusions     []struct{} `json:"exclusions"`
			} `json:"coverage"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Data.DataAvailable)
	require.Equal(t, "company", body.Data.Scope)
	require.Equal(t, "AVAILABLE", body.Data.Status)
	require.Equal(t, "THB", body.Data.Currency)
	require.Equal(t, "1234567", body.Data.AUMToday)
	require.Equal(t, "12345", body.Data.TodayPnL)
	require.NotNil(t, body.Data.TodayPnLPercent)
	require.Empty(t, body.Data.Coverage.Exclusions)
}

func TestGetValuationSummary_IncompleteEnvelopeOmitsNumericTotalsAndIncludesCoverage(t *testing.T) {
	t.Parallel()
	bd := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	prov := &fakeHandlerValuationProvider{result: &contract.ValuationSummaryResult{
		Status:        contract.ValuationSummaryStatusIncomplete,
		DataAvailable: false,
		Currency:      "THB",
		BusinessDate:  bd,
		Coverage: contract.ValuationSummaryCoverage{
			TotalFundCount:         2,
			IncludedFundCount:      1,
			ExcludedFundCount:      1,
			TotalPortfolioCount:    2,
			IncludedPortfolioCount: 1,
			ExcludedPortfolioCount: 1,
			ExcludedCurrencies:     []string{"USD"},
			ExcludedBusinessDates:  []time.Time{bd},
			ExclusionReasons:       []contract.ValuationSummaryExclusionReason{contract.ValuationSummaryExclusionMissingFX},
			Exclusions: []contract.ValuationSummaryExclusion{{
				FundCode: "F-USD", PortfolioCode: "P-USD", Currency: "USD", BusinessDate: bd,
				RequiredBusinessDate: bd, Reason: contract.ValuationSummaryExclusionMissingFX,
			}},
		},
	}}
	h := NewDashboardHandler(nil, nil, query.NewGetValuationSummaryHandler(&fakeHandlerIAMPort{hasPerm: true, contracts: []string{"*"}}, prov))
	req := withHandlerUserClaims(httptest.NewRequest(http.MethodGet, "/integration/dashboard/valuation-summary?scope=company", nil))
	rec := httptest.NewRecorder()

	h.GetValuationSummary(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "INCOMPLETE", body.Data["status"])
	require.Equal(t, false, body.Data["data_available"])
	require.Equal(t, "THB", body.Data["currency"])
	require.Equal(t, "2026-07-14", body.Data["business_date"])
	require.NotContains(t, body.Data, "aum_today")
	require.NotContains(t, body.Data, "today_pnl")
	require.NotContains(t, body.Data, "today_pnl_percent")
	coverage := body.Data["coverage"].(map[string]any)
	require.Equal(t, float64(2), coverage["total_fund_count"])
	require.Equal(t, []any{"USD"}, coverage["excluded_currencies"])
	require.Equal(t, []any{"MISSING_FX_RATE"}, coverage["exclusion_reasons"])
	exclusions := coverage["exclusions"].([]any)
	require.Len(t, exclusions, 1)
	require.NotContains(t, exclusions[0].(map[string]any), "fund_id")
	require.Equal(t, "F-USD", exclusions[0].(map[string]any)["fund_code"])
}

func TestGetValuationSummary_ProviderFailureReturns500(t *testing.T) {
	t.Parallel()
	prov := &fakeHandlerValuationProvider{err: errors.New("db down")}
	h := NewDashboardHandler(nil, nil, query.NewGetValuationSummaryHandler(&fakeHandlerIAMPort{hasPerm: true, contracts: []string{"*"}}, prov))

	req := withHandlerUserClaims(httptest.NewRequest(http.MethodGet, "/integration/dashboard/valuation-summary?scope=company", nil))
	rec := httptest.NewRecorder()

	h.GetValuationSummary(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─── test helpers ────────────────────────────────────────────────────────────

func withHandlerUserClaims(req *http.Request) *http.Request {
	claims := &middleware.UserClaims{
		Username:         "alice",
		RegisteredClaims: jwt.RegisteredClaims{Subject: uuid.NewString()},
	}
	return req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, claims))
}

type fakeHandlerIAMPort struct {
	hasPerm   bool
	contracts []string
}

func (p *fakeHandlerIAMPort) HasFunctionPermission(context.Context, string, string) (bool, error) {
	return p.hasPerm, nil
}
func (p *fakeHandlerIAMPort) GetAccessibleContracts(context.Context, string) ([]string, error) {
	return p.contracts, nil
}

var _ query.IAMPort = (*fakeHandlerIAMPort)(nil)

type fakeHandlerValuationProvider struct {
	result *contract.ValuationSummaryResult
	err    error
}

func (p *fakeHandlerValuationProvider) GetValuationSummary(context.Context, contract.ValuationSummaryRequest) (*contract.ValuationSummaryResult, error) {
	return p.result, p.err
}

func (p *fakeHandlerValuationProvider) ReportingCurrency() string { return "THB" }

var _ contract.ValuationSummaryProvider = (*fakeHandlerValuationProvider)(nil)
