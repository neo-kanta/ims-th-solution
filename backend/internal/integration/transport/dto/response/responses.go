package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
)

// TaskDTO is the API representation of a single dashboard task.
//
// Title and Description are English fallbacks; Subject and Severity are
// language-neutral identifiers the frontend uses to build localised titles
// via i18n templates (see dashboard.taskCard.* in the i18n messages).
type TaskDTO struct {
	TaskID         string   `json:"taskId"`
	Type           string   `json:"type"`
	Module         string   `json:"module"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	Status         string   `json:"status"`
	BusinessDate   string   `json:"businessDate"`
	SourceRecordID string   `json:"sourceRecordId"`
	SourceType     string   `json:"sourceType"`
	ActionURL      string   `json:"actionUrl"`
	Reason         string   `json:"reason"`
	Subject        string   `json:"subject"`
	Severity       string   `json:"severity"`
	CanAct         bool     `json:"canAct"`
	AllowedActions []string `json:"allowedActions"`
	ContractID     string   `json:"contractId"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// WorkflowStateDTO is the API representation of a workflow state row.
type WorkflowStateDTO struct {
	ContractID   string `json:"contractId"`
	BusinessDate string `json:"businessDate"`
	CurrentState string `json:"currentState"`
	UpdatedAt    string `json:"updatedAt"`
}

// TaskSummaryDTO is the API representation of aggregate task counts.
type TaskSummaryDTO struct {
	Total        int            `json:"total"`
	ByModule     map[string]int `json:"byModule"`
	ByPriority   map[string]int `json:"byPriority"`
	HighPriority int            `json:"highPriority"`
}

// DashboardSnapshotDTO is the API response for GET /integration/dashboard/me.
type DashboardSnapshotDTO struct {
	Tasks          []TaskDTO          `json:"tasks"`
	Summary        TaskSummaryDTO     `json:"summary"`
	WorkflowStates []WorkflowStateDTO `json:"workflowStates"`
	LastRefreshed  string             `json:"lastRefreshed"`
}

// TaskListDTO is the API response for GET /integration/tasks/my.
type TaskListDTO struct {
	Tasks   []TaskDTO      `json:"tasks"`
	Summary TaskSummaryDTO `json:"summary"`
}

// FromDashboardTask converts a domain task to its DTO.
func FromDashboardTask(t domain.DashboardTask) TaskDTO {
	actions := t.AllowedActions
	if actions == nil {
		actions = []string{}
	}
	return TaskDTO{
		TaskID:         t.TaskID,
		Type:           string(t.Type),
		Module:         t.Module,
		Title:          t.Title,
		Description:    t.Description,
		Priority:       string(t.Priority),
		Status:         string(t.Status),
		BusinessDate:   t.BusinessDate,
		SourceRecordID: t.SourceRecordID,
		SourceType:     t.SourceType,
		ActionURL:      t.ActionURL,
		Reason:         t.Reason,
		Subject:        t.Subject,
		Severity:       t.Severity,
		CanAct:         t.CanAct,
		AllowedActions: actions,
		ContractID:     t.ContractID,
		CreatedAt:      t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.Format(time.RFC3339),
	}
}

// FromWorkflowStateRow converts a domain workflow state row to its DTO.
func FromWorkflowStateRow(w domain.WorkflowStateRow) WorkflowStateDTO {
	return WorkflowStateDTO{
		ContractID:   w.ContractID,
		BusinessDate: w.BusinessDate,
		CurrentState: w.CurrentState,
		UpdatedAt:    w.UpdatedAt.Format(time.RFC3339),
	}
}

// FromTaskSummary converts a domain summary to its DTO.
func FromTaskSummary(s domain.TaskSummary) TaskSummaryDTO {
	byModule := s.ByModule
	if byModule == nil {
		byModule = map[string]int{}
	}
	byPriority := s.ByPriority
	if byPriority == nil {
		byPriority = map[string]int{}
	}
	return TaskSummaryDTO{
		Total:        s.Total,
		ByModule:     byModule,
		ByPriority:   byPriority,
		HighPriority: s.HighPriority,
	}
}

// FromDashboardSnapshot converts a domain snapshot to its DTO.
func FromDashboardSnapshot(snap *domain.DashboardSnapshot) DashboardSnapshotDTO {
	tasks := make([]TaskDTO, 0, len(snap.Tasks))
	for _, t := range snap.Tasks {
		tasks = append(tasks, FromDashboardTask(t))
	}
	states := make([]WorkflowStateDTO, 0, len(snap.WorkflowStates))
	for _, w := range snap.WorkflowStates {
		states = append(states, FromWorkflowStateRow(w))
	}
	return DashboardSnapshotDTO{
		Tasks:          tasks,
		Summary:        FromTaskSummary(snap.Summary),
		WorkflowStates: states,
		LastRefreshed:  snap.LastRefreshed.Format(time.RFC3339),
	}
}
