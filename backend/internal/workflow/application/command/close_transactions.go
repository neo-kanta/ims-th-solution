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
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
)

type CloseTransactionsRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
}

type CloseTransactionsResult struct {
	TransitionID        uuid.UUID
	WorkflowDayID       uuid.UUID
	ContractID          uuid.UUID
	BusinessDate        time.Time
	FromState           vo.WorkflowState
	ToState             vo.WorkflowState
	OccurredAt          time.Time
	TransactionClosedAt time.Time

	// ComplianceCheckGroupID references the post-trade verification group that
	// cleared (or blocked) this close. Always set when the verifier is wired.
	ComplianceCheckGroupID *uuid.UUID
}

// EvaluatePostTradeGate is the pure-function gate applied before transitioning
// a workflow day to TRANSACTION_CLOSED. Returns nil when the transition is
// permitted; otherwise a typed *domain.ErrPostTradeBreachesBlockClose.
//
// Exposed for unit testing — production callers go through CloseTransactionsHandler.
func EvaluatePostTradeGate(
	contractID uuid.UUID,
	businessDate time.Time,
	verification *contract.PostTradeVerificationResult,
) error {
	if verification == nil {
		// Verifier returned nil → treat as no breaches (fail-open only for nil,
		// not for error; errors are handled at the caller with fail-closed).
		return nil
	}
	if !verification.HasBlockingBreach {
		return nil
	}
	blockCount := 0
	for _, b := range verification.Breaches {
		if b.Verdict == contract.ComplianceVerdictBlock {
			blockCount++
		}
	}
	if blockCount == 0 {
		// HasBlockingBreach=true but no BLOCK entries — treat the summary as
		// authoritative and still refuse the close. Count at least 1 for UX.
		blockCount = 1
	}
	return &domain.ErrPostTradeBreachesBlockClose{
		ContractID:   contractID.String(),
		BusinessDate: businessDate.Format("2006-01-02"),
		CheckGroupID: verification.CheckGroupID.String(),
		BreachCount:  blockCount,
	}
}

type CloseTransactionsHandler struct {
	pool         *pgxpool.Pool
	dayRepo      domain.WorkflowDayRepository
	logRepo      domain.TransitionLogRepository
	policy       *policy.TransitionPolicy
	postTradeVer contract.PostTradeVerifier
}

// NewCloseTransactionsHandler wires the handler.
//
// `postTradeVer` is optional — nil disables the IRG post-trade gate (useful in
// unit tests and during early bring-up, before the compliance module ships).
// In production it MUST be a non-nil adapter, wired in cmd/server/main.go.
func NewCloseTransactionsHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	pol *policy.TransitionPolicy,
	postTradeVer contract.PostTradeVerifier,
) *CloseTransactionsHandler {
	return &CloseTransactionsHandler{
		pool:         pool,
		dayRepo:      dayRepo,
		logRepo:      logRepo,
		policy:       pol,
		postTradeVer: postTradeVer,
	}
}

// Handle runs the close-transactions transition with the IRG post-trade gate.
//
// Order of operations (all inside a single DB transaction):
//  1. SELECT ... FOR UPDATE the workflow day row.
//  2. Enforce transition policy (current state must be MANAGER_APPROVED).
//  3. Run the post-trade verifier. On any BLOCK breach, abort with
//     *domain.ErrPostTradeBreachesBlockClose — the TX is rolled back and
//     the day stays in MANAGER_APPROVED.
//  4. Persist the new state + transition log.
//
// The verifier call sits INSIDE the transaction so a concurrent override
// (which itself takes a short TX on the breach row) cannot race with us: once
// we hold FOR UPDATE on the workflow day, any override racing to resolve a
// breach will complete before we read the breach set. Conversely, if we block,
// rollback makes the intended transition invisible to other readers.
func (h *CloseTransactionsHandler) Handle(ctx context.Context, req CloseTransactionsRequest) (*CloseTransactionsResult, error) {
	var result *CloseTransactionsResult
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

		if err := h.policy.CanCloseTransactions(policy.CloseTransactionsInput{CurrentDay: day}); err != nil {
			return err
		}

		// IRG post-trade gate — refuse to close while BLOCK breaches remain.
		var checkGroupID *uuid.UUID
		if h.postTradeVer != nil {
			verification, verr := h.postTradeVer.RunPostTradeVerification(ctx, req.ContractID, req.BusinessDate)
			if verr != nil {
				return fmt.Errorf("running IRG post-trade verification: %w", verr)
			}
			if gateErr := EvaluatePostTradeGate(req.ContractID, req.BusinessDate, verification); gateErr != nil {
				return gateErr
			}
			if verification != nil {
				gid := verification.CheckGroupID
				checkGroupID = &gid
			}
		}

		now := time.Now().UTC()
		fromState := day.CurrentState
		day.CurrentState = vo.StateTransactionClosed
		day.TransactionClosedAt = &now
		day.TransactionClosedBy = &req.Actor.UserID
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID

		if err := h.dayRepo.UpdateState(ctx, tx, day); err != nil {
			return err
		}

		metadata := map[string]any{}
		if checkGroupID != nil {
			metadata["compliance_check_group_id"] = checkGroupID.String()
		}
		transition := &entity.WorkflowTransition{
			ID:            uuid.New(),
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     fromState,
			ToState:       vo.StateTransactionClosed,
			Action:        vo.ActionCloseTransactions,
			ActorID:       &req.Actor.UserID,
			ActorType:     req.Actor.ActorType,
			ActorUsername: req.Actor.Username,
			Metadata:      metadata,
			OccurredAt:    now,
			RequestID:     req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &CloseTransactionsResult{
			TransitionID:           transition.ID,
			WorkflowDayID:          day.ID,
			ContractID:             req.ContractID,
			BusinessDate:           req.BusinessDate,
			FromState:              fromState,
			ToState:                vo.StateTransactionClosed,
			OccurredAt:             now,
			TransactionClosedAt:    now,
			ComplianceCheckGroupID: checkGroupID,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
