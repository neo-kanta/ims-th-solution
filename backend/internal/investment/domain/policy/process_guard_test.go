package policy

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func TestCanExecuteInvestmentProcess_AllowsWhenAllGuardsPass(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 10, 0),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       true,
		TransactionLocked:  false,
		Assignment: &entity.ProcessAssignmentMatch{
			AssignmentID: uuid.New(),
			ProcessStep:  vo.ProcessStepAnalysisReport,
			UserID:       uuid.New(),
			CanExecute:   true,
		},
	})

	if !result.Allowed {
		t.Fatalf("expected allowed, got reasons: %+v", result.Reasons)
	}
}

func TestCanExecuteInvestmentProcess_BlocksOutsideWorkingWindow(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 7, 59),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       true,
		Assignment:         executableAssignment(),
	})

	assertHasReason(t, result, ReasonOutsideWorkingWindow)
}

func TestCanExecuteInvestmentProcess_BlocksWhenWorkflowNotOpen(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 10, 0),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       false,
		Assignment:         executableAssignment(),
	})

	assertHasReason(t, result, ReasonWorkflowDayNotOpen)
}

func TestCanExecuteInvestmentProcess_BlocksWhenTransactionsLocked(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 10, 0),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       false,
		TransactionLocked:  true,
		Assignment:         executableAssignment(),
	})

	assertHasReason(t, result, ReasonTransactionsLocked)
}

func TestCanExecuteInvestmentProcess_BlocksRejectedDecision(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 10, 0),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       true,
		Assignment:         executableAssignment(),
		BlockingDecision: &entity.BlockingControlDecision{
			ID:             uuid.New(),
			DecisionScope:  "INVESTMENT_PROCESS",
			DecisionLevel:  "MANAGER",
			DecisionStatus: "REJECTED",
			Reason:         "manager rejected the decision",
		},
	})

	assertHasReason(t, result, ReasonBlockedByRejection)
}

func TestCanExecuteInvestmentProcess_BlocksUnassignedUser(t *testing.T) {
	result := CanExecuteInvestmentProcess(ProcessGuardInput{
		Now:                bangkokTime(2026, 4, 24, 10, 0),
		WorkflowDaySetting: defaultSetting(),
		TradeAllowed:       true,
	})

	assertHasReason(t, result, ReasonUserNotAssigned)
}

func defaultSetting() *entity.WorkflowDaySetting {
	return &entity.WorkflowDaySetting{
		ID:               uuid.New(),
		Name:             "Default Thai Business Day",
		Timezone:         "Asia/Bangkok",
		WorkStartTime:    8*time.Hour + 30*time.Minute,
		WorkEndTime:      17*time.Hour + 30*time.Minute,
		BlockOnRejection: true,
		EffectiveFrom:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func executableAssignment() *entity.ProcessAssignmentMatch {
	return &entity.ProcessAssignmentMatch{
		AssignmentID: uuid.New(),
		ProcessStep:  vo.ProcessStepAnalysisReport,
		UserID:       uuid.New(),
		CanExecute:   true,
	}
}

func bangkokTime(year int, month time.Month, day int, hour int, minute int) time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return time.Date(year, month, day, hour, minute, 0, 0, loc).UTC()
}

func assertHasReason(t *testing.T, result ProcessGuardResult, code string) {
	t.Helper()
	if result.Allowed {
		t.Fatalf("expected blocked result")
	}
	for _, reason := range result.Reasons {
		if reason.Code == code {
			return
		}
	}
	t.Fatalf("expected reason %s, got %+v", code, result.Reasons)
}
