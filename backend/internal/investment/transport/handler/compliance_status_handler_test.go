package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type statusComplianceChecker struct {
	result *contract.ProposedOrderResult
}

func (c *statusComplianceChecker) CheckProposedOrder(context.Context, contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return c.result, nil
}

type statusAuditLogger struct {
	entries []contract.AuditEntry
}

func (a *statusAuditLogger) LogAction(entry contract.AuditEntry) error {
	a.entries = append(a.entries, entry)
	return nil
}

func (a *statusAuditLogger) LogActionStrict(_ context.Context, entry contract.AuditEntry) error {
	a.entries = append(a.entries, entry)
	return nil
}

type statusDecisionRepo struct {
	decision *entity.Decision
	updates  int
}

func (r *statusDecisionRepo) GetByID(context.Context, uuid.UUID) (*entity.Decision, error) {
	return r.decision, nil
}
func (*statusDecisionRepo) GetByDecisionNumber(context.Context, string) (*entity.Decision, error) {
	return nil, nil
}
func (*statusDecisionRepo) FindDecisionSubjectRefByNumber(context.Context, string) (*domain.DecisionSubjectRef, error) {
	return nil, nil
}
func (*statusDecisionRepo) Create(context.Context, pgx.Tx, *entity.Decision) error { return nil }
func (r *statusDecisionRepo) Update(context.Context, pgx.Tx, *entity.Decision) error {
	r.updates++
	return nil
}
func (*statusDecisionRepo) List(context.Context, domain.DecisionListFilter) ([]*entity.Decision, int, error) {
	return nil, 0, nil
}
func (*statusDecisionRepo) NextDecisionNumber(context.Context, pgx.Tx, time.Time) (string, error) {
	return "", nil
}
func (*statusDecisionRepo) UpdateStatus(context.Context, uuid.UUID, vo.DecisionStatus, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}

type statusExecutionRepo struct {
	creates int
}

func (r *statusExecutionRepo) Create(context.Context, pgx.Tx, *entity.Execution) error {
	r.creates++
	return nil
}
func (*statusExecutionRepo) Update(context.Context, pgx.Tx, *entity.Execution) error { return nil }
func (*statusExecutionRepo) GetByID(context.Context, uuid.UUID) (*entity.Execution, error) {
	return nil, nil
}
func (*statusExecutionRepo) ListByDecision(context.Context, uuid.UUID) ([]*entity.Execution, error) {
	return nil, nil
}
func (*statusExecutionRepo) ListByFundDate(context.Context, uuid.UUID, time.Time) ([]*entity.Execution, error) {
	return nil, nil
}

func complianceStatusDecision(status contract.ComplianceStatus) *entity.Decision {
	qty := decimal.NewFromInt(100)
	price := decimal.NewFromInt(35)
	return &entity.Decision{
		ID:             uuid.New(),
		FundID:         uuid.New(),
		PortfolioID:    uuid.New(),
		InstrumentCode: "PTT",
		BusinessDate:   time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
		Side:           vo.OrderSideBuy,
		Quantity:       &qty,
		LimitPrice:     &price,
		Currency:       "THB",
		Exchange:       "SET",
		Status:         statusDecisionLifecycle(status),
	}
}

func statusDecisionLifecycle(status contract.ComplianceStatus) vo.DecisionLifecycleStatus {
	// Tests use NOT_CONFIGURED for submit (DRAFT) and UNAVAILABLE for execution
	// (APPROVED) to cover both typed 422 codes through real handlers.
	if status == contract.ComplianceStatusUnavailable {
		return vo.DecisionLifecycleApproved
	}
	return vo.DecisionLifecycleDraft
}

func decodeComplianceError(t *testing.T, recorder *httptest.ResponseRecorder, wantCode, wantCheckGroupID string) {
	t.Helper()
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", recorder.Code, recorder.Body.String())
	}
	var body httputil.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != wantCode {
		t.Fatalf("code = %q, want %q", body.Code, wantCode)
	}
	details, ok := body.Details.(map[string]any)
	if !ok {
		t.Fatalf("details type = %T, want map[string]any", body.Details)
	}
	if details["check_group_id"] != wantCheckGroupID {
		t.Fatalf("check_group_id = %v, want %s", details["check_group_id"], wantCheckGroupID)
	}
}

func TestSubmitDecisionByCode_LiveNotConfigured_ReturnsTyped422WithoutPersistence(t *testing.T) {
	decision := complianceStatusDecision(contract.ComplianceStatusNotConfigured)
	decisionRepo := &statusDecisionRepo{decision: decision}
	checkGroupID := uuid.New()
	checker := &statusComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: checkGroupID,
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}}
	portfolioCode := "LIVE-01"
	portfolioRepo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		portfolioCode: {
			ID:            decision.PortfolioID,
			Code:          portfolioCode,
			PortfolioType: vo.PortfolioTypeLive,
		},
	}}
	audit := &statusAuditLogger{}
	cmd := command.NewDecisionCommandHandler(nil, decisionRepo, nil, nil, audit, nil)
	cmd.SetComplianceChecker(checker)
	cmd.SetPortfolioRepository(portfolioRepo)
	h := NewDecisionHandler(decisionRepo, cmd)
	h.SetPortfolioRepository(portfolioRepo)

	router := chi.NewRouter()
	router.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/submit", h.SubmitDecisionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost,
		"/portfolios/"+portfolioCode+"/decisions/"+decision.ID.String()+"/submit", nil))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	decodeComplianceError(t, recorder, command.ComplianceErrorCodeNotConfigured, checkGroupID.String())
	if decisionRepo.updates != 0 {
		t.Fatalf("decision updates = %d, want 0", decisionRepo.updates)
	}
}

