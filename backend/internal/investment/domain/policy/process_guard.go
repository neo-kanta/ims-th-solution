// Package policy contains pure investment-domain guard logic.
package policy

import (
	"fmt"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

const (
	ReasonWorkflowDayNotOpen      = "INVESTMENT_WORKFLOW_DAY_NOT_OPEN"
	ReasonTransactionsLocked      = "INVESTMENT_TRANSACTIONS_LOCKED"
	ReasonOutsideWorkingWindow    = "INVESTMENT_OUTSIDE_WORKING_WINDOW"
	ReasonBlockedByRejection      = "INVESTMENT_BLOCKED_BY_REJECTION"
	ReasonUserNotAssigned         = "INVESTMENT_USER_NOT_ASSIGNED"
	ReasonWorkflowSettingNotFound = "INVESTMENT_WORKFLOW_SETTING_NOT_FOUND"
)

// BlockingReason explains one condition that prevents investment work.
type BlockingReason struct {
	Code    string
	Message string
	Details map[string]any
}

// ProcessGuardInput contains all facts needed to decide if investment work may
// execute. It is deliberately I/O-free so the policy is simple to test.
type ProcessGuardInput struct {
	Now time.Time

	WorkflowDaySetting *entity.WorkflowDaySetting
	TradeAllowed       bool
	TransactionLocked  bool

	BlockingDecision *entity.BlockingControlDecision
	Assignment       *entity.ProcessAssignmentMatch
}

// ProcessGuardResult is the pure decision returned by CanExecuteInvestmentProcess.
type ProcessGuardResult struct {
	Allowed bool
	Reasons []BlockingReason
}

// CanExecuteInvestmentProcess evaluates the investment-process execution guard.
func CanExecuteInvestmentProcess(in ProcessGuardInput) ProcessGuardResult {
	var reasons []BlockingReason

	if in.WorkflowDaySetting == nil {
		reasons = append(reasons, BlockingReason{
			Code:    ReasonWorkflowSettingNotFound,
			Message: "no active workflow day setting was found for this contract and business date",
		})
	} else if !isWithinWorkingWindow(in.Now, in.WorkflowDaySetting) {
		reasons = append(reasons, BlockingReason{
			Code: ReasonOutsideWorkingWindow,
			Message: fmt.Sprintf(
				"investment process is outside the configured working window %s-%s %s",
				formatHHMM(in.WorkflowDaySetting.WorkStartTime),
				formatHHMM(in.WorkflowDaySetting.WorkEndTime),
				in.WorkflowDaySetting.Timezone,
			),
			Details: map[string]any{
				"timezone":  in.WorkflowDaySetting.Timezone,
				"startTime": formatHHMM(in.WorkflowDaySetting.WorkStartTime),
				"endTime":   formatHHMM(in.WorkflowDaySetting.WorkEndTime),
			},
		})
	}

	if !in.TradeAllowed {
		reasons = append(reasons, BlockingReason{
			Code:    ReasonWorkflowDayNotOpen,
			Message: "workflow day is not open for investment processing",
		})
	}

	if in.TransactionLocked {
		reasons = append(reasons, BlockingReason{
			Code:    ReasonTransactionsLocked,
			Message: "transactions are locked for this contract and business date",
		})
	}

	if in.BlockingDecision != nil {
		reasons = append(reasons, BlockingReason{
			Code:    ReasonBlockedByRejection,
			Message: "manager or high-level rejection blocks this investment process",
			Details: map[string]any{
				"decisionId":     in.BlockingDecision.ID.String(),
				"decisionScope":  in.BlockingDecision.DecisionScope,
				"decisionLevel":  in.BlockingDecision.DecisionLevel,
				"decisionStatus": in.BlockingDecision.DecisionStatus,
				"reason":         in.BlockingDecision.Reason,
			},
		})
	}

	if in.Assignment == nil || !in.Assignment.CanExecute {
		reasons = append(reasons, BlockingReason{
			Code:    ReasonUserNotAssigned,
			Message: "user is not assigned to execute this investment process step",
		})
	}

	return ProcessGuardResult{
		Allowed: len(reasons) == 0,
		Reasons: reasons,
	}
}

func isWithinWorkingWindow(now time.Time, setting *entity.WorkflowDaySetting) bool {
	if setting == nil {
		return false
	}
	loc, err := time.LoadLocation(setting.Timezone)
	if err != nil {
		loc = time.FixedZone(setting.Timezone, 7*60*60)
	}
	local := now.In(loc)
	elapsed := time.Duration(local.Hour())*time.Hour +
		time.Duration(local.Minute())*time.Minute +
		time.Duration(local.Second())*time.Second
	return elapsed >= setting.WorkStartTime && elapsed <= setting.WorkEndTime
}

func formatHHMM(d time.Duration) string {
	totalMinutes := int(d / time.Minute)
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}
