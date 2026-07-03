package command

// Tests for the P0 compliance pre-trade gate and COMPLIANCE_RELEASE flow wired
// into DecisionCommandHandler.Submit and ApplyComplianceReleaseDecision.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ─────────────────────────────────────────────────────────────────────────────
// test-local fakes (shared across the tests in this file)
// ─────────────────────────────────────────────────────────────────────────────

// fakeFundRepo satisfies domain.FundRepository.
type fakeFundRepo struct {
	fund *entity.Fund
	err  error
}

func (r *fakeFundRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Fund, error) {
	return r.fund, r.err
}
func (r *fakeFundRepo) Create(_ context.Context, _ pgx.Tx, _ *entity.Fund) error { return nil }
func (r *fakeFundRepo) GetByCode(_ context.Context, _ string) (*entity.Fund, error) {
	return nil, nil
}
func (r *fakeFundRepo) GetByContractCode(_ context.Context, _ string) (*entity.Fund, error) {
	return nil, nil
}
func (r *fakeFundRepo) List(_ context.Context, _ domain.FundListFilter) ([]*entity.Fund, int, error) {
	return nil, 0, nil
}
func (r *fakeFundRepo) Update(_ context.Context, _ pgx.Tx, _ *entity.Fund) error { return nil }
func (r *fakeFundRepo) SoftDelete(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ int, _ uuid.UUID) error {
	return nil
}
func (r *fakeFundRepo) CountActivePortfolios(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

// fakeApprovalSubmitter records the last submission and returns a preset result.
type fakeApprovalSubmitter struct {
	mu     sync.Mutex
	result *contract.ApprovalSubmissionResult
	err    error
	calls  int
	lastIn contract.ApprovalSubmission
}

func (f *fakeApprovalSubmitter) SubmitForApproval(_ context.Context, in contract.ApprovalSubmission) (*contract.ApprovalSubmissionResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastIn = in
	return f.result, f.err
}

// trackingDecisionRepo wraps fakeDecisionRepo and records entity state on Update.
type trackingDecisionRepo struct {
	*fakeDecisionRepo
	updates []*entity.Decision
}

func (r *trackingDecisionRepo) Update(_ context.Context, _ pgx.Tx, d *entity.Decision) error {
	cp := *d
	r.updates = append(r.updates, &cp)
	r.mu.Lock()
	defer r.mu.Unlock()
	cp2 := *d
	r.items[d.ID] = &cp2
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func newDraftDec() *entity.Decision {
	qty := decimal.NewFromFloat(100)
	price := decimal.NewFromFloat(35.5)
	return &entity.Decision{
		ID:              uuid.New(),
		FundID:          uuid.New(),
		PortfolioID:     uuid.New(),
		ContractID:      uuid.New(),
		InstrumentCode:  "PTT",
		Side:            vo.OrderSideBuy,
		Status:          vo.DecisionLifecycleDraft,
		Quantity:        &qty,
		LimitPrice:      &price,
		Currency:        "THB",
		Exchange:        "SET",
		DecisionNumber:  "DEC-2026-001",
		BusinessDate:    time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC),
		CreatedBy:       uuid.New(),
		CreatedAt:       time.Now().UTC(),
		SubmitterUserID: uuid.New(),
	}
}

func defaultApprovalResult() *contract.ApprovalSubmissionResult {
	return &contract.ApprovalSubmissionResult{
		RequestID: uuid.New(),
		Status:    "PENDING_APPROVAL",
	}
}

// approvedReport returns a ResearchReport that passes the reference policy:
// ACTIVE status, review completed, BUY recommendation (compatible with Buy side).
// ApplicableContractID set to nil → accepted for any contract.
func approvedReport() *entity.ResearchReport {
	return &entity.ResearchReport{
		ID:             uuid.New(),
		ReportStatus:   vo.ReportStatusActive,
		ReviewStatus:   vo.ReviewStatusReviewCompleted,
		Recommendation: vo.RecommendationBuy,
	}
}

func handlerWithCompliance(
	decRepo domain.DecisionRepository,
	checker contract.ComplianceChecker,
	submitter *fakeApprovalSubmitter,
) *DecisionCommandHandler {
	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	if checker != nil {
		h.SetComplianceChecker(checker)
	}
	if submitter != nil {
		h.SetApprovalSubmitter(submitter)
	}
	return h
}

// ─────────────────────────────────────────────────────────────────────────────
// Config-driven report gate tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmit_FundRequiresReport_NoReport_Rejected(t *testing.T) {
	d := newDraftDec()
	d.ResearchReportID = nil

	decRepo := newFakeDecisionRepo(d)
	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetFundRepository(&fakeFundRepo{fund: &entity.Fund{
		ID:                               d.FundID,
		RequireResearchReportForDecision: true,
	}})

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error when fund requires report and none is linked")
	}
	var le *domain.ErrDecisionLifecycle
	if !errors.As(err, &le) {
		t.Fatalf("expected *ErrDecisionLifecycle, got %T: %v", err, err)
	}
}

