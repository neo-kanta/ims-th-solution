// Package valueobject holds the immutable enumerations and value types for the
// approval module. These string types are stable contracts shared with the
// database CHECK constraints and the frontend; renaming a value is a breaking
// change.
package valueobject

// ProcessType identifies the business process an approval configuration governs.
type ProcessType string

const (
	ProcessInvestmentAnalysisReport ProcessType = "INVESTMENT_ANALYSIS_REPORT"
	ProcessInvestmentDecision       ProcessType = "INVESTMENT_DECISION"
	ProcessInvestmentCancellation   ProcessType = "INVESTMENT_CANCELLATION"
	ProcessWorkflowOperation        ProcessType = "WORKFLOW_OPERATION"
	ProcessLeaveRequest             ProcessType = "LEAVE_REQUEST"
	ProcessLeaveCancellation        ProcessType = "LEAVE_CANCELLATION"
	ProcessDelegationRequest        ProcessType = "DELEGATION_REQUEST"
	ProcessPortfolioOnboarding      ProcessType = "PORTFOLIO_ONBOARDING"
	ProcessComplianceRelease        ProcessType = "COMPLIANCE_RELEASE"
)

// ValidProcessType reports whether the value is a recognised process type.
func ValidProcessType(v ProcessType) bool {
	switch v {
	case ProcessInvestmentAnalysisReport, ProcessInvestmentDecision,
		ProcessInvestmentCancellation, ProcessWorkflowOperation,
		ProcessLeaveRequest, ProcessLeaveCancellation, ProcessDelegationRequest,
		ProcessPortfolioOnboarding, ProcessComplianceRelease:
		return true
	}
	return false
}

// ContractType is the scope a process configuration applies to.
type ContractType string

const (
	ContractTypeFund          ContractType = "FUND"
	ContractTypeDiscretionary ContractType = "DISCRETIONARY"
	ContractTypeCompany       ContractType = "COMPANY"
)

// ValidContractType reports whether the value is a recognised contract type.
func ValidContractType(v ContractType) bool {
	switch v {
	case ContractTypeFund, ContractTypeDiscretionary, ContractTypeCompany:
		return true
	}
	return false
}

// SubjectType identifies the kind of business object under approval.
type SubjectType string

const (
	SubjectResearchReport     SubjectType = "RESEARCH_REPORT"
	SubjectInvestmentDecision SubjectType = "INVESTMENT_DECISION"
	SubjectWorkflowOperation  SubjectType = "WORKFLOW_OPERATION"
	SubjectLeaveRequest       SubjectType = "LEAVE_REQUEST"
	SubjectDelegationRequest  SubjectType = "DELEGATION_REQUEST"
	SubjectPortfolio          SubjectType = "PORTFOLIO"
	SubjectFund               SubjectType = "FUND"
	SubjectComplianceRelease  SubjectType = "COMPLIANCE_RELEASE"
)

// ValidSubjectType reports whether the value is a recognised subject type.
func ValidSubjectType(v SubjectType) bool {
	switch v {
	case SubjectResearchReport, SubjectInvestmentDecision, SubjectWorkflowOperation,
		SubjectLeaveRequest, SubjectDelegationRequest,
		SubjectPortfolio, SubjectFund, SubjectComplianceRelease:
		return true
	}
	return false
}

// ApproverMode controls how approvers are resolved for a stage.
type ApproverMode string

const (
	// ApproverModeSingleUser assigns one configured user.
	ApproverModeSingleUser ApproverMode = "SINGLE_USER"
	// ApproverModeGroupPriority assigns the highest-priority eligible group member.
	ApproverModeGroupPriority ApproverMode = "GROUP_PRIORITY"
	// ApproverModeGroupAny creates tasks for all eligible group members; any one
	// approval completes the stage.
	ApproverModeGroupAny ApproverMode = "GROUP_ANY"
	// ApproverModeTeamMinimum uses the contract's approval team and requires a
	// minimum number of stamps before the stage completes.
	ApproverModeTeamMinimum ApproverMode = "TEAM_MINIMUM"
)

