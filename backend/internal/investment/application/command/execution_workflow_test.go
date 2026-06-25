package command

// Tests for ExecutionCommandHandler.Create with the workflow transaction-lock gate.

import (
	"context"
	"errors"
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
func (r *fakeExecutionRepo) ListByContractDate(_ context.Context, _ uuid.UUID, _ time.Time) ([]*entity.Execution, error) {
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
		ContractID:     uuid.New(),
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
	decRepo := newFakeDecisionRepo(d)
	h := NewExecutionCommandHandler(nil, decRepo, execRepo, &recordingAudit{}, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
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
