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

// LogAction implements contract.AuditLogger.
//
// The contract interface is fire-and-forget (returns error); the audit
// module's Recorder is also fire-and-forget but logs failures internally.
// We always return nil and rely on the recorder to deal with errors.
func (a *AuditLoggerAdapter) LogAction(entry contract.AuditEntry) error {
	if a == nil || a.recorder == nil {
		return nil
	}

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

var _ contract.AuditLogger = (*AuditLoggerAdapter)(nil)
