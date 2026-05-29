package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// ApprovalConfigService manages approval groups, teams and process configs.
type ApprovalConfigService struct {
	pool  *pgxpool.Pool
	repo  domain.Repository
	now   func() time.Time
}

// NewApprovalConfigService wires the configuration service.
func NewApprovalConfigService(pool *pgxpool.Pool, repo domain.Repository) *ApprovalConfigService {
	return &ApprovalConfigService{pool: pool, repo: repo, now: nowUTC}
}

func (s *ApprovalConfigService) runTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return withTransaction(ctx, s.pool, fn)
}

// ───────────────────────── Groups ─────────────────────────

// GroupInput is the payload for creating/updating an approval group.
type GroupInput struct {
	GroupCode string
	GroupName string
	Remarks   string
	IsActive  bool
}

// CreateGroup creates an approval group.
func (s *ApprovalConfigService) CreateGroup(ctx context.Context, in GroupInput, actor uuid.UUID) (*entity.ApprovalGroup, error) {
	code := strings.ToUpper(strings.TrimSpace(in.GroupCode))
	name := strings.TrimSpace(in.GroupName)
	if code == "" {
		return nil, domain.Validation("group_code is required")
	}
	if name == "" {
		return nil, domain.Validation("group_name is required")
	}
	g := &entity.ApprovalGroup{
		ID: uuid.New(), GroupCode: code, GroupName: name, Remarks: strings.TrimSpace(in.Remarks),
		IsActive: in.IsActive, CreatedBy: &actor, UpdatedBy: &actor,
	}
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.CreateGroup(ctx, tx, g) }); err != nil {
		return nil, err
	}
	return s.repo.GetGroup(ctx, g.ID)
}

// UpdateGroup updates an approval group.
func (s *ApprovalConfigService) UpdateGroup(ctx context.Context, id uuid.UUID, in GroupInput, actor uuid.UUID) (*entity.ApprovalGroup, error) {
	g, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.NotFound("approval group not found")
	}
	if name := strings.TrimSpace(in.GroupName); name != "" {
		g.GroupName = name
	}
	g.Remarks = strings.TrimSpace(in.Remarks)
	g.IsActive = in.IsActive
	g.UpdatedBy = &actor
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.UpdateGroup(ctx, tx, g) }); err != nil {
		return nil, err
	}
	return s.repo.GetGroup(ctx, id)
}

// ListGroups lists approval groups.
func (s *ApprovalConfigService) ListGroups(ctx context.Context, f domain.GroupListFilter) ([]*entity.ApprovalGroup, error) {
	return s.repo.ListGroups(ctx, f)
}

// GetGroup returns a group by id.
func (s *ApprovalConfigService) GetGroup(ctx context.Context, id uuid.UUID) (*entity.ApprovalGroup, error) {
	g, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.NotFound("approval group not found")
	}
	return g, nil
}

// GroupMemberInput is the payload for adding/updating a group member.
type GroupMemberInput struct {
	UserID        uuid.UUID
	PriorityOrder int
	MemberType    vo.GroupMemberType
	Status        vo.GroupMemberStatus
	IsActive      bool
}

// AddGroupMember adds a member to a group (defaults to PENDING).
func (s *ApprovalConfigService) AddGroupMember(ctx context.Context, groupID uuid.UUID, in GroupMemberInput, actor uuid.UUID) (*entity.ApprovalGroupMember, error) {
	g, err := s.repo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.NotFound("approval group not found")
	}
	if in.UserID == uuid.Nil {
		return nil, domain.Validation("user_id is required")
	}
	mt := in.MemberType
	if mt == "" {
		mt = vo.GroupMemberTypeMember
	}
	if !vo.ValidGroupMemberType(mt) {
		return nil, domain.Validation("invalid member_type")
	}
	st := in.Status
	if st == "" {
		st = vo.GroupMemberStatusPending
	}
	if !vo.ValidGroupMemberStatus(st) {
		return nil, domain.Validation("invalid status")
	}
	m := &entity.ApprovalGroupMember{
		ID: uuid.New(), GroupID: groupID, UserID: in.UserID, PriorityOrder: in.PriorityOrder,
		MemberType: mt, Status: st, IsActive: in.IsActive, CreatedBy: &actor, UpdatedBy: &actor,
	}
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.AddGroupMember(ctx, tx, m) }); err != nil {
		return nil, err
	}
	return s.repo.GetGroupMember(ctx, m.ID)
}

