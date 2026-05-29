package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// ConfigHandler exposes approval configuration endpoints (groups, teams, processes).
type ConfigHandler struct {
	svc *service.ApprovalConfigService
}

// NewConfigHandler wires the config handler.
func NewConfigHandler(svc *service.ApprovalConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// ───────────────────────── Groups ─────────────────────────

// ListGroups handles GET /approval-config/groups.
// @Summary List approval groups
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param active_only query bool false "Only active groups"
// @Success 200 {array} response.GroupResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/groups [get]
func (h *ConfigHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	f := domain.GroupListFilter{
		ActiveOnly: r.URL.Query().Get("active_only") == "true",
		Search:     r.URL.Query().Get("search"),
	}
	groups, err := h.svc.ListGroups(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroups(groups))
}

// CreateGroup handles POST /approval-config/groups.
// @Summary Create an approval group
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body request.GroupRequest true "Group payload"
// @Success 201 {object} response.GroupResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/groups [post]
func (h *ConfigHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	var body request.GroupRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	g, err := h.svc.CreateGroup(r.Context(), service.GroupInput{
		GroupCode: body.GroupCode, GroupName: body.GroupName, Remarks: body.Remarks, IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromGroup(g))
}

// UpdateGroup handles PUT /approval-config/groups/{id}.
// @Summary Update an approval group
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group UUID"
// @Param payload body request.GroupRequest true "Group payload"
// @Success 200 {object} response.GroupResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id} [put]
func (h *ConfigHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid group id")
		return
	}
	var body request.GroupRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	g, err := h.svc.UpdateGroup(r.Context(), id, service.GroupInput{
		GroupName: body.GroupName, Remarks: body.Remarks, IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroup(g))
}

// ListGroupMembers handles GET /approval-config/groups/{id}/members.
// @Summary List approval group members
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group UUID"
// @Success 200 {array} response.GroupMemberResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members [get]
func (h *ConfigHandler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid group id")
		return
	}
	members, err := h.svc.ListGroupMembers(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroupMembers(members))
}

// AddGroupMember handles POST /approval-config/groups/{id}/members.
// @Summary Add an approval group member
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group UUID"
// @Param payload body request.GroupMemberRequest true "Member payload"
// @Success 201 {object} response.GroupMemberResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members [post]
func (h *ConfigHandler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid group id")
		return
	}
	var body request.GroupMemberRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		httputil.BadRequest(w, "invalid user_id")
		return
	}
	m, err := h.svc.AddGroupMember(r.Context(), id, service.GroupMemberInput{
		UserID: userID, PriorityOrder: body.PriorityOrder,
		MemberType: vo.GroupMemberType(body.MemberType), Status: vo.GroupMemberStatus(body.Status),
		IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromGroupMember(m))
}

// UpdateGroupMember handles PUT /approval-config/groups/{id}/members/{memberId}.
// @Summary Update an approval group member
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group UUID"
// @Param memberId path string true "Member UUID"
// @Param payload body request.GroupMemberRequest true "Member payload"
// @Success 200 {object} response.GroupMemberResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members/{memberId} [put]
func (h *ConfigHandler) UpdateGroupMember(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	memberID, err := parseUUIDParam(r, "memberId")
	if err != nil {
		httputil.BadRequest(w, "invalid member id")
		return
	}
	var body request.GroupMemberRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	m, err := h.svc.UpdateGroupMember(r.Context(), memberID, service.GroupMemberInput{
		PriorityOrder: body.PriorityOrder, MemberType: vo.GroupMemberType(body.MemberType),
		IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroupMember(m))
}

// ApproveGroupMember handles POST /approval-config/groups/{id}/members/{memberId}/approve.
// @Summary Approve an approval group member
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group UUID"
// @Param memberId path string true "Member UUID"
// @Success 200 {object} response.GroupMemberResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members/{memberId}/approve [post]
func (h *ConfigHandler) ApproveGroupMember(w http.ResponseWriter, r *http.Request) {
	h.memberStatus(w, r, true)
}

// RevokeGroupMember handles POST /approval-config/groups/{id}/members/{memberId}/revoke.
// @Summary Revoke an approval group member
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group UUID"
// @Param memberId path string true "Member UUID"
// @Success 200 {object} response.GroupMemberResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members/{memberId}/revoke [post]
func (h *ConfigHandler) RevokeGroupMember(w http.ResponseWriter, r *http.Request) {
	h.memberStatus(w, r, false)
}

func (h *ConfigHandler) memberStatus(w http.ResponseWriter, r *http.Request, approve bool) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	memberID, err := parseUUIDParam(r, "memberId")
	if err != nil {
		httputil.BadRequest(w, "invalid member id")
		return
	}
	if approve {
		mm, err := h.svc.ApproveGroupMember(r.Context(), memberID, actor)
		if err != nil {
			writeError(w, err)
			return
		}
		httputil.OK(w, response.FromGroupMember(mm))
		return
	}
	mm, err := h.svc.RevokeGroupMember(r.Context(), memberID, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroupMember(mm))
}

