package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Milestone 5 (docs/handoff/portfolio-v2-claude-implementation-prompt.md):
// "Add tests for portfolio ownership consistency." These handler-level
// tests cover exactly that — every V2 decision/execution/confirmation
// write handler resolves {portfolioCode} and then loads the target entity
// by ID *before* touching the concrete command handler (h.cmd), so the
// 404/409 ownership paths are fully reachable here without a live DB. The
// success paths (create/submit/fill/resolve actually mutating state) need
// a real pgx transaction and are covered by
// tests/e2e/portfolio_v2_decision_test.go.

type stubDecisionRepo struct {
	byID map[uuid.UUID]*entity.Decision
}

func (r *stubDecisionRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Decision, error) {
	return r.byID[id], nil
}
func (r *stubDecisionRepo) GetByDecisionNumber(context.Context, string) (*entity.Decision, error) {
	return nil, nil
}
func (r *stubDecisionRepo) FindDecisionSubjectRefByNumber(context.Context, string) (*domain.DecisionSubjectRef, error) {
	return nil, nil
}
func (r *stubDecisionRepo) Create(context.Context, pgx.Tx, *entity.Decision) error { return nil }
func (r *stubDecisionRepo) Update(context.Context, pgx.Tx, *entity.Decision) error { return nil }
func (r *stubDecisionRepo) List(context.Context, domain.DecisionListFilter) ([]*entity.Decision, int, error) {
	return nil, 0, nil
}
func (r *stubDecisionRepo) NextDecisionNumber(context.Context, pgx.Tx, time.Time) (string, error) {
	return "", nil
}
func (r *stubDecisionRepo) UpdateStatus(context.Context, uuid.UUID, vo.DecisionStatus, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}

type stubExecutionRepo struct {
	byID map[uuid.UUID]*entity.Execution
}

func (r *stubExecutionRepo) Create(context.Context, pgx.Tx, *entity.Execution) error { return nil }
func (r *stubExecutionRepo) Update(context.Context, pgx.Tx, *entity.Execution) error { return nil }
func (r *stubExecutionRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Execution, error) {
	return r.byID[id], nil
}
func (r *stubExecutionRepo) ListByDecision(context.Context, uuid.UUID) ([]*entity.Execution, error) {
	return nil, nil
}
func (r *stubExecutionRepo) ListByFundDate(context.Context, uuid.UUID, time.Time) ([]*entity.Execution, error) {
	return nil, nil
}
func (r *stubExecutionRepo) ListByPortfolio(context.Context, uuid.UUID, domain.ExecutionListFilter) ([]*entity.Execution, int, error) {
	return nil, 0, nil
}

type stubConfirmationRepo struct {
	byID map[uuid.UUID]*entity.TradeConfirmation
}

func (r *stubConfirmationRepo) Create(context.Context, pgx.Tx, *entity.TradeConfirmation) error {
	return nil
}
func (r *stubConfirmationRepo) Update(context.Context, pgx.Tx, *entity.TradeConfirmation) error {
	return nil
}
func (r *stubConfirmationRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.TradeConfirmation, error) {
	return r.byID[id], nil
}
func (r *stubConfirmationRepo) ListByExecution(context.Context, uuid.UUID) ([]*entity.TradeConfirmation, error) {
	return nil, nil
}
func (r *stubConfirmationRepo) ListByFundDate(context.Context, uuid.UUID, time.Time) ([]*entity.TradeConfirmation, error) {
	return nil, nil
}
func (r *stubConfirmationRepo) ListByPortfolio(context.Context, uuid.UUID, domain.TradeConfirmationListFilter) ([]*entity.TradeConfirmation, int, error) {
	return nil, 0, nil
}
func (r *stubConfirmationRepo) GetByBrokerReference(context.Context, string) (*entity.TradeConfirmation, error) {
	return nil, nil
}

func newOwnedPortfolioRepo(code string, portfolioID uuid.UUID) *stubPortfolioRepo {
	return &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		code: {ID: portfolioID, Code: code, FundID: uuid.New()},
	}}
}

// ─── Decisions ──────────────────────────────────────────────────────────────

