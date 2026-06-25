// Package adapter holds cross-module adapters that translate the investment
// module's required ports into concrete implementations sourced from other
// modules — without leaking those modules' internal types.
package adapter

import (
	"context"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// AuditLoggerAdapter bridges the audit module's Recorder port to the
// pkg/contract.AuditLogger interface used by investment commands.
type AuditLoggerAdapter struct {
	recorder auditdomain.Recorder
}

// NewAuditLoggerAdapter wires the adapter.
func NewAuditLoggerAdapter(recorder auditdomain.Recorder) *AuditLoggerAdapter {
	if recorder == nil {
		recorder = auditdomain.NopRecorder{}
	}
	return &AuditLoggerAdapter{recorder: recorder}
}

// LogAction implements contract.AuditLogger using the fire-and-forget path.
// Failures are swallowed by the recorder's internal logging; callers in
// low-risk CRUD paths use this so an audit-store outage does not block the
// business action.
func (a *AuditLoggerAdapter) LogAction(entry contract.AuditEntry) error {
	if a == nil || a.recorder == nil {
		return nil
	}
	actor, metadata := toRecorderInputs(entry)
	a.recorder.Record(
		context.Background(),
		actor,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		"", // ip_address — handler does not propagate this yet
		"", // user_agent
		metadata,
	)
	return nil
}

// LogActionStrict implements contract.AuditLogger using the synchronous,
// error-returning path. Financial-grade actions invoke this and surface the
// failure so the operation is never silently de-audited.
func (a *AuditLoggerAdapter) LogActionStrict(ctx context.Context, entry contract.AuditEntry) error {
	if a == nil || a.recorder == nil {
		return nil
	}
	actor, metadata := toRecorderInputs(entry)
	return a.recorder.RecordStrict(
		ctx,
		actor,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		"", // ip_address — handler does not propagate this yet
		"", // user_agent
		metadata,
	)
}

// toRecorderInputs translates a contract.AuditEntry into the parameter set
// expected by the recorder. Centralised so the fire-and-forget and strict
// paths emit identical events.
func toRecorderInputs(entry contract.AuditEntry) (*uuid.UUID, map[string]any) {
	var actor *uuid.UUID
	if entry.ActorID != "" {
		if id, err := uuid.Parse(entry.ActorID); err == nil {
			actor = &id
		}
	}
	metadata := map[string]any{
		"module":        entry.Module,
		"resource_type": entry.ResourceType,
		"resource_id":   entry.ResourceID,
	}
	if !entry.BusinessDate.IsZero() {
		metadata["business_date"] = entry.BusinessDate.Format("2006-01-02")
	}
	if entry.Details != nil {
		metadata["details"] = entry.Details
	}
	return actor, metadata
}

var _ contract.AuditLogger = (*AuditLoggerAdapter)(nil)
