// Package response holds the approval module's HTTP response DTOs. Handlers map
// domain entities to these shapes; domain entities are never exposed directly.
package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
)

// GroupResponse is an approval group.
type GroupResponse struct {
	ID        string    `json:"id"`
	GroupCode string    `json:"group_code"`
	GroupName string    `json:"group_name"`
	Remarks   string    `json:"remarks"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupMemberResponse is an approval group member.
type GroupMemberResponse struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	UserID        string    `json:"user_id"`
	DisplayName   string    `json:"display_name"`
	PriorityOrder int       `json:"priority_order"`
	MemberType    string    `json:"member_type"`
	Status        string    `json:"status"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TeamResponse is an approval team.
type TeamResponse struct {
	ID                string    `json:"id"`
	TeamCode          string    `json:"team_code"`
	TeamName          string    `json:"team_name"`
	Remarks           string    `json:"remarks"`
	HasCoManager      bool      `json:"has_co_manager"`
	MinRequiredStamps int       `json:"min_required_stamps"`
	MaxAllowedStamps  int       `json:"max_allowed_stamps"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TeamContractResponse is a team-to-contract assignment.
type TeamContractResponse struct {
	ID            string    `json:"id"`
	TeamID        string    `json:"team_id"`
	ContractID    string    `json:"contract_id"`
	EffectiveDate time.Time `json:"effective_date"`
	IsActive      bool      `json:"is_active"`
}

// TeamMemberResponse is an approval team member.
type TeamMemberResponse struct {
	ID            string `json:"id"`
	TeamID        string `json:"team_id"`
	UserID        string `json:"user_id"`
	DisplayName   string `json:"display_name"`
	MemberType    string `json:"member_type"`
	PriorityOrder int    `json:"priority_order"`
	IsActive      bool   `json:"is_active"`
}

// StageResponse is one stage of a process config.
type StageResponse struct {
	ID                    string `json:"id"`
	StageNumber           int    `json:"stage_number"`
	StageName             string `json:"stage_name"`
	ApproverMode          string `json:"approver_mode"`
	ApproverUserID        string `json:"approver_user_id,omitempty"`
	ApprovalGroupID       string `json:"approval_group_id,omitempty"`
	RequiredApprovalCount int    `json:"required_approval_count"`
	IsFinalStage          bool   `json:"is_final_stage"`
	RejectPolicy          string `json:"reject_policy"`
}

// ProcessConfigResponse is an approval process configuration with stages.
type ProcessConfigResponse struct {
	ID                   string          `json:"id"`
	ProcessCode          string          `json:"process_code"`
	ProcessName          string          `json:"process_name"`
	ProcessType          string          `json:"process_type"`
	ContractType         string          `json:"contract_type"`
	ContractID           string          `json:"contract_id,omitempty"`
	EffectiveDate        time.Time       `json:"effective_date"`
	IsActive             bool            `json:"is_active"`
	GroupApprovalEnabled bool            `json:"group_approval_enabled"`
	RequireTeamApproval  bool            `json:"require_team_approval"`
	Stages               []StageResponse `json:"stages"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// RequestResponse is an approval request header.
type RequestResponse struct {
	ID                 string     `json:"id"`
	RequestNumber      string     `json:"request_number"`
	ProcessType        string     `json:"process_type"`
	ProcessConfigID    string     `json:"process_config_id,omitempty"`
	SubjectType        string     `json:"subject_type"`
	SubjectID          string     `json:"subject_id"`
	SubjectTitle       string     `json:"subject_title"`
	SubjectReference   string     `json:"subject_reference"`
	ContractID         string     `json:"contract_id,omitempty"`
	PortfolioID        string     `json:"portfolio_id,omitempty"`
	SubmitterID        string     `json:"submitter_id"`
	SubmitterName      string     `json:"submitter_name"`
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"`
	CurrentStageNumber int        `json:"current_stage_number"`
	Status             string     `json:"status"`
	FinalDecisionBy    string     `json:"final_decision_by,omitempty"`
	FinalDecisionAt    *time.Time `json:"final_decision_at,omitempty"`
	RejectionReason    string     `json:"rejection_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// TaskResponse is an approval task.
type TaskResponse struct {
	ID                  string     `json:"id"`
	ApprovalRequestID   string     `json:"approval_request_id"`
	StageNumber         int        `json:"stage_number"`
	AssignedUserID      string     `json:"assigned_user_id,omitempty"`
	AssignedUserName    string     `json:"assigned_user_name,omitempty"`
	AssignedGroupID     string     `json:"assigned_group_id,omitempty"`
	AssignedTeamID      string     `json:"assigned_team_id,omitempty"`
	DelegatedFromUserID string     `json:"delegated_from_user_id,omitempty"`
	Status              string     `json:"status"`
	ActedBy             string     `json:"acted_by,omitempty"`
	ActedAt             *time.Time `json:"acted_at,omitempty"`
	ActionComment       string     `json:"action_comment,omitempty"`
	IsDelegatedAction   bool       `json:"is_delegated_action"`
	DueAt               *time.Time `json:"due_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// EventResponse is an immutable approval timeline event.
type EventResponse struct {
	ID                  string         `json:"id"`
	EventType           string         `json:"event_type"`
	StageNumber         *int           `json:"stage_number,omitempty"`
	ActorUserID         string         `json:"actor_user_id,omitempty"`
	ActorName           string         `json:"actor_name,omitempty"`
	DelegatedFromUserID string         `json:"delegated_from_user_id,omitempty"`
	Comment             string         `json:"comment,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
}

// SignatureResponse is an approval signature/stamp record.
type SignatureResponse struct {
	ID                string    `json:"id"`
	StageNumber       int       `json:"stage_number"`
	SignerUserID      string    `json:"signer_user_id"`
	SignerDisplayName string    `json:"signer_display_name"`
	SignerTitle       string    `json:"signer_title,omitempty"`
	SignedAt          time.Time `json:"signed_at"`
	IsProxySignature  bool      `json:"is_proxy_signature"`
	ProxyForUserID    string    `json:"proxy_for_user_id,omitempty"`
	SignatureLabel    string    `json:"signature_label"`
}

// RequestDetailResponse bundles a request with tasks, timeline and signatures.
type RequestDetailResponse struct {
	Request    RequestResponse     `json:"request"`
	Tasks      []TaskResponse      `json:"tasks"`
	Timeline   []EventResponse     `json:"timeline"`
	Signatures []SignatureResponse `json:"signatures"`
	ViewerTask *TaskResponse       `json:"viewer_task,omitempty"`
}

// InboxItemResponse is a single approver work item joined with its request.
type InboxItemResponse struct {
	Task    TaskResponse    `json:"task"`
	Request RequestResponse `json:"request"`
}

// InboxListResponse is the paginated inbox payload.
type InboxListResponse struct {
	Items []InboxItemResponse `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// RequestListResponse is the paginated approval-request payload.
type RequestListResponse struct {
	Items []RequestResponse `json:"items"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// SubjectStatusResponse reports the latest approval status for a subject.
type SubjectStatusResponse struct {
	HasRequest bool             `json:"has_request"`
	Request    *RequestResponse `json:"request,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Mappers
// ─────────────────────────────────────────────────────────────────────────────

// FromGroup maps a group entity.
func FromGroup(g *entity.ApprovalGroup) GroupResponse {
	return GroupResponse{
		ID: g.ID.String(), GroupCode: g.GroupCode, GroupName: g.GroupName, Remarks: g.Remarks,
		IsActive: g.IsActive, CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

// FromGroups maps a slice of group entities.
func FromGroups(gs []*entity.ApprovalGroup) []GroupResponse {
	out := make([]GroupResponse, 0, len(gs))
	for _, g := range gs {
		out = append(out, FromGroup(g))
	}
	return out
}

// FromGroupMember maps a group member entity.
func FromGroupMember(m *entity.ApprovalGroupMember) GroupMemberResponse {
	return GroupMemberResponse{
		ID: m.ID.String(), GroupID: m.GroupID.String(), UserID: m.UserID.String(), DisplayName: m.DisplayName,
		PriorityOrder: m.PriorityOrder, MemberType: string(m.MemberType), Status: string(m.Status),
		IsActive: m.IsActive, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// FromGroupMembers maps a slice of group members.
func FromGroupMembers(ms []*entity.ApprovalGroupMember) []GroupMemberResponse {
	out := make([]GroupMemberResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, FromGroupMember(m))
	}
	return out
}

// FromTeam maps a team entity.
func FromTeam(t *entity.ApprovalTeam) TeamResponse {
	return TeamResponse{
		ID: t.ID.String(), TeamCode: t.TeamCode, TeamName: t.TeamName, Remarks: t.Remarks,
		HasCoManager: t.HasCoManager, MinRequiredStamps: t.MinRequiredStamps, MaxAllowedStamps: t.MaxAllowedStamps,
		IsActive: t.IsActive, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

// FromTeams maps a slice of team entities.
func FromTeams(ts []*entity.ApprovalTeam) []TeamResponse {
	out := make([]TeamResponse, 0, len(ts))
	for _, t := range ts {
		out = append(out, FromTeam(t))
	}
	return out
}

// FromTeamContract maps a team-contract assignment.
func FromTeamContract(c *entity.ApprovalTeamContract) TeamContractResponse {
	return TeamContractResponse{
		ID: c.ID.String(), TeamID: c.TeamID.String(), ContractID: c.ContractID.String(),
		EffectiveDate: c.EffectiveDate, IsActive: c.IsActive,
	}
}

// FromTeamContracts maps a slice of team-contract assignments.
func FromTeamContracts(cs []*entity.ApprovalTeamContract) []TeamContractResponse {
	out := make([]TeamContractResponse, 0, len(cs))
	for _, c := range cs {
		out = append(out, FromTeamContract(c))
	}
	return out
}

// FromTeamMember maps a team member.
func FromTeamMember(m *entity.ApprovalTeamMember) TeamMemberResponse {
	return TeamMemberResponse{
		ID: m.ID.String(), TeamID: m.TeamID.String(), UserID: m.UserID.String(), DisplayName: m.DisplayName,
		MemberType: string(m.MemberType), PriorityOrder: m.PriorityOrder, IsActive: m.IsActive,
	}
}

// FromTeamMembers maps a slice of team members.
func FromTeamMembers(ms []*entity.ApprovalTeamMember) []TeamMemberResponse {
	out := make([]TeamMemberResponse, 0, len(ms))
	for _, m := range ms {
		out = append(out, FromTeamMember(m))
	}
	return out
}

// FromProcessConfig maps a process config with stages.
func FromProcessConfig(c *entity.ApprovalProcessConfig) ProcessConfigResponse {
	resp := ProcessConfigResponse{
		ID: c.ID.String(), ProcessCode: c.ProcessCode, ProcessName: c.ProcessName,
		ProcessType: string(c.ProcessType), ContractType: string(c.ContractType),
		EffectiveDate: c.EffectiveDate, IsActive: c.IsActive,
		GroupApprovalEnabled: c.GroupApprovalEnabled, RequireTeamApproval: c.RequireTeamApproval,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
	if c.ContractID != nil {
		resp.ContractID = c.ContractID.String()
	}
	resp.Stages = make([]StageResponse, 0, len(c.Stages))
	for _, st := range c.Stages {
		sr := StageResponse{
			ID: st.ID.String(), StageNumber: st.StageNumber, StageName: st.StageName,
			ApproverMode: string(st.ApproverMode), RequiredApprovalCount: st.RequiredApprovalCount,
			IsFinalStage: st.IsFinalStage, RejectPolicy: st.RejectPolicy,
		}
		if st.ApproverUserID != nil {
			sr.ApproverUserID = st.ApproverUserID.String()
		}
		if st.ApprovalGroupID != nil {
			sr.ApprovalGroupID = st.ApprovalGroupID.String()
		}
		resp.Stages = append(resp.Stages, sr)
	}
	return resp
}

// FromProcessConfigs maps a slice of process configs.
func FromProcessConfigs(cs []*entity.ApprovalProcessConfig) []ProcessConfigResponse {
	out := make([]ProcessConfigResponse, 0, len(cs))
	for _, c := range cs {
		out = append(out, FromProcessConfig(c))
	}
	return out
}

// FromRequest maps a request header.
func FromRequest(r *entity.ApprovalRequest) RequestResponse {
	resp := RequestResponse{
		ID: r.ID.String(), RequestNumber: r.RequestNumber, ProcessType: string(r.ProcessType),
		SubjectType: string(r.SubjectType), SubjectID: r.SubjectID.String(), SubjectTitle: r.SubjectTitle,
		SubjectReference: r.SubjectReference, SubmitterID: r.SubmitterID.String(), SubmitterName: r.SubmitterName,
		SubmittedAt: r.SubmittedAt, CurrentStageNumber: r.CurrentStageNumber, Status: string(r.Status),
		FinalDecisionAt: r.FinalDecisionAt, RejectionReason: r.RejectionReason,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.ProcessConfigID != nil {
		resp.ProcessConfigID = r.ProcessConfigID.String()
	}
	if r.ContractID != nil {
		resp.ContractID = r.ContractID.String()
	}
	if r.PortfolioID != nil {
		resp.PortfolioID = r.PortfolioID.String()
	}
	if r.FinalDecisionBy != nil {
		resp.FinalDecisionBy = r.FinalDecisionBy.String()
	}
	return resp
}

// FromRequests maps a slice of request headers.
func FromRequests(rs []*entity.ApprovalRequest) []RequestResponse {
	out := make([]RequestResponse, 0, len(rs))
	for _, r := range rs {
		out = append(out, FromRequest(r))
	}
	return out
}

// FromTask maps a task entity.
func FromTask(t *entity.ApprovalTask) TaskResponse {
	resp := TaskResponse{
		ID: t.ID.String(), ApprovalRequestID: t.ApprovalRequestID.String(), StageNumber: t.StageNumber,
		AssignedUserName: t.AssignedUserName, Status: string(t.Status), ActionComment: t.ActionComment,
		IsDelegatedAction: t.IsDelegatedAction, ActedAt: t.ActedAt, DueAt: t.DueAt, CreatedAt: t.CreatedAt,
	}
	if t.AssignedUserID != nil {
		resp.AssignedUserID = t.AssignedUserID.String()
	}
	if t.AssignedGroupID != nil {
		resp.AssignedGroupID = t.AssignedGroupID.String()
	}
	if t.AssignedTeamID != nil {
		resp.AssignedTeamID = t.AssignedTeamID.String()
	}
	if t.DelegatedFromUserID != nil {
		resp.DelegatedFromUserID = t.DelegatedFromUserID.String()
	}
	if t.ActedBy != nil {
		resp.ActedBy = t.ActedBy.String()
	}
	return resp
}

// FromTasks maps a slice of tasks.
func FromTasks(ts []*entity.ApprovalTask) []TaskResponse {
	out := make([]TaskResponse, 0, len(ts))
	for _, t := range ts {
		out = append(out, FromTask(t))
	}
	return out
}

// FromEvent maps a timeline event.
func FromEvent(e *entity.ApprovalEvent) EventResponse {
	resp := EventResponse{
		ID: e.ID.String(), EventType: string(e.EventType), StageNumber: e.StageNumber,
		ActorName: e.ActorName, Comment: e.Comment, Metadata: e.Metadata, CreatedAt: e.CreatedAt,
	}
	if e.ActorUserID != nil {
		resp.ActorUserID = e.ActorUserID.String()
	}
	if e.DelegatedFromUserID != nil {
		resp.DelegatedFromUserID = e.DelegatedFromUserID.String()
	}
	return resp
}

// FromEvents maps a slice of events.
func FromEvents(es []*entity.ApprovalEvent) []EventResponse {
	out := make([]EventResponse, 0, len(es))
	for _, e := range es {
		out = append(out, FromEvent(e))
	}
	return out
}

// FromSignature maps a signature record.
func FromSignature(s *entity.ApprovalSignatureRecord) SignatureResponse {
	resp := SignatureResponse{
		ID: s.ID.String(), StageNumber: s.StageNumber, SignerUserID: s.SignerUserID.String(),
		SignerDisplayName: s.SignerDisplayName, SignedAt: s.SignedAt, IsProxySignature: s.IsProxySignature,
		SignatureLabel: string(s.SignatureLabel),
	}
	if s.SignerTitle != nil {
		resp.SignerTitle = *s.SignerTitle
	}
	if s.ProxyForUserID != nil {
		resp.ProxyForUserID = s.ProxyForUserID.String()
	}
	return resp
}

// FromSignatures maps a slice of signatures.
func FromSignatures(ss []*entity.ApprovalSignatureRecord) []SignatureResponse {
	out := make([]SignatureResponse, 0, len(ss))
	for _, s := range ss {
		out = append(out, FromSignature(s))
	}
	return out
}

// FromRequestDetail maps a runtime RequestDetail.
func FromRequestDetail(d *service.RequestDetail) RequestDetailResponse {
	resp := RequestDetailResponse{
		Request:    FromRequest(d.Request),
		Tasks:      FromTasks(d.Tasks),
		Timeline:   FromEvents(d.Events),
		Signatures: FromSignatures(d.Signatures),
	}
	if d.ViewerTask != nil {
		vt := FromTask(d.ViewerTask)
		resp.ViewerTask = &vt
	}
	return resp
}

// FromInboxItems maps inbox items.
func FromInboxItems(items []*domain.InboxItem) []InboxItemResponse {
	out := make([]InboxItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, InboxItemResponse{
			Task:    FromTask(&it.Task),
			Request: FromRequest(&it.Request),
		})
	}
	return out
}
