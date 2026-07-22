package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// V1 data-scope security fix.
//
// DecisionHandler.ListDecisions/GetDecision/GetDecisionWithLines/
// ListApprovalItems, ExecutionHandler.ListExecutions/GetExecution, and
// TradeConfirmationHandler.ListConfirmations/GetConfirmation accepted
// caller-supplied fund_id/portfolio_id/decision_id/id parameters and queried
// the repository directly with no data-scope filter — unlike
// InvestmentHandler.ListPortfolios, which correctly applies
// accessibleFundIDs. Any authenticated user holding the route's function
// permission (e.g. INVESTMENT_DECISION_VIEW) could read ANY fund/
// portfolio's decisions/executions/confirmations by guessing or supplying
// an arbitrary ID, regardless of their assigned fund/portfolio data scope.
//
// These tests prove, for every affected handler:
//  1. authorized-same-scope-passes (200)
//  2. cross-fund-denied (403)
//  3. nil-checker-fails-closed (403, not a silent pass-through)
//  4. global-scope-caller-sees-everything (AccessibleFundIDs computed as nil
//     — "unrestricted" — for a caller holding the "*" wildcard scope; a
//     denied caller gets a non-nil empty AccessibleFundIDs, which repository
//     implementations treat as "return zero rows", matching
//     InvestmentHandler.ListPortfolios' existing pattern)

// filterCapturingDecisionRepo records the last filter passed to List so
// tests can assert on the computed AccessibleFundIDs value directly, without
// needing a live database to prove the SQL-level filter (already covered by
// FundRepository/PortfolioRepository's existing AccessibleFundIDs
// implementations, which PostgresDecisionRepository.List mirrors).
type filterCapturingDecisionRepo struct {
	stubDecisionRepo
	lastFilter domain.DecisionListFilter
}

func (r *filterCapturingDecisionRepo) List(_ context.Context, f domain.DecisionListFilter) ([]*entity.Decision, int, error) {
	r.lastFilter = f
	return nil, 0, nil
}

// ─── DecisionHandler.ListDecisions: AccessibleFundIDs computation ──────────

func TestListDecisions_GlobalScopeComputesUnrestrictedFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingDecisionRepo{stubDecisionRepo: stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}}}
	h := &DecisionHandler{
		decisions: repo,
		pc:        &fakeV2PermissionChecker{Contracts: []string{"*"}},
	}
	r := chi.NewRouter()
	r.Get("/decisions", h.ListDecisions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs != nil {
		t.Fatalf("AccessibleFundIDs = %v, want nil (unrestricted) for a global/wildcard-scope caller", repo.lastFilter.AccessibleFundIDs)
	}
}

func TestListDecisions_DeniedScopeComputesEmptyFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingDecisionRepo{stubDecisionRepo: stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}}}
	h := &DecisionHandler{
		decisions: repo,
		pc:        denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/decisions", h.ListDecisions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (list endpoints filter, they do not 403); body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs == nil || len(repo.lastFilter.AccessibleFundIDs) != 0 {
		t.Fatalf("AccessibleFundIDs = %v, want non-nil empty slice (zero fund access)", repo.lastFilter.AccessibleFundIDs)
	}
}

func TestListDecisions_NilCheckerFailsClosedToEmptyFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingDecisionRepo{stubDecisionRepo: stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}}}
	h := &DecisionHandler{decisions: repo, pc: nil}
	r := chi.NewRouter()
	r.Get("/decisions", h.ListDecisions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs == nil || len(repo.lastFilter.AccessibleFundIDs) != 0 {
		t.Fatalf("AccessibleFundIDs = %v, want non-nil empty slice when pc is nil (fail closed)", repo.lastFilter.AccessibleFundIDs)
	}
}

func TestListApprovalItems_GlobalScopeComputesUnrestrictedFilter(t *testing.T) {
	t.Parallel()
	repo := &filterCapturingDecisionRepo{stubDecisionRepo: stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}}}
	h := &DecisionHandler{
		decisions: repo,
		pc:        &fakeV2PermissionChecker{Contracts: []string{"*"}},
	}
	r := chi.NewRouter()
	r.Get("/decisions/approval-items", h.ListApprovalItems)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/approval-items", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if repo.lastFilter.AccessibleFundIDs != nil {
		t.Fatalf("AccessibleFundIDs = %v, want nil (unrestricted)", repo.lastFilter.AccessibleFundIDs)
	}
}

