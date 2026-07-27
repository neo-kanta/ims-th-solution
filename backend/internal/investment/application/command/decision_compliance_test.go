package command

// Tests for the P0 compliance pre-trade gate and COMPLIANCE_RELEASE flow wired
// into DecisionCommandHandler.Submit and ApplyComplianceReleaseDecision.

import (
	"context"
	"errors"
	"strings"
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
		FundID:          func() *uuid.UUID { v := uuid.New(); return &v }(),
		PortfolioID:     uuid.New(),
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
	return handlerWithComplianceAudit(decRepo, checker, submitter, &recordingAudit{})
}

func handlerWithComplianceAudit(
	decRepo domain.DecisionRepository,
	checker contract.ComplianceChecker,
	submitter *fakeApprovalSubmitter,
	audit contract.AuditLogger,
) *DecisionCommandHandler {
	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, audit, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: vo.PortfolioTypeLive}})
	if checker != nil {
		h.SetComplianceChecker(checker)
	}
	if submitter != nil {
		h.SetApprovalSubmitter(submitter)
	}
	return h
}

func TestSubmitDecision_LiveFailsClosedOnComplianceControlStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		result     *contract.ProposedOrderResult
		assertType func(*testing.T, error)
	}{
		{
			name: "not configured cannot be mistaken for pass",
			result: &contract.ProposedOrderResult{
				CheckGroupID: uuid.New(),
				Status:       contract.ComplianceStatusNotConfigured,
				Verdict:      contract.ComplianceVerdictPass,
			},
			assertType: func(t *testing.T, err error) {
				var target *ErrComplianceNotConfigured
				if !errors.As(err, &target) {
					t.Fatalf("error = %T %v, want *ErrComplianceNotConfigured", err, err)
				}
			},
		},
		{
			name: "classification unavailable blocks even with warn verdict",
			result: &contract.ProposedOrderResult{
				CheckGroupID:   uuid.New(),
				Status:         contract.ComplianceStatusUnavailable,
				Verdict:        contract.ComplianceVerdictWarn,
				RulesEvaluated: 1,
				Breaches: []contract.ProposedOrderBreach{{
					RuleTypeID: "allocation.asset_class_min",
					Message:    "asset class classification unavailable for instruments: NEW-BOND",
				}},
			},
			assertType: func(t *testing.T, err error) {
				var target *ErrComplianceUnavailable
				if !errors.As(err, &target) {
					t.Fatalf("error = %T %v, want *ErrComplianceUnavailable", err, err)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDraftDec()
			repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
			approval := &fakeApprovalSubmitter{result: defaultApprovalResult()}
			audit := &recordingAudit{}
			h := handlerWithComplianceAudit(repo, &fakeComplianceChecker{result: tt.result}, approval, audit)

			_, err := h.Submit(context.Background(), d.ID, uuid.New())
			if err == nil {
				t.Fatal("expected LIVE compliance status to fail closed")
			}
			tt.assertType(t, err)
			if approval.calls != 0 {
				t.Fatalf("approval calls = %d, want 0", approval.calls)
			}
			if len(repo.updates) != 0 {
				t.Fatalf("decision updates = %d, want no persistence", len(repo.updates))
			}
			current, _ := repo.GetByID(context.Background(), d.ID)
			if current.Status != vo.DecisionLifecycleDraft {
				t.Fatalf("decision status = %q, want DRAFT", current.Status)
			}
			audit.mu.Lock()
			defer audit.mu.Unlock()
			if len(audit.entries) != 1 {
				t.Fatalf("audit entries = %d, want one strict control-gap event", len(audit.entries))
			}
			entry := audit.entries[0]
			if entry.Action != complianceControlGapAuditAction || entry.ResourceID != d.ID.String() {
				t.Fatalf("audit entry = %#v, want correlated decision control-gap action", entry)
			}
			details, ok := entry.Details.(map[string]any)
			if !ok {
				t.Fatalf("audit details type = %T, want map[string]any", entry.Details)
			}
			if details["check_group_id"] != tt.result.CheckGroupID.String() {
				t.Fatalf("audit check_group_id = %v, want %s", details["check_group_id"], tt.result.CheckGroupID)
			}
			if details["phase"] != "DECISION_SUBMISSION" {
				t.Fatalf("audit phase = %v, want DECISION_SUBMISSION", details["phase"])
			}
		})
	}
}

