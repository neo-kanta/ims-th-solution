// Package query contains read-side use cases for the compliance module.
package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
)

// CheckGroupResult bundles all records and breaches for one check group.
type CheckGroupResult struct {
	CheckGroupID uuid.UUID            `json:"check_group_id"`
	Records      []entity.CheckRecord `json:"records"`
	Breaches     []entity.Breach      `json:"breaches"`
}

// GetCheckGroupHandler retrieves all check records and breaches for a group ID.
type GetCheckGroupHandler struct {
	checkRepo  domain.CheckRecordRepository
	breachRepo domain.BreachRepository
}

// NewGetCheckGroupHandler creates the handler.
func NewGetCheckGroupHandler(
	checkRepo domain.CheckRecordRepository,
	breachRepo domain.BreachRepository,
) *GetCheckGroupHandler {
	return &GetCheckGroupHandler{checkRepo: checkRepo, breachRepo: breachRepo}
}

// Handle returns the full check group result.
func (h *GetCheckGroupHandler) Handle(ctx context.Context, groupID uuid.UUID) (*CheckGroupResult, error) {
	if groupID == uuid.Nil {
		return nil, fmt.Errorf("group_id is required")
	}

	records, err := h.checkRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("fetching check records: %w", err)
	}

	status := entity.BreachStatusOpen
	breaches, _, err := h.breachRepo.List(ctx, domain.BreachFilter{
		// We want all breach statuses for this group — filter by group_id via records.
		// ASSUMPTION: list by check_group_id isn't in the interface; use portfolio_id + date range?
		// For now: fetch OPEN breaches and include those whose check_record_id matches.
		Status: &status,
		Limit:  500,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching breaches: %w", err)
	}

	// Filter breaches belonging to this check group.
	recordIDs := make(map[uuid.UUID]bool, len(records))
	for _, r := range records {
		recordIDs[r.ID] = true
	}
	var groupBreaches []entity.Breach
	for _, b := range breaches {
		if recordIDs[b.CheckRecordID] {
			groupBreaches = append(groupBreaches, b)
		}
	}

	return &CheckGroupResult{
		CheckGroupID: groupID,
		Records:      records,
		Breaches:     groupBreaches,
	}, nil
}
