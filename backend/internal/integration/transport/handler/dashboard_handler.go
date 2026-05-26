package handler

import (
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// DashboardHandler handles dashboard and task feed HTTP endpoints.
type DashboardHandler struct {
	snapshot *query.GetDashboardSnapshotHandler
	tasks    *query.GetMyTasksHandler
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	snapshot *query.GetDashboardSnapshotHandler,
	tasks *query.GetMyTasksHandler,
) *DashboardHandler {
	return &DashboardHandler{
		snapshot: snapshot,
		tasks:    tasks,
	}
}

// GetDashboardSnapshot handles GET /integration/dashboard/me.
//
// @Summary      Personal dashboard snapshot
// @Description  Returns the caller's full dashboard read model: tasks, workflow states, and aggregate counts.
// @Tags         Integration
// @Produce      json
// @Success      200  {object}  httputil.SuccessResponse{data=response.DashboardSnapshotDTO}
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      500  {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /integration/dashboard/me [get]
func (h *DashboardHandler) GetDashboardSnapshot(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "authentication required")
		return
	}

	snap, err := h.snapshot.Execute(r.Context(), claims.Subject)
	if err != nil {
		httputil.InternalError(w, "failed to load dashboard")
		return
	}

	httputil.OK(w, response.FromDashboardSnapshot(snap))
}

// GetMyTasks handles GET /integration/tasks/my.
//
// @Summary      Personal task list
// @Description  Returns the caller's task feed, optionally filtered by module, priority, or status.
// @Tags         Integration
// @Produce      json
// @Param        module    query  string  false  "Filter by module (investment|workflow|compliance)"
// @Param        priority  query  string  false  "Filter by priority (HIGH|MEDIUM|LOW|INFO)"
// @Param        status    query  string  false  "Filter by status (PENDING|IN_PROGRESS|COMPLETED)"
// @Success      200  {object}  httputil.SuccessResponse{data=response.TaskListDTO}
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      500  {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /integration/tasks/my [get]
func (h *DashboardHandler) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "authentication required")
		return
	}

	q := r.URL.Query()
	filter := query.MyTasksFilter{
		Module:   q.Get("module"),
		Priority: q.Get("priority"),
		Status:   q.Get("status"),
	}

	tasks, summary, err := h.tasks.Execute(r.Context(), claims.Subject, filter)
	if err != nil {
		httputil.InternalError(w, "failed to load tasks")
		return
	}

	dtos := make([]response.TaskDTO, 0, len(tasks))
	for _, t := range tasks {
		dtos = append(dtos, response.FromDashboardTask(t))
	}

	httputil.OK(w, response.TaskListDTO{
		Tasks:   dtos,
		Summary: response.FromTaskSummary(summary),
	})
}

// GetMyTasksSummary handles GET /integration/tasks/my/summary.
//
// @Summary      Personal task summary
// @Description  Returns aggregate task counts for the caller without the full task list.
// @Tags         Integration
// @Produce      json
// @Success      200  {object}  httputil.SuccessResponse{data=response.TaskSummaryDTO}
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      500  {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /integration/tasks/my/summary [get]
func (h *DashboardHandler) GetMyTasksSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "authentication required")
		return
	}

	_, summary, err := h.tasks.Execute(r.Context(), claims.Subject, query.MyTasksFilter{})
	if err != nil {
		httputil.InternalError(w, "failed to load task summary")
		return
	}

	httputil.OK(w, response.FromTaskSummary(summary))
}
