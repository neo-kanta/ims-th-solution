package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// New Portfolio V2 endpoints: GET/POST /portfolios, PATCH
// /portfolios/{portfolioCode}, GET .../executions[/{executionId}],
// GET .../confirmations[/{confirmationId}]. Every endpoint must be scoped
// from creation — these tests prove the fund-scope gate matches the
// existing resolvePortfolioByCode/hasFundAccess/accessibleFundIDs
// conventions, not a new authorization mechanism.

// ─── stubFundRepo ───────────────────────────────────────────────────────────

type stubFundRepo struct {
	byCode map[string]*entity.Fund
}

func (r *stubFundRepo) Create(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r *stubFundRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Fund, error) {
	for _, f := range r.byCode {
		if f != nil && f.ID == id {
			return f, nil
		}
	}
	return nil, nil
}
func (r *stubFundRepo) GetByCode(_ context.Context, code string) (*entity.Fund, error) {
	return r.byCode[code], nil
}
func (r *stubFundRepo) GetByContractCode(context.Context, string) (*entity.Fund, error) {
	return nil, nil
}
func (r *stubFundRepo) List(context.Context, domain.FundListFilter) ([]*entity.Fund, int, error) {
	return nil, 0, nil
}
func (r *stubFundRepo) Update(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r *stubFundRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *stubFundRepo) CountActivePortfolios(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func activeFund(code string, id uuid.UUID) *entity.Fund {
	return &entity.Fund{
		ID:             id,
		Code:           code,
		Name:           "Fund " + code,
		FundCategoryID: uuid.New(),
		BaseCurrency:   "THB",
		InceptionDate:  time.Now(),
		Status:         vo.FundStatusActive,
		Version:        1,
	}
}

// ─── ListPortfoliosV2 ───────────────────────────────────────────────────────

func TestListPortfoliosV2_GlobalScopeComputesUnrestrictedFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingPortfolioRepo{}
	h := &InvestmentHandler{portfolios: repo, pc: &fakeV2PermissionChecker{Contracts: []string{"*"}}}
	r := chi.NewRouter()
	r.Get("/portfolios", h.ListPortfoliosV2)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs != nil {
		t.Fatalf("AccessibleFundIDs = %v, want nil (unrestricted)", repo.lastFilter.AccessibleFundIDs)
	}
}

func TestListPortfoliosV2_DeniedScopeComputesEmptyFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingPortfolioRepo{}
	h := &InvestmentHandler{portfolios: repo, pc: denyAll()}
	r := chi.NewRouter()
	r.Get("/portfolios", h.ListPortfoliosV2)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs == nil || len(repo.lastFilter.AccessibleFundIDs) != 0 {
		t.Fatalf("AccessibleFundIDs = %v, want non-nil empty slice", repo.lastFilter.AccessibleFundIDs)
	}
}

