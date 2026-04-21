package query

import (
	"context"
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// ListRuleInstancesRequest is the filter for rule instance listing.
type ListRuleInstancesRequest struct {
	RuleTypeID *string
	IsActive   *bool
	Offset     int
	Limit      int
}

// RuleInstanceDetail augments the entity with live registry metadata.
type RuleInstanceDetail struct {
	entity.RuleInstance
	TypeMetadata *spi.RuleMetadata `json:"type_metadata,omitempty"`
}

// ListRuleInstancesResult is the paginated response.
type ListRuleInstancesResult struct {
	Instances []RuleInstanceDetail `json:"instances"`
	Total     int64                `json:"total"`
	Offset    int                  `json:"offset"`
	Limit     int                  `json:"limit"`
}

// ListRuleInstancesHandler lists configured rule instances with SPI metadata.
type ListRuleInstancesHandler struct {
	instanceRepo domain.RuleInstanceRepository
	registry     *spi.RuleRegistry
}

// NewListRuleInstancesHandler creates the handler.
func NewListRuleInstancesHandler(
	instanceRepo domain.RuleInstanceRepository,
	registry *spi.RuleRegistry,
) *ListRuleInstancesHandler {
	return &ListRuleInstancesHandler{instanceRepo: instanceRepo, registry: registry}
}

// Handle returns a paginated list of rule instances enriched with SPI metadata.
func (h *ListRuleInstancesHandler) Handle(ctx context.Context, req ListRuleInstancesRequest) (*ListRuleInstancesResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		return nil, fmt.Errorf("limit cannot exceed 200")
	}

	instances, total, err := h.instanceRepo.List(ctx, domain.RuleInstanceFilter{
		RuleTypeID: req.RuleTypeID,
		IsActive:   req.IsActive,
		Offset:     req.Offset,
		Limit:      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("listing rule instances: %w", err)
	}

	details := make([]RuleInstanceDetail, 0, len(instances))
	for _, inst := range instances {
		d := RuleInstanceDetail{RuleInstance: inst}
		if evaluator, ok := h.registry.Get(inst.RuleTypeID); ok {
			meta := evaluator.Metadata()
			d.TypeMetadata = &meta
		}
		details = append(details, d)
	}

	return &ListRuleInstancesResult{
		Instances: details,
		Total:     total,
		Offset:    req.Offset,
		Limit:     limit,
	}, nil
}
