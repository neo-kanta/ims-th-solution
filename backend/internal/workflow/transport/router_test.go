package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/handler"
)

type routeSmokeDayRepo struct{}

func (routeSmokeDayRepo) GetByBusinessDate(context.Context, time.Time) (*entity.WorkflowDay, error) {
	return nil, nil
}

func (r routeSmokeDayRepo) GetByContractDate(ctx context.Context, _ uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error) {
	return r.GetByBusinessDate(ctx, businessDate)
}

func (routeSmokeDayRepo) GetForUpdateByBusinessDate(context.Context, pgx.Tx, time.Time) (*entity.WorkflowDay, error) {
	return nil, nil
}

func (r routeSmokeDayRepo) GetForUpdate(ctx context.Context, tx pgx.Tx, _ uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error) {
	return r.GetForUpdateByBusinessDate(ctx, tx, businessDate)
}

func (routeSmokeDayRepo) Insert(context.Context, pgx.Tx, *entity.WorkflowDay) error {
	return nil
}

func (routeSmokeDayRepo) UpdateState(context.Context, pgx.Tx, *entity.WorkflowDay) error {
	return nil
}

func (routeSmokeDayRepo) ResetToNotStarted(context.Context, pgx.Tx, *entity.WorkflowDay) error {
	return nil
}

func (routeSmokeDayRepo) ListByState(context.Context, vo.WorkflowState, time.Time, int, int) ([]*entity.WorkflowDay, int64, error) {
	return nil, 0, nil
}

type routeSmokeLogRepo struct{}

func (routeSmokeLogRepo) Append(context.Context, pgx.Tx, *entity.WorkflowTransition) error {
	return nil
}

func (routeSmokeLogRepo) ListByBusinessDate(context.Context, time.Time) ([]*entity.WorkflowTransition, error) {
	return nil, nil
}

func (r routeSmokeLogRepo) ListByContractDate(ctx context.Context, _ uuid.UUID, businessDate time.Time) ([]*entity.WorkflowTransition, error) {
	return r.ListByBusinessDate(ctx, businessDate)
}

func (routeSmokeLogRepo) ListByBusinessDatePaginated(_ context.Context, req domain.TransitionLogPageRequest) (*domain.TransitionLogPageResult, error) {
	return &domain.TransitionLogPageResult{Page: req.Page, PageSize: req.PageSize}, nil
}

func (r routeSmokeLogRepo) ListByContractDatePaginated(ctx context.Context, req domain.TransitionLogPageRequest) (*domain.TransitionLogPageResult, error) {
	return r.ListByBusinessDatePaginated(ctx, req)
}

type routeSmokeSettingsRepo struct{}

func (routeSmokeSettingsRepo) ListByOperationType(context.Context, string) ([]*entity.WorkflowApprovalSetting, error) {
	return nil, nil
}

func (routeSmokeSettingsRepo) ListAll(context.Context) ([]*entity.WorkflowApprovalSetting, error) {
	return nil, nil
}

func (routeSmokeSettingsRepo) Upsert(context.Context, pgx.Tx, string, []*entity.WorkflowApprovalSetting) error {
	return nil
}

type routeSmokeCalendar struct{}

func (routeSmokeCalendar) IsBusinessDay(context.Context, time.Time) (bool, string, error) {
	return true, "", nil
}

func (routeSmokeCalendar) PreviousBusinessDay(_ context.Context, date time.Time) (time.Time, error) {
	return date.AddDate(0, 0, -1), nil
}

func TestRegisterRoutes_DailyRouteMountedUnderAPIV1(t *testing.T) {
	dayRepo := routeSmokeDayRepo{}
	logRepo := routeSmokeLogRepo{}
	settingsRepo := routeSmokeSettingsRepo{}
	getDaily := query.NewGetDailyWorkflowHandler(dayRepo, logRepo, settingsRepo, routeSmokeCalendar{})

	workflowHandler := handler.NewWorkflowHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	dailyHandler := handler.NewDailyHandler(
		getDaily,
		settingsRepo,
		logRepo,
		nil,
		nil,
		nil, nil, nil, nil, nil, nil, nil, nil,
	)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		transport.RegisterRoutes(r, workflowHandler, dailyHandler, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflow/daily?businessDate=2026-04-24", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("expected route to be registered, got 404 body: %s", w.Body.String())
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data, _ := resp["data"].(map[string]any)
	if data["businessDate"] != "2026-04-24" {
		t.Fatalf("expected businessDate 2026-04-24, got %v", data["businessDate"])
	}
}
