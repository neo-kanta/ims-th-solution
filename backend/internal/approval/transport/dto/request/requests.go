// Package request holds the approval module's HTTP request DTOs.
package request

// GroupRequest creates or updates an approval group.
type GroupRequest struct {
	GroupCode string `json:"group_code"`
	GroupName string `json:"group_name"`
	Remarks   string `json:"remarks"`
	IsActive  *bool  `json:"is_active"`
}

// GroupMemberRequest adds or updates a group member.
type GroupMemberRequest struct {
	UserID        string `json:"user_id"`
	PriorityOrder int    `json:"priority_order"`
	MemberType    string `json:"member_type"`
	Status        string `json:"status"`
	IsActive      *bool  `json:"is_active"`
}

// ReorderRequest carries an ordered list of member ids (index = priority).
type ReorderRequest struct {
	MemberIDs []string `json:"member_ids"`
}

// TeamRequest creates or updates an approval team.
type TeamRequest struct {
	TeamCode          string `json:"team_code"`
	TeamName          string `json:"team_name"`
	Remarks           string `json:"remarks"`
	HasCoManager      bool   `json:"has_co_manager"`
	MinRequiredStamps int    `json:"min_required_stamps"`
	MaxAllowedStamps  int    `json:"max_allowed_stamps"`
	IsActive          *bool  `json:"is_active"`
}

// TeamContractRequest assigns a contract/fund to a team.
type TeamContractRequest struct {
	ContractID    string `json:"contract_id"`
	EffectiveDate string `json:"effective_date"`
}

// TeamMemberRequest adds or updates a team member.
type TeamMemberRequest struct {
	UserID        string `json:"user_id"`
	MemberType    string `json:"member_type"`
	PriorityOrder int    `json:"priority_order"`
	IsActive      *bool  `json:"is_active"`
}

// StageRequest describes one stage of a process config.
type StageRequest struct {
	StageNumber           int    `json:"stage_number"`
	StageName             string `json:"stage_name"`
	ApproverMode          string `json:"approver_mode"`
	ApproverUserID        string `json:"approver_user_id"`
	ApprovalGroupID       string `json:"approval_group_id"`
	RequiredApprovalCount int    `json:"required_approval_count"`
	IsFinalStage          bool   `json:"is_final_stage"`
}

// ProcessConfigRequest creates or updates an approval process config.
type ProcessConfigRequest struct {
	ProcessCode          string         `json:"process_code"`
	ProcessName          string         `json:"process_name"`
	ProcessType          string         `json:"process_type"`
	ContractType         string         `json:"contract_type"`
	ContractID           string         `json:"contract_id"`
	EffectiveDate        string         `json:"effective_date"`
	IsActive             *bool          `json:"is_active"`
	GroupApprovalEnabled bool           `json:"group_approval_enabled"`
	RequireTeamApproval  bool           `json:"require_team_approval"`
	Stages               []StageRequest `json:"stages"`
}

// SubmitRequest submits a business object into the approval workflow.
type SubmitRequest struct {
	ProcessType      string `json:"process_type"`
	SubjectType      string `json:"subject_type"`
	SubjectID        string `json:"subject_id"`
	SubjectTitle     string `json:"subject_title"`
	SubjectReference string `json:"subject_reference"`
	ContractType     string `json:"contract_type"`
	ContractID       string `json:"contract_id"`
	PortfolioID      string `json:"portfolio_id"`
}

// ActionRequest carries an optional comment for approve / a reason for reject.
type ActionRequest struct {
	Comment string `json:"comment"`
	Reason  string `json:"reason"`
}
