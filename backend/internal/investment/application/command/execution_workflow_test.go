package command

// Tests for ExecutionCommandHandler.Create with the workflow transaction-lock gate.

import (
	"context"
	"errors"
	"strings"
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
// fakes
// ─────────────────────────────────────────────────────────────────────────────

// fakeExecutionRepo satisfies domain.ExecutionRepository.
type fakeExecutionRepo struct {
	created []*entity.Execution
}

func (r *fakeExecutionRepo) Create(_ context.Context, _ pgx.Tx, e *entity.Execution) error {
	cp := *e
	r.created = append(r.created, &cp)
	return nil
}
func (r *fakeExecutionRepo) Update(_ context.Context, _ pgx.Tx, _ *entity.Execution) error {
	return nil
}
func (r *fakeExecutionRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Execution, error) {
	return nil, nil
}
func (r *fakeExecutionRepo) ListByDecision(_ context.Context, _ uuid.UUID) ([]*entity.Execution, error) {
	return nil, nil
}
func (r *fakeExecutionRepo) ListByPortfolio(_ context.Context, _ uuid.UUID, _ domain.ExecutionListFilter) ([]*entity.Execution, int, error) {
	return nil, 0, nil
}
func (r *fakeExecutionRepo) ListByFundDate(_ context.Context, _ uuid.UUID, _ time.Time) ([]*entity.Execution, error) {
	return nil, nil
}

// fakeWorkflowProvider returns configurable responses for workflow gate calls.
type fakeWorkflowProvider struct {
	tradeAllowed bool
	txLocked     bool
	err          error
}

func (f *fakeWorkflowProvider) IsTradeAllowed(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.tradeAllowed, f.err
}
func (f *fakeWorkflowProvider) IsTransactionLocked(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return f.txLocked, f.err
}

var _ contract.WorkflowStateProvider = (*fakeWorkflowProvider)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func approvedDecision() *entity.Decision {
	qty := decimal.NewFromFloat(100)
	return &entity.Decision{
		ID:             uuid.New(),
		FundID:         uuid.New(),
		PortfolioID:    uuid.New(),
		InstrumentCode: "PTT",
		Side:           vo.OrderSideBuy,
		Status:         vo.DecisionLifecycleApproved,
		DecisionNumber: "DEC-2026-001",
		BusinessDate:   time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC),
		Quantity:       &qty,
		Currency:       "THB",
		CreatedBy:      uuid.New(),
		CreatedAt:      time.Now().UTC(),
	}
}

func buildExecutionHandler(d *entity.Decision, execRepo *fakeExecutionRepo) *ExecutionCommandHandler {
	return buildExecutionHandlerWithAudit(d, execRepo, &recordingAudit{})
}

func buildExecutionHandlerWithAudit(
	d *entity.Decision,
	execRepo *fakeExecutionRepo,
	audit contract.AuditLogger,
) *ExecutionCommandHandler {
	decRepo := newFakeDecisionRepo(d)
	h := NewExecutionCommandHandler(nil, decRepo, execRepo, audit, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: vo.PortfolioTypeLive}})
	return h
}

// ─────────────────────────────────────────────────────────────────────────────
// tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateExecution_WorkflowNil_Proceeds(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	// no workflow port wired

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("nil workflow: execution must proceed: %v", err)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_WorkflowNotLocked_Proceeds(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	h.SetWorkflowStateProvider(&fakeWorkflowProvider{txLocked: false})

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("unlocked workflow: execution must proceed: %v", err)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_WorkflowLocked_Blocked(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	h.SetWorkflowStateProvider(&fakeWorkflowProvider{txLocked: true})

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error when workflow transaction is locked")
	}
	var le *domain.ErrDecisionLifecycle
	if !errors.As(err, &le) {
		t.Fatalf("expected *ErrDecisionLifecycle on locked workflow, got %T: %v", err, err)
	}
	if len(execRepo.created) != 0 {
		t.Errorf("no execution must be persisted when workflow is locked; got %d", len(execRepo.created))
	}
}

func TestCreateExecution_ComplianceBlock_Blocked(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictBlock,
		RulesEvaluated: 1,
		Breaches: []contract.ProposedOrderBreach{
			{RuleTypeID: "MAX_EXPOSURE", Message: "exceeds limit"},
		},
	}}
	h.SetComplianceChecker(checker)

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error when compliance returns BLOCK")
	}
	var ce *domain.ErrComplianceRejected
	if !errors.As(err, &ce) {
		t.Fatalf("expected *ErrComplianceRejected, got %T: %v", err, err)
	}
	if len(execRepo.created) != 0 {
		t.Errorf("no execution must be created on BLOCK; got %d", len(execRepo.created))
	}
	if checker.calls != 1 {
		t.Fatalf("expected compliance checker to be called once, got %d", checker.calls)
	}
	if checker.lastIn.PortfolioID != d.PortfolioID {
		t.Errorf("expected compliance request portfolio_id %s, got %s", d.PortfolioID, checker.lastIn.PortfolioID)
	}
	if checker.lastIn.ContractID != d.FundID {
		t.Errorf("expected compliance request contract_id (legacy fund_id) %s, got %s", d.FundID, checker.lastIn.ContractID)
	}
}

