package command

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// fakeDecisionRepo is an in-memory DecisionRepository for tests.
type fakeDecisionRepo struct {
	mu         sync.Mutex
	items      map[uuid.UUID]*entity.Decision
	getErr     error
	updateErr  error
	lastStatus vo.DecisionStatus
	lastGroup  uuid.UUID
	updates    int
}

func newFakeDecisionRepo(seed ...*entity.Decision) *fakeDecisionRepo {
	r := &fakeDecisionRepo{items: make(map[uuid.UUID]*entity.Decision)}
	for _, d := range seed {
		cp := *d
		r.items[d.ID] = &cp
	}
	return r
}

func (r *fakeDecisionRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Decision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.getErr != nil {
		return nil, r.getErr
	}
	d, ok := r.items[id]
	if !ok {
		return nil, &domain.ErrDecisionNotFound{DecisionID: id.String()}
	}
	cp := *d
	return &cp, nil
}

func (r *fakeDecisionRepo) UpdateStatus(
	_ context.Context, id uuid.UUID, to vo.DecisionStatus,
	checkGroupID uuid.UUID, updatedBy uuid.UUID, updatedAt time.Time,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.updateErr != nil {
		return r.updateErr
	}
	d, ok := r.items[id]
	if !ok {
		return &domain.ErrDecisionNotFound{DecisionID: id.String()}
	}
	d.Status = to
	g := checkGroupID
	d.ComplianceCheckGroupID = &g
	d.UpdatedAt = updatedAt
	d.UpdatedBy = updatedBy
	r.lastStatus = to
	r.lastGroup = checkGroupID
	r.updates++
	return nil
}

// fakeComplianceChecker returns a pre-seeded result or error.
type fakeComplianceChecker struct {
	result *contract.ProposedOrderResult
	err    error
	calls  int
	lastIn contract.ProposedOrderCheck
}

func (c *fakeComplianceChecker) CheckProposedOrder(
	_ context.Context, req contract.ProposedOrderCheck,
) (*contract.ProposedOrderResult, error) {
	c.calls++
	c.lastIn = req
	if c.err != nil {
		return nil, c.err
	}
	return c.result, nil
}

// helpers

func draftDecision() *entity.Decision {
	return &entity.Decision{
		ID:           uuid.New(),
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		Ticker:       "PTT",
		Side:         vo.OrderSideBuy,
		Quantity:     decimal.NewFromInt(1000),
		Price:        decimal.NewFromFloat(35.5),
		Currency:     "THB",
		Exchange:     "SET",
		BusinessDate: time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC),
		Status:       vo.DecisionStatusDraft,
		CreatedBy:    uuid.New(),
		CreatedAt:    time.Now().UTC().Add(-time.Hour),
	}
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// ─────────────────────────────────────────────────────────────────────────────
// tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitDecision_HappyPath_Pass(t *testing.T) {
	d := draftDecision()
	repo := newFakeDecisionRepo(d)
	groupID := uuid.New()
	checker := &fakeComplianceChecker{
		result: &contract.ProposedOrderResult{
			CheckGroupID:   groupID,
			Verdict:        contract.ComplianceVerdictPass,
			RulesEvaluated: 4,
		},
	}
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	h := NewSubmitDecisionForExecutionHandler(repo, checker, fixedClock(now))

	res, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != vo.DecisionStatusSubmitted {
		t.Fatalf("want SUBMITTED, got %s", res.Status)
	}
	if res.CheckGroupID != groupID {
		t.Fatalf("check group not propagated")
	}
	if res.SubmittedAt != now {
		t.Fatalf("clock not applied")
	}
	if repo.lastStatus != vo.DecisionStatusSubmitted || repo.updates != 1 {
		t.Fatalf("repo state wrong: status=%s updates=%d", repo.lastStatus, repo.updates)
	}
	if checker.lastIn.Ticker != "PTT" || checker.lastIn.Side != contract.ComplianceOrderSideBuy {
		t.Fatalf("checker input not populated: %+v", checker.lastIn)
	}
}

func TestSubmitDecision_Warn_StillSubmitted(t *testing.T) {
	// WARN verdict should NOT block submission — traders acknowledge downstream.
	d := draftDecision()
	repo := newFakeDecisionRepo(d)
	checker := &fakeComplianceChecker{
		result: &contract.ProposedOrderResult{
			CheckGroupID: uuid.New(),
			Verdict:      contract.ComplianceVerdictWarn,
			Breaches: []contract.ProposedOrderBreach{{
				RuleTypeID: "concentration.sector",
				Verdict:    contract.ComplianceVerdictWarn,
				Severity:   "WARN",
				Message:    "sector exposure 45%",
			}},
		},
	}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)

	res, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("WARN must not error: %v", err)
	}
	if res.Status != vo.DecisionStatusSubmitted {
		t.Fatalf("WARN should still submit, got %s", res.Status)
	}
	if len(res.Breaches) != 1 {
		t.Fatalf("warn breach not surfaced")
	}
}