// UpdateGroupMember updates a member's priority/type/active flag.
func (s *ApprovalConfigService) UpdateGroupMember(ctx context.Context, memberID uuid.UUID, in GroupMemberInput, actor uuid.UUID) (*entity.ApprovalGroupMember, error) {
	m, err := s.repo.GetGroupMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.NotFound("group member not found")
	}
	m.PriorityOrder = in.PriorityOrder
	if in.MemberType != "" {
		if !vo.ValidGroupMemberType(in.MemberType) {
			return nil, domain.Validation("invalid member_type")
		}
		m.MemberType = in.MemberType
	}
	m.IsActive = in.IsActive
	m.UpdatedBy = &actor
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.UpdateGroupMember(ctx, tx, m) }); err != nil {
		return nil, err
	}
	return s.repo.GetGroupMember(ctx, memberID)
}

// ApproveGroupMember marks a member APPROVED (eligible to approve).
func (s *ApprovalConfigService) ApproveGroupMember(ctx context.Context, memberID uuid.UUID, actor uuid.UUID) (*entity.ApprovalGroupMember, error) {
	return s.setGroupMemberStatus(ctx, memberID, vo.GroupMemberStatusApproved, actor)
}

// RevokeGroupMember marks a member REVOKED (no longer eligible).
func (s *ApprovalConfigService) RevokeGroupMember(ctx context.Context, memberID uuid.UUID, actor uuid.UUID) (*entity.ApprovalGroupMember, error) {
	return s.setGroupMemberStatus(ctx, memberID, vo.GroupMemberStatusRevoked, actor)
}

func (s *ApprovalConfigService) setGroupMemberStatus(ctx context.Context, memberID uuid.UUID, status vo.GroupMemberStatus, actor uuid.UUID) (*entity.ApprovalGroupMember, error) {
	m, err := s.repo.GetGroupMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.NotFound("group member not found")
	}
	m.Status = status
	if status == vo.GroupMemberStatusRevoked {
		m.IsActive = false
	}
	m.UpdatedBy = &actor
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.UpdateGroupMember(ctx, tx, m) }); err != nil {
		return nil, err
	}
	return s.repo.GetGroupMember(ctx, memberID)
}

// ListGroupMembers lists all members of a group.
func (s *ApprovalConfigService) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	return s.repo.ListGroupMembers(ctx, groupID)
}

// ReorderGroupMembers applies new priority order for the given member ids (order = priority).
func (s *ApprovalConfigService) ReorderGroupMembers(ctx context.Context, groupID uuid.UUID, orderedMemberIDs []uuid.UUID, actor uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	g, err := s.repo.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.NotFound("approval group not found")
	}
	err = s.runTx(ctx, func(tx pgx.Tx) error {
		for i, id := range orderedMemberIDs {
			m, err := s.repo.GetGroupMember(ctx, id)
			if err != nil {
				return err
			}
			if m == nil || m.GroupID != groupID {
				return domain.Validation("member does not belong to this group")
			}
			m.PriorityOrder = i + 1
			m.UpdatedBy = &actor
			if err := s.repo.UpdateGroupMember(ctx, tx, m); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.ListGroupMembers(ctx, groupID)
}

// ───────────────────────── Teams ─────────────────────────

// TeamInput is the payload for creating/updating an approval team.
type TeamInput struct {
	TeamCode          string
	TeamName          string
	Remarks           string
	HasCoManager      bool
	MinRequiredStamps int
	MaxAllowedStamps  int
	IsActive          bool
}

// CreateTeam creates an approval team.
func (s *ApprovalConfigService) CreateTeam(ctx context.Context, in TeamInput, actor uuid.UUID) (*entity.ApprovalTeam, error) {
	code := strings.ToUpper(strings.TrimSpace(in.TeamCode))
	name := strings.TrimSpace(in.TeamName)
	if code == "" {
		return nil, domain.Validation("team_code is required")
	}
	if name == "" {
		return nil, domain.Validation("team_name is required")
	}
	minS, maxS := in.MinRequiredStamps, in.MaxAllowedStamps
	if minS < 1 {
		minS = 1
	}
	if maxS < minS {
		maxS = minS
	}
	t := &entity.ApprovalTeam{
		ID: uuid.New(), TeamCode: code, TeamName: name, Remarks: strings.TrimSpace(in.Remarks),
		HasCoManager: in.HasCoManager, MinRequiredStamps: minS, MaxAllowedStamps: maxS,
		IsActive: in.IsActive, CreatedBy: &actor, UpdatedBy: &actor,
	}
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.CreateTeam(ctx, tx, t) }); err != nil {
		return nil, err
	}
	return s.repo.GetTeam(ctx, t.ID)
}