// ReorderGroupMembers handles POST /approval-config/groups/{id}/members/reorder.
// @Summary Reorder approval group members
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group UUID"
// @Param payload body request.ReorderRequest true "Ordered member ids"
// @Success 200 {array} response.GroupMemberResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/groups/{id}/members/reorder [post]
func (h *ConfigHandler) ReorderGroupMembers(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid group id")
		return
	}
	var body request.ReorderRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	ids := make([]uuid.UUID, 0, len(body.MemberIDs))
	for _, s := range body.MemberIDs {
		mid, err := uuid.Parse(s)
		if err != nil {
			httputil.BadRequest(w, "invalid member id in list")
			return
		}
		ids = append(ids, mid)
	}
	members, err := h.svc.ReorderGroupMembers(r.Context(), id, ids, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromGroupMembers(members))
}

// ───────────────────────── Teams ─────────────────────────

// ListTeams handles GET /approval-config/teams.
// @Summary List approval teams
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Success 200 {array} response.TeamResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/teams [get]
func (h *ConfigHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.svc.ListTeams(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromTeams(teams))
}

// CreateTeam handles POST /approval-config/teams.
// @Summary Create an approval team
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body request.TeamRequest true "Team payload"
// @Success 201 {object} response.TeamResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/teams [post]
func (h *ConfigHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	var body request.TeamRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	t, err := h.svc.CreateTeam(r.Context(), service.TeamInput{
		TeamCode: body.TeamCode, TeamName: body.TeamName, Remarks: body.Remarks,
		HasCoManager: body.HasCoManager, MinRequiredStamps: body.MinRequiredStamps,
		MaxAllowedStamps: body.MaxAllowedStamps, IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromTeam(t))
}

// UpdateTeam handles PUT /approval-config/teams/{id}.
// @Summary Update an approval team
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Team UUID"
// @Param payload body request.TeamRequest true "Team payload"
// @Success 200 {object} response.TeamResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id} [put]
func (h *ConfigHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid team id")
		return
	}
	var body request.TeamRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	t, err := h.svc.UpdateTeam(r.Context(), id, service.TeamInput{
		TeamName: body.TeamName, Remarks: body.Remarks, HasCoManager: body.HasCoManager,
		MinRequiredStamps: body.MinRequiredStamps, MaxAllowedStamps: body.MaxAllowedStamps,
		IsActive: boolOr(body.IsActive, true),
	}, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromTeam(t))
}

// AssignTeamContract handles POST /approval-config/teams/{id}/contracts.
// @Summary Assign a contract/fund to an approval team
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Team UUID"
// @Param payload body request.TeamContractRequest true "Contract assignment"
// @Success 201 {object} response.TeamContractResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id}/contracts [post]
func (h *ConfigHandler) AssignTeamContract(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	teamID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid team id")
		return
	}
	var body request.TeamContractRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	contractID, err := uuid.Parse(body.ContractID)
	if err != nil {
		httputil.BadRequest(w, "invalid contract_id")
		return
	}
	eff, err := optDate(body.EffectiveDate)
	if err != nil {
		httputil.BadRequest(w, "invalid effective_date (expected YYYY-MM-DD)")
		return
	}
	c, err := h.svc.AssignTeamToContract(r.Context(), teamID, contractID, eff, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromTeamContract(c))
}

