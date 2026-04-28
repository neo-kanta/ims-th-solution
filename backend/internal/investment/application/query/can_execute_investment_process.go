package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CanExecuteInvestmentProcessRequest is the application boundary for checking
// whether a user may execute one investment process step.
type CanExecuteInvestmentProcessRequest struct {
	UserID       uuid.UUID
	ContractID   uuid.UUID
	BusinessDate time.Time
	ProcessStep  vo.ProcessStepKey
}

type CanExecuteInvestmentProcessResult struct {
	Allowed          bool
	Reasons          []policy.BlockingReason
	Assignment       *entity.ProcessAssignmentMatch
	BlockingDecision *entity.BlockingControlDecision
	DaySetting       *entity.WorkflowDaySetting
}

type CanExecuteInvestmentProcessHandler struct {
	repo     domain.InvestmentProcessGuardRepository
	workflow contract.WorkflowStateProvider
	now      func() time.Time
}

func NewCanExecuteInvestmentProcessHandler(
	repo domain.InvestmentProcessGuardRepository,
	workflow contract.WorkflowStateProvider,
	now func() time.Time,
) *CanExecuteInvestmentProcessHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &CanExecuteInvestmentProcessHandler{
		repo:     repo,
		workflow: workflow,
		now:      now,
	}
}

// Handle answers: Can this user execute this process step for this contract/date?
func (h *CanExecuteInvestmentProcessHandler) Handle(
	ctx context.Context,
	req CanExecuteInvestmentProcessRequest,
) (*CanExecuteInvestmentProcessResult, error) {
	if h == nil || h.repo == nil || h.workflow == nil {
		return nil, fmt.Errorf("can-execute investment process handler not initialised")
	}
	if err := validateCanExecuteRequest(req); err != nil {
		return nil, err
	}

	setting, err := h.repo.GetActiveDaySetting(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("reading active workflow day setting: %w", err)
	}

	tradeAllowed, err := h.workflow.IsTradeAllowed(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("checking workflow trade allowance: %w", err)
	}

	locked, err := h.workflow.IsTransactionLocked(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("checking workflow transaction lock: %w", err)
	}

	blockingDecision, err := h.repo.FindBlockingControlDecision(
		ctx, req.ContractID, req.BusinessDate, req.ProcessStep,
	)
	if err != nil {
		return nil, fmt.Errorf("checking workflow rejection controls: %w", err)
	}

	assignment, err := h.repo.FindUserProcessAssignment(
		ctx, req.UserID, req.ContractID, req.BusinessDate, req.ProcessStep,
	)
	if err != nil {
		return nil, fmt.Errorf("checking process step assignment: %w", err)
	}

	decision := policy.CanExecuteInvestmentProcess(policy.ProcessGuardInput{
		Now:                h.now(),
		WorkflowDaySetting: setting,
		TradeAllowed:       tradeAllowed,
		TransactionLocked:  locked,
		BlockingDecision:   blockingDecision,
		Assignment:         assignment,
	})

	return &CanExecuteInvestmentProcessResult{
		Allowed:          decision.Allowed,
		Reasons:          decision.Reasons,
		Assignment:       assignment,
		BlockingDecision: blockingDecision,
		DaySetting:       setting,
	}, nil
}

func validateCanExecuteRequest(req CanExecuteInvestmentProcessRequest) error {
	if req.UserID == uuid.Nil {
		return &domain.ErrInvalidProcessGuardRequest{Field: "user_id", Detail: "is required"}
	}
	if req.ContractID == uuid.Nil {
		return &domain.ErrInvalidProcessGuardRequest{Field: "contract_id", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return &domain.ErrInvalidProcessGuardRequest{Field: "business_date", Detail: "is required"}
	}
	if !req.ProcessStep.IsValid() {
		return &domain.ErrInvalidProcessGuardRequest{Field: "process_step", Detail: "is invalid"}
	}
	return nil
}