func TestCreateExecution_LiveFailsClosedOnComplianceControlStatus(t *testing.T) {
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
			d := approvedDecision()
			repo := &fakeExecutionRepo{}
			audit := &recordingAudit{}
			h := buildExecutionHandlerWithAudit(d, repo, audit)
			h.SetComplianceChecker(&fakeComplianceChecker{result: tt.result})

			_, err := h.Create(context.Background(), CreateExecutionRequest{
				DecisionID: d.ID,
				ActorID:    uuid.New(),
			})
			if err == nil {
				t.Fatal("expected LIVE compliance status to fail closed")
			}
			tt.assertType(t, err)
			if len(repo.created) != 0 {
				t.Fatalf("execution rows = %d, want 0", len(repo.created))
			}
			audit.mu.Lock()
			defer audit.mu.Unlock()
			if len(audit.entries) != 1 {
				t.Fatalf("audit entries = %d, want one strict control-gap event", len(audit.entries))
			}
			entry := audit.entries[0]
			details, ok := entry.Details.(map[string]any)
			if entry.Action != complianceControlGapAuditAction || !ok {
				t.Fatalf("audit entry = %#v, want control-gap event with details", entry)
			}
			if details["check_group_id"] != tt.result.CheckGroupID.String() || details["phase"] != "EXECUTION_CREATION" {
				t.Fatalf("audit correlation details = %#v", details)
			}
		})
	}
}

func TestCreateExecution_NonLivePortfolioPolicyUnchanged(t *testing.T) {
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
							RuleTypeID: "allocation.asset_class_max",
							Message:    "classified equity exposure exceeds the configured maximum",
						}},
					},
					wantBlocked: true,
				},
			}
			for _, tc := range cases {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					d := approvedDecision()
					repo := &fakeExecutionRepo{}
					h := buildExecutionHandler(d, repo)
					h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: portfolioType}})
					h.SetComplianceChecker(&fakeComplianceChecker{result: tc.result})

					_, err := h.Create(context.Background(), CreateExecutionRequest{DecisionID: d.ID, ActorID: uuid.New()})
					if tc.wantBlocked {
						var blocked *domain.ErrComplianceRejected
						if !errors.As(err, &blocked) {
							t.Fatalf("%s valid-rule error = %T %v, want existing ErrComplianceRejected", portfolioType, err, err)
						}
						var unavailable *ErrComplianceUnavailable
						if errors.As(err, &unavailable) {
							t.Fatalf("%s must not apply the LIVE-only typed status gate", portfolioType)
						}
						if len(repo.created) != 0 {
							t.Fatalf("execution rows = %d, want 0 on BLOCK", len(repo.created))
						}
						return
					}
					if err != nil || len(repo.created) != 1 {
						t.Fatalf("%s no-rules result error=%v executions=%d, want success/1", portfolioType, err, len(repo.created))
					}
				})
			}
		})
	}
}

