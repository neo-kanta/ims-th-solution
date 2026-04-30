package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// WorkflowTransition is an immutable record of a single state-machine step.
//
// Rows in workflow__transition_log are NEVER updated or deleted.
// Every transition — including cancellations and rollbacks — appends a new row.
// This table is the forensic source of truth for "who did what, when, and why".
type WorkflowTransition struct {
	ID            uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time

	FromState vo.WorkflowState
	ToState   vo.WorkflowState
	Action    vo.WorkflowAction

	// Actor — nil ActorID means a system-automated action
	ActorID       *uuid.UUID
	ActorType     vo.ActorType
	ActorUsername string // snapshot at transition time

	// Reason is mandatory for cancel/rollback actions; optional otherwise
	Reason *string

	// Metadata holds action-specific payload: attestation data, external refs,
	// IRG snapshots, etc. Stored as JSONB. Never nil; use empty map as default.
	Metadata map[string]any

	OccurredAt time.Time
	RequestID  string // X-Request-Id for tracing
}