func TestSubmitDecision_NilComplianceResultFailsClosedWithGeneratedCorrelation(t *testing.T) {
	t.Parallel()
	d := newDraftDec()
	repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	approval := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	checker := &fakeComplianceChecker{}
	audit := &recordingAudit{}
	h := handlerWithComplianceAudit(repo, checker, approval, audit)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	var unavailable *ErrComplianceUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("error = %T %v, want *ErrComplianceUnavailable", err, err)
	}
	if unavailable.CheckGroupID == "" || unavailable.CheckGroupID == uuid.Nil.String() {
		t.Fatalf("nil-result check_group_id = %q, want generated non-empty correlation", unavailable.CheckGroupID)
	}
	if checker.lastIn.CheckGroupID.String() != unavailable.CheckGroupID {
		t.Fatalf("request check_group_id = %s, error correlation = %s", checker.lastIn.CheckGroupID, unavailable.CheckGroupID)
	}
	if approval.calls != 0 || len(repo.updates) != 0 {
		t.Fatalf("nil result must create no approval/decision write; approvals=%d updates=%d", approval.calls, len(repo.updates))
	}
	audit.mu.Lock()
	defer audit.mu.Unlock()
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	details := audit.entries[0].Details.(map[string]any)
	if details["check_group_id"] != unavailable.CheckGroupID {
		t.Fatalf("audit check_group_id = %v, want %s", details["check_group_id"], unavailable.CheckGroupID)
	}
}

func TestSubmitDecision_EmptyOrUnknownVerdictFailsClosed(t *testing.T) {
	t.Parallel()
	for _, verdict := range []contract.ComplianceVerdict{"", "ALLOW"} {
		verdict := verdict
		t.Run(string(verdict), func(t *testing.T) {
			t.Parallel()
			d := newDraftDec()
			repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
			approval := &fakeApprovalSubmitter{result: defaultApprovalResult()}
			groupID := uuid.New()
			audit := &recordingAudit{}
			h := handlerWithComplianceAudit(repo, &fakeComplianceChecker{result: &contract.ProposedOrderResult{
				CheckGroupID:   groupID,
				Status:         contract.ComplianceStatusEvaluated,
				Verdict:        verdict,
				RulesEvaluated: 1,
			}}, approval, audit)

			_, err := h.Submit(context.Background(), d.ID, uuid.New())
			var unavailable *ErrComplianceUnavailable
			if !errors.As(err, &unavailable) {
				t.Fatalf("verdict %q error = %T %v, want *ErrComplianceUnavailable", verdict, err, err)
			}
			if unavailable.CheckGroupID != groupID.String() {
				t.Fatalf("check_group_id = %q, want %s", unavailable.CheckGroupID, groupID)
			}
			if approval.calls != 0 || len(repo.updates) != 0 {
				t.Fatalf("unsupported verdict must create no approval/decision write; approvals=%d updates=%d", approval.calls, len(repo.updates))
			}
			if len(audit.entries) != 1 {
				t.Fatalf("audit entries = %d, want 1", len(audit.entries))
			}
		})
	}
}

func TestSubmitDecision_ControlGapAuditFailureCannotSucceed(t *testing.T) {
	t.Parallel()
	d := newDraftDec()
	repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	approval := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithComplianceAudit(repo, &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}}, approval, &failingAudit{strictErr: errors.New("audit store down")})

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil || !strings.Contains(err.Error(), "audit store down") {
		t.Fatalf("audit failure error = %v, want surfaced strict audit failure", err)
	}
	if approval.calls != 0 || len(repo.updates) != 0 {
		t.Fatalf("audit failure must create no approval/decision write; approvals=%d updates=%d", approval.calls, len(repo.updates))
	}
}

func TestSubmitDecision_EmptyAuthoritativePortfolioTypeFailsClosed(t *testing.T) {
	t.Parallel()
	d := newDraftDec()
	repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}}
	h := handlerWithCompliance(repo, checker, &fakeApprovalSubmitter{result: defaultApprovalResult()})
	h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: ""}})

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil || !strings.Contains(err.Error(), "unsupported portfolio type") {
		t.Fatalf("empty authoritative portfolio type error = %v, want fail-closed error", err)
	}
	if checker.calls != 0 || len(repo.updates) != 0 {
		t.Fatalf("empty authoritative type must fail before compliance/write; checker=%d updates=%d", checker.calls, len(repo.updates))
	}
}

