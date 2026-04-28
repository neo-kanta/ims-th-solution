package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
)

// GetHistoryRequest is the input for the transition history query.
type GetHistoryRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
}

// GetHistoryResult is the query output.
type GetHistoryResult struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Transitions  []*entity.WorkflowTransition
}

// GetHistoryHandler retrieves the full transition log for a contract+date.
type GetHistoryHandler struct {
	logRepo domain.TransitionLogRepository
}

// NewGetHistoryHandler creates the handler.
func NewGetHistoryHandler(logRepo domain.TransitionLogRepository) *GetHistoryHandler {
	return &GetHistoryHandler{logRepo: logRepo}
}

// Handle returns the chronological transition history. An empty slice (not an
// error) is returned when the day has never been opened.
func (h *GetHistoryHandler) Handle(
	ctx context.Context,
	req GetHistoryRequest,
) (*GetHistoryResult, error) {
	transitions, err := h.logRepo.ListByContractDate(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("listing transition history: %w", err)
	}
	if transitions == nil {
		transitions = []*entity.WorkflowTransition{}
	}
	return &GetHistoryResult{
		ContractID:   req.ContractID,
		BusinessDate: req.BusinessDate,
		Transitions:  transitions,
	}, nil
}
