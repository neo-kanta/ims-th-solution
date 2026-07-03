package query

import (
	"context"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	integrationperm "github.com/neo-kanta/ims-th-solution/backend/internal/integration/permission"
)

// TaskRepository is the persistence port consumed by dashboard queries.
type TaskRepository interface {
	FetchResearchTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error)
	FetchWorkflowTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error)
	FetchWorkflowStates(ctx context.Context, contractIDs []string) ([]domain.WorkflowStateRow, error)
	FetchComplianceTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error)
}

// GetDashboardSnapshotHandler assembles the full dashboard read model.
type GetDashboardSnapshotHandler struct {
	iam  IAMPort
	repo TaskRepository
}

// NewGetDashboardSnapshotHandler creates the handler.
func NewGetDashboardSnapshotHandler(iam IAMPort, repo TaskRepository) *GetDashboardSnapshotHandler {
	return &GetDashboardSnapshotHandler{iam: iam, repo: repo}
}

// Execute fetches and assembles the dashboard snapshot for the given user.
func (h *GetDashboardSnapshotHandler) Execute(ctx context.Context, userID string) (*domain.DashboardSnapshot, error) {
	ok, err := h.iam.HasFunctionPermission(ctx, userID, integrationperm.DashboardView)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &domain.DashboardSnapshot{
			Tasks:          []domain.DashboardTask{},
			WorkflowStates: []domain.WorkflowStateRow{},
			LastRefreshed:  time.Now().UTC(),
		}, nil
	}

	contractIDs, err := h.iam.GetAccessibleContracts(ctx, userID)
	if err != nil {
		return nil, err
	}

	research, err := h.repo.FetchResearchTasks(ctx, contractIDs)
	if err != nil {
		return nil, err
	}

	workflowTasks, err := h.repo.FetchWorkflowTasks(ctx, contractIDs)
	if err != nil {
		return nil, err
	}

	workflowStates, err := h.repo.FetchWorkflowStates(ctx, contractIDs)
	if err != nil {
		return nil, err
	}

	compliance, err := h.repo.FetchComplianceTasks(ctx, contractIDs)
	if err != nil {
		return nil, err
	}

	tasks := make([]domain.DashboardTask, 0, len(research)+len(workflowTasks)+len(compliance))
	tasks = append(tasks, research...)
	tasks = append(tasks, workflowTasks...)
	tasks = append(tasks, compliance...)
	sortTasksByPriority(tasks)

	return &domain.DashboardSnapshot{
		Tasks:          tasks,
		Summary:        buildSummary(tasks),
		WorkflowStates: workflowStates,
		LastRefreshed:  time.Now().UTC(),
	}, nil
}
