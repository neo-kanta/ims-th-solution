package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
)

type CancelDayStartRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
	Reason       string
}

type CancelDayStartResult struct {
	TransitionID  uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time
	FromState     vo.WorkflowState
	ToState       vo.WorkflowState
	OccurredAt    time.Time
}

type CancelDayStartHandler struct {
	pool            *pgxpool.Pool
	dayRepo         domain.WorkflowDayRepository
	logRepo         domain.TransitionLogRepository
	investmentQuery ports.InvestmentQueryPort
	policy          *policy.TransitionPolicy
}

func NewCancelDayStartHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	investmentQuery ports.InvestmentQueryPort,
	pol *policy.TransitionPolicy,
) *CancelDayStartHandler {
	return &CancelDayStartHandler{
		pool:            pool,
		dayRepo:         dayRepo,
		logRepo:         logRepo,
		investmentQuery: investmentQuery,
		policy:          pol,
	}
}

// Handle reverts a DAY_OPEN row back to NOT_STARTED.
//
// Storage decision: NOT_STARTED is documented as a synthetic state never stored
// in workflow__day_states, but the schema CHECK constraint allows it. Hard-deleting
// the day row would orphan the FK from workflow__transition_log (NOT NULL REFERENCES),
// destroying the audit trail. We therefore keep the row, set current_state back
// to NOT_STARTED, and null out the opened_* stamps so subsequent reads (and
// IsTradeAllowed) treat it as never-opened. This is the only intentional
// exception to the "NOT_STARTED is never persisted" invariant.
func (h *CancelDayStartHandler) Handle(ctx context.Context, req CancelDayStartRequest) (*CancelDayStartResult, error) {
	txSummary, err := h.investmentQuery.GetTransactionSummary(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("fetching transaction summary: %w", err)
	}

	var result *CancelDayStartResult
	txErr := database.WithTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		day, err := h.dayRepo.GetForUpdateByBusinessDate(ctx, tx, req.BusinessDate)
		if err != nil {
			return fmt.Errorf("locking workflow day: %w", err)
		}
		if day == nil {
			return &domain.ErrNotFound{
				ContractID:   req.ContractID.String(),
				BusinessDate: req.BusinessDate.Format("2006-01-02"),
			}
		}

		if err := h.policy.CanCancelDayStart(policy.CancelDayStartInput{
			CurrentDay:       day,
			Reason:           req.Reason,
			TransactionCount: txSummary.Count,
		}); err != nil {
			return err
		}

		now := time.Now().UTC()
		fromState := day.CurrentState
		day.CurrentState = vo.StateNotStarted
		day.OpenedAt = nil
		day.OpenedBy = nil
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID

		if err := h.dayRepo.ResetToNotStarted(ctx, tx, day); err != nil {
			return err
		}

		reason := req.Reason
		transition := &entity.WorkflowTransition{
			ID:               uuid.New(),
			WorkflowDayID:    day.ID,
			ContractID:       req.ContractID,
			BusinessDate:     req.BusinessDate,
			FromState:        fromState,
			ToState:          vo.StateNotStarted,
			Action:           vo.ActionCancelDayStart,
			ActorID:          &req.Actor.UserID,
			ActorType:        req.Actor.ActorType,
			ActorUsername:    req.Actor.Username,
			ActorAccountCode: req.Actor.AccountCode,
			IsAdminOverride:  req.Actor.IsAdminOverride,
			Reason:           &reason,
			Metadata:         map[string]any{},
			OccurredAt:       now,
			RequestID:        req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &CancelDayStartResult{
			TransitionID:  transition.ID,
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateNotStarted,
			OccurredAt:    now,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
