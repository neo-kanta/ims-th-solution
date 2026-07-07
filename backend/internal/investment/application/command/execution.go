package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CreateExecutionRequest opens a new execution for an APPROVED decision.
type CreateExecutionRequest struct {
	DecisionID      uuid.UUID
	OrderedQuantity *decimal.Decimal
	OrderedAmount   *decimal.Decimal
	BrokerReference string
	ActorID         uuid.UUID
}

// FillExecutionRequest records a fill against an existing execution.
type FillExecutionRequest struct {
	ExecutionID      uuid.UUID
	ExecutedQuantity *decimal.Decimal
	ExecutedAmount   *decimal.Decimal
	ExecutionPrice   *decimal.Decimal
	Status           vo.ExecutionStatus // EXECUTED or PARTIALLY_EXECUTED
	BrokerReference  string
	ActorID          uuid.UUID
}

// CancelExecutionRequest cancels an in-flight execution.
type CancelExecutionRequest struct {
	ExecutionID uuid.UUID
	Reason      string
	ActorID     uuid.UUID
}

type ExecutionCommandHandler struct {
	pool       *pgxpool.Pool
	decisions  domain.DecisionRepository
	executions domain.ExecutionRepository
	audit      contract.AuditLogger
	now        func() time.Time
	runTx      func(ctx context.Context, fn func(pgx.Tx) error) error
	workflow   contract.WorkflowStateProvider
}

// SetWorkflowStateProvider injects the workflow state port post-construction.
// When wired, Create() refuses to open an execution while the trading day is
// transaction-locked (e.g. during manager approval or EOD processing).
func (h *ExecutionCommandHandler) SetWorkflowStateProvider(w contract.WorkflowStateProvider) {
	if h != nil {
		h.workflow = w
	}
}

func NewExecutionCommandHandler(
	pool *pgxpool.Pool,
	decisions domain.DecisionRepository,
	executions domain.ExecutionRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *ExecutionCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	h := &ExecutionCommandHandler{
		pool:       pool,
		decisions:  decisions,
		executions: executions,
		audit:      audit,
		now:        now,
	}
	h.runTx = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return withTransaction(ctx, pool, fn)
	}
	return h
}

