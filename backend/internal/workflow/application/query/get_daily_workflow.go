package query

import (
	"context"
	"fmt"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

// DailyWorkflowRequest is the input for the aggregated daily workflow query.
type DailyWorkflowRequest struct {
	BusinessDate time.Time
}

// DailyTimelineEntry is one normalised row from the transition log.
type DailyTimelineEntry struct {
	TransitionID          string
	OperationType         string // canonical API name (e.g. "START_INVESTMENT_DAY")
	FromState             string // normalised API name
	ToState               string // normalised API name
	ExecutedByUsername    string
	ExecutedByAccountCode string
	IsAdminOverride       bool
	OccurredAt            time.Time
	Reason                *string
}

// DailyApproverEntry is one configured approver for an operation type.
type DailyApproverEntry struct {
	AccountCode string
	Username    string
	Role        string
}

// DailyModuleStatus reports whether a supporting module is fully operational.
type DailyModuleStatus struct {
	Ready  bool
	Reason string
}

// DailyAuditSummary is a compact summary of today's activity.
type DailyAuditSummary struct {
	TotalTransitions int
	LastTransitionAt *time.Time
	LastTransitionBy string
}

// DailyWorkflowResult is the aggregated response for GET /api/workflow/daily.
type DailyWorkflowResult struct {
	BusinessDate      string
	CurrentState      string // canonical API name
	IsToday           bool
	Persisted         bool
	AllowedOperations []string // canonical API operation names
	BlockedReasons    []BlockingReasonItem

	Timeline []DailyTimelineEntry

	// Approvers maps operationType → configured approver list.
	Approvers map[string][]DailyApproverEntry

	// Settings maps operationType → approval mode + approvers.
	Settings map[string]DailyOperationSetting

	AuditSummary DailyAuditSummary

	ModuleReadiness map[string]DailyModuleStatus
}

// BlockingReasonItem explains why an operation is blocked.
type BlockingReasonItem struct {
	Code    string
	Message string
}

// DailyOperationSetting holds the approval configuration for one operation type.
type DailyOperationSetting struct {
	ApprovalMode string
	Approvers    []DailyApproverEntry
}

// GetDailyWorkflowHandler aggregates workflow state, timeline, approver settings,
// and module readiness into a single response for the daily workflow UI.
type GetDailyWorkflowHandler struct {
	dayRepo      domain.WorkflowDayRepository
	logRepo      domain.TransitionLogRepository
	settingRepo  domain.WorkflowApprovalSettingRepository
	calendarPort ports.HolidayCalendarPort
}

// NewGetDailyWorkflowHandler creates the handler.
func NewGetDailyWorkflowHandler(
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	settingRepo domain.WorkflowApprovalSettingRepository,
	calendarPort ports.HolidayCalendarPort,
) *GetDailyWorkflowHandler {
	return &GetDailyWorkflowHandler{
		dayRepo:      dayRepo,
		logRepo:      logRepo,
		settingRepo:  settingRepo,
		calendarPort: calendarPort,
	}
}

// Handle executes the aggregated daily workflow query.
func (h *GetDailyWorkflowHandler) Handle(
	ctx context.Context,
	req DailyWorkflowRequest,
) (*DailyWorkflowResult, error) {
	// Current day state
	day, err := h.dayRepo.GetByBusinessDate(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("reading workflow day: %w", err)
	}

	// Previous business date (for blocked-reason computation)
	prevDate, err := h.calendarPort.PreviousBusinessDay(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("computing previous business day: %w", err)
	}

	// Determine allowed operations and blocked reasons
	currentState := vo.StateNotStarted
	persisted := false
	var blockedReasons []BlockingReasonItem

	if day != nil {
		currentState = day.CurrentState
		persisted = true
	} else {
		// Check previous day blocking conditions
		prevDay, err := h.dayRepo.GetByBusinessDate(ctx, prevDate)
		if err != nil {
			return nil, fmt.Errorf("reading previous workflow day: %w", err)
		}
		if prevDay != nil {
			switch {
			case prevDay.CurrentState.IsOpenForTrading():
				blockedReasons = append(blockedReasons, BlockingReasonItem{
					Code: "WORKFLOW_PREVIOUS_DAY_NOT_APPROVED",
					Message: fmt.Sprintf(
						"previous business day %s is still open for trading",
						prevDate.Format("2006-01-02"),
					),
				})
			case prevDay.CurrentState.IsManagerApproved():
				blockedReasons = append(blockedReasons, BlockingReasonItem{
					Code: "WORKFLOW_PREVIOUS_DAY_NOT_TRANSACTION_CLOSED",
					Message: fmt.Sprintf(
						"previous business day %s awaits transaction close",
						prevDate.Format("2006-01-02"),
					),
				})
			}
		}
	}

	internalAllowed := vo.AllowedActions(currentState)
	if len(blockedReasons) > 0 {
		internalAllowed = nil
	}
	allowedOps := make([]string, 0, len(internalAllowed))
	for _, a := range internalAllowed {
		allowedOps = append(allowedOps, a.ToAPIName())
	}

	// Determine isToday
	nowBKK := time.Now().In(clock.BangkokLocation)
	todayBKK := time.Date(nowBKK.Year(), nowBKK.Month(), nowBKK.Day(), 0, 0, 0, 0, clock.BangkokLocation)
	reqDateBKK := req.BusinessDate.In(clock.BangkokLocation)
	reqDateOnly := time.Date(reqDateBKK.Year(), reqDateBKK.Month(), reqDateBKK.Day(), 0, 0, 0, 0, clock.BangkokLocation)
	isToday := reqDateOnly.Equal(todayBKK)

	// Transition timeline
	transitions, err := h.logRepo.ListByBusinessDate(ctx, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("reading transition timeline: %w", err)
	}
	timeline := make([]DailyTimelineEntry, 0, len(transitions))
	for _, t := range transitions {
		entry := DailyTimelineEntry{
			TransitionID:          t.ID.String(),
			OperationType:         t.Action.ToAPIName(),
			FromState:             t.FromState.ToAPIName(),
			ToState:               t.ToState.ToAPIName(),
			ExecutedByUsername:    t.ActorUsername,
			ExecutedByAccountCode: t.ActorAccountCode,
			IsAdminOverride:       t.IsAdminOverride,
			OccurredAt:            t.OccurredAt,
			Reason:                t.Reason,
		}
		timeline = append(timeline, entry)
	}

	// Audit summary
	auditSummary := buildAuditSummary(transitions)

	// Approval settings — load for all allowed operations
	approvers := make(map[string][]DailyApproverEntry)
	settings := make(map[string]DailyOperationSetting)
	for _, opName := range allowedOps {
		settingList, err := h.settingRepo.ListByOperationType(ctx, opName)
		if err != nil {
			return nil, fmt.Errorf("reading approval settings for %s: %w", opName, err)
		}
		entries := make([]DailyApproverEntry, 0, len(settingList))
		for _, s := range settingList {
			entries = append(entries, DailyApproverEntry{
				AccountCode: s.ApproverAccountCode,
				Username:    s.ApproverUsername,
				Role:        s.ApproverRole,
			})
		}
		approvers[opName] = entries
		mode := "ANY_OF"
		if len(settingList) > 0 {
			mode = string(settingList[0].ApprovalMode)
		}
		settings[opName] = DailyOperationSetting{
			ApprovalMode: mode,
			Approvers:    entries,
		}
	}

	// Module readiness — approval module is scaffold
	moduleReadiness := map[string]DailyModuleStatus{
		"approval": {Ready: false, Reason: "approval module is scaffold"},
	}

	return &DailyWorkflowResult{
		BusinessDate:      req.BusinessDate.Format("2006-01-02"),
		CurrentState:      currentState.ToAPIName(),
		IsToday:           isToday,
		Persisted:         persisted,
		AllowedOperations: allowedOps,
		BlockedReasons:    blockedReasons,
		Timeline:          timeline,
		Approvers:         approvers,
		Settings:          settings,
		AuditSummary:      auditSummary,
		ModuleReadiness:   moduleReadiness,
	}, nil
}

func buildAuditSummary(transitions []*entity.WorkflowTransition) DailyAuditSummary {
	if len(transitions) == 0 {
		return DailyAuditSummary{}
	}
	last := transitions[len(transitions)-1]
	lastAt := last.OccurredAt
	return DailyAuditSummary{
		TotalTransitions: len(transitions),
		LastTransitionAt: &lastAt,
		LastTransitionBy: last.ActorUsername,
	}
}