func TestSubmitDecision_NonLivePortfolioPolicyUnchanged(t *testing.T) {
	t.Parallel()
	for _, portfolioType := range []vo.PortfolioType{vo.PortfolioTypeSimulation, vo.PortfolioTypeModel} {
		portfolioType := portfolioType
		t.Run(string(portfolioType), func(t *testing.T) {
			t.Parallel()
			cases := []struct {
				name        string
				result      *contract.ProposedOrderResult
				wantBlocked bool
			}{
				{
					name: "no active effective bindings",
					result: &contract.ProposedOrderResult{
						CheckGroupID: uuid.New(),
						Status:       contract.ComplianceStatusNotConfigured,
						Verdict:      contract.ComplianceVerdictPass,
					},
				},
				{
					name: "missing asset classification preserves prior pass evaluation",
					result: &contract.ProposedOrderResult{
						CheckGroupID:   uuid.New(),
						Status:         contract.ComplianceStatusEvaluated,
						Verdict:        contract.ComplianceVerdictPass,
						RulesEvaluated: 1,
					},
				},
				{
					name: "valid classified allocation breach still blocks",
					result: &contract.ProposedOrderResult{
						CheckGroupID:   uuid.New(),
						Status:         contract.ComplianceStatusEvaluated,
						Verdict:        contract.ComplianceVerdictBlock,
						RulesEvaluated: 1,
						Breaches: []contract.ProposedOrderBreach{{
							RuleTypeID:  "allocation.asset_class_max",
							Message:     "classified equity exposure exceeds the configured maximum",
							Overridable: false,
						}},
					},
					wantBlocked: true,
				},
			}
			for _, tc := range cases {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					d := newDraftDec()
					repo := &trackingDecisionRepo{fakeDecisionRepo: newFakeDecisionRepo(d)}
					approval := &fakeApprovalSubmitter{result: defaultApprovalResult()}
					h := handlerWithCompliance(repo, &fakeComplianceChecker{result: tc.result}, approval)
					h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: portfolioType}})

					result, err := h.Submit(context.Background(), d.ID, uuid.New())
					if tc.wantBlocked {
						var blocked *domain.ErrComplianceRejected
						if !errors.As(err, &blocked) {
							t.Fatalf("%s valid-rule error = %T %v, want existing ErrComplianceRejected", portfolioType, err, err)
						}
						var unavailable *ErrComplianceUnavailable
						if errors.As(err, &unavailable) {
							t.Fatalf("%s must not apply the LIVE-only typed status gate", portfolioType)
						}
						if approval.calls != 0 {
							t.Fatalf("approval calls = %d, want 0 on BLOCK", approval.calls)
						}
						return
					}
					if err != nil {
						t.Fatalf("%s no-rules policy changed unexpectedly: %v", portfolioType, err)
					}
					if result.Status != vo.DecisionLifecyclePendingApproval || approval.calls != 1 {
						t.Fatalf("no-rules result status=%q approvals=%d, want PENDING_APPROVAL/1", result.Status, approval.calls)
					}
				})
			}
		})
	}
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
		ID:                               *d.FundID,
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
		ID:                               *d.FundID,
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
		ID:                               *d.FundID,
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
		Status:         contract.ComplianceStatusEvaluated,
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
	if d.LimitPrice == nil || !checker.lastIn.Price.Equal(*d.LimitPrice) {
		t.Fatalf("quantity-only compliance price = %s, want decision limit price %v", checker.lastIn.Price, d.LimitPrice)
	}
	if submitter.calls != 1 {
		t.Fatalf("expected 1 approval submission on PASS, got %d", submitter.calls)
	}
}

func TestSubmit_ComplianceUsesAmountAsProposedNotional(t *testing.T) {
	d := newDraftDec()
	quantity := decimal.NewFromInt(1)
	amount := decimal.NewFromInt(11)
	d.Quantity = &quantity
	d.Amount = &amount
	d.LimitPrice = nil

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 1,
	}}
	decRepo := newFakeDecisionRepo(d)
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(decRepo, checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err != nil {
		t.Fatalf("quantity + amount without limit price must submit on PASS: %v", err)
	}
	if !checker.lastIn.Quantity.Equal(quantity) {
		t.Fatalf("compliance quantity = %s, want %s", checker.lastIn.Quantity, quantity)
	}
	wantPrice := amount.Div(quantity)
	if !checker.lastIn.Price.Equal(wantPrice) {
		t.Fatalf("compliance price = %s, want amount/quantity = %s", checker.lastIn.Price, wantPrice)
	}
	if !checker.lastIn.Quantity.Mul(checker.lastIn.Price).Equal(amount) {
		t.Fatalf("compliance notional = %s, want entered amount %s", checker.lastIn.Quantity.Mul(checker.lastIn.Price), amount)
	}
}