// UpdateTeam updates an approval team.
func (s *ApprovalConfigService) UpdateTeam(ctx context.Context, id uuid.UUID, in TeamInput, actor uuid.UUID) (*entity.ApprovalTeam, error) {
	t, err := s.repo.GetTeam(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.NotFound("approval team not found")
	}
	if name := strings.TrimSpace(in.TeamName); name != "" {
		t.TeamName = name
	}
	t.Remarks = strings.TrimSpace(in.Remarks)
	t.HasCoManager = in.HasCoManager
	if in.MinRequiredStamps >= 1 {
		t.MinRequiredStamps = in.MinRequiredStamps
	}
	if in.MaxAllowedStamps >= t.MinRequiredStamps {
		t.MaxAllowedStamps = in.MaxAllowedStamps
	}
	t.IsActive = in.IsActive
	t.UpdatedBy = &actor
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.UpdateTeam(ctx, tx, t) }); err != nil {
		return nil, err
	}
	return s.repo.GetTeam(ctx, id)
}

// ListTeams lists approval teams.
func (s *ApprovalConfigService) ListTeams(ctx context.Context) ([]*entity.ApprovalTeam, error) {
	return s.repo.ListTeams(ctx)
}

// GetTeam returns a team by id.
func (s *ApprovalConfigService) GetTeam(ctx context.Context, id uuid.UUID) (*entity.ApprovalTeam, error) {
	t, err := s.repo.GetTeam(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.NotFound("approval team not found")
	}
	return t, nil
}

// AssignTeamToContract assigns a contract/fund to a team (deactivating prior assignment).
func (s *ApprovalConfigService) AssignTeamToContract(ctx context.Context, teamID, contractID uuid.UUID, effective *time.Time, actor uuid.UUID) (*entity.ApprovalTeamContract, error) {
	t, err := s.repo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.NotFound("approval team not found")
	}
	if contractID == uuid.Nil {
		return nil, domain.Validation("contract_id is required")
	}
	eff := s.now()
	if effective != nil {
		eff = *effective
	}
	c := &entity.ApprovalTeamContract{ID: uuid.New(), TeamID: teamID, ContractID: contractID, EffectiveDate: eff, IsActive: true}
	err = s.runTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.DeactivateContractAssignment(ctx, tx, contractID); err != nil {
			return err
		}
		return s.repo.AssignContract(ctx, tx, c)
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListTeamContracts lists a team's contract assignments.
func (s *ApprovalConfigService) ListTeamContracts(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamContract, error) {
	return s.repo.ListTeamContracts(ctx, teamID)
}

// TeamMemberInput is the payload for adding/updating a team member.
type TeamMemberInput struct {
	UserID        uuid.UUID
	MemberType    vo.TeamMemberType
	PriorityOrder int
	IsActive      bool
}

// AddTeamMember adds a member to a team.
func (s *ApprovalConfigService) AddTeamMember(ctx context.Context, teamID uuid.UUID, in TeamMemberInput) (*entity.ApprovalTeamMember, error) {
	t, err := s.repo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.NotFound("approval team not found")
	}
	if in.UserID == uuid.Nil {
		return nil, domain.Validation("user_id is required")
	}
	mt := in.MemberType
	if mt == "" {
		mt = vo.TeamMemberReviewerAgent
	}
	if !vo.ValidTeamMemberType(mt) {
		return nil, domain.Validation("invalid member_type")
	}
	m := &entity.ApprovalTeamMember{ID: uuid.New(), TeamID: teamID, UserID: in.UserID, MemberType: mt, PriorityOrder: in.PriorityOrder, IsActive: in.IsActive}
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.AddTeamMember(ctx, tx, m) }); err != nil {
		return nil, err
	}
	return s.repo.GetTeamMember(ctx, m.ID)
}

