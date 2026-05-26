package query

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	integrationperm "github.com/neo-kanta/ims-th-solution/backend/internal/integration/permission"
)

// MyTasksFilter carries optional filter criteria for the task list endpoint.
type MyTasksFilter struct {
	Module   string
	Priority string
	Status   string
}

// GetMyTasksHandler returns a filtered, sorted task list with a summary.
type GetMyTasksHandler struct {
	iam  IAMPort
	repo TaskRepository
}

// NewGetMyTasksHandler creates the handler.
func NewGetMyTasksHandler(iam IAMPort, repo TaskRepository) *GetMyTasksHandler {
	return &GetMyTasksHandler{iam: iam, repo: repo}
}

// Execute returns tasks filtered by the given criteria.
func (h *GetMyTasksHandler) Execute(
	ctx context.Context,
	userID string,
	filter MyTasksFilter,
) ([]domain.DashboardTask, domain.TaskSummary, error) {
	ok, err := h.iam.HasFunctionPermission(ctx, userID, integrationperm.DashboardView)
	if err != nil {
		return nil, domain.TaskSummary{}, err
	}
	if !ok {
		return []domain.DashboardTask{}, domain.TaskSummary{}, nil
	}

	contractIDs, err := h.iam.GetAccessibleContracts(ctx, userID)
	if err != nil {
		return nil, domain.TaskSummary{}, err
	}

	research, err := h.repo.FetchResearchTasks(ctx, contractIDs)
	if err != nil {
		return nil, domain.TaskSummary{}, err
	}

	workflowTasks, err := h.repo.FetchWorkflowTasks(ctx, contractIDs)
	if err != nil {
		return nil, domain.TaskSummary{}, err
	}

	compliance, err := h.repo.FetchComplianceTasks(ctx, contractIDs)
	if err != nil {
		return nil, domain.TaskSummary{}, err
	}

	all := make([]domain.DashboardTask, 0, len(research)+len(workflowTasks)+len(compliance))
	all = append(all, research...)
	all = append(all, workflowTasks...)
	all = append(all, compliance...)
	sortTasksByPriority(all)

	filtered := applyTaskFilter(all, filter)
	return filtered, buildSummary(filtered), nil
}

func applyTaskFilter(tasks []domain.DashboardTask, f MyTasksFilter) []domain.DashboardTask {
	if f.Module == "" && f.Priority == "" && f.Status == "" {
		return tasks
	}
	out := make([]domain.DashboardTask, 0, len(tasks))
	for _, t := range tasks {
		if f.Module != "" && t.Module != f.Module {
			continue
		}
		if f.Priority != "" && string(t.Priority) != f.Priority {
			continue
		}
		if f.Status != "" && string(t.Status) != f.Status {
			continue
		}
		out = append(out, t)
	}
	return out
}