// ValidApproverMode reports whether the value is a recognised approver mode.
func ValidApproverMode(v ApproverMode) bool {
	switch v {
	case ApproverModeSingleUser, ApproverModeGroupPriority,
		ApproverModeGroupAny, ApproverModeTeamMinimum:
		return true
	}
	return false
}

// RequestStatus is the lifecycle status of an approval request.
type RequestStatus string

const (
	RequestStatusDraft           RequestStatus = "DRAFT"
	RequestStatusSubmitted       RequestStatus = "SUBMITTED"
	RequestStatusPendingApproval RequestStatus = "PENDING_APPROVAL"
	RequestStatusApproved        RequestStatus = "APPROVED"
	RequestStatusRejected        RequestStatus = "REJECTED"
	RequestStatusCancelled       RequestStatus = "CANCELLED"
	RequestStatusWithdrawn       RequestStatus = "WITHDRAWN"
	RequestStatusRevoked         RequestStatus = "REVOKED"
)

// IsTerminal reports whether the request can no longer transition.
func (s RequestStatus) IsTerminal() bool {
	switch s {
	case RequestStatusApproved, RequestStatusRejected,
		RequestStatusCancelled, RequestStatusWithdrawn, RequestStatusRevoked:
		return true
	}
	return false
}

// IsActionable reports whether approve/reject actions can be taken on the request.
func (s RequestStatus) IsActionable() bool {
	return s == RequestStatusPendingApproval
}

// TaskStatus is the lifecycle status of an approval task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusApproved  TaskStatus = "APPROVED"
	TaskStatusRejected  TaskStatus = "REJECTED"
	TaskStatusSkipped   TaskStatus = "SKIPPED"
	TaskStatusCancelled TaskStatus = "CANCELLED"
)

// EventType enumerates immutable approval timeline events.
type EventType string

const (
	EventSubmitted        EventType = "SUBMITTED"
	EventTaskCreated      EventType = "TASK_CREATED"
	EventApproved         EventType = "APPROVED"
	EventRejected         EventType = "REJECTED"
	EventCancelled        EventType = "CANCELLED"
	EventWithdrawn        EventType = "WITHDRAWN"
	EventRevoked          EventType = "REVOKED"
	EventDelegated        EventType = "DELEGATED"
	EventStageCompleted   EventType = "STAGE_COMPLETED"
	EventRequestCompleted EventType = "REQUEST_COMPLETED"
)

// GroupMemberType distinguishes ordinary members from supervisors.
type GroupMemberType string

const (
	GroupMemberTypeMember     GroupMemberType = "MEMBER"
	GroupMemberTypeSupervisor GroupMemberType = "SUPERVISOR"
)

// ValidGroupMemberType reports whether the value is recognised.
func ValidGroupMemberType(v GroupMemberType) bool {
	return v == GroupMemberTypeMember || v == GroupMemberTypeSupervisor
}

// GroupMemberStatus is the approval lifecycle for a group member entry.
type GroupMemberStatus string

const (
	GroupMemberStatusPending  GroupMemberStatus = "PENDING"
	GroupMemberStatusApproved GroupMemberStatus = "APPROVED"
	GroupMemberStatusRevoked  GroupMemberStatus = "REVOKED"
)

// ValidGroupMemberStatus reports whether the value is recognised.
func ValidGroupMemberStatus(v GroupMemberStatus) bool {
	switch v {
	case GroupMemberStatusPending, GroupMemberStatusApproved, GroupMemberStatusRevoked:
		return true
	}
	return false
}

// TeamMemberType distinguishes order submitters from reviewer agents.
type TeamMemberType string

const (
	TeamMemberOrderSubmitter TeamMemberType = "ORDER_SUBMITTER"
	TeamMemberReviewerAgent  TeamMemberType = "REVIEWER_AGENT"
)

// ValidTeamMemberType reports whether the value is recognised.
func ValidTeamMemberType(v TeamMemberType) bool {
	return v == TeamMemberOrderSubmitter || v == TeamMemberReviewerAgent
}

// SignatureLabel marks a signature as normal or delegated (proxy).
type SignatureLabel string

const (
	SignatureLabelNormal    SignatureLabel = "NORMAL"
	SignatureLabelDelegated SignatureLabel = "DELEGATED"
)