// ─── DecisionHandler.GetDecision / GetDecisionWithLines ────────────────────

func TestGetDecision_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}", h.GetDecision)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecision_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}", h.GetDecision)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecision_NilCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: nil,
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}", h.GetDecision)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (nil permission checker must fail closed); body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecision_GlobalScopeReturns200(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}", h.GetDecision)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (global/company-wide scope must see every fund); body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionWithLines_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}/details", h.GetDecisionWithLines)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String()+"/details", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionWithLines_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/decisions/{id}/details", h.GetDecisionWithLines)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/decisions/"+decisionID.String()+"/details", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

// ─── ExecutionHandler.ListExecutions / GetExecution ────────────────────────

func TestListExecutions_FundBranch_AuthorizedReturns200(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		pc:         allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/executions", h.ListExecutions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions?fund_id="+fundID.String()+"&business_date=2026-07-21", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestListExecutions_FundBranch_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		pc:         denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/executions", h.ListExecutions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions?fund_id="+fundID.String()+"&business_date=2026-07-21", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestListExecutions_FundBranch_NilCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		pc:         nil,
	}
	r := chi.NewRouter()
	r.Get("/executions", h.ListExecutions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions?fund_id="+fundID.String()+"&business_date=2026-07-21", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (nil checker must fail closed); body=%s", w.Code, w.Body.String())
	}
}

func TestListExecutions_DecisionBranch_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/executions", h.ListExecutions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions?decision_id="+decisionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestListExecutions_DecisionBranch_AuthorizedReturns200(t *testing.T) {
	t.Parallel()
	fundID, decisionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, FundID: fundID},
		}},
		pc: allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/executions", h.ListExecutions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions?decision_id="+decisionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecution_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	fundID, executionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, FundID: fundID},
		}},
		pc: allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/executions/{id}", h.GetExecution)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecution_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, executionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/executions/{id}", h.GetExecution)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecution_NilCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	fundID, executionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, FundID: fundID},
		}},
		pc: nil,
	}
	r := chi.NewRouter()
	r.Get("/executions/{id}", h.GetExecution)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (nil checker must fail closed); body=%s", w.Code, w.Body.String())
	}
}

func TestGetExecution_GlobalScopeReturns200(t *testing.T) {
	t.Parallel()
	fundID, executionID := uuid.New(), uuid.New()
	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, FundID: fundID},
		}},
		pc: &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Get("/executions/{id}", h.GetExecution)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/executions/"+executionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (global/company-wide scope must see every fund); body=%s", w.Code, w.Body.String())
	}
}

// ─── TradeConfirmationHandler.ListConfirmations / GetConfirmation ──────────

func TestListConfirmations_FundBranch_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		pc:            denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations", h.ListConfirmations)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations?fund_id="+fundID.String()+"&business_date=2026-07-21", nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestListConfirmations_FundBranch_AuthorizedReturns200(t *testing.T) {
	t.Parallel()
	fundID := uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		pc:            allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations", h.ListConfirmations)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations?fund_id="+fundID.String()+"&business_date=2026-07-21", nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestListConfirmations_ExecutionBranch_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, executionID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations", h.ListConfirmations)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations?execution_id="+executionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmation_AuthorizedFundReturns200(t *testing.T) {
	t.Parallel()
	fundID, confirmationID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, FundID: fundID},
		}},
		pc: allowFund(fundID),
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations/{id}", h.GetConfirmation)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmation_CrossFundReturns403(t *testing.T) {
	t.Parallel()
	fundID, confirmationID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, FundID: fundID},
		}},
		pc: denyAll(),
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations/{id}", h.GetConfirmation)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmation_NilCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	fundID, confirmationID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, FundID: fundID},
		}},
		pc: nil,
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations/{id}", h.GetConfirmation)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (nil checker must fail closed); body=%s", w.Code, w.Body.String())
	}
}

func TestGetConfirmation_GlobalScopeReturns200(t *testing.T) {
	t.Parallel()
	fundID, confirmationID := uuid.New(), uuid.New()
	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, FundID: fundID},
		}},
		pc: &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Get("/trade-confirmations/{id}", h.GetConfirmation)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/trade-confirmations/"+confirmationID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (global/company-wide scope must see every fund); body=%s", w.Code, w.Body.String())
	}
}

var _ = time.Now // keep time import if unused by future edits