func TestCreateExecution_CompliancePass_Proceeds(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 1,
	}}
	h.SetComplianceChecker(checker)

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("PASS verdict: execution must proceed: %v", err)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_ComplianceWarn_Proceeds(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictWarn,
		RulesEvaluated: 1,
	}}
	h.SetComplianceChecker(checker)

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("WARN verdict: execution must proceed: %v", err)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_ComplianceUsesActualOrderedAmount(t *testing.T) {
	d := approvedDecision()
	decisionLimitPrice := decimal.NewFromInt(35)
	d.LimitPrice = &decisionLimitPrice
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 1,
	}}
	h.SetComplianceChecker(checker)

	orderedQuantity := decimal.NewFromInt(40)
	orderedAmount := decimal.NewFromInt(4_000)
	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID:      d.ID,
		ActorID:         uuid.New(),
		OrderedQuantity: &orderedQuantity,
		OrderedAmount:   &orderedAmount,
	})
	if err != nil {
		t.Fatalf("execution with explicit notional must proceed on PASS: %v", err)
	}
	if !checker.lastIn.Quantity.Equal(orderedQuantity) {
		t.Fatalf("compliance quantity = %s, want %s", checker.lastIn.Quantity, orderedQuantity)
	}
	wantEffectivePrice := orderedAmount.Div(orderedQuantity)
	if !checker.lastIn.Price.Equal(wantEffectivePrice) {
		t.Fatalf("compliance price = %s, want amount/quantity = %s", checker.lastIn.Price, wantEffectivePrice)
	}
	if checker.lastIn.Price.Equal(decisionLimitPrice) {
		t.Fatalf("compliance must not reuse decision limit price %s when ordered_amount is explicit", decisionLimitPrice)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_ComplianceNilResult_Blocked(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	audit := &recordingAudit{}
	h := buildExecutionHandlerWithAudit(d, execRepo, audit)
	checker := &fakeComplianceChecker{}
	h.SetComplianceChecker(checker)

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	var unavailable *ErrComplianceUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("error = %T %v, want *ErrComplianceUnavailable", err, err)
	}
	if unavailable.CheckGroupID == "" || unavailable.CheckGroupID == uuid.Nil.String() {
		t.Fatalf("nil-result check_group_id = %q, want generated correlation", unavailable.CheckGroupID)
	}
	if checker.lastIn.CheckGroupID.String() != unavailable.CheckGroupID {
		t.Fatalf("request check_group_id = %s, error correlation = %s", checker.lastIn.CheckGroupID, unavailable.CheckGroupID)
	}
	if len(execRepo.created) != 0 {
		t.Fatalf("nil compliance result must create no execution; got %d", len(execRepo.created))
	}
	if len(audit.entries) != 1 || audit.entries[0].Details.(map[string]any)["check_group_id"] != unavailable.CheckGroupID {
		t.Fatalf("nil-result audit correlation = %#v, want %s", audit.entries, unavailable.CheckGroupID)
	}
}

func TestCreateExecution_ComplianceUnknownVerdict_Blocked(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	h.SetComplianceChecker(&fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Status:         contract.ComplianceStatusEvaluated,
		Verdict:        contract.ComplianceVerdict("REVIEW"),
		RulesEvaluated: 1,
	}})

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	var unavailable *ErrComplianceUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("error = %T %v, want *ErrComplianceUnavailable", err, err)
	}
	if len(execRepo.created) != 0 {
		t.Fatalf("unsupported verdict must create no execution; got %d", len(execRepo.created))
	}
}

func TestCreateExecution_ControlGapAuditFailureCannotSucceed(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandlerWithAudit(d, execRepo, &failingAudit{strictErr: errors.New("audit store down")})
	h.SetComplianceChecker(&fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}})

	_, err := h.Create(context.Background(), CreateExecutionRequest{DecisionID: d.ID, ActorID: uuid.New()})
	if err == nil || !strings.Contains(err.Error(), "audit store down") {
		t.Fatalf("audit failure error = %v, want surfaced strict audit failure", err)
	}
	if len(execRepo.created) != 0 {
		t.Fatalf("audit failure must create no execution; got %d", len(execRepo.created))
	}
}

func TestCreateExecution_EmptyAuthoritativePortfolioTypeFailsClosed(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	checker := &fakeComplianceChecker{result: &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Status:       contract.ComplianceStatusNotConfigured,
		Verdict:      contract.ComplianceVerdictPass,
	}}
	h := buildExecutionHandler(d, execRepo)
	h.SetPortfolioRepository(postPortfolioRepo{portfolio: &entity.Portfolio{PortfolioType: ""}})
	h.SetComplianceChecker(checker)

	_, err := h.Create(context.Background(), CreateExecutionRequest{DecisionID: d.ID, ActorID: uuid.New()})
	if err == nil || !strings.Contains(err.Error(), "unsupported portfolio type") {
		t.Fatalf("empty authoritative portfolio type error = %v, want fail-closed error", err)
	}
	if checker.calls != 0 || len(execRepo.created) != 0 {
		t.Fatalf("empty authoritative type must fail before compliance/write; checker=%d executions=%d", checker.calls, len(execRepo.created))
	}
}

func TestCreateExecution_ComplianceNil_Proceeds(t *testing.T) {
	d := approvedDecision()
	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	// no compliance checker wired

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("nil compliance checker: execution must proceed: %v", err)
	}
	if len(execRepo.created) != 1 {
		t.Fatalf("expected 1 execution created, got %d", len(execRepo.created))
	}
}

func TestCreateExecution_DecisionNotApproved_Blocked(t *testing.T) {
	d := approvedDecision()
	d.Status = vo.DecisionLifecyclePendingApproval // not yet approved

	execRepo := &fakeExecutionRepo{}
	h := buildExecutionHandler(d, execRepo)
	h.SetWorkflowStateProvider(&fakeWorkflowProvider{txLocked: false})

	_, err := h.Create(context.Background(), CreateExecutionRequest{
		DecisionID: d.ID,
		ActorID:    uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for non-approved decision")
	}
	if len(execRepo.created) != 0 {
		t.Errorf("no execution must be created for non-approved decision; got %d", len(execRepo.created))
	}
}
