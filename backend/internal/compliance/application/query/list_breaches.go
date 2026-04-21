package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
)

// ListBreachesRequest is the filter for the breach list query.
type ListBreachesRequest struct {
	PortfolioID *uuid.UUID
	ContractID  *uuid.UUID
	Status      *entity.BreachStatus
	RuleTypeID  string
	DateFrom    *time.Time
	DateTo      *time.Time
	Offset      int
	Limit       int
}

// ListBreachesResult is the paginated response.
type ListBreachesResult struct {
	Breaches []entity.Breach `json:"breaches"`
	Total    int64           `json:"total"`
	Offset   int             `json:"offset"`
	Limit    int             `json:"limit"`
}

// ListBreachesHandler retrieves breach records with optional filtering.
type ListBreachesHandler struct {
	breachRepo domain.BreachRepository
}

// NewListBreachesHandler creates the handler.
func NewListBreachesHandler(breachRepo domain.BreachRepository) *ListBreachesHandler {
	return &ListBreachesHandler{breachRepo: breachRepo}
}

// Handle returns a paginated breach list.
func (h *ListBreachesHandler) Handle(ctx context.Context, req ListBreachesRequest) (*ListBreachesResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		return nil, fmt.Errorf("limit cannot exceed 200")
	}

	filter := domain.BreachFilter{
		PortfolioID: req.PortfolioID,
		ContractID:  req.ContractID,
		Status:      req.Status,
		RuleTypeID:  req.RuleTypeID,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Offset:      req.Offset,
		Limit:       limit,
	}

	breaches, total, err := h.breachRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing breaches: %w", err)
	}

	return &ListBreachesResult{
		Breaches: breaches,
		Total:    total,
		Offset:   req.Offset,
		Limit:    limit,
	}, nil
}