func TestListPortfoliosV2_UnknownFundCodeReturns404(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingPortfolioRepo{}
	h := &InvestmentHandler{
		portfolios: repo,
		funds:      &stubFundRepo{byCode: map[string]*entity.Fund{}},
		pc:         allowFund(uuid.New()),
	}
	r := chi.NewRouter()
	r.Get("/portfolios", h.ListPortfoliosV2)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios?fund_code=NOPE", nil)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

// filterCapturingPortfolioRepo records the last filter passed to List.
type filterCapturingPortfolioRepo struct {
	lastFilter domain.PortfolioListFilter
}

func (r *filterCapturingPortfolioRepo) Create(context.Context, pgx.Tx, *entity.Portfolio) error {
	return nil
}
func (r *filterCapturingPortfolioRepo) GetByID(context.Context, uuid.UUID) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *filterCapturingPortfolioRepo) GetByFundCode(context.Context, uuid.UUID, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *filterCapturingPortfolioRepo) GetByCode(context.Context, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *filterCapturingPortfolioRepo) List(_ context.Context, f domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	r.lastFilter = f
	return nil, 0, nil
}
func (r *filterCapturingPortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error {
	return nil
}
func (r *filterCapturingPortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *filterCapturingPortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

// ─── CreatePortfolioV2 ──────────────────────────────────────────────────────

func validCreatePortfolioV2Body(fundCode string) string {
	return `{
		"fund_code":"` + fundCode + `",
		"portfolio_type":"LIVE",
		"code":"NEW-PORT",
		"name":"New Portfolio",
		"base_currency":"THB",
		"valuation_currency":"THB",
		"inception_date":"2026-01-01"
	}`
}

func TestCreatePortfolioV2_UnknownFundCodeReturns404(t *testing.T) {
	t.Parallel()
	h := &InvestmentHandler{
		funds: &stubFundRepo{byCode: map[string]*entity.Fund{}},
		pc:    allowFund(uuid.New()),
	}
	r := chi.NewRouter()
	r.Post("/portfolios", h.CreatePortfolioV2)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios", strings.NewReader(validCreatePortfolioV2Body("NOPE"))))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestCreatePortfolioV2_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &InvestmentHandler{
		funds: &stubFundRepo{byCode: map[string]*entity.Fund{"FUND-A": activeFund("FUND-A", fundID)}},
		pc:    denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios", h.CreatePortfolioV2)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios", strings.NewReader(validCreatePortfolioV2Body("FUND-A"))))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestCreatePortfolioV2_InvalidPortfolioTypeReturns400(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &InvestmentHandler{
		funds: &stubFundRepo{byCode: map[string]*entity.Fund{"FUND-A": activeFund("FUND-A", fundID)}},
		pc:    allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios", h.CreatePortfolioV2)
	body := `{"fund_code":"FUND-A","portfolio_type":"BOGUS","code":"X","name":"X","base_currency":"THB","valuation_currency":"THB","inception_date":"2026-01-01"}`
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestCreatePortfolioV2_InvalidCurrencyReturns400(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &InvestmentHandler{
		funds: &stubFundRepo{byCode: map[string]*entity.Fund{"FUND-A": activeFund("FUND-A", fundID)}},
		pc:    allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios", h.CreatePortfolioV2)
	body := `{"fund_code":"FUND-A","portfolio_type":"LIVE","code":"X","name":"X","base_currency":"12B","valuation_currency":"THB","inception_date":"2026-01-01"}`
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// "12B" is not 3 letters, so it fails ^[A-Z]{3}$ even after
	// case-normalization — expect a clean 400 from the handler's own
	// ISO-4217-shape check, not a raw DB constraint 500.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

// ─── PatchPortfolioByCode ───────────────────────────────────────────────────

func TestPatchPortfolioByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &InvestmentHandler{
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Patch("/portfolios/{portfolioCode}", h.PatchPortfolioByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPatch, "/portfolios/PORT-A", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestPatchPortfolioByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &InvestmentHandler{
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Patch("/portfolios/{portfolioCode}", h.PatchPortfolioByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPatch, "/portfolios/PORT-A", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// json.Decode fails strictly after resolvePortfolioCode, so 400 here
	// (not 403) proves the fund-scope check passed.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (permission gate must pass before JSON decode); body=%s", w.Code, w.Body.String())
	}
}

// Note on lifecycle-mutation safety: request.PatchPortfolioV2Request
// (transport/dto/request/requests_v2.go) has no `status`, `fund_id`, or
// `code` field by construction — encoding/json silently drops unknown JSON
// keys on decode, so a caller supplying status/fund_id in the PATCH body has
// no way to reach PortfolioCommandHandler.Update's Status/fund-mutation
// path (which cmdReq never populates in PatchPortfolioByCode). This is a
// structural property of the DTO, verified by code review rather than a
// runtime test (exercising it end-to-end would require a live pgx
// transaction, out of scope for this package's unit tests per the existing
// convention — see portfolio_v2_decision_handler_test.go).

// ─── ListExecutionsByCode / GetExecutionByCode ──────────────────────────────

func TestListExecutionsByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/executions", h.ListExecutionsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/executions", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestListExecutionsByCode_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/executions", h.ListExecutionsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/executions", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecutionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA, portfolioB, fundID, executionID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioB},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioA, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/executions/{executionId}", h.GetExecutionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (execution belongs to a different portfolio); body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecutionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, executionID := uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/executions/{executionId}", h.GetExecutionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

// ─── ListConfirmationsByCode / GetConfirmationByCode ────────────────────────

func TestListConfirmationsByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		portfolios:    scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:            denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/confirmations", h.ListConfirmationsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/confirmations", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestListConfirmationsByCode_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		portfolios:    scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:            allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/confirmations", h.ListConfirmationsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/confirmations", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmationByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA, portfolioB, fundID, confirmationID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, PortfolioID: portfolioB},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioA, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/confirmations/{confirmationId}", h.GetConfirmationByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (confirmation belongs to a different portfolio); body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmationByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, confirmationID := uuid.New(), uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/confirmations/{confirmationId}", h.GetConfirmationByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

// ─── RunValuationByCode: MODEL portfolio rejection ─────────────────────────

func TestRunValuationByCode_ModelPortfolioReturns422(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	portfolios := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		"MODEL-A": {ID: portfolioID, Code: "MODEL-A", FundID: &fundID, PortfolioType: vo.PortfolioTypeModel},
	}}
	h := &InvestmentHandler{
		portfolios: portfolios,
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/valuations/run", h.RunValuationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/MODEL-A/valuations/run", strings.NewReader(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 (MODEL portfolios reject official valuation runs); body=%s", w.Code, w.Body.String())
	}
}

func TestRunValuationByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	portfolios := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		"PORT-A": {ID: portfolioID, Code: "PORT-A", FundID: &fundID, PortfolioType: vo.PortfolioTypeLive},
	}}
	h := &InvestmentHandler{
		portfolios: portfolios,
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/valuations/run", h.RunValuationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/valuations/run", strings.NewReader(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}
