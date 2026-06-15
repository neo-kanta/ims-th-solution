package response

import "time"

// DailyWorkflowResponse is returned by GET /api/v1/workflow/daily
// @Description Aggregated workflow state for one contract on one business date.
type DailyWorkflowResponse struct {
	BusinessDate string `json:"businessDate"` // YYYY-MM-DD

	// CurrentState is one of: NOT_STARTED, INVESTMENT_DAY_STARTED,
	// MANAGER_APPROVED, TRANSACTION_CLOSED, ACCOUNTING_CLOSED
	CurrentState string `json:"currentState"`
	StateLabel   string `json:"stateLabel"`
	IsToday      bool   `json:"isToday"`
	Persisted    bool   `json:"persisted"`

	// AllowedOperations lists operation types the caller may attempt (API names).
	AllowedOperations []string              `json:"allowedOperations"`
	BlockedOperations []string              `json:"blockedOperations"`
	BlockedReasons    []DailyBlockingReason `json:"blockedReasons,omitempty"`
	TransitionRules   []TransitionRule      `json:"transitionRules"`
	Timeline          []DailyTimelineEntry  `json:"timeline"`

	// Approvers maps operationType → configured approver list.
	Approvers map[string][]DailyApproverEntry `json:"approvers,omitempty"`

	// Settings maps operationType → approval configuration.
	Settings        map[string]DailyOperationSetting `json:"settings,omitempty"`
	SettingsSummary map[string]DailyOperationSetting `json:"settingsSummary"`

	AuditSummary    DailyAuditSummary            `json:"auditSummary"`
	ModuleReadiness map[string]DailyModuleStatus `json:"moduleReadiness"`
}

// DailyBlockingReason explains why an operation is blocked.
type DailyBlockingReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DailyTimelineEntry is one normalised row from the transition log.
type DailyTimelineEntry struct {
	TransitionID          string    `json:"transitionId"`
	OperationType         string    `json:"operationType"` // canonical API name
	FromState             string    `json:"fromState"`
	ToState               string    `json:"toState"`
	ExecutedByUsername    string    `json:"executedByUsername"`
	ExecutedByAccountCode string    `json:"executedByAccountCode"`
	IsAdminOverride       bool      `json:"isAdminOverride"`
	OccurredAt            time.Time `json:"executedAt"`
	Reason                *string   `json:"reason,omitempty"`
}

// DailyApproverEntry is one configured approver for an operation type.
type DailyApproverEntry struct {
	AccountCode string `json:"accountCode"`
	Username    string `json:"username"`
	Role        string `json:"role,omitempty"`
}

// DailyOperationSetting holds the approval configuration for one operation.
type DailyOperationSetting struct {
	ApprovalMode string               `json:"approvalMode"`
	Approvers    []DailyApproverEntry `json:"approvers"`
}

// DailyAuditSummary is a compact summary of today's transitions.
type DailyAuditSummary struct {
	TotalTransitions int        `json:"totalTransitions"`
	LastTransitionAt *time.Time `json:"lastTransitionAt,omitempty"`
	LastTransitionBy string     `json:"lastTransitionBy,omitempty"`
}

// DailyModuleStatus reports whether a supporting module is operational.
type DailyModuleStatus struct {
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}

// DailyExecuteResponse is returned by POST /api/v1/workflow/daily/execute
// @Description Result of a workflow transition executed via the daily API.
type DailyExecuteResponse struct {
	TransitionID          string `json:"transitionId"`
	FromState             string `json:"fromState"`  // normalised API name
	ToState               string `json:"toState"`    // normalised API name
	ExecutedAt            string `json:"executedAt"` // RFC3339 UTC
	ExecutedByAccountCode string `json:"executedByAccountCode"`
	IsAdminOverride       bool   `json:"isAdminOverride"`
}

// DailyTransitionsResponse is returned by GET /api/v1/workflow/daily/transitions
type DailyTransitionsResponse struct {
	BusinessDate string               `json:"businessDate"`
	Total        int64                `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"pageSize"`
	Transitions  []DailyTimelineEntry `json:"transitions"`
}

// TransitionRulesResponse is returned by GET /api/v1/workflow/transition-rules
type TransitionRulesResponse struct {
	Rules []TransitionRule `json:"rules"`
}

// TransitionRule describes one valid state→operation→state transition.
type TransitionRule struct {
	FromState      string `json:"fromState"`
	OperationType  string `json:"operationType"`
	ToState        string `json:"toState"`
	RequiresReason bool   `json:"requiresReason"`
}

// WorkflowSettingsResponse is returned by GET and PUT /api/v1/workflow/settings
type WorkflowSettingsResponse struct {
	Settings []OperationSettingEntry `json:"settings"`
}

// OperationSettingEntry is one operation type's full approver configuration.
type OperationSettingEntry struct {
	OperationType string               `json:"operationType"`
	ApprovalMode  string               `json:"approvalMode"`
	Approvers     []DailyApproverEntry `json:"approvers"`
}
