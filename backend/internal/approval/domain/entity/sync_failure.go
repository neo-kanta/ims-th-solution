package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// SyncFailure is a durable, operator-visible record of an approval decision
// whose post-commit subject-sync callback (OnApproved/OnRejected) failed to
// reach the owning business module. The approval decision itself already
// committed and is never rolled back for a sync failure; this record is the
// mechanism an operator uses to discover and replay the missed callback so the
// business module's state (e.g. a materialized ledger transaction) eventually
// catches up with the approval outcome.
type SyncFailure struct {
	ID                uuid.UUID
	ApprovalRequestID uuid.UUID
	SubjectType       vo.SubjectType
	SubjectID         uuid.UUID
	Outcome           vo.SyncFailureOutcome
	// Reason carries the rejection/revoke reason for REJECTED/REVOKED outcomes;
	// empty for APPROVED.
	Reason string

	AttemptCount int
	MaxAttempts  int
	LastError    string
	Status       vo.SyncFailureStatus

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ResolvedAt *time.Time
	ResolvedBy *uuid.UUID
}

// CanRetry reports whether an operator-triggered retry is allowed: the record
// must still be PENDING (not already resolved) and have attempts remaining.
func (f *SyncFailure) CanRetry() bool {
	return f != nil && f.Status == vo.SyncFailureStatusPending && f.AttemptCount < f.MaxAttempts
}
