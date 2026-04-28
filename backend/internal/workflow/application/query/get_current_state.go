package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
)

// GetCurrentStateRequest is the input for the current-state query.
type GetCurrentStateRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time // pre-parsed Bangkok calendar date
}

// GetCurrentStateResult is the query output.
// When Persisted is false, the Day field is nil and the response represents
// the synthetic NOT_STARTED state (no row exists yet).
type GetCurrentStateResult struct {
	// Persisted indicates whether a workflow__day_states row actually exists.
	// false → caller renders a synthetic NOT_STARTED response.
	Persisted bool

	// Day is the current workflow day record. Nil when Persisted = false.
	Day *entity.WorkflowDay

	// AllowedActions is the set of actions structurally allowed from CurrentState.
	AllowedActions []vo.WorkflowAction

	// PrevDay is the previous business day's record.
	// Populated only when Persisted = false (NOT_STARTED), to let the UI
	// display the previous-day status before the operator opens the day.
	// Nil when not applicable or when the previous day has no record.
	PrevDay *entity.WorkflowDay

	// PreviousBusinessDate is the business date of PrevDay (always populated).
	PreviousBusinessDate time.Time
}

// GetCurrentStateHandler retrieves the current workflow state for a contract+date.
type GetCurrentStateHandler struct {
	dayRepo      domain.WorkflowDayRepository
	calendarPort ports.HolidayCalendarPort
}

// NewGetCurrentStateHandler creates the handler.
func NewGetCurrentStateHandler(
	dayRepo domain.WorkflowDayRepository,
	calendarPort ports.HolidayCalendarPort,
) *GetCurrentStateHandler {
	return &GetCurrentStateHandler{dayRepo: dayRepo, calendarPort: calendarPort}
}

// Handle executes the query.
func (h *GetCurrentStateHandler) Handle(
	ctx context.Context,
	req GetCurrentStateRequest,
) (*GetCurrentStateResult, error) {
	day, err := h.dayRepo.GetByContractDate(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("reading workflow day: %w", err)
	}

	// Compute previous business date for all responses (used in NOT_STARTED path).
	prevDate, err := h.calendarPort.PreviousBusinessDay(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("computing previous business day: %w", err)
	}

	// Case A: no row exists → synthetic NOT_STARTED
	if day == nil {
		prevDay, err := h.dayRepo.GetByContractDate(ctx, req.ContractID, prevDate)
		if err != nil {
			return nil, fmt.Errorf("reading previous workflow day: %w", err)
		}
		return &GetCurrentStateResult{
			Persisted:            false,
			Day:                  nil,
			AllowedActions:       vo.AllowedActions(vo.StateNotStarted),
			PrevDay:              prevDay,
			PreviousBusinessDate: prevDate,
		}, nil
	}

	// Case B: row exists
	return &GetCurrentStateResult{
		Persisted:            true,
		Day:                  day,
		AllowedActions:       vo.AllowedActions(day.CurrentState),
		PrevDay:              nil, // not shown once the day is open
		PreviousBusinessDate: prevDate,
	}, nil
}