// ListTeamMembers handles GET /approval-config/teams/{id}/members.
// @Summary List approval team members
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Team UUID"
// @Success 200 {array} response.TeamMemberResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id}/members [get]
func (h *ConfigHandler) ListTeamMembers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid team id")
		return
	}
	members, err := h.svc.ListTeamMembers(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromTeamMembers(members))
}

// AddTeamMember handles POST /approval-config/teams/{id}/members.
// @Summary Add an approval team member
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Team UUID"
// @Param payload body request.TeamMemberRequest true "Member payload"
// @Success 201 {object} response.TeamMemberResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id}/members [post]
func (h *ConfigHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid team id")
		return
	}
	var body request.TeamMemberRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	userID, err := uuid.Parse(body.UserID)
	if err != nil {
		httputil.BadRequest(w, "invalid user_id")
		return
	}
	m, err := h.svc.AddTeamMember(r.Context(), id, service.TeamMemberInput{
		UserID: userID, MemberType: vo.TeamMemberType(body.MemberType),
		PriorityOrder: body.PriorityOrder, IsActive: boolOr(body.IsActive, true),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromTeamMember(m))
}

// UpdateTeamMember handles PUT /approval-config/teams/{id}/members/{memberId}.
// @Summary Update an approval team member
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Team UUID"
// @Param memberId path string true "Member UUID"
// @Param payload body request.TeamMemberRequest true "Member payload"
// @Success 200 {object} response.TeamMemberResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id}/members/{memberId} [put]
func (h *ConfigHandler) UpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	memberID, err := parseUUIDParam(r, "memberId")
	if err != nil {
		httputil.BadRequest(w, "invalid member id")
		return
	}
	var body request.TeamMemberRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	m, err := h.svc.UpdateTeamMember(r.Context(), memberID, service.TeamMemberInput{
		MemberType: vo.TeamMemberType(body.MemberType), PriorityOrder: body.PriorityOrder,
		IsActive: boolOr(body.IsActive, true),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromTeamMember(m))
}

// RemoveTeamMember handles DELETE /approval-config/teams/{id}/members/{memberId}.
// @Summary Remove an approval team member
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Team UUID"
// @Param memberId path string true "Member UUID"
// @Success 204 "No Content"
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/teams/{id}/members/{memberId} [delete]
func (h *ConfigHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	memberID, err := parseUUIDParam(r, "memberId")
	if err != nil {
		httputil.BadRequest(w, "invalid member id")
		return
	}
	if err := h.svc.RemoveTeamMember(r.Context(), memberID); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

// ───────────────────────── Process configs ─────────────────────────

// ListProcesses handles GET /approval-config/processes.
// @Summary List approval process configs
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param process_type query string false "Process type filter"
// @Param active_only query bool false "Only active configs"
// @Success 200 {array} response.ProcessConfigResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approval-config/processes [get]
func (h *ConfigHandler) ListProcesses(w http.ResponseWriter, r *http.Request) {
	f := domain.ProcessListFilter{
		ProcessType: vo.ProcessType(r.URL.Query().Get("process_type")),
		ActiveOnly:  r.URL.Query().Get("active_only") == "true",
	}
	configs, err := h.svc.ListProcessConfigs(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromProcessConfigs(configs))
}

// GetProcess handles GET /approval-config/processes/{id}.
// @Summary Get an approval process config
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Process config UUID"
// @Success 200 {object} response.ProcessConfigResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/processes/{id} [get]
func (h *ConfigHandler) GetProcess(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid process id")
		return
	}
	cfg, err := h.svc.GetProcessConfigDetail(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromProcessConfig(cfg))
}

