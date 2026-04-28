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
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
)

type CancelTransactionCloseRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
	Reason       string
}

type CancelTransactionCloseResult struct {
	TransitionID  uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time
	FromState     vo.WorkflowState
	ToState       vo.WorkflowState
	OccurredAt    time.Time
}

type CancelTransactionCloseHandler struct {
	pool    *pgxpool.Pool
	dayRepo domain.WorkflowDayRepository
	logRepo domain.TransitionLogRepository
	policy  *policy.TransitionPolicy
}

func NewCancelTransactionCloseHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	pol *policy.TransitionPolicy,
) *CancelTransactionCloseHandler {
	return &CancelTransactionCloseHandler{
		pool:    pool,
		dayRepo: dayRepo,
		logRepo: logRepo,
		policy:  pol,
	}
}

func (h *CancelTransactionCloseHandler) Handle(ctx context.Context, req CancelTransactionCloseRequest) (*CancelTransactionCloseResult, error) {
	var result *CancelTransactionCloseResult
	txErr := database.WithTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		day, err := h.dayRepo.GetForUpdate(ctx, tx, req.ContractID, req.BusinessDate)
		if err != nil {
			return fmt.Errorf("locking workflow day: %w", err)
		}
		if day == nil {
			return &domain.ErrNotFound{
				ContractID:   req.ContractID.String(),
				BusinessDate: req.BusinessDate.Format("2006-01-02"),
			}
		}

		if err := h.policy.CanCancelTransactionClose(policy.CancelTransactionCloseInput{
			CurrentDay: day,
			Reason:     req.Reason,
		}); err != nil {
			return err
		}

		now := time.Now().UTC()
		fromState := day.CurrentState
		day.CurrentState = vo.StateManagerApproved
		day.TransactionClosedAt = nil
		day.TransactionClosedBy = nil
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID

		if err := h.dayRepo.UpdateState(ctx, tx, day); err != nil {
			return err
		}

		reason := req.Reason
		transition := &entity.WorkflowTransition{
			ID:            uuid.New(),
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateManagerApproved,
			Action:        vo.ActionCancelTransactionClose,
			ActorID:       &req.Actor.UserID,
			ActorType:     req.Actor.ActorType,
			ActorUsername: req.Actor.Username,
			Reason:        &reason,
			Metadata:      map[string]any{},
			OccurredAt:    now,
			RequestID:     req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &CancelTransactionCloseResult{
			TransitionID:  transition.ID,
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateManagerApproved,
			OccurredAt:    now,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
