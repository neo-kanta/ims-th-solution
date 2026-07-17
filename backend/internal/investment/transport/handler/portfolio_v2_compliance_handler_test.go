package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// fakePortfolioCompliance is a minimal contract.PortfolioComplianceContract
// double. It records the last request it received so tests can assert on
// what the handler resolved (in particular, ContractID/FundID).
type fakePortfolioCompliance struct {
	lastCheckReq            contract.ProposedOrderCheck
	lastPostTrade           contract.PortfolioPostTradeRequest
	lastBindReq             contract.PortfolioRuleBindingRequest
	lastDeactivatePortfolio uuid.UUID
	lastDeactivateBinding   uuid.UUID
	rules                   []contract.PortfolioRuleCatalogEntry
	bindErr                 error
	deactivateErr           error
	checkErr                error
}

func (f *fakePortfolioCompliance) CheckProposedOrder(_ context.Context, req contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	f.lastCheckReq = req
	if f.checkErr != nil {
		return nil, f.checkErr
	}
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

func (f *fakePortfolioCompliance) SimulateProposedOrder(_ context.Context, req contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	f.lastCheckReq = req
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

func (f *fakePortfolioCompliance) RunPortfolioPostTradeCheck(_ context.Context, req contract.PortfolioPostTradeRequest) (*contract.ProposedOrderResult, error) {
	f.lastPostTrade = req
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

func (f *fakePortfolioCompliance) ListPortfolioRules(context.Context, uuid.UUID) ([]contract.PortfolioRuleCatalogEntry, error) {
	return f.rules, nil
}

func (f *fakePortfolioCompliance) BindPortfolioRule(_ context.Context, req contract.PortfolioRuleBindingRequest) (*contract.PortfolioRuleBindingView, error) {
	f.lastBindReq = req
	if f.bindErr != nil {
		return nil, f.bindErr
	}
	return &contract.PortfolioRuleBindingView{
		BindingID:     uuid.New(),
		Severity:      req.Severity,
		IsActive:      true,
		EffectiveFrom: req.EffectiveFrom,
	}, nil
}

func (f *fakePortfolioCompliance) DeactivatePortfolioRuleBinding(_ context.Context, portfolioID, bindingID uuid.UUID) error {
	f.lastDeactivatePortfolio = portfolioID
	f.lastDeactivateBinding = bindingID
	return f.deactivateErr
}

func (f *fakePortfolioCompliance) ListPortfolioBreaches(context.Context, uuid.UUID, contract.PortfolioBreachFilter) ([]contract.PortfolioBreachView, error) {
	return nil, nil
}

var _ contract.PortfolioComplianceContract = (*fakePortfolioCompliance)(nil)

func newComplianceV2Handler(repo *stubPortfolioRepo, comp *fakePortfolioCompliance, fundID uuid.UUID) *InvestmentHandler {
	pc := &fakeV2PermissionChecker{Allowed: map[string]bool{fundID.String(): true}}
	return &InvestmentHandler{pc: pc, portfolios: repo, compliance: comp}
}

func serveCompliancePreTrade(h *InvestmentHandler, portfolioCode string, body []byte) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Post("/portfolios/{portfolioCode}/compliance/checks/pre-trade", h.RunPreTradeCheckByCode)

	req := withUserClaims(httptest.NewRequest(http.MethodPost, "/portfolios/"+portfolioCode+"/compliance/checks/pre-trade", bytes.NewReader(body)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validPreTradeBody() []byte {
	b, _ := json.Marshal(PortfolioPreTradeRequest{
		BusinessDate: "2026-07-07",
		Ticker:       "PTT",
		Side:         "BUY",
		Quantity:     "100",
		Price:        "35.5",
		Currency:     "THB",
		Exchange:     "SET",
	})
	return b
}

// TestRunPreTradeCheckByCode_ResolvesPortfolioCode verifies the V2 handler
// resolves {portfolioCode} to the portfolio's internal ID and passes it
// through as PortfolioID.
func TestRunPreTradeCheckByCode_ResolvesPortfolioCode(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, fundID)

	w := serveCompliancePreTrade(h, "TH-EQ-01", validPreTradeBody())

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if comp.lastCheckReq.PortfolioID != portfolioID {
		t.Errorf("PortfolioID = %s, want %s", comp.lastCheckReq.PortfolioID, portfolioID)
	}
	if comp.lastCheckReq.ContractID != fundID {
		t.Errorf("ContractID = %s, want fund_id %s", comp.lastCheckReq.ContractID, fundID)
	}
}

// TestRunPreTradeCheckByCode_NilFundID_PassesNilContractID verifies that a
// portfolio with no fund_id results in ContractID=uuid.Nil being forwarded —
// the "GLOBAL + PORTFOLIO scope only" path documented in
// portfolio_v2_compliance_handler.go's portfolioFundID.
func TestRunPreTradeCheckByCode_NilFundID_PassesNilContractID(t *testing.T) {
	t.Parallel()

	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: uuid.Nil, Code: "PF-NOFUND",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"PF-NOFUND": p}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, uuid.Nil)

	w := serveCompliancePreTrade(h, "PF-NOFUND", validPreTradeBody())

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if comp.lastCheckReq.PortfolioID != portfolioID {
		t.Errorf("PortfolioID = %s, want %s", comp.lastCheckReq.PortfolioID, portfolioID)
	}
	if comp.lastCheckReq.ContractID != uuid.Nil {
		t.Errorf("ContractID = %s, want uuid.Nil for a fund-less portfolio", comp.lastCheckReq.ContractID)
	}
}

// TestRunPreTradeCheckByCode_UnknownPortfolioCode_404 verifies an unknown
// code returns 404 without ever reaching the compliance contract.
func TestRunPreTradeCheckByCode_UnknownPortfolioCode_404(t *testing.T) {
	t.Parallel()

	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, uuid.Nil)

	w := serveCompliancePreTrade(h, "NOT-A-CODE", validPreTradeBody())

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
	if comp.lastCheckReq.PortfolioID != uuid.Nil {
		t.Error("compliance contract must not be called for an unresolved portfolio code")
	}
}

// TestRunPreTradeCheckByCode_ValidationError_400 verifies that a validation
// failure surfaced from the compliance module (wrapped in
// contract.ErrInvalidProposedOrder) maps to HTTP 400, not 500.
func TestRunPreTradeCheckByCode_ValidationError_400(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{
		checkErr: fmt.Errorf("%w: ticker is required", contract.ErrInvalidProposedOrder),
	}
	h := newComplianceV2Handler(repo, comp, fundID)

	w := serveCompliancePreTrade(h, "TH-EQ-01", validPreTradeBody())

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

// TestRunPreTradeCheckByCode_InvalidOrderID_400 verifies malformed optional
// order IDs are rejected instead of being silently replaced with a new UUID.
func TestRunPreTradeCheckByCode_InvalidOrderID_400(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, fundID)
	var payload map[string]any
	if err := json.Unmarshal(validPreTradeBody(), &payload); err != nil {
		t.Fatalf("decode test payload: %v", err)
	}
	payload["order_id"] = "not-a-uuid"
	body, _ := json.Marshal(payload)

	w := serveCompliancePreTrade(h, "TH-EQ-01", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
	if comp.lastCheckReq.PortfolioID != uuid.Nil {
		t.Fatal("compliance contract must not be called for an invalid order_id")
	}
}

// TestRunPreTradeCheckByCode_InfraError_500 verifies that a non-validation
// (infrastructure) failure still surfaces as HTTP 500.
func TestRunPreTradeCheckByCode_InfraError_500(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{checkErr: fmt.Errorf("db: connection reset")}
	h := newComplianceV2Handler(repo, comp, fundID)

	w := serveCompliancePreTrade(h, "TH-EQ-01", validPreTradeBody())

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", w.Code, w.Body.String())
	}
}

// TestListRulesByCode_ResolvesPortfolioCode verifies the rules endpoint
// resolves the code and returns the catalog from the contract as-is.
func TestListRulesByCode_ResolvesPortfolioCode(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{rules: []contract.PortfolioRuleCatalogEntry{
		{RuleInstanceID: uuid.New(), RuleTypeID: "cash.availability", Name: "Cash floor", IsActive: true},
	}}
	h := newComplianceV2Handler(repo, comp, fundID)

	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}/compliance/rules", h.ListRulesByCode)
	req := withUserClaims(httptest.NewRequest(http.MethodGet, "/portfolios/TH-EQ-01/compliance/rules", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	var envelope struct {
		Data []contract.PortfolioRuleCatalogEntry `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].RuleTypeID != "cash.availability" {
		t.Errorf("unexpected rules payload: %+v", envelope.Data)
	}
}

func TestBindRuleByCode_ResolvesPortfolioAndRule(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	ruleInstanceID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, fundID)

	router := chi.NewRouter()
	router.Post("/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings", h.BindRuleByCode)
	body, _ := json.Marshal(BindRuleRequest{
		Severity:      "BLOCK",
		Priority:      25,
		EffectiveFrom: "2026-07-10",
	})
	req := withUserClaims(httptest.NewRequest(
		http.MethodPost,
		"/portfolios/TH-EQ-01/compliance/rules/"+ruleInstanceID.String()+"/bindings",
		bytes.NewReader(body),
	))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", w.Code, w.Body.String())
	}
	if comp.lastBindReq.PortfolioID != portfolioID {
		t.Fatalf("PortfolioID = %s, want %s", comp.lastBindReq.PortfolioID, portfolioID)
	}
	if comp.lastBindReq.RuleInstanceID != ruleInstanceID {
		t.Fatalf("RuleInstanceID = %s, want %s", comp.lastBindReq.RuleInstanceID, ruleInstanceID)
	}
	if comp.lastBindReq.Severity != "BLOCK" || comp.lastBindReq.Priority != 25 {
		t.Fatalf("unexpected binding request: %+v", comp.lastBindReq)
	}
	if comp.lastBindReq.ActorID == uuid.Nil {
		t.Fatal("authenticated actor must be forwarded to the binding command")
	}
}

func TestDeactivateRuleBindingByCode_ResolvesPortfolioAndBinding(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolioID := uuid.New()
	ruleInstanceID := uuid.New()
	bindingID := uuid.New()
	p := &entity.Portfolio{
		ID: portfolioID, FundID: fundID, Code: "TH-EQ-01",
		BaseCurrency: "THB", ValuationCurrency: "THB",
		Status: vo.PortfolioStatusActive, InceptionDate: time.Now(),
	}
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": p}}
	comp := &fakePortfolioCompliance{}
	h := newComplianceV2Handler(repo, comp, fundID)

	router := chi.NewRouter()
	router.Delete("/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}", h.DeactivateRuleBindingByCode)
	req := withUserClaims(httptest.NewRequest(
		http.MethodDelete,
		"/portfolios/TH-EQ-01/compliance/rules/"+ruleInstanceID.String()+"/bindings/"+bindingID.String(),
		nil,
	))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", w.Code, w.Body.String())
	}
	if comp.lastDeactivatePortfolio != portfolioID {
		t.Fatalf("portfolio ID = %s, want %s", comp.lastDeactivatePortfolio, portfolioID)
	}
	if comp.lastDeactivateBinding != bindingID {
		t.Fatalf("binding ID = %s, want %s", comp.lastDeactivateBinding, bindingID)
	}
}