func TestSubmit_FundRequiresReport_ReportLinked_Proceeds(t *testing.T) {
	d := newDraftDec()
	rep := approvedReport()
	d.ResearchReportID = &rep.ID

	repRepo := newFakeResearchReportRepo()
	repRepo.seed(rep)

	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}

	h := NewDecisionCommandHandler(nil, decRepo, repRepo, nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetFundRepository(&fakeFundRepo{fund: &entity.Fund{
		ID:                               d.FundID,
		RequireResearchReportForDecision: true,
	}})
	h.SetApprovalSubmitter(submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("submit with required report linked: %v", err)
	}
}

func TestSubmit_FundNotRequireReport_NoReport_Proceeds(t *testing.T) {
	d := newDraftDec()
	d.ResearchReportID = nil

	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}

	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetFundRepository(&fakeFundRepo{fund: &entity.Fund{
		ID:                               d.FundID,
		RequireResearchReportForDecision: false,
	}})
	h.SetApprovalSubmitter(submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("fund does not require report — submit must succeed: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Compliance pre-trade check gate tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmit_ComplianceNil_ContinuesToApproval(t *testing.T) {
	d := newDraftDec()
	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(decRepo, nil, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("nil compliance checker: submit must proceed to approval, got: %v", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("expected 1 approval submission, got %d", submitter.calls)
	}
}

func TestSubmit_CompliancePass_ContinuesToApproval(t *testing.T) {
	d := newDraftDec()
	checkGroupID := uuid.New()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   checkGroupID,
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 3,
	}}
	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(decRepo, checker, submitter)

	result, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("compliance PASS: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil decision on PASS")
	}
	if result.ComplianceCheckGroupID == nil || *result.ComplianceCheckGroupID != checkGroupID {
		t.Errorf("ComplianceCheckGroupID must be stored; got %v", result.ComplianceCheckGroupID)
	}
	if submitter.calls != 1 {
		t.Fatalf("expected 1 approval submission on PASS, got %d", submitter.calls)
	}
}

func TestSubmit_ComplianceWarn_ContinuesToApproval(t *testing.T) {
	d := newDraftDec()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictWarn,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.limit",
			Verdict:     contract.ComplianceVerdictWarn,
			Message:     "approaching limit",
			Overridable: true,
		}},
	}}
	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(decRepo, checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("compliance WARN must not block submission: %v", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("expected 1 approval submission on WARN, got %d", submitter.calls)
	}
}

func TestSubmit_ComplianceBlock_NonReleasable_ReturnsError(t *testing.T) {
	d := newDraftDec()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.hard_limit",
			Verdict:     contract.ComplianceVerdictBlock,
			Message:     "hard limit breached",
			Overridable: false,
		}},
	}}
	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(decRepo, checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("expected ErrComplianceRejected on BLOCK with non-overridable breach")
	}
	var compErr *domain.ErrComplianceRejected
	if !errors.As(err, &compErr) {
		t.Fatalf("expected *ErrComplianceRejected, got %T: %v", err, err)
	}
	if submitter.calls != 0 {
		t.Errorf("approval engine must NOT be called on non-releasable BLOCK; got %d calls", submitter.calls)
	}
}

func TestSubmit_ComplianceBlock_AllReleasable_CreatesComplianceReleaseApproval(t *testing.T) {
	d := newDraftDec()
	checkGroupID := uuid.New()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: checkGroupID,
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.soft_limit",
			Verdict:     contract.ComplianceVerdictBlock,
			Message:     "soft limit — can be released",
			Overridable: true,
		}},
	}}
	releaseReqID := uuid.New()
	submitter := &fakeApprovalSubmitter{result: &contract.ApprovalSubmissionResult{
		RequestID: releaseReqID,
		Status:    "PENDING_APPROVAL",
	}}
	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	h := handlerWithCompliance(tr, checker, submitter)

	result, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("all-releasable BLOCK must not return error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil decision")
	}
	if result.Status != vo.DecisionLifecyclePendingComplianceRelease {
		t.Errorf("decision status must be PENDING_COMPLIANCE_RELEASE, got %s", result.Status)
	}
	if result.ComplianceCheckGroupID == nil || *result.ComplianceCheckGroupID != checkGroupID {
		t.Errorf("ComplianceCheckGroupID must be stored; got %v", result.ComplianceCheckGroupID)
	}
	if submitter.lastIn.ProcessType != "COMPLIANCE_RELEASE" {
		t.Errorf("expected COMPLIANCE_RELEASE process type, got %q", submitter.lastIn.ProcessType)
	}
	if submitter.lastIn.SubjectType != "COMPLIANCE_RELEASE" {
		t.Errorf("expected COMPLIANCE_RELEASE subject type, got %q", submitter.lastIn.SubjectType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ApplyComplianceReleaseDecision tests
// ─────────────────────────────────────────────────────────────────────────────

func TestApplyComplianceReleaseDecision_Rejected_CancelsDecision(t *testing.T) {
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingComplianceRelease

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }

	err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, false, "compliance breach not acceptable")
	if err != nil {
		t.Fatalf("ApplyComplianceReleaseDecision (rejected): %v", err)
	}
	updated, _ := tr.GetByID(context.Background(), d.ID)
	if updated.Status != vo.DecisionLifecycleCancelled {
		t.Errorf("expected CANCELLED after release rejection, got %s", updated.Status)
	}
}

