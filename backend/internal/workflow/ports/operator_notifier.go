// Package ports holds the cross-module contracts the workflow module consumes.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// StuckDayWarning describes a workflow day that the scheduler has decided
// warrants operator attention. Emitted by the watcher each tick; consumed
// by an injected OperatorNotifier — typically backed by the in-app
// notification service.
type StuckDayWarning struct {
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time
	CurrentState  string
	OpenedBy      *uuid.UUID
	ApprovedBy    *uuid.UUID
	// AgeDays is the number of full days the row has been in CurrentState.
	AgeDays int
}

// OperatorNotifier is implemented by an outbound channel adapter (e.g. the
// notification module) and consumed by the workflow watcher to warn
// operators when a day is stuck in DAY_OPEN or MANAGER_APPROVED past the
// configured cutoff. Implementations MUST be dedupe-aware so the warning
// is not re-sent on every scheduler tick.
type OperatorNotifier interface {
	NotifyStuckDay(ctx context.Context, w StuckDayWarning) error
}

// NopOperatorNotifier is the default when no notifier is wired. The watcher
// still runs (and logs the warning); no in-app notification is created.
type NopOperatorNotifier struct{}

// NotifyStuckDay implements OperatorNotifier.
func (NopOperatorNotifier) NotifyStuckDay(context.Context, StuckDayWarning) error {
	return nil
}
