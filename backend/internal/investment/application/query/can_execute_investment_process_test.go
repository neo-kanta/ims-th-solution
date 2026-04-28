package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

type fakeProcessGuardRepo struct {
	setting          *entity.WorkflowDaySetting
	blockingDecision *entity.BlockingControlDecision
	assignment       *entity.ProcessAssignmentMatch
	err              error
}

func (r *fakeProcessGuardRepo) GetActiveDaySetting(context.Context, uuid.UUID, time.Time) (*entity.WorkflowDaySetting, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.setting, nil
}

func (r *fakeProcessGuardRepo) FindBlockingControlDecision(context.Context, uuid.UUID, time.Time, vo.ProcessStepKey) (*entity.BlockingControlDecision, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.blockingDecision, nil
}

func (r *fakeProcessGuardRepo) FindUserProcessAssignment(context.Context, uuid.UUID, uuid.UUID, time.Time, vo.ProcessStepKey) (*entity.ProcessAssignmentMatch, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.assignment, nil
}

type fakeWorkflowState struct {
	tradeAllowed bool
	locked       bool
	err          error
}

func (w fakeWorkflowState) IsTradeAllowed(context.Context, uuid.UUID, time.Time) (bool, error) {
	if w.err != nil {
		return false, w.err
	}
	return w.tradeAllowed, nil
}

func (w fakeWorkflowState) IsTransactionLocked(context.Context, uuid.UUID, time.Time) (bool, error) {
	if w.err != nil {
		return false, w.err
	}
	return w.locked, nil
}

func TestCanExecuteInvestmentProcessHandler_Allows(t *testing.T) {
	userID := uuid.New()
	repo := &fakeProcessGuardRepo{
		setting: defaultQuerySetting(),
		assignment: &entity.ProcessAssignmentMatch{
			AssignmentID: uuid.New(),
			ProcessStep:  vo.ProcessStepAnalysisReport,
			UserID:       userID,
			CanExecute:   true,
		},
	}
	h := NewCanExecuteInvestmentProcessHandler(
		repo,
		fakeWorkflowState{tradeAllowed: true},
		func() time.Time { return bangkokQueryTime(2026, 4, 24, 10, 0) },
	)

	res, err := h.Handle(context.Background(), validCanExecuteRequest(userID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Fatalf("expected allowed, got reasons %+v", res.Reasons)
	}
	if res.Assignment == nil {
		t.Fatalf("expected assignment snapshot")
	}
}

func TestCanExecuteInvestmentProcessHandler_BlocksWhenUnassigned(t *testing.T) {
	h := NewCanExecuteInvestmentProcessHandler(
		&fakeProcessGuardRepo{setting: defaultQuerySetting()},
		fakeWorkflowState{tradeAllowed: true},
		func() time.Time { return bangkokQueryTime(2026, 4, 24, 10, 0) },
	)

	res, err := h.Handle(context.Background(), validCanExecuteRequest(uuid.New()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertQueryReason(t, res, policy.ReasonUserNotAssigned)
}

func TestCanExecuteInvestmentProcessHandler_Validation(t *testing.T) {
	h := NewCanExecuteInvestmentProcessHandler(&fakeProcessGuardRepo{}, fakeWorkflowState{}, nil)
	_, err := h.Handle(context.Background(), CanExecuteInvestmentProcessRequest{
		ContractID:   uuid.New(),
		BusinessDate: time.Now(),
		ProcessStep:  vo.ProcessStepAnalysisReport,
	})
	var invalid *domain.ErrInvalidProcessGuardRequest
	if !errors.As(err, &invalid) {
		t.Fatalf("expected invalid request error, got %v", err)
	}
	if invalid.Field != "user_id" {
		t.Fatalf("expected user_id field, got %s", invalid.Field)
	}
}

func validCanExecuteRequest(userID uuid.UUID) CanExecuteInvestmentProcessRequest {
	return CanExecuteInvestmentProcessRequest{
		UserID:       userID,
		ContractID:   uuid.New(),
		BusinessDate: time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC),
		ProcessStep:  vo.ProcessStepAnalysisReport,
	}
}

func defaultQuerySetting() *entity.WorkflowDaySetting {
	return &entity.WorkflowDaySetting{
		ID:            uuid.New(),
		Name:          "Default Thai Business Day",
		Timezone:      "Asia/Bangkok",
		WorkStartTime: 8*time.Hour + 30*time.Minute,
		WorkEndTime:   17*time.Hour + 30*time.Minute,
	}
}

func bangkokQueryTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return time.Date(year, month, day, hour, minute, 0, 0, loc).UTC()
}

func assertQueryReason(t *testing.T, res *CanExecuteInvestmentProcessResult, code string) {
	t.Helper()
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.Allowed {
		t.Fatalf("expected blocked result")
	}
	for _, reason := range res.Reasons {
		if reason.Code == code {
			return
		}
	}
	t.Fatalf("expected reason %s, got %+v", code, res.Reasons)
}
