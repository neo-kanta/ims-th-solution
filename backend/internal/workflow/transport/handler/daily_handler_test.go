package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// ─────────────────────────────────────────────────────────────────────────────
// Stubs
// ─────────────────────────────────────────────────────────────────────────────

type stubApprovalSettingRepo struct {
	byOp map[string][]*entity.WorkflowApprovalSetting
	all  []*entity.WorkflowApprovalSetting
}

func (s *stubApprovalSettingRepo) ListByOperationType(_ context.Context, op string) ([]*entity.WorkflowApprovalSetting, error) {
	return s.byOp[op], nil
}

func (s *stubApprovalSettingRepo) ListAll(_ context.Context) ([]*entity.WorkflowApprovalSetting, error) {
	return s.all, nil
}

func (s *stubApprovalSettingRepo) Upsert(_ context.Context, _ pgx.Tx, op string, settings []*entity.WorkflowApprovalSetting) error {
	if s.byOp == nil {
		s.byOp = make(map[string][]*entity.WorkflowApprovalSetting)
	}
	for k := range s.byOp {
		if k == op {
			delete(s.byOp, k)
		}
	}
	s.byOp[op] = settings
	s.all = nil
	for _, v := range s.byOp {
		s.all = append(s.all, v...)
	}
	return nil
}

var _ domain.WorkflowApprovalSettingRepository = (*stubApprovalSettingRepo)(nil)

type stubTransitionLogRepo struct {
	rows []*entity.WorkflowTransition
}

func (s *stubTransitionLogRepo) Append(_ context.Context, _ pgx.Tx, _ *entity.WorkflowTransition) error {
	return nil
}

func (s *stubTransitionLogRepo) ListByBusinessDate(_ context.Context, _ time.Time) ([]*entity.WorkflowTransition, error) {
	return s.rows, nil
}

func (s *stubTransitionLogRepo) ListByContractDate(_ context.Context, _ uuid.UUID, _ time.Time) ([]*entity.WorkflowTransition, error) {
	return s.rows, nil
}

