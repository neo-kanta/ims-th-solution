package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
			Currency        string  `json:"currency"`
			AUMToday        string  `json:"aum_today"`
			TodayPnL        string  `json:"today_pnl"`
			TodayPnLPercent *string `json:"today_pnl_percent"`
			DataAvailable   bool    `json:"data_available"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Data.DataAvailable)
	require.Equal(t, "company", body.Data.Scope)
	require.Equal(t, "THB", body.Data.Currency)
	require.Equal(t, "1234567", body.Data.AUMToday)
	require.Equal(t, "12345", body.Data.TodayPnL)
	require.NotNil(t, body.Data.TodayPnLPercent)
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

var _ contract.ValuationSummaryProvider = (*fakeHandlerValuationProvider)(nil)
