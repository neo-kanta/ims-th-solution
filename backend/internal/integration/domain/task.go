package domain

import "time"

// TaskType identifies the business source of a work item.
type TaskType string

const (
	TaskTypeResearchReview   TaskType = "RESEARCH_REVIEW"
	TaskTypeWorkflowPending  TaskType = "WORKFLOW_PENDING"
	TaskTypeComplianceBreach TaskType = "COMPLIANCE_BREACH"
)

// Priority reflects urgency.
type Priority string

const (
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
	PriorityInfo   Priority = "INFO"
)

// TaskStatus is the lifecycle state of the work item.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "PENDING"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
)

// DashboardTask is the normalised cross-module work item.
//
// Display semantics:
//   - Title and Description hold reasonable English text so non-localised
//     consumers (alerts, exports, notifications) can render them directly.
//   - Subject and Severity are language-neutral identifiers the frontend
//     uses to build localised titles via i18n templates:
//     research   → Subject = report_no
//     workflow   → Subject = workflow state code (NOT_STARTED, DAY_OPEN, …)
//     compliance → Subject = rule_type_id, Severity = BLOCK|WARN|REQUIRE_APPROVAL|MONITOR
//
// Frontend display is responsible for choosing between the English Title
// fallback and the localised template derived from Subject/Severity/Type.
type DashboardTask struct {
	TaskID         string
	Type           TaskType
	Module         string
	Title          string
	Description    string
	Priority       Priority
	Status         TaskStatus
	BusinessDate   string
	SourceRecordID string
	SourceType     string
	ActionURL      string
	Reason         string
	Subject        string
	Severity       string
	CanAct         bool
	AllowedActions []string
	ContractID     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// WorkflowStateRow is a lightweight snapshot of a contract's current workflow state.
type WorkflowStateRow struct {
	ContractID   string
	BusinessDate string
	CurrentState string
	UpdatedAt    time.Time
}

// TaskSummary holds aggregate counts for the task feed.
type TaskSummary struct {
	Total        int
	ByModule     map[string]int
	ByPriority   map[string]int
	HighPriority int
}

// DashboardSnapshot is the complete read model for GET /dashboard/me.
type DashboardSnapshot struct {
	Tasks          []DashboardTask
	Summary        TaskSummary
	WorkflowStates []WorkflowStateRow
	LastRefreshed  time.Time
}
