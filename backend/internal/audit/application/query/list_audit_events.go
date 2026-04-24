package query

import (
	"context"
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
)

// ListAuditEventsQuery retrieves paginated audit events.
type ListAuditEventsQuery struct {
	auditRepo domain.AuditRepository
}

// NewListAuditEventsQuery creates a new ListAuditEventsQuery.
func NewListAuditEventsQuery(auditRepo domain.AuditRepository) *ListAuditEventsQuery {
	return &ListAuditEventsQuery{auditRepo: auditRepo}
}

// Execute returns paginated audit events matching the filter.
func (q *ListAuditEventsQuery) Execute(ctx context.Context, filter domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}

	events, total, err := q.auditRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit events: %w", err)
	}
	return events, total, nil
}

// ExecuteForExport returns audit events for export with a higher limit.
func (q *ListAuditEventsQuery) ExecuteForExport(ctx context.Context, filter domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10000
	}
	if filter.Limit > 10000 {
		filter.Limit = 10000
	}
	filter.Offset = 0

	events, total, err := q.auditRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("exporting audit events: %w", err)
	}
	return events, total, nil
}