func TestSubmitDecisionV1_LiveNotConfigured_ReturnsCorrelatedTyped422WithoutPersistence(t *testing.T) {
	decision := complianceStatusDecision(contract.ComplianceStatusNotConfigured)
	decisionRepo := &statusDecisionRepo{decision: decision}
	checkGroupID := uuid.New()
	checker := &statusComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: checkGroupID,
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}}
	portfolioCode := "LIVE-01"
	portfolioRepo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		portfolioCode: {ID: decision.PortfolioID, Code: portfolioCode, PortfolioType: vo.PortfolioTypeLive},
	}}
	cmd := command.NewDecisionCommandHandler(nil, decisionRepo, nil, nil, &statusAuditLogger{}, nil)
	cmd.SetComplianceChecker(checker)
	cmd.SetPortfolioRepository(portfolioRepo)
	h := NewDecisionHandler(decisionRepo, cmd)

	router := chi.NewRouter()
	router.Post("/investment/decisions/{id}/submit", h.SubmitDecision)
	req := withUserClaims(httptest.NewRequest(http.MethodPost,
		"/investment/decisions/"+decision.ID.String()+"/submit", nil))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	decodeComplianceError(t, recorder, command.ComplianceErrorCodeNotConfigured, checkGroupID.String())
	if decisionRepo.updates != 0 {
		t.Fatalf("decision updates = %d, want 0", decisionRepo.updates)
	}
}

func TestCreateExecutionByCode_LiveUnavailable_ReturnsTyped422WithoutPersistence(t *testing.T) {
	decision := complianceStatusDecision(contract.ComplianceStatusUnavailable)
	decisionRepo := &statusDecisionRepo{decision: decision}
	executionRepo := &statusExecutionRepo{}
	checkGroupID := uuid.New()
	checker := &statusComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   checkGroupID,
		Status:         contract.ComplianceStatusUnavailable,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
		Breaches: []contract.ProposedOrderBreach{{
			RuleTypeID: "allocation.asset_class_min",
			Message:    "asset class classification unavailable for instruments: NEW-BOND",
		}},
	}}
	portfolioCode := "LIVE-01"
	portfolioRepo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		portfolioCode: {
			ID:            decision.PortfolioID,
			Code:          portfolioCode,
			PortfolioType: vo.PortfolioTypeLive,
		},
	}}
	audit := &statusAuditLogger{}
	cmd := command.NewExecutionCommandHandler(nil, decisionRepo, executionRepo, audit, nil)
	cmd.SetComplianceChecker(checker)
	cmd.SetPortfolioRepository(portfolioRepo)
	h := NewExecutionHandler(executionRepo, cmd)
	h.SetDecisionRepository(decisionRepo)
	h.SetPortfolioRepository(portfolioRepo)

	router := chi.NewRouter()
	router.Post("/portfolios/{portfolioCode}/decisions/{decisionId}/executions", h.CreateExecutionByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodPost,
		"/portfolios/"+portfolioCode+"/decisions/"+decision.ID.String()+"/executions",
		bytes.NewBufferString(`{}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	decodeComplianceError(t, recorder, command.ComplianceErrorCodeUnavailable, checkGroupID.String())
	if executionRepo.creates != 0 {
		t.Fatalf("execution creates = %d, want 0", executionRepo.creates)
	}
	if decisionRepo.updates != 0 {
		t.Fatalf("decision updates = %d, want 0", decisionRepo.updates)
	}
}

func TestCreateExecutionV1_LiveUnavailable_ReturnsCorrelatedTyped422WithoutPersistence(t *testing.T) {
	decision := complianceStatusDecision(contract.ComplianceStatusUnavailable)
	decisionRepo := &statusDecisionRepo{decision: decision}
	executionRepo := &statusExecutionRepo{}
	checkGroupID := uuid.New()
	checker := &statusComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   checkGroupID,
		Status:         contract.ComplianceStatusUnavailable,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
		Breaches: []contract.ProposedOrderBreach{{
			RuleTypeID: "allocation.asset_class_min",
			Message:    "asset class classification unavailable for instruments: NEW-BOND",
		}},
	}}
	portfolioCode := "LIVE-01"
	portfolioRepo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		portfolioCode: {ID: decision.PortfolioID, Code: portfolioCode, PortfolioType: vo.PortfolioTypeLive},
	}}
	cmd := command.NewExecutionCommandHandler(nil, decisionRepo, executionRepo, &statusAuditLogger{}, nil)
	cmd.SetComplianceChecker(checker)
	cmd.SetPortfolioRepository(portfolioRepo)
	h := NewExecutionHandler(executionRepo, cmd)

	router := chi.NewRouter()
	router.Post("/investment/executions", h.CreateExecution)
	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/investment/executions",
		bytes.NewBufferString(`{"decision_id":"`+decision.ID.String()+`"}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	decodeComplianceError(t, recorder, command.ComplianceErrorCodeUnavailable, checkGroupID.String())
	if executionRepo.creates != 0 || decisionRepo.updates != 0 {
		t.Fatalf("writes after unavailable status: executions=%d decisions=%d", executionRepo.creates, decisionRepo.updates)
	}
}
