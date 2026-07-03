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

type RollbackAccountingCloseRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
	Reason       string
}

type RollbackAccountingCloseResult struct {
	TransitionID  uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time
	FromState     vo.WorkflowState
	ToState       vo.WorkflowState
	OccurredAt    time.Time
	RecloseCount  int
}

type RollbackAccountingCloseHandler struct {
	pool    *pgxpool.Pool
	dayRepo domain.WorkflowDayRepository
	logRepo domain.TransitionLogRepository
	policy  *policy.TransitionPolicy
}

func NewRollbackAccountingCloseHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	pol *policy.TransitionPolicy,
) *RollbackAccountingCloseHandler {
	return &RollbackAccountingCloseHandler{
		pool:    pool,
		dayRepo: dayRepo,
		logRepo: logRepo,
		policy:  pol,
	}
}

func (h *RollbackAccountingCloseHandler) Handle(ctx context.Context, req RollbackAccountingCloseRequest) (*RollbackAccountingCloseResult, error) {
	var result *RollbackAccountingCloseResult
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

		if err := h.policy.CanRollbackAccountingClose(policy.RollbackAccountingCloseInput{
			CurrentDay: day,
			Reason:     req.Reason,
		}); err != nil {
			return err
		}

		// Stash the prior accounting date so the audit trail preserves which
		// NAV cycle was reversed. Cleared the next time CLOSE_ACCOUNTING runs.
		var priorAccountingDate *time.Time
		if day.AccountingDate != nil {
			t := *day.AccountingDate
			priorAccountingDate = &t
		}

		now := time.Now().UTC()
		fromState := day.CurrentState
		day.CurrentState = vo.StateTransactionClosed
		day.AccountingClosedAt = nil
		day.AccountingClosedBy = nil
		day.PrevAccountingDate = priorAccountingDate
		day.AccountingDate = nil
		day.PendingReclose = true
		day.RecloseCount = day.RecloseCount + 1
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID

		if err := h.dayRepo.UpdateState(ctx, tx, day); err != nil {
			return err
		}

		reason := req.Reason
		metadata := map[string]any{
			"recloseCount": day.RecloseCount,
		}
		if priorAccountingDate != nil {
			metadata["prev_accounting_date"] = priorAccountingDate.Format("2006-01-02")
		}
		transition := &entity.WorkflowTransition{
			ID:               uuid.New(),
			WorkflowDayID:    day.ID,
			ContractID:       req.ContractID,
			BusinessDate:     req.BusinessDate,
			FromState:        fromState,
			ToState:          vo.StateTransactionClosed,
			Action:           vo.ActionRollbackAccountingClose,
			ActorID:          &req.Actor.UserID,
			ActorType:        req.Actor.ActorType,
			ActorUsername:    req.Actor.Username,
			ActorAccountCode: req.Actor.AccountCode,
			IsAdminOverride:  req.Actor.IsAdminOverride,
			Reason:           &reason,
			Metadata:         metadata,
			OccurredAt:       now,
			RequestID:        req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &RollbackAccountingCloseResult{
			TransitionID:  transition.ID,
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateTransactionClosed,
			OccurredAt:    now,
			RecloseCount:  day.RecloseCount,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