// UpdateTeamMember updates a team member.
func (s *ApprovalConfigService) UpdateTeamMember(ctx context.Context, memberID uuid.UUID, in TeamMemberInput) (*entity.ApprovalTeamMember, error) {
	m, err := s.repo.GetTeamMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.NotFound("team member not found")
	}
	if in.MemberType != "" {
		if !vo.ValidTeamMemberType(in.MemberType) {
			return nil, domain.Validation("invalid member_type")
		}
		m.MemberType = in.MemberType
	}
	m.PriorityOrder = in.PriorityOrder
	m.IsActive = in.IsActive
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.UpdateTeamMember(ctx, tx, m) }); err != nil {
		return nil, err
	}
	return s.repo.GetTeamMember(ctx, memberID)
}

// RemoveTeamMember deletes a team member.
func (s *ApprovalConfigService) RemoveTeamMember(ctx context.Context, memberID uuid.UUID) error {
	m, err := s.repo.GetTeamMember(ctx, memberID)
	if err != nil {
		return err
	}
	if m == nil {
		return domain.NotFound("team member not found")
	}
	return s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.RemoveTeamMember(ctx, tx, memberID) })
}

// ListTeamMembers lists a team's members.
func (s *ApprovalConfigService) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error) {
	return s.repo.ListTeamMembers(ctx, teamID)
}

// ───────────────────────── Process configs ─────────────────────────

// StageInput describes one stage of an approval process.
type StageInput struct {
	StageNumber           int
	StageName             string
	ApproverMode          vo.ApproverMode
	ApproverUserID        *uuid.UUID
	ApprovalGroupID       *uuid.UUID
	RequiredApprovalCount int
	IsFinalStage          bool
}

// ProcessConfigInput is the payload for creating/updating a process config.
type ProcessConfigInput struct {
	ProcessCode          string
	ProcessName          string
	ProcessType          vo.ProcessType
	ContractType         vo.ContractType
	ContractID           *uuid.UUID
	EffectiveDate        *time.Time
	IsActive             bool
	GroupApprovalEnabled bool
	RequireTeamApproval  bool
	Stages               []StageInput
}

func validateStages(stages []StageInput) ([]entity.ApprovalProcessStage, error) {
	if len(stages) == 0 {
		return nil, domain.Validation("at least one approval stage is required")
	}
	if len(stages) > 3 {
		return nil, domain.Validation("a maximum of 3 approval stages is allowed")
	}
	seen := map[int]bool{}
	out := make([]entity.ApprovalProcessStage, 0, len(stages))
	for _, st := range stages {
		if st.StageNumber < 1 || st.StageNumber > 3 {
			return nil, domain.Validation("stage_number must be between 1 and 3")
		}
		if seen[st.StageNumber] {
			return nil, domain.Validation("duplicate stage_number")
		}
		seen[st.StageNumber] = true
		if !vo.ValidApproverMode(st.ApproverMode) {
			return nil, domain.Validation("invalid approver_mode")
		}
		switch st.ApproverMode {
		case vo.ApproverModeSingleUser:
			if st.ApproverUserID == nil {
				return nil, domain.Validation("approver_user_id is required for SINGLE_USER stage")
			}
		case vo.ApproverModeGroupPriority, vo.ApproverModeGroupAny:
			if st.ApprovalGroupID == nil {
				return nil, domain.Validation("approval_group_id is required for group stages")
			}
		}
		req := st.RequiredApprovalCount
		if req < 1 {
			req = 1
		}
		out = append(out, entity.ApprovalProcessStage{
			ID:                    uuid.New(),
			StageNumber:           st.StageNumber,
			StageName:             strings.TrimSpace(st.StageName),
			ApproverMode:          st.ApproverMode,
			ApproverUserID:        st.ApproverUserID,
			ApprovalGroupID:       st.ApprovalGroupID,
			RequiredApprovalCount: req,
			IsFinalStage:          st.IsFinalStage,
			RejectPolicy:          "STOP",
		})
	}
	// Mark the highest stage number as final if none flagged.
	hasFinal := false
	maxIdx := 0
	for i, st := range out {
		if st.IsFinalStage {
			hasFinal = true
		}
		if st.StageNumber > out[maxIdx].StageNumber {
			maxIdx = i
		}
	}
	if !hasFinal {
		out[maxIdx].IsFinalStage = true
	}
	return out, nil
}