func TestSubmitDecision_Block_RejectsAndMarksBlocked(t *testing.T) {
	d := draftDecision()
	repo := newFakeDecisionRepo(d)
	groupID := uuid.New()
	checker := &fakeComplianceChecker{
		result: &contract.ProposedOrderResult{
			CheckGroupID: groupID,
			Verdict:      contract.ComplianceVerdictBlock,
			Breaches: []contract.ProposedOrderBreach{{
				RuleTypeID: "restriction.blacklist",
				Verdict:    contract.ComplianceVerdictBlock,
				Severity:   "BLOCK",
				Message:    "ticker is blacklisted",
			}},
		},
	}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)

	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected BLOCK error, got nil")
	}
	var rejected *domain.ErrComplianceRejected
	if !errors.As(err, &rejected) {
		t.Fatalf("expected *ErrComplianceRejected, got %T", err)
	}
	if rejected.CheckGroupID != groupID.String() {
		t.Fatalf("check group not propagated in error: %s", rejected.CheckGroupID)
	}
	if repo.lastStatus != vo.DecisionStatusBlocked {
		t.Fatalf("decision not marked BLOCKED: %s", repo.lastStatus)
	}
}

func TestSubmitDecision_NotFound(t *testing.T) {
	repo := newFakeDecisionRepo() // empty
	checker := &fakeComplianceChecker{}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)
	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: uuid.New(),
		ActorID:    uuid.New(),
	})
	var nf *domain.ErrDecisionNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ErrDecisionNotFound, got %v", err)
	}
	if checker.calls != 0 {
		t.Fatalf("compliance must not be called when decision is missing")
	}
}

func TestSubmitDecision_NotDraft_Rejected(t *testing.T) {
	d := draftDecision()
	d.Status = vo.DecisionStatusSubmitted
	repo := newFakeDecisionRepo(d)
	checker := &fakeComplianceChecker{}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)

	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	var notDraft *domain.ErrDecisionNotDraft
	if !errors.As(err, &notDraft) {
		t.Fatalf("expected ErrDecisionNotDraft, got %v", err)
	}
	if checker.calls != 0 {
		t.Fatalf("compliance must not be called when decision is non-DRAFT")
	}
}

func TestSubmitDecision_ValidationErrors(t *testing.T) {
	repo := newFakeDecisionRepo()
	checker := &fakeComplianceChecker{}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)

	cases := []struct {
		name  string
		req   SubmitDecisionForExecutionRequest
		field string
	}{
		{"missing_decision_id", SubmitDecisionForExecutionRequest{ActorID: uuid.New()}, "decision_id"},
		{"missing_actor_id", SubmitDecisionForExecutionRequest{DecisionID: uuid.New()}, "actor_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(context.Background(), tc.req)
			var invalid *domain.ErrInvalidDecisionRequest
			if !errors.As(err, &invalid) {
				t.Fatalf("want *ErrInvalidDecisionRequest, got %v", err)
			}
			if invalid.Field != tc.field {
				t.Fatalf("want field=%q, got %q", tc.field, invalid.Field)
			}
		})
	}
}

func TestSubmitDecision_ComplianceInfraError_Propagates(t *testing.T) {
	d := draftDecision()
	repo := newFakeDecisionRepo(d)
	checker := &fakeComplianceChecker{err: errors.New("pipeline: DB timeout")}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)

	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected infra error")
	}
	// Must NOT be a typed business error — callers map these to 500.
	var rejected *domain.ErrComplianceRejected
	if errors.As(err, &rejected) {
		t.Fatalf("infra error must not masquerade as business rejection")
	}
	if repo.updates != 0 {
		t.Fatalf("decision must not be mutated on infra failure")
	}
}

func TestSubmitDecision_NilChecker_NotInitialised(t *testing.T) {
	h := NewSubmitDecisionForExecutionHandler(newFakeDecisionRepo(), nil, nil)
	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: uuid.New(), ActorID: uuid.New(),
	})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected not-initialised error, got %v", err)
	}
}

func TestSubmitDecision_PersistFailure_BlockPath(t *testing.T) {
	d := draftDecision()
	repo := newFakeDecisionRepo(d)
	repo.updateErr = fmt.Errorf("db down")
	checker := &fakeComplianceChecker{
		result: &contract.ProposedOrderResult{
			CheckGroupID: uuid.New(),
			Verdict:      contract.ComplianceVerdictBlock,
		},
	}
	h := NewSubmitDecisionForExecutionHandler(repo, checker, nil)
	_, err := h.Handle(context.Background(), SubmitDecisionForExecutionRequest{
		DecisionID: d.ID, ActorID: uuid.New(),
	})
	var rejected *domain.ErrComplianceRejected
	if errors.As(err, &rejected) {
		t.Fatalf("persist failure on block path should surface as infra error, not business rejection")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}
