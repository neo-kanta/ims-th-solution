package entity

import (
	"time"

	"github.com/google/uuid"
)

// SchedulerAction is the action configured in workflow__schedule_rules.
// It is intentionally separate from WorkflowAction because END_DAY is a
// scheduler concern until the workflow state machine has an explicit command.
type SchedulerAction string

const (
	SchedulerActionOpenDay SchedulerAction = "OPEN_DAY"
	SchedulerActionEndDay  SchedulerAction = "END_DAY"
)

// SchedulerRunStatus mirrors workflow__scheduler_runs.status.
type SchedulerRunStatus string

const (
	SchedulerRunStatusRunning        SchedulerRunStatus = "RUNNING"
	SchedulerRunStatusSuccess        SchedulerRunStatus = "SUCCESS"
	SchedulerRunStatusPartialSuccess SchedulerRunStatus = "PARTIAL_SUCCESS"
	SchedulerRunStatusFailed         SchedulerRunStatus = "FAILED"
	SchedulerRunStatusSkipped        SchedulerRunStatus = "SKIPPED"
)

// SchedulerItemStatus mirrors workflow__scheduler_run_items.status.
type SchedulerItemStatus string

const (
	SchedulerItemStatusSuccess SchedulerItemStatus = "SUCCESS"
	SchedulerItemStatusSkipped SchedulerItemStatus = "SKIPPED"
	SchedulerItemStatusFailed  SchedulerItemStatus = "FAILED"
)

// ScheduleRule is the runtime view of workflow__schedule_rules.
type ScheduleRule struct {
	ID               uuid.UUID
	DaySettingID     uuid.UUID
	Name             string
	Action           SchedulerAction
	TriggerTimeLocal time.Duration
	Timezone         string
	DaysOfWeek       []int
	SkipHolidays     bool
	Priority         int
	EffectiveFrom    time.Time
	EffectiveTo      *time.Time
}

// SchedulerRun is the tick-level audit header written for every scheduler
// execution attempt. Rule/action detail is stored in SchedulerRunItem rows and
// in Summary for rule-level skips that have no contract item.
type SchedulerRun struct {
	ID            uuid.UUID
	SchedulerName string
	RuleID        *uuid.UUID
	Action        *SchedulerAction
	BusinessDate  time.Time
	Timezone      string
	StartedAt     time.Time
	FinishedAt    *time.Time
	Status        SchedulerRunStatus
	TriggeredBy   string
	LockedBy      string
	Summary       map[string]any
	ErrorMessage  *string
	CreatedAt     time.Time
}

// NewSchedulerTickRun creates a RUNNING tick-level audit row model.
func NewSchedulerTickRun(
	schedulerName string,
	businessDate time.Time,
	timezone string,
	startedAt time.Time,
	lockedBy string,
) *SchedulerRun {
	return &SchedulerRun{
		ID:            uuid.New(),
		SchedulerName: schedulerName,
		BusinessDate:  businessDate,
		Timezone:      timezone,
		StartedAt:     startedAt.UTC(),
		Status:        SchedulerRunStatusRunning,
		TriggeredBy:   "SYSTEM",
		LockedBy:      lockedBy,
		Summary:       map[string]any{},
		CreatedAt:     startedAt.UTC(),
	}
}

// SchedulerRunItem is the per-contract audit detail for one scheduler run.
type SchedulerRunItem struct {
	ID             uuid.UUID
	SchedulerRunID uuid.UUID
	RuleID         uuid.UUID
	ContractID     uuid.UUID
	BusinessDate   time.Time
	Action         SchedulerAction
	Status         SchedulerItemStatus
	SkipReason     *string
	ErrorMessage   *string
	WorkflowDayID  *uuid.UUID
	TransitionID   *uuid.UUID
	CreatedAt      time.Time
}

// NewSchedulerRunItem creates an audit detail row model.
func NewSchedulerRunItem(
	runID uuid.UUID,
	ruleID uuid.UUID,
	contractID uuid.UUID,
	businessDate time.Time,
	action SchedulerAction,
	createdAt time.Time,
) *SchedulerRunItem {
	return &SchedulerRunItem{
		ID:             uuid.New(),
		SchedulerRunID: runID,
		RuleID:         ruleID,
		ContractID:     contractID,
		BusinessDate:   businessDate,
		Action:         action,
		CreatedAt:      createdAt.UTC(),
	}
}