func (s *stubTransitionLogRepo) ListByBusinessDatePaginated(_ context.Context, req domain.TransitionLogPageRequest) (*domain.TransitionLogPageResult, error) {
	return &domain.TransitionLogPageResult{
		Transitions: s.rows,
		Total:       int64(len(s.rows)),
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (s *stubTransitionLogRepo) ListByContractDatePaginated(_ context.Context, req domain.TransitionLogPageRequest) (*domain.TransitionLogPageResult, error) {
	return &domain.TransitionLogPageResult{
		Transitions: s.rows,
		Total:       int64(len(s.rows)),
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

var _ domain.TransitionLogRepository = (*stubTransitionLogRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

type stubWorkflowDayRepo struct {
	byDate map[string]*entity.WorkflowDay
}

func (s *stubWorkflowDayRepo) GetByBusinessDate(_ context.Context, businessDate time.Time) (*entity.WorkflowDay, error) {
	return s.byDate[businessDate.Format("2006-01-02")], nil
}

func (s *stubWorkflowDayRepo) GetByContractDate(_ context.Context, _ uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error) {
	return s.GetByBusinessDate(context.Background(), businessDate)
}

func (s *stubWorkflowDayRepo) GetForUpdateByBusinessDate(_ context.Context, _ pgx.Tx, businessDate time.Time) (*entity.WorkflowDay, error) {
	return s.GetByBusinessDate(context.Background(), businessDate)
}

func (s *stubWorkflowDayRepo) GetForUpdate(_ context.Context, _ pgx.Tx, _ uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error) {
	return s.GetByBusinessDate(context.Background(), businessDate)
}

func (s *stubWorkflowDayRepo) Insert(_ context.Context, _ pgx.Tx, day *entity.WorkflowDay) error {
	if s.byDate == nil {
		s.byDate = make(map[string]*entity.WorkflowDay)
	}
	s.byDate[day.BusinessDate.Format("2006-01-02")] = day
	return nil
}

func (s *stubWorkflowDayRepo) UpdateState(_ context.Context, _ pgx.Tx, day *entity.WorkflowDay) error {
	s.byDate[day.BusinessDate.Format("2006-01-02")] = day
	return nil
}

func (s *stubWorkflowDayRepo) ResetToNotStarted(_ context.Context, _ pgx.Tx, day *entity.WorkflowDay) error {
	s.byDate[day.BusinessDate.Format("2006-01-02")] = day
	return nil
}

func (s *stubWorkflowDayRepo) ListByState(_ context.Context, _ vo.WorkflowState, _ time.Time, _, _ int) ([]*entity.WorkflowDay, int64, error) {
	return nil, 0, nil
}

var _ domain.WorkflowDayRepository = (*stubWorkflowDayRepo)(nil)

type stubCalendar struct{}

func (stubCalendar) IsBusinessDay(context.Context, time.Time) (bool, string, error) {
	return true, "", nil
}

func (stubCalendar) PreviousBusinessDay(_ context.Context, date time.Time) (time.Time, error) {
	return date.AddDate(0, 0, -1), nil
}

func withJWT(r *http.Request, userID uuid.UUID, username string, roles []string) *http.Request {
	claims := &middleware.UserClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{Subject: userID.String()},
		Username:         username,
		Roles:            roles,
	}
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, claims)
	return r.WithContext(ctx)
}

func buildMinimalDailyHandler(settingRepo domain.WorkflowApprovalSettingRepository, logRepo domain.TransitionLogRepository) *handler.DailyHandler {
	h := handler.NewDailyHandler(
		nil, // getDailyWorkflow — not tested here
		settingRepo,
		logRepo,
		nil,
		nil, // pool — not needed for pure-auth tests
		nil, nil, nil, nil, nil, nil, nil, nil,
	)
	return h
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests: GET /workflow/settings
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDailyWorkflow_WorksWithoutContractCode(t *testing.T) {
	dayRepo := &stubWorkflowDayRepo{byDate: map[string]*entity.WorkflowDay{}}
	logRepo := &stubTransitionLogRepo{}
	settingRepo := &stubApprovalSettingRepo{byOp: map[string][]*entity.WorkflowApprovalSetting{}}
	getDaily := query.NewGetDailyWorkflowHandler(dayRepo, logRepo, settingRepo, stubCalendar{})
	h := handler.NewDailyHandler(
		getDaily,
		settingRepo,
		logRepo,
		nil,
		nil,
		nil, nil, nil, nil, nil, nil, nil, nil,
	)

	r := httptest.NewRequest(http.MethodGet, "/workflow/daily?businessDate=2026-04-24", nil)
	r = withJWT(r, uuid.New(), "user", []string{"Staff"})
	w := httptest.NewRecorder()

	h.GetDailyWorkflow(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d - body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data, _ := resp["data"].(map[string]any)
	if data["businessDate"] != "2026-04-24" {
		t.Fatalf("expected businessDate 2026-04-24, got %v", data["businessDate"])
	}
	if data["currentState"] != "NOT_STARTED" {
		t.Fatalf("expected NOT_STARTED, got %v", data["currentState"])
	}
	if _, ok := data["contractCode"]; ok {
		t.Fatal("response must not contain contractCode")
	}
}

func TestGetSettings_ReturnsActiveApprovers(t *testing.T) {
	repo := &stubApprovalSettingRepo{
		all: []*entity.WorkflowApprovalSetting{
			{
				ID:                  uuid.New(),
				OperationType:       vo.APIOpManagerApprove,
				ApprovalMode:        entity.ApprovalModeAnyOf,
				ApproverAccountCode: "jane.smith",
				ApproverUsername:    "Jane Smith",
				IsActive:            true,
				UpdatedAt:           time.Now(),
			},
		},
	}
	h := buildMinimalDailyHandler(repo, &stubTransitionLogRepo{})

	r := httptest.NewRequest(http.MethodGet, "/workflow/settings", nil)
	r = withJWT(r, uuid.New(), "admin.user", []string{"Admin"})
	w := httptest.NewRecorder()

	h.GetSettings(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal("response is not JSON:", err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data envelope, got body: %s", w.Body.String())
	}
	settings, ok := data["settings"].([]any)
	if !ok || len(settings) == 0 {
		t.Fatal("expected at least one entry in settings")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests: PUT /workflow/settings — admin gate
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateSettings_AdminSucceeds(t *testing.T) {
	repo := &stubApprovalSettingRepo{byOp: map[string][]*entity.WorkflowApprovalSetting{}}
	h := buildMinimalDailyHandler(repo, &stubTransitionLogRepo{})

	body, _ := json.Marshal(map[string]any{
		"operationType": vo.APIOpManagerApprove,
		"approvers": []map[string]any{
			{"accountCode": "jane.smith", "username": "Jane Smith", "role": "FundManager"},
		},
	})
	r := httptest.NewRequest(http.MethodPut, "/workflow/settings", bytes.NewReader(body))
	r = withJWT(r, uuid.New(), "admin.user", []string{"Admin"})
	w := httptest.NewRecorder()

	h.UpdateSettings(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUpdateSettings_NonAdminForbidden(t *testing.T) {
	repo := &stubApprovalSettingRepo{}
	h := buildMinimalDailyHandler(repo, &stubTransitionLogRepo{})

	body, _ := json.Marshal(map[string]any{
		"operationType": vo.APIOpManagerApprove,
		"approvers":     []any{},
	})
	r := httptest.NewRequest(http.MethodPut, "/workflow/settings", bytes.NewReader(body))
	r = withJWT(r, uuid.New(), "regular.user", []string{"FundManager"})
	w := httptest.NewRecorder()

	h.UpdateSettings(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests: GET /workflow/transition-rules
// ─────────────────────────────────────────────────────────────────────────────

func TestGetTransitionRules_Returns8Rules(t *testing.T) {
	h := buildMinimalDailyHandler(&stubApprovalSettingRepo{}, &stubTransitionLogRepo{})

	r := httptest.NewRequest(http.MethodGet, "/workflow/transition-rules", nil)
	r = withJWT(r, uuid.New(), "user", []string{"Staff"})
	w := httptest.NewRecorder()

	h.GetTransitionRules(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data, _ := resp["data"].(map[string]any)
	rules, _ := data["rules"].([]any)
	if len(rules) != 8 {
		t.Errorf("expected 8 transition rules, got %d", len(rules))
	}
}

func TestGetTransitionRules_UsesCanonicalAPINames(t *testing.T) {
	h := buildMinimalDailyHandler(&stubApprovalSettingRepo{}, &stubTransitionLogRepo{})

	r := httptest.NewRequest(http.MethodGet, "/workflow/transition-rules", nil)
	r = withJWT(r, uuid.New(), "user", []string{"Staff"})
	w := httptest.NewRecorder()

	h.GetTransitionRules(w, r)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	rules, _ := data["rules"].([]any)

	// Verify none of the legacy internal names appear
	forbidden := map[string]bool{
		"OPEN_DAY": true, "APPROVE": true, "CLOSE_TRANSACTIONS": true,
		"ROLLBACK_ACCOUNTING_CLOSE": true, "CANCEL_DAY_START": true,
		"CANCEL_APPROVAL": true, "DAY_OPEN": true, "MANAGER_APPROVED_END_OF_DAY": true,
	}
	for _, rawRule := range rules {
		rule := rawRule.(map[string]any)
		opType := rule["operationType"].(string)
		if forbidden[opType] {
			t.Errorf("transition rules contain legacy name %q; expected canonical API name", opType)
		}
		from := rule["fromState"].(string)
		if forbidden[from] {
			t.Errorf("transition rules contain legacy fromState %q", from)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests: GET /workflow/daily/transitions pagination
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDailyTransitions_PaginationDefaults(t *testing.T) {
	logRepo := &stubTransitionLogRepo{
		rows: []*entity.WorkflowTransition{
			{
				ID:               uuid.New(),
				Action:           vo.ActionOpenDay,
				FromState:        vo.StateNotStarted,
				ToState:          vo.StateDayOpen,
				ActorUsername:    "john.doe",
				ActorAccountCode: "john.doe",
				OccurredAt:       time.Now().UTC(),
			},
		},
	}
	h := buildMinimalDailyHandler(&stubApprovalSettingRepo{}, logRepo)

	r := httptest.NewRequest(http.MethodGet, "/workflow/daily/transitions?businessDate=2026-06-15", nil)
	r = withJWT(r, uuid.New(), "user", []string{"Staff"})
	w := httptest.NewRecorder()

	h.GetDailyTransitions(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	if _, ok := data["contractCode"]; ok {
		t.Fatal("response must not contain contractCode")
	}

	// Response must include pagination fields
	if _, ok := data["total"]; !ok {
		t.Error("response missing 'total' field")
	}
	if _, ok := data["page"]; !ok {
		t.Error("response missing 'page' field")
	}
	if _, ok := data["pageSize"]; !ok {
		t.Error("response missing 'pageSize' field")
	}

	// Action must be API-normalised
	transitions, _ := data["transitions"].([]any)
	if len(transitions) > 0 {
		entry := transitions[0].(map[string]any)
		opType := entry["operationType"].(string)
		if opType != vo.APIOpStartInvestmentDay {
			t.Errorf("expected operationType %q, got %q", vo.APIOpStartInvestmentDay, opType)
		}
		fromState := entry["fromState"].(string)
		if fromState != "NOT_STARTED" {
			t.Errorf("expected fromState NOT_STARTED, got %q", fromState)
		}
		toState := entry["toState"].(string)
		if toState != "INVESTMENT_DAY_STARTED" {
			t.Errorf("expected toState INVESTMENT_DAY_STARTED (normalized from DAY_OPEN), got %q", toState)
		}
	}
}

func TestGetDailyTransitions_ActorAccountCodeIsPopulated(t *testing.T) {
	logRepo := &stubTransitionLogRepo{
		rows: []*entity.WorkflowTransition{
			{
				ID:               uuid.New(),
				Action:           vo.ActionOpenDay,
				FromState:        vo.StateNotStarted,
				ToState:          vo.StateInvestmentDayStarted,
				ActorUsername:    "john.doe",
				ActorAccountCode: "john.doe",
				IsAdminOverride:  false,
				OccurredAt:       time.Now().UTC(),
			},
		},
	}
	h := buildMinimalDailyHandler(&stubApprovalSettingRepo{}, logRepo)

	r := httptest.NewRequest(http.MethodGet, "/workflow/daily/transitions?businessDate=2026-06-15", nil)
	r = withJWT(r, uuid.New(), "user", []string{"Staff"})
	w := httptest.NewRecorder()

	h.GetDailyTransitions(w, r)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]any)
	transitions, _ := data["transitions"].([]any)
	if len(transitions) == 0 {
		t.Fatal("expected at least one transition entry")
	}
	entry := transitions[0].(map[string]any)
	accountCode, _ := entry["executedByAccountCode"].(string)
	if accountCode == "" {
		t.Error("executedByAccountCode must not be empty in transition history")
	}
}