func TestSubmit_ComplianceExplicitAmountOverridesLimitPrice(t *testing.T) {
	d := newDraftDec()
	quantity := decimal.NewFromInt(40)
	amount := decimal.NewFromInt(4_000)
	limitPrice := decimal.NewFromInt(35)
	d.Quantity = &quantity
	d.Amount = &amount
	d.LimitPrice = &limitPrice

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 1,
	}}
	h := handlerWithCompliance(
		newFakeDecisionRepo(d),
		checker,
		&fakeApprovalSubmitter{result: defaultApprovalResult()},
	)

	if _, err := h.Submit(context.Background(), d.ID, uuid.New()); err != nil {
		t.Fatalf("quantity + amount with limit price must submit on PASS: %v", err)
	}
	wantPrice := amount.Div(quantity)
	if !checker.lastIn.Price.Equal(wantPrice) {
		t.Fatalf("compliance price = %s, want amount/quantity = %s", checker.lastIn.Price, wantPrice)
	}
	if checker.lastIn.Price.Equal(limitPrice) {
		t.Fatalf("compliance must not reuse limit price %s when amount is explicit", limitPrice)
	}
}

func TestSubmit_ComplianceAmountOnlyFailsClosedBeforeChecker(t *testing.T) {
	d := newDraftDec()
	amount := decimal.NewFromInt(11)
	d.Quantity = nil
	d.Amount = &amount
	d.LimitPrice = nil

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictPass,
	}}
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(newFakeDecisionRepo(d), checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("amount-only decision must fail closed before compliance evaluation")
	}
	var invalid *domain.ErrInvalidDecisionRequest
	if !errors.As(err, &invalid) || invalid.Field != "quantity" {
		t.Fatalf("expected quantity ErrInvalidDecisionRequest, got %T: %v", err, err)
	}
	if checker.calls != 0 {
		t.Fatalf("invalid amount-only decision must not reach compliance; got %d calls", checker.calls)
	}
	if submitter.calls != 0 {
		t.Fatalf("invalid amount-only decision must not reach approval; got %d calls", submitter.calls)
	}
}

func TestSubmit_ComplianceQuantityWithoutAmountOrLimitFailsClosedBeforeChecker(t *testing.T) {
	d := newDraftDec()
	d.Amount = nil
	d.LimitPrice = nil

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictPass,
	}}
	submitter := &fakeApprovalSubmitter{result: defaultApprovalResult()}
	h := handlerWithCompliance(newFakeDecisionRepo(d), checker, submitter)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("quantity-only decision without limit price must fail closed")
	}
	var invalid *domain.ErrInvalidDecisionRequest
	if !errors.As(err, &invalid) || invalid.Field != "limit_price" {
		t.Fatalf("expected limit_price ErrInvalidDecisionRequest, got %T: %v", err, err)
	}
	if checker.calls != 0 {
		t.Fatalf("invalid quantity-only decision must not reach compliance; got %d calls", checker.calls)
	}
	if submitter.calls != 0 {
		t.Fatalf("invalid quantity-only decision must not reach approval; got %d calls", submitter.calls)
	}
}

func TestSubmit_ComplianceWarn_ContinuesToApproval(t *testing.T) {
	d := newDraftDec()

	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictWarn,
		RulesEvaluated: 1,
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
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
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
		CheckGroupID:   checkGroupID,
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
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
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
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
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
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

func TestValidateCreateDecisionRejectsNonPositiveNumericFields(t *testing.T) {
	positive := decimal.NewFromInt(1)
	base := func() CreateDecisionRequest {
		return CreateDecisionRequest{
			ActorID:        uuid.New(),
			FundID:         func() *uuid.UUID { v := uuid.New(); return &v }(),
			PortfolioID:    uuid.New(),
			BusinessDate:   time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
			Side:           vo.OrderSideBuy,
			InstrumentCode: "AOT",
			Currency:       "THB",
			Quantity:       &positive,
		}
	}

	tests := []struct {
		name  string
		field string
		set   func(*CreateDecisionRequest)
	}{
		{name: "zero quantity", field: "quantity", set: func(req *CreateDecisionRequest) {
			zero := decimal.Zero
			req.Quantity = &zero
		}},
		{name: "negative amount", field: "amount", set: func(req *CreateDecisionRequest) {
			negative := decimal.NewFromInt(-1)
			req.Amount = &negative
		}},
		{name: "zero limit price", field: "limit_price", set: func(req *CreateDecisionRequest) {
			zero := decimal.Zero
			req.LimitPrice = &zero
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base()
			tt.set(&req)
			err := validateCreateDecision(req)
			var invalid *domain.ErrInvalidDecisionRequest
			if !errors.As(err, &invalid) || invalid.Field != tt.field {
				t.Fatalf("expected %s ErrInvalidDecisionRequest, got %T: %v", tt.field, err, err)
			}
		})
	}
}
