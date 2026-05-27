package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	RequestStatusDraft          = "DRAFT"
	RequestStatusReadyForReview = "READY_FOR_REVIEW"
	RequestStatusChanges        = "CHANGES_REQUESTED"
	RequestStatusApproved       = "APPROVED"
	RequestStatusRejected       = "REJECTED"
	RequestStatusMerged         = "MERGED"
	RequestStatusClosed         = "CLOSED"
	RequestStatusCancelled      = "CANCELLED"

	StepStatusNotStarted      = "NOT_STARTED"
	StepStatusPending         = "PENDING"
	StepStatusApproved        = "APPROVED"
	StepStatusChanges         = "CHANGES_REQUESTED"
	StepStatusRejected        = "REJECTED"
	StepStatusSkipped         = "SKIPPED"
	ReviewerStatusPending     = "PENDING"
	ReviewerStatusApproved    = "APPROVED"
	ReviewerStatusChanges     = "CHANGES_REQUESTED"
	ReviewerStatusRejected    = "REJECTED"
	CheckStatusPassed         = "PASSED"
	CheckStatusWarning        = "WARNING"
	CheckStatusFailed         = "FAILED"
	CheckSeverityInfo         = "INFO"
	CheckSeverityWarning      = "WARNING"
	CheckSeverityBlocker      = "BLOCKER"
	RiskLow                   = "LOW"
	RiskMedium                = "MEDIUM"
	RiskHigh                  = "HIGH"
	RiskCritical              = "CRITICAL"
	RequestTypePermission     = "PERMISSION_CHANGE"
	RequestTypeRoleAssignment = "ROLE_ASSIGNMENT"
	RequestTypeDataRights     = "DATA_RIGHTS_CHANGE"
)

type ChangeRequest struct {
	ID               uuid.UUID  `json:"id"`
	RequestNo        string     `json:"request_no"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	RequestType      string     `json:"request_type"`
	Status           string     `json:"status"`
	RiskLevel        string     `json:"risk_level"`
	TargetEntityType string     `json:"target_entity_type"`
	TargetEntityID   string     `json:"target_entity_id"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	CreatedByName    string     `json:"created_by_name"`
	AssignedTo       *uuid.UUID `json:"assigned_to,omitempty"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	ApprovedBy       *uuid.UUID `json:"approved_by,omitempty"`
	MergedAt         *time.Time `json:"merged_at,omitempty"`
	MergedBy         *uuid.UUID `json:"merged_by,omitempty"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty"`
	RejectedBy       *uuid.UUID `json:"rejected_by,omitempty"`
	RejectionReason  string     `json:"rejection_reason"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	ClosedBy         *uuid.UUID `json:"closed_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	Items     []ChangeItem    `json:"items"`
	Steps     []ApprovalStep  `json:"steps"`
	Approvers []StepApprover  `json:"approvers"`
	Comments  []Comment       `json:"comments"`
	Checks    []Check         `json:"checks"`
	Labels    []Label         `json:"labels"`
	Events    []WorkflowEvent `json:"events"`
}

type ChangeRequestFilter struct {
	Status    string
	RiskLevel string
	Label     string
	Requester *uuid.UUID
	Reviewer  *uuid.UUID
	Module    string
	Search    string
	DateFrom  *time.Time
	DateTo    *time.Time
	Page      int
	Limit     int
}

