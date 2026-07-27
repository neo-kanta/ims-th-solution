package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// P0-A: Portfolio V2 data-scope security fix.
//
// resolvePortfolioByCode (portfolio_v2_handler.go) is the shared
// {portfolioCode} -> portfolio resolver used by every Portfolio V2 handler.
// Before this fix, DecisionHandler, ExecutionHandler, and
// TradeConfirmationHandler passed a literal nil for the pc
// (contract.PermissionChecker) argument at all 10 of their portfolioCode
// call sites, which permanently disabled the `pc != nil && !hasFundAccess`
// check — any authenticated caller with the route-level function permission
// could read or mutate another fund's decisions/executions/confirmations by
// guessing or discovering its portfolio code, regardless of their fund data
// scope.
//
// These tests prove, for all 10 call sites, that:
//  1. A caller without data-permission on the resolved portfolio's fund is
//     rejected with 403 Forbidden (httputil.Forbidden), matching
//     resolvePortfolioByCode's existing enforced path — never a 404 (which
//     would leak whether the portfolio code exists) and never a silent
//     pass-through.
//  2. A caller WITH data-permission on the fund is not blocked by the
//     permission gate — request handling proceeds past
//     resolvePortfolioByCode into the handler's own logic.
//
// For write endpoints, reaching a genuine 200/201 requires a live pgx
// transaction (h.cmd's command handlers call pool.Begin under the hood) —
// out of scope for this package's unit tests per the existing convention in
// portfolio_v2_decision_handler_test.go ("success paths ... need a real pgx
// transaction and are covered by tests/e2e/portfolio_v2_decision_test.go").
// Instead, "the permission gate did not block this call" is proved by a
// deterministic downstream signal reached only after
// resolvePortfolioByCode + any ownership check succeed:
//   - For endpoints with a JSON request body, an intentionally malformed
//     body yields 400 (json decode happens strictly after the permission
//     and ownership checks in every one of these handlers) instead of 403.
//   - For SubmitDecisionByCode (the one write endpoint with no request
//     body), a real *command.DecisionCommandHandler is wired against a
//     decision stub whose status is not DRAFT, so DecisionCommandHandler.Submit
//     rejects with a 409 domain lifecycle error before ever touching a
//     pool/transaction (domain.Decision.CanSubmit() is checked before any
//     repository write) — proving the call reached genuine business logic,
//     not a permission short-circuit.

// ─── Test fixtures ──────────────────────────────────────────────────────────

// scopedPortfolio builds a single-entry stubPortfolioRepo whose portfolio has
// an explicit, caller-controlled FundID (unlike newOwnedPortfolioRepo, which
// always assigns a random FundID and so cannot be used to test fund-scope
// enforcement).
func scopedPortfolio(code string, portfolioID, fundID uuid.UUID) *stubPortfolioRepo {
	return &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		code: {ID: portfolioID, Code: code, FundID: &fundID},
	}}
}

// allowFund returns a fakeV2PermissionChecker granting data access only to
// fundID.
func allowFund(fundID uuid.UUID) *fakeV2PermissionChecker {
	return &fakeV2PermissionChecker{Allowed: map[string]bool{fundID.String(): true}}
}

// denyAll returns a fakeV2PermissionChecker granting no fund access at all —
// simulates a caller with the route's function permission but zero fund data
// scope, exactly the attacker profile in the original vulnerability.
func denyAll() *fakeV2PermissionChecker {
	return &fakeV2PermissionChecker{Allowed: map[string]bool{}}
}

const malformedBody = `{not-json`

// ─── Decisions: ListDecisionsByCode / GetDecisionByCode (no cmd; full 200 reachable) ──

func TestListDecisionsByCode_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions", h.ListDecisionsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestListDecisionsByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions", h.ListDecisionsByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionByCode_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
	// The 403 must come from the permission gate before ownership lookup —
	// confirm it is not a 404 (which would look identical to "unknown
	// decision" and 200/other otherwise).
	if w.Code == http.StatusNotFound {
		t.Fatalf("must not fall back to 404 when the fund access check fails")
	}
}

// ─── Decisions: CreateDecisionByCode / CancelDecisionByCode (JSON body precedes cmd) ──

func TestCreateDecisionByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions", h.CreateDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// json.Decode fails strictly after resolvePortfolioByCode, so 400 here
	// (not 403) proves the fund-scope check passed.
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (permission gate must pass before JSON decode); body=%s", w.Code, w.Body.String())
	}
}

func TestCreateDecisionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions", h.CreateDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestCancelDecisionByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/cancel", h.CancelDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/cancel", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (permission + ownership gates must pass before JSON decode); body=%s", w.Code, w.Body.String())
	}
}

func TestCancelDecisionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/cancel", h.CancelDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/cancel", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

// ─── Decisions: SubmitDecisionByCode (no request body — needs a real cmd) ──

func TestSubmitDecisionByCode_AuthorizedFundReachesBusinessLogic(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	decisions := &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
		// CANCELLED (not DRAFT) so DecisionCommandHandler.Submit's
		// !d.CanSubmit() guard rejects with 409 *before* touching h.pool —
		// this reaches real business logic without a live DB.
		decisionID: {ID: decisionID, PortfolioID: portfolioID, Status: vo.DecisionLifecycleCancelled},
	}}
	cmd := command.NewDecisionCommandHandler(nil, decisions, nil, nil, nil, nil)
	h := &DecisionHandler{
		decisions:  decisions,
		cmd:        cmd,
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/submit", h.SubmitDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/submit", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (real lifecycle rejection proves the permission gate passed); body=%s", w.Code, w.Body.String())
	}
}

func TestSubmitDecisionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID, Status: vo.DecisionLifecycleDraft},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/submit", h.SubmitDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/submit", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (h.cmd must never be reached for a denied fund); body=%s", w.Code, w.Body.String())
	}
}

// ─── Executions: CreateExecutionByCode / FillExecutionByCode / CancelExecutionByCode ──

func TestCreateExecutionByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/executions", h.CreateExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/executions", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestCreateExecutionByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, decisionID := uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioID},
		}},
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/executions", h.CreateExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/executions", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestFillExecutionByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, executionID := uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/fill", h.FillExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/fill", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestFillExecutionByCode_CrossFundReturns403(t *testing.T) {
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
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/fill", h.FillExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/fill", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestCancelExecutionByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, executionID := uuid.New(), uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/cancel", h.CancelExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/cancel", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestCancelExecutionByCode_CrossFundReturns403(t *testing.T) {
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
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/cancel", h.CancelExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/cancel", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

// ─── Trade confirmations: RecordConfirmationByCode / ResolveConfirmationByCode ──

func TestRecordConfirmationByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, executionID := uuid.New(), uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioID},
		}},
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		portfolios:    scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:            allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/confirmations", h.RecordConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/confirmations", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestRecordConfirmationByCode_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, executionID := uuid.New(), uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioID},
		}},
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		portfolios:    scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:            denyAll(),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/confirmations", h.RecordConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/confirmations", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestResolveConfirmationByCode_AuthorizedFundPassesGate(t *testing.T) {
	t.Parallel()
	portfolioID, fundID, confirmationID := uuid.New(), uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, PortfolioID: portfolioID},
		}},
		portfolios: scopedPortfolio("PORT-A", portfolioID, fundID),
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve", h.ResolveConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/confirmations/"+confirmationID.String()+"/resolve", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestResolveConfirmationByCode_CrossFundReturns403(t *testing.T) {
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
	r.Post("/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve", h.ResolveConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/confirmations/"+confirmationID.String()+"/resolve", strings.NewReader(malformedBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}