func TestGetDecisionByCode_UnknownPortfolioCodeReturns404(t *testing.T) {
	t.Parallel()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}},
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/portfolios/NOPE/decisions/"+uuid.NewString(), nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionByCode_AmbiguousCodeReturns409(t *testing.T) {
	t.Parallel()
	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: &stubPortfolioRepo{
			byCodeErr: &domain.ErrAmbiguousPortfolioCode{Code: "DUP", Count: 2},
		},
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/portfolios/DUP/decisions/"+uuid.NewString(), nil))
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

// TestGetDecisionByCode_CrossPortfolioReturns404 is the ownership-consistency
// assertion: a decision that belongs to portfolio B must 404 when accessed
// through portfolio A's code, even though the decision ID itself is valid.
func TestGetDecisionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	decisionID := uuid.New()

	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (cross-portfolio decision must not be visible); body=%s", w.Code, w.Body.String())
	}
}

func TestGetDecisionByCode_SamePortfolioReturns200(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	decisionID := uuid.New()

	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioA},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestSubmitDecisionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	decisionID := uuid.New()

	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/submit", h.SubmitDecisionByCode)
	w := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/submit", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, w)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCancelDecisionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	decisionID := uuid.New()

	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/cancel", h.CancelDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/cancel", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

// TestGetDecisionByCode_NilPermissionCheckerFailsClosed proves
// resolvePortfolioByCode (the shared {portfolioCode} resolver behind every
// V2 decision/execution/confirmation route) denies access rather than
// silently allowing it when the handler's permission checker was never
// wired — mirroring the fail-closed guarantee v1_data_scope_test.go already
// proves for the V1 routes. Production wiring in module.go always sets a
// real checker, but this test pins the resolver's own contract regardless
// of wiring correctness elsewhere.
func TestGetDecisionByCode_NilPermissionCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	decisionID := uuid.New()

	h := &DecisionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioA},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		// pc intentionally left nil.
	}
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/decisions/{decisionId}", h.GetDecisionByCode)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/PORT-A/decisions/"+decisionID.String(), nil)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (nil permission checker must deny, not silently allow); body=%s", w.Code, w.Body.String())
	}
}

func TestCreateDecisionByCode_UnknownPortfolioCodeReturns404(t *testing.T) {
	t.Parallel()
	h := &DecisionHandler{
		decisions:  &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{}},
		portfolios: &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions", h.CreateDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/NOPE/decisions", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

// ─── Executions ─────────────────────────────────────────────────────────────

func TestCreateExecutionByCode_DecisionFromOtherPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	decisionID := uuid.New()

	h := &ExecutionHandler{
		decisions: &stubDecisionRepo{byID: map[uuid.UUID]*entity.Decision{
			decisionID: {ID: decisionID, PortfolioID: portfolioB},
		}},
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/executions", h.CreateExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/decisions/"+decisionID.String()+"/executions", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestFillExecutionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	executionID := uuid.New()

	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/fill", h.FillExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/fill", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestCancelExecutionByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	executionID := uuid.New()

	h := &ExecutionHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/cancel", h.CancelExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/cancel", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

// ─── Trade confirmations ────────────────────────────────────────────────────

func TestRecordConfirmationByCode_ExecutionFromOtherPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	executionID := uuid.New()

	h := &TradeConfirmationHandler{
		executions: &stubExecutionRepo{byID: map[uuid.UUID]*entity.Execution{
			executionID: {ID: executionID, PortfolioID: portfolioB},
		}},
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{}},
		portfolios:    newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:            &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/executions/{executionId}/confirmations", h.RecordConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/executions/"+executionID.String()+"/confirmations", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestResolveConfirmationByCode_CrossPortfolioReturns404(t *testing.T) {
	t.Parallel()
	portfolioA := uuid.New()
	portfolioB := uuid.New()
	confirmationID := uuid.New()

	h := &TradeConfirmationHandler{
		confirmations: &stubConfirmationRepo{byID: map[uuid.UUID]*entity.TradeConfirmation{
			confirmationID: {ID: confirmationID, PortfolioID: portfolioB},
		}},
		portfolios: newOwnedPortfolioRepo("PORT-A", portfolioA),
		pc:         &fakeV2PermissionChecker{Global: true},
	}
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve", h.ResolveConfirmationByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/PORT-A/confirmations/"+confirmationID.String()+"/resolve", nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}
