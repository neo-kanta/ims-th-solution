package command

import (
	"context"
	"fmt"
	"log/slog"
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

type CloseAccountingRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
}

type CloseAccountingResult struct {
	TransitionID       uuid.UUID
	WorkflowDayID      uuid.UUID
	ContractID         uuid.UUID
	BusinessDate       time.Time
	FromState          vo.WorkflowState
	ToState            vo.WorkflowState
	OccurredAt         time.Time
	AccountingClosedAt time.Time
}

type CloseAccountingHandler struct {
	pool    *pgxpool.Pool
	dayRepo domain.WorkflowDayRepository
	logRepo domain.TransitionLogRepository
	policy  *policy.TransitionPolicy
}

func NewCloseAccountingHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	pol *policy.TransitionPolicy,
) *CloseAccountingHandler {
	return &CloseAccountingHandler{
		pool:    pool,
		dayRepo: dayRepo,
		logRepo: logRepo,
		policy:  pol,
	}
}

func (h *CloseAccountingHandler) Handle(ctx context.Context, req CloseAccountingRequest) (*CloseAccountingResult, error) {
	var result *CloseAccountingResult
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

		if err := h.policy.CanCloseAccounting(policy.CloseAccountingInput{CurrentDay: day}); err != nil {
			return err
		}

		now := time.Now().UTC()
		fromState := day.CurrentState
		day.CurrentState = vo.StateAccountingClosed
		day.AccountingClosedAt = &now
		day.AccountingClosedBy = &req.Actor.UserID
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID

		if err := h.dayRepo.UpdateState(ctx, tx, day); err != nil {
			return err
		}

		transition := &entity.WorkflowTransition{
			ID:            uuid.New(),
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateAccountingClosed,
			Action:        vo.ActionCloseAccounting,
			ActorID:       &req.Actor.UserID,
			ActorType:     req.Actor.ActorType,
			ActorUsername: req.Actor.Username,
			Metadata:      map[string]any{},
			OccurredAt:    now,
			RequestID:     req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &CloseAccountingResult{
			TransitionID:       transition.ID,
			WorkflowDayID:      day.ID,
			ContractID:         req.ContractID,
			BusinessDate:       req.BusinessDate,
			FromState:          fromState,
			ToState:            vo.StateAccountingClosed,
			OccurredAt:         now,
			AccountingClosedAt: now,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// TODO(batch 3): invoke compliance post-trade IRG check here.
	// Contract: compliance.RunPostTradeCheck(ctx, contractID, businessDate, workflowDayID).
	// Must not block the HTTP response — dispatch via the outbox pattern once the
	// domain-event dispatcher lands.
	slog.InfoContext(ctx, "accounting_closed: post-trade IRG check required",
		"contractId", req.ContractID.String(),
		"businessDate", req.BusinessDate.Format("2006-01-02"),
		"workflowDayId", result.WorkflowDayID.String(),
		"transitionId", result.TransitionID.String(),
	)

	return result, nil
}