// CreateProcessConfig creates a process config with its stages.
func (s *ApprovalConfigService) CreateProcessConfig(ctx context.Context, in ProcessConfigInput, actor uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	code := strings.ToUpper(strings.TrimSpace(in.ProcessCode))
	if code == "" {
		return nil, domain.Validation("process_code is required")
	}
	if strings.TrimSpace(in.ProcessName) == "" {
		return nil, domain.Validation("process_name is required")
	}
	if !vo.ValidProcessType(in.ProcessType) {
		return nil, domain.Validation("invalid process_type")
	}
	ct := in.ContractType
	if ct == "" {
		ct = vo.ContractTypeCompany
	}
	if !vo.ValidContractType(ct) {
		return nil, domain.Validation("invalid contract_type")
	}
	stages, err := validateStages(in.Stages)
	if err != nil {
		return nil, err
	}
	eff := s.now()
	if in.EffectiveDate != nil {
		eff = *in.EffectiveDate
	}
	cfg := &entity.ApprovalProcessConfig{
		ID: uuid.New(), ProcessCode: code, ProcessName: strings.TrimSpace(in.ProcessName),
		ProcessType: in.ProcessType, ContractType: ct, ContractID: in.ContractID, EffectiveDate: eff,
		IsActive: in.IsActive, GroupApprovalEnabled: in.GroupApprovalEnabled, RequireTeamApproval: in.RequireTeamApproval,
		CreatedBy: &actor, UpdatedBy: &actor,
	}
	err = s.runTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.CreateConfig(ctx, tx, cfg); err != nil {
			return err
		}
		return s.repo.ReplaceStages(ctx, tx, cfg.ID, stages)
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetConfig(ctx, cfg.ID)
}

// UpdateProcessConfig updates a process config and replaces its stages.
func (s *ApprovalConfigService) UpdateProcessConfig(ctx context.Context, id uuid.UUID, in ProcessConfigInput, actor uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	cfg, err := s.repo.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, domain.NotFound("approval process config not found")
	}
	if strings.TrimSpace(in.ProcessName) != "" {
		cfg.ProcessName = strings.TrimSpace(in.ProcessName)
	}
	if vo.ValidProcessType(in.ProcessType) {
		cfg.ProcessType = in.ProcessType
	}
	if in.ContractType != "" {
		if !vo.ValidContractType(in.ContractType) {
			return nil, domain.Validation("invalid contract_type")
		}
		cfg.ContractType = in.ContractType
	}
	cfg.ContractID = in.ContractID
	if in.EffectiveDate != nil {
		cfg.EffectiveDate = *in.EffectiveDate
	}
	cfg.IsActive = in.IsActive
	cfg.GroupApprovalEnabled = in.GroupApprovalEnabled
	cfg.RequireTeamApproval = in.RequireTeamApproval
	cfg.UpdatedBy = &actor

	var stages []entity.ApprovalProcessStage
	if len(in.Stages) > 0 {
		stages, err = validateStages(in.Stages)
		if err != nil {
			return nil, err
		}
	}
	err = s.runTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.UpdateConfig(ctx, tx, cfg); err != nil {
			return err
		}
		if stages != nil {
			return s.repo.ReplaceStages(ctx, tx, cfg.ID, stages)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.GetConfig(ctx, id)
}

// SetProcessConfigActive activates/deactivates a process config.
func (s *ApprovalConfigService) SetProcessConfigActive(ctx context.Context, id uuid.UUID, active bool, actor uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	cfg, err := s.repo.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, domain.NotFound("approval process config not found")
	}
	if err := s.runTx(ctx, func(tx pgx.Tx) error { return s.repo.SetConfigActive(ctx, tx, id, active, &actor) }); err != nil {
		return nil, err
	}
	return s.repo.GetConfig(ctx, id)
}

// ListProcessConfigs lists process configs.
func (s *ApprovalConfigService) ListProcessConfigs(ctx context.Context, f domain.ProcessListFilter) ([]*entity.ApprovalProcessConfig, error) {
	return s.repo.ListConfigs(ctx, f)
}

// GetProcessConfigDetail returns a config with stages.
func (s *ApprovalConfigService) GetProcessConfigDetail(ctx context.Context, id uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	cfg, err := s.repo.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, domain.NotFound("approval process config not found")
	}
	return cfg, nil
}
