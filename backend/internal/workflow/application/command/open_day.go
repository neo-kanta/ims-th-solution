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

// OpenDayRequest is the input for the OPEN_DAY command.
// BusinessDate must be an explicit date parsed from the API request;
// it must NEVER be derived from time.Now().
type OpenDayRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext
}

// OpenDayResult is returned on success.
type OpenDayResult struct {
	TransitionID  uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time
	FromState     vo.WorkflowState
	ToState       vo.WorkflowState
	OccurredAt    time.Time
}

// OpenDayHandler executes the OPEN_DAY workflow transition.
//
// Transaction boundary: this handler begins and commits the transaction.
// The transaction spans: workflow__day_states INSERT + workflow__transition_log INSERT.
//
// External calls (calendar check, previous-day read) run BEFORE the transaction
// to avoid holding the database lock during I/O.
type OpenDayHandler struct {
	pool         *pgxpool.Pool
	dayRepo      domain.WorkflowDayRepository
	logRepo      domain.TransitionLogRepository
	calendarPort ports.HolidayCalendarPort
	policy       *policy.TransitionPolicy
}

// NewOpenDayHandler creates the handler with its dependencies.
func NewOpenDayHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	calendarPort ports.HolidayCalendarPort,
	pol *policy.TransitionPolicy,
) *OpenDayHandler {
	return &OpenDayHandler{
		pool:         pool,
		dayRepo:      dayRepo,
		logRepo:      logRepo,
		calendarPort: calendarPort,
		policy:       pol,
	}
}

// Handle executes the OPEN_DAY transition.
//
// Flow:
//  1. Business-day guard (no DB, fail fast).
//  2. Compute previous business date via calendar port.
//  3. BEGIN TRANSACTION.
//  4. Read previous business day record (non-locking, inside tx for consistency).
//  5. Run policy checks (pure, no I/O).
//  6. INSERT workflow day row (ON CONFLICT DO NOTHING — race-safe).
//  7. INSERT transition log row.
//  8. COMMIT.
func (h *OpenDayHandler) Handle(ctx context.Context, req OpenDayRequest) (*OpenDayResult, error) {
	// Step 1 — business-day guard (pure calendar check, no database needed)
	isBusinessDay, holidayName, err := h.calendarPort.IsBusinessDay(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("checking business day: %w", err)
	}

	// Step 2 — previous business date (needed inside the transaction)
	prevDate, err := h.calendarPort.PreviousBusinessDay(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("computing previous business day: %w", err)
	}

	var result *OpenDayResult
	txErr := database.WithTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		// Step 4 — read previous day state (non-locking read inside tx)
		prevDay, err := h.dayRepo.GetByBusinessDate(ctx, prevDate)
		if err != nil {
			return fmt.Errorf("reading previous business day: %w", err)
		}

		// Read current day to detect existing row (for the ErrWorkflowDayExists path)
		currentDay, err := h.dayRepo.GetByBusinessDate(ctx, req.BusinessDate)
		if err != nil {
			return fmt.Errorf("reading current workflow day: %w", err)
		}

		// Step 5 — policy checks (pure, no I/O)
		if err := h.policy.CanOpenDay(policy.OpenDayInput{
			IsBusinessDay: isBusinessDay,
			HolidayName:   holidayName,
			CurrentDay:    currentDay,
			PrevDay:       prevDay,
		}); err != nil {
			return err // typed domain error; transport layer maps to HTTP status
		}

		// Step 6 — build and insert the workflow day entity
		now := time.Now().UTC()
		day := &entity.WorkflowDay{
			ID:             uuid.New(),
			ContractID:     req.ContractID,
			BusinessDate:   req.BusinessDate,
			CurrentState:   vo.StateInvestmentDayStarted,
			OpenedAt:       &now,
			OpenedBy:       &req.Actor.UserID,
			PendingReclose: false,
			RecloseCount:   0,
			Version:        1,
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedBy:      req.Actor.UserID,
			UpdatedBy:      req.Actor.UserID,
		}
		if err := h.dayRepo.Insert(ctx, tx, day); err != nil {
			return err // may be ErrWorkflowDayExists if a concurrent request won
		}

		// Step 7 — append immutable transition log
		transition := &entity.WorkflowTransition{
			ID:               uuid.New(),
			WorkflowDayID:    day.ID,
			ContractID:       req.ContractID,
			BusinessDate:     req.BusinessDate,
			FromState:        vo.StateNotStarted,
			ToState:          vo.StateInvestmentDayStarted,
			Action:           vo.ActionOpenDay,
			ActorID:          &req.Actor.UserID,
			ActorType:        req.Actor.ActorType,
			ActorUsername:    req.Actor.Username,
			ActorAccountCode: req.Actor.AccountCode,
			IsAdminOverride:  req.Actor.IsAdminOverride,
			Metadata:         map[string]any{},
			OccurredAt:       now,
			RequestID:        req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &OpenDayResult{
			TransitionID:  transition.ID,
			WorkflowDayID: day.ID,
			ContractID:    req.ContractID,
			BusinessDate:  req.BusinessDate,
			FromState:     vo.StateNotStarted,
			ToState:       vo.StateInvestmentDayStarted,
			OccurredAt:    now,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