// CreateProcess handles POST /approval-config/processes.
// @Summary Create an approval process config
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body request.ProcessConfigRequest true "Process config payload"
// @Success 201 {object} response.ProcessConfigResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approval-config/processes [post]
func (h *ConfigHandler) CreateProcess(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	var body request.ProcessConfigRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	in, err := h.toProcessInput(body)
	if err != nil {
		writeError(w, err)
		return
	}
	cfg, err := h.svc.CreateProcessConfig(r.Context(), in, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromProcessConfig(cfg))
}

// UpdateProcess handles PUT /approval-config/processes/{id}.
// @Summary Update an approval process config
// @Tags Approval - Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Process config UUID"
// @Param payload body request.ProcessConfigRequest true "Process config payload"
// @Success 200 {object} response.ProcessConfigResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/processes/{id} [put]
func (h *ConfigHandler) UpdateProcess(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid process id")
		return
	}
	var body request.ProcessConfigRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	in, err := h.toProcessInput(body)
	if err != nil {
		writeError(w, err)
		return
	}
	cfg, err := h.svc.UpdateProcessConfig(r.Context(), id, in, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromProcessConfig(cfg))
}

// ActivateProcess handles POST /approval-config/processes/{id}/activate.
// @Summary Activate an approval process config
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Process config UUID"
// @Success 200 {object} response.ProcessConfigResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/processes/{id}/activate [post]
func (h *ConfigHandler) ActivateProcess(w http.ResponseWriter, r *http.Request) {
	h.setProcessActive(w, r, true)
}

// DeactivateProcess handles POST /approval-config/processes/{id}/deactivate.
// @Summary Deactivate an approval process config
// @Tags Approval - Config
// @Security BearerAuth
// @Produce json
// @Param id path string true "Process config UUID"
// @Success 200 {object} response.ProcessConfigResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approval-config/processes/{id}/deactivate [post]
func (h *ConfigHandler) DeactivateProcess(w http.ResponseWriter, r *http.Request) {
	h.setProcessActive(w, r, false)
}

func (h *ConfigHandler) setProcessActive(w http.ResponseWriter, r *http.Request, active bool) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid process id")
		return
	}
	cfg, err := h.svc.SetProcessConfigActive(r.Context(), id, active, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromProcessConfig(cfg))
}

func (h *ConfigHandler) toProcessInput(body request.ProcessConfigRequest) (service.ProcessConfigInput, error) {
	var in service.ProcessConfigInput
	contractID, err := optUUID(body.ContractID)
	if err != nil {
		return in, domain.Validation("invalid contract_id")
	}
	eff, err := optDate(body.EffectiveDate)
	if err != nil {
		return in, domain.Validation("invalid effective_date (expected YYYY-MM-DD)")
	}
	stages := make([]service.StageInput, 0, len(body.Stages))
	for _, st := range body.Stages {
		userID, err := optUUID(st.ApproverUserID)
		if err != nil {
			return in, domain.Validation("invalid approver_user_id in stage")
		}
		groupID, err := optUUID(st.ApprovalGroupID)
		if err != nil {
			return in, domain.Validation("invalid approval_group_id in stage")
		}
		stages = append(stages, service.StageInput{
			StageNumber: st.StageNumber, StageName: st.StageName,
			ApproverMode: vo.ApproverMode(st.ApproverMode), ApproverUserID: userID, ApprovalGroupID: groupID,
			RequiredApprovalCount: st.RequiredApprovalCount, IsFinalStage: st.IsFinalStage,
		})
	}
	in = service.ProcessConfigInput{
		ProcessCode: body.ProcessCode, ProcessName: body.ProcessName,
		ProcessType: vo.ProcessType(body.ProcessType), ContractType: vo.ContractType(body.ContractType),
		ContractID: contractID, EffectiveDate: eff, IsActive: boolOr(body.IsActive, true),
		GroupApprovalEnabled: body.GroupApprovalEnabled, RequireTeamApproval: body.RequireTeamApproval,
		Stages: stages,
	}
	return in, nil
}