func (h *ExecutionCommandHandler) Create(ctx context.Context, req CreateExecutionRequest) (*entity.Execution, error) {
	if req.DecisionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "decision_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	d, err := h.decisions.GetByID(ctx, req.DecisionID)
	if err != nil {
		return nil, fmt.Errorf("loading decision: %w", err)
	}
	if d == nil {
		return nil, &domain.ErrDecisionNotFound{DecisionID: req.DecisionID.String()}
	}
	if d.Status != vo.DecisionLifecycleApproved && d.Status != vo.DecisionLifecycleReadyForExecution {
		return nil, &domain.ErrDecisionLifecycle{
			DecisionID:    d.ID.String(),
			CurrentStatus: string(d.Status),
			Detail:        "decision must be APPROVED before an execution can be opened",
		}
	}
	// Workflow lock gate — refuse to open an execution while the trading day is
	// locked (e.g. manager approval lock or EOD). Uses IsTransactionLocked, not
	// IsTradeAllowed, because execution is an operational act against an already-
	// approved decision, not a new trade submission.
	if h.workflow != nil {
		locked, err := h.workflow.IsTransactionLocked(ctx, d.FundID, d.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("checking workflow transaction lock: %w", err)
		}
		if locked {
			return nil, &domain.ErrDecisionLifecycle{
				DecisionID:    d.ID.String(),
				CurrentStatus: string(d.Status),
				Detail:        "workflow transaction lock is active — execution cannot be opened",
			}
		}
	}
	now := h.now()
	ordQty := req.OrderedQuantity
	if ordQty == nil {
		ordQty = d.Quantity
	}
	ordAmt := req.OrderedAmount
	if ordAmt == nil {
		ordAmt = d.Amount
	}
	e := &entity.Execution{
		ID:              uuid.New(),
		DecisionID:      d.ID,
		FundID:          d.FundID,
		PortfolioID:     d.PortfolioID,
		InstrumentID:    d.InstrumentID,
		InstrumentCode:  d.InstrumentCode,
		BusinessDate:    d.BusinessDate,
		Side:            d.Side,
		OrderedQuantity: ordQty,
		OrderedAmount:   ordAmt,
		Currency:        d.Currency,
		Status:          vo.ExecutionStatusPending,
		BrokerReference: strings.TrimSpace(req.BrokerReference),
		CreatedAt:       now,
		CreatedBy:       req.ActorID,
		UpdatedAt:       now,
		UpdatedBy:       req.ActorID,
	}
	err = h.runTx(ctx, func(tx pgx.Tx) error {
		if err := h.executions.Create(ctx, tx, e); err != nil {
			return err
		}
		// Mark the decision READY_FOR_EXECUTION on first execution open.
		if d.Status == vo.DecisionLifecycleApproved {
			d.Status = vo.DecisionLifecycleReadyForExecution
			d.UpdatedAt = now
			d.UpdatedBy = req.ActorID
			d.ReadyForExecutionAt = &now
			return h.decisions.Update(ctx, tx, d)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_EXECUTION_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_EXECUTION",
		ResourceID:   e.ID.String(),
		Details: map[string]any{
			"decision_id":    d.ID.String(),
			"instrument":     e.InstrumentCode,
			"side":           string(e.Side),
			"ordered_qty":    e.OrderedQuantity,
			"ordered_amount": e.OrderedAmount,
		},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing execution create: %w", err)
	}
	return e, nil
}

func (h *ExecutionCommandHandler) Fill(ctx context.Context, req FillExecutionRequest) (*entity.Execution, error) {
	if req.ExecutionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "execution_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if req.Status != "" && !req.Status.IsValid() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "status", Detail: "invalid execution status"}
	}
	e, err := h.executions.GetByID(ctx, req.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("loading execution: %w", err)
	}
	if e == nil {
		return nil, &domain.ErrExecutionNotFound{ExecutionID: req.ExecutionID.String()}
	}
	if !e.CanUpdate() {
		return nil, &domain.ErrExecutionLifecycle{
			ExecutionID:   e.ID.String(),
			CurrentStatus: string(e.Status),
		}
	}
	now := h.now()
	if req.ExecutedQuantity != nil {
		e.ExecutedQuantity = req.ExecutedQuantity
	}
	if req.ExecutedAmount != nil {
		e.ExecutedAmount = req.ExecutedAmount
	}
	if req.ExecutionPrice != nil {
		e.ExecutionPrice = req.ExecutionPrice
	}
	if req.BrokerReference != "" {
		e.BrokerReference = strings.TrimSpace(req.BrokerReference)
	}
	target := req.Status
	if target == "" {
		target = vo.ExecutionStatusExecuted
	}
	e.Status = target
	e.ExecutedAt = &now
	e.UpdatedAt = now
	e.UpdatedBy = req.ActorID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.executions.Update(ctx, tx, e)
	}); err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_EXECUTION_FILLED",
		Module:       "investment",
		ResourceType: "INVESTMENT_EXECUTION",
		ResourceID:   e.ID.String(),
		Details: map[string]any{
			"status":          string(e.Status),
			"executed_qty":    e.ExecutedQuantity,
			"executed_amount": e.ExecutedAmount,
			"executed_price":  e.ExecutionPrice,
		},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing execution fill: %w", err)
	}
	return e, nil
}

func (h *ExecutionCommandHandler) Cancel(ctx context.Context, req CancelExecutionRequest) (*entity.Execution, error) {
	if req.ExecutionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "execution_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "reason", Detail: "is required"}
	}
	e, err := h.executions.GetByID(ctx, req.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("loading execution: %w", err)
	}
	if e == nil {
		return nil, &domain.ErrExecutionNotFound{ExecutionID: req.ExecutionID.String()}
	}
	if e.IsTerminal() {
		return nil, &domain.ErrExecutionLifecycle{
			ExecutionID:   e.ID.String(),
			CurrentStatus: string(e.Status),
			Detail:        "execution already terminal",
		}
	}
	now := h.now()
	e.Status = vo.ExecutionStatusCancelled
	e.CancelledAt = &now
	e.CancelledBy = &req.ActorID
	e.CancellationReason = reason
	e.UpdatedAt = now
	e.UpdatedBy = req.ActorID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.executions.Update(ctx, tx, e)
	}); err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_EXECUTION_CANCELLED",
		Module:       "investment",
		ResourceType: "INVESTMENT_EXECUTION",
		ResourceID:   e.ID.String(),
		Details:      map[string]any{"reason": reason},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing execution cancel: %w", err)
	}
	return e, nil
}