func TestApplyComplianceReleaseDecision_Approved_ResubmitsToInvestmentDecisionApproval(t *testing.T) {
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingComplianceRelease

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}

	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetApprovalSubmitter(submitter)

	err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, true, "")
	if err != nil {
		t.Fatalf("ApplyComplianceReleaseDecision (approved): %v", err)
	}
	if submitter.calls != 1 {
		t.Fatalf("expected 1 approval submission after release, got %d", submitter.calls)
	}
	if submitter.lastIn.ProcessType != "INVESTMENT_DECISION" {
		t.Errorf("re-submission must use INVESTMENT_DECISION process type, got %q", submitter.lastIn.ProcessType)
	}
	updated, _ := tr.GetByID(context.Background(), d.ID)
	if updated.Status != vo.DecisionLifecyclePendingApproval {
		t.Errorf("expected PENDING_APPROVAL after release approval, got %s", updated.Status)
	}
}

func TestApplyComplianceReleaseDecision_Approved_NoApprovalEngine_ReturnsError(t *testing.T) {
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingComplianceRelease

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	// approval engine intentionally NOT wired

	err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, true, "")
	if err == nil {
		t.Fatal("expected error when approval engine is not wired")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BLOCKER 4: ApplyComplianceReleaseDecision re-validation tests
// ─────────────────────────────────────────────────────────────────────────────

func TestApplyComplianceReleaseDecision_WrongStatus_ReturnsError(t *testing.T) {
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingApproval // wrong: not PENDING_COMPLIANCE_RELEASE

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }

	err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, true, "")
	if err == nil {
		t.Fatal("expected ErrDecisionLifecycle when decision is not in PENDING_COMPLIANCE_RELEASE")
	}
	var le *domain.ErrDecisionLifecycle
	if !errors.As(err, &le) {
		t.Fatalf("expected *ErrDecisionLifecycle, got %T: %v", err, err)
	}
}