type ChangeItem struct {
	ID          uuid.UUID       `json:"id"`
	RequestID   uuid.UUID       `json:"request_id"`
	ItemType    string          `json:"item_type"`
	TargetTable string          `json:"target_table"`
	TargetID    string          `json:"target_id"`
	ActionType  string          `json:"action_type"`
	BeforeJSON  json.RawMessage `json:"before_json"`
	AfterJSON   json.RawMessage `json:"after_json"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ApprovalSetting struct {
	ID                      uuid.UUID             `json:"id"`
	SettingCode             string                `json:"setting_code"`
	SettingName             string                `json:"setting_name"`
	Module                  string                `json:"module"`
	RequestType             string                `json:"request_type"`
	RiskLevel               string                `json:"risk_level"`
	IsActive                bool                  `json:"is_active"`
	IsDefault               bool                  `json:"is_default"`
	SequentialApproval      bool                  `json:"sequential_approval"`
	AllowCreatorApproval    bool                  `json:"allow_creator_approval"`
	FailedCheckBlocksSubmit bool                  `json:"failed_check_blocks_submit"`
	FailedCheckBlocksMerge  bool                  `json:"failed_check_blocks_merge"`
	Description             string                `json:"description"`
	CreatedAt               time.Time             `json:"created_at"`
	UpdatedAt               time.Time             `json:"updated_at"`
	Steps                   []ApprovalSettingStep `json:"steps"`
}

type ApprovalSettingStep struct {
	ID                   uuid.UUID  `json:"id"`
	WorkflowSettingID    uuid.UUID  `json:"workflow_setting_id"`
	StepNo               int        `json:"step_no"`
	StepName             string     `json:"step_name"`
	StepDescription      string     `json:"step_description"`
	ApprovalMode         string     `json:"approval_mode"`
	ApproverType         string     `json:"approver_type"`
	RequiredRoleCode     string     `json:"required_role_code"`
	RequiredGroupID      *uuid.UUID `json:"required_group_id,omitempty"`
	RequiredUserID       *uuid.UUID `json:"required_user_id,omitempty"`
	MinApprovalsRequired int        `json:"min_approvals_required"`
	AllowDelegation      bool       `json:"allow_delegation"`
	IsRequired           bool       `json:"is_required"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ApprovalStep struct {
	ID                   uuid.UUID      `json:"id"`
	RequestID            uuid.UUID      `json:"request_id"`
	WorkflowSettingID    uuid.UUID      `json:"workflow_setting_id"`
	StepNo               int            `json:"step_no"`
	StepName             string         `json:"step_name"`
	ApprovalMode         string         `json:"approval_mode"`
	Status               string         `json:"status"`
	MinApprovalsRequired int            `json:"min_approvals_required"`
	ApprovalsReceived    int            `json:"approvals_received"`
	StartedAt            *time.Time     `json:"started_at,omitempty"`
	CompletedAt          *time.Time     `json:"completed_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	Approvers            []StepApprover `json:"approvers"`
}

type StepApprover struct {
	ID                  uuid.UUID  `json:"id"`
	RequestID           uuid.UUID  `json:"request_id"`
	ApprovalStepID      uuid.UUID  `json:"approval_step_id"`
	ApproverUserID      *uuid.UUID `json:"approver_user_id,omitempty"`
	ApproverDisplayName string     `json:"approver_display_name"`
	ApproverRoleCode    string     `json:"approver_role_code"`
	ApprovalStatus      string     `json:"approval_status"`
	ApprovalComment     string     `json:"approval_comment"`
	IsDelegated         bool       `json:"is_delegated"`
	DelegatedFromUserID *uuid.UUID `json:"delegated_from_user_id,omitempty"`
	ApprovedAt          *time.Time `json:"approved_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type Comment struct {
	ID        uuid.UUID  `json:"id"`
	RequestID uuid.UUID  `json:"request_id"`
	UserID    uuid.UUID  `json:"user_id"`
	UserName  string     `json:"user_name"`
	Comment   string     `json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Check struct {
	ID        uuid.UUID `json:"id"`
	RequestID uuid.UUID `json:"request_id"`
	CheckCode string    `json:"check_code"`
	CheckName string    `json:"check_name"`
	Status    string    `json:"status"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Label struct {
	ID          uuid.UUID `json:"id"`
	LabelCode   string    `json:"label_code"`
	LabelName   string    `json:"label_name"`
	LabelType   string    `json:"label_type"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkflowEvent struct {
	ID             uuid.UUID       `json:"id"`
	RequestID      uuid.UUID       `json:"request_id"`
	ApprovalStepID *uuid.UUID      `json:"approval_step_id,omitempty"`
	ActorUserID    *uuid.UUID      `json:"actor_user_id,omitempty"`
	ActorName      string          `json:"actor_name"`
	EventType      string          `json:"event_type"`
	BeforeJSON     json.RawMessage `json:"before_json"`
	AfterJSON      json.RawMessage `json:"after_json"`
	Comment        string          `json:"comment"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Role struct {
	ID                       uuid.UUID `json:"id"`
	RoleCode                 string    `json:"role_code"`
	RoleName                 string    `json:"role_name"`
	Department               string    `json:"department"`
	RoleCategory             string    `json:"role_category"`
	PriorityRank             int       `json:"priority_rank"`
	AssignmentScope          string    `json:"assignment_scope"`
	CanRequestRoleAssignment bool      `json:"can_request_role_assignment"`
	CanApproveRoleAssignment bool      `json:"can_approve_role_assignment"`
	IsHighRisk               bool      `json:"is_high_risk"`
	IsActive                 bool      `json:"is_active"`
	Description              string    `json:"description"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type UserSummary struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	DisplayName         string     `json:"display_name"`
	Email               string     `json:"email"`
	IsActive            bool       `json:"is_active"`
	IsLocked            bool       `json:"is_locked"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	ForcePasswordChange bool       `json:"force_password_change"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	Groups              []string   `json:"groups"`
	Roles               []string   `json:"roles"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type GroupSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	MembersCount int       `json:"members_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FunctionDefinition struct {
	Code         string     `json:"code"`
	Module       string     `json:"module"`
	Screen       string     `json:"screen"`
	Action       string     `json:"action"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`
}

type FunctionRight struct {
	ID                uuid.UUID  `json:"id"`
	SubjectType       string     `json:"subject_type"`
	SubjectID         uuid.UUID  `json:"subject_id"`
	SubjectName       string     `json:"subject_name"`
	PermissionCode    string     `json:"permission_code"`
	CanView           bool       `json:"can_view"`
	CanSearch         bool       `json:"can_search"`
	CanAdd            bool       `json:"can_add"`
	CanEdit           bool       `json:"can_edit"`
	CanDelete         bool       `json:"can_delete"`
	CanApprove        bool       `json:"can_approve"`
	CanRevokeApproval bool       `json:"can_revoke_approval"`
	CanExport         bool       `json:"can_export"`
	CanConfigure      bool       `json:"can_configure"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty"`
	ApprovedBy        *uuid.UUID `json:"approved_by,omitempty"`
}

type DataRight struct {
	ID          uuid.UUID  `json:"id"`
	SubjectType string     `json:"subject_type"`
	SubjectID   uuid.UUID  `json:"subject_id"`
	SubjectName string     `json:"subject_name"`
	FundID      *uuid.UUID `json:"fund_id,omitempty"`
	ContractID  *uuid.UUID `json:"contract_id,omitempty"`
	PortfolioID *uuid.UUID `json:"portfolio_id,omitempty"`
	AccessLevel string     `json:"access_level"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  *uuid.UUID `json:"approved_by,omitempty"`
}

type EffectivePermissions struct {
	UserID              uuid.UUID       `json:"user_id"`
	DirectFunctions     []FunctionRight `json:"direct_function_permissions"`
	GroupFunctions      []FunctionRight `json:"group_function_permissions"`
	RoleFunctions       []FunctionRight `json:"role_function_permissions"`
	DirectDataRights    []DataRight     `json:"direct_data_permissions"`
	GroupDataRights     []DataRight     `json:"group_data_permissions"`
	RoleDataRights      []DataRight     `json:"role_data_permissions"`
	FinalFunctionCodes  []string        `json:"final_function_permissions"`
	FinalContractScopes []string        `json:"final_contract_permissions"`
	Roles               []Role          `json:"roles"`
	Groups              []GroupSummary  `json:"groups"`
}

type AuditLog struct {
	ID            uuid.UUID       `json:"id"`
	ActorUserID   *uuid.UUID      `json:"actor_user_id,omitempty"`
	ActorName     string          `json:"actor_name"`
	Action        string          `json:"action"`
	Module        string          `json:"module"`
	EntityType    string          `json:"entity_type"`
	EntityID      string          `json:"entity_id"`
	BeforeJSON    json.RawMessage `json:"before_json"`
	AfterJSON     json.RawMessage `json:"after_json"`
	IPAddress     string          `json:"ip_address"`
	UserAgent     string          `json:"user_agent"`
	CorrelationID string          `json:"correlation_id"`
	CreatedAt     time.Time       `json:"created_at"`
}

type NotificationSetting struct {
	ID              uuid.UUID `json:"id"`
	EventCode       string    `json:"event_code"`
	Channel         string    `json:"channel"`
	Enabled         bool      `json:"enabled"`
	TargetScope     string    `json:"target_scope"`
	TemplateSubject string    `json:"template_subject"`
	TemplateBody    string    `json:"template_body"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