func TestApplyComplianceReleaseDecision_WorkflowLocked_ReturnsError(t *testing.T) {
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingComplianceRelease

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	wf := &fakeWorkflowProvider{tradeAllowed: false}
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}

	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), wf, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetApprovalSubmitter(submitter)

	err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, true, "")
	if err == nil {
		t.Fatal("expected error when workflow day is locked at compliance release time")
	}
	if submitter.calls != 0 {
		t.Errorf("approval engine must NOT be called when workflow is locked; got %d calls", submitter.calls)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BLOCKER 5: Report validation before releasable-BLOCK compliance path
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmit_ReleasableBlock_InvalidReport_ReturnsError(t *testing.T) {
	// The linked research report fails the reference policy. Even though the
	// compliance check returns a releasable BLOCK, the decision must be rejected
	// because the report validation now runs before the compliance check.
	d := newDraftDec()
	rep := approvedReport()
	// Use a different contract so the report fails the contract-scope policy.
	otherContractID := uuid.New()
	rep.ApplicableContractID = &otherContractID // contract mismatch → VIOLATION_CONTRACT_MISMATCH
	d.ResearchReportID = &rep.ID

	repRepo := newFakeResearchReportRepo()
	repRepo.seed(rep)

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.soft_limit",
			Verdict:     contract.ComplianceVerdictBlock,
			Overridable: true,
		}},
	}}
	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}

	h := NewDecisionCommandHandler(nil, tr, repRepo, nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetComplianceChecker(checker)
	h.SetApprovalSubmitter(submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error when linked report fails reference policy")
	}
	var refErr *domain.ErrDecisionReferenceInvalid
	if !errors.As(err, &refErr) {
		t.Fatalf("expected *ErrDecisionReferenceInvalid, got %T: %v", err, err)
	}
	if submitter.calls != 0 {
		t.Errorf("COMPLIANCE_RELEASE approval must not be submitted when report is invalid; got %d calls", submitter.calls)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HIGH 1: Approval ID traceability
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmit_ComplianceBlock_AllReleasable_ApprovalIDStoredInCorrectField(t *testing.T) {
	// After the fix the COMPLIANCE_RELEASE request ID must be stored in
	// ComplianceReleaseApprovalRequestID, NOT in ApprovalRequestID.
	// ApprovalRequestID is reserved for the subsequent INVESTMENT_DECISION request.
	d := newDraftDec()

	releaseReqID := uuid.New()
	submitter := &fakeApprovalSubmitter{result: &contract.ApprovalSubmissionResult{
		RequestID: releaseReqID,
		Status:    "PENDING_APPROVAL",
	}}
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.soft_limit",
			Verdict:     contract.ComplianceVerdictBlock,
			Overridable: true,
		}},
	}}
	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}

	h := handlerWithCompliance(tr, checker, submitter)

	result, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("submit releasable-BLOCK: %v", err)
	}
	if result.ComplianceReleaseApprovalRequestID == nil {
		t.Fatal("ComplianceReleaseApprovalRequestID must be set after COMPLIANCE_RELEASE submission")
	}
	if *result.ComplianceReleaseApprovalRequestID != releaseReqID {
		t.Errorf("ComplianceReleaseApprovalRequestID = %v, want %v", result.ComplianceReleaseApprovalRequestID, releaseReqID)
	}
	if result.ApprovalRequestID != nil {
		t.Errorf("ApprovalRequestID must remain nil during PENDING_COMPLIANCE_RELEASE; got %v", result.ApprovalRequestID)
	}
}

func TestApplyComplianceReleaseDecision_Approved_SetsApprovalRequestIDForInvestmentDecision(t *testing.T) {
	// After compliance release is approved, the INVESTMENT_DECISION approval request
	// ID must be stored in ApprovalRequestID (not ComplianceReleaseApprovalRequestID).
	d := newDraftDec()
	d.Status = vo.DecisionLifecyclePendingComplianceRelease
	complianceReleaseID := uuid.New()
	d.ComplianceReleaseApprovalRequestID = &complianceReleaseID

	investDecisionReqID := uuid.New()
	submitter := &fakeApprovalSubmitter{result: &contract.ApprovalSubmissionResult{
		RequestID: investDecisionReqID,
		Status:    "PENDING_APPROVAL",
	}}

	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	h := NewDecisionCommandHandler(nil, tr, newFakeResearchReportRepo(), nil, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetApprovalSubmitter(submitter)

	if err := h.ApplyComplianceReleaseDecision(context.Background(), d.ID, true, ""); err != nil {
		t.Fatalf("ApplyComplianceReleaseDecision (approved): %v", err)
	}

	updated, _ := tr.GetByID(context.Background(), d.ID)
	if updated.ApprovalRequestID == nil || *updated.ApprovalRequestID != investDecisionReqID {
		t.Errorf("ApprovalRequestID must be set to the INVESTMENT_DECISION request; got %v", updated.ApprovalRequestID)
	}
	// Compliance release ID must be preserved, not overwritten.
	if updated.ComplianceReleaseApprovalRequestID == nil || *updated.ComplianceReleaseApprovalRequestID != complianceReleaseID {
		t.Errorf("ComplianceReleaseApprovalRequestID must be preserved; got %v", updated.ComplianceReleaseApprovalRequestID)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HIGH 2: Status persisted only after successful SubmitForApproval
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmit_ComplianceBlock_AllReleasable_ApprovalFails_DecisionStaysInDraft(t *testing.T) {
	d := newDraftDec()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			BreachID:    uuid.New(),
			RuleTypeID:  "concentration.soft_limit",
			Verdict:     contract.ComplianceVerdictBlock,
			Overridable: true,
		}},
	}}
	// SubmitForApproval returns an error, simulating a missing process config.
	submitter := &fakeApprovalSubmitter{err: errors.New("no active process config for COMPLIANCE_RELEASE")}
	tr := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}

	h := handlerWithCompliance(tr, checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("expected error when approval submission fails")
	}

	// Decision must remain in DRAFT — not stuck in PENDING_COMPLIANCE_RELEASE.
	current, _ := tr.GetByID(context.Background(), d.ID)
	if current.Status != vo.DecisionLifecycleDraft {
		t.Errorf("decision must remain DRAFT when approval submission fails; got %s", current.Status)
	}
}
