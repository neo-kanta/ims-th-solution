package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// fakeRepo is an in-memory implementation of domain.Repository for unit tests.
// The pgx.Tx parameter is ignored (tests inject a no-op tx runner).
type fakeRepo struct {
	groups        map[uuid.UUID]*entity.ApprovalGroup
	groupMembers  map[uuid.UUID]*entity.ApprovalGroupMember
	teams         map[uuid.UUID]*entity.ApprovalTeam
	teamContracts map[uuid.UUID]*entity.ApprovalTeamContract
	teamMembers   map[uuid.UUID]*entity.ApprovalTeamMember
	configs       map[uuid.UUID]*entity.ApprovalProcessConfig
	requests      map[uuid.UUID]*entity.ApprovalRequest
	tasks         map[uuid.UUID]*entity.ApprovalTask
	events        []*entity.ApprovalEvent
	signatures    []*entity.ApprovalSignatureRecord
	seq           int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		groups:        map[uuid.UUID]*entity.ApprovalGroup{},
		groupMembers:  map[uuid.UUID]*entity.ApprovalGroupMember{},
		teams:         map[uuid.UUID]*entity.ApprovalTeam{},
		teamContracts: map[uuid.UUID]*entity.ApprovalTeamContract{},
		teamMembers:   map[uuid.UUID]*entity.ApprovalTeamMember{},
		configs:       map[uuid.UUID]*entity.ApprovalProcessConfig{},
		requests:      map[uuid.UUID]*entity.ApprovalRequest{},
		tasks:         map[uuid.UUID]*entity.ApprovalTask{},
	}
}

var _ domain.Repository = (*fakeRepo)(nil)

func copyRequest(r *entity.ApprovalRequest) *entity.ApprovalRequest {
	if r == nil {
		return nil
	}
	c := *r
	return &c
}

func copyTask(t *entity.ApprovalTask) *entity.ApprovalTask {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}

// ── Groups ──
func (f *fakeRepo) CreateGroup(_ context.Context, _ pgx.Tx, g *entity.ApprovalGroup) error {
	f.groups[g.ID] = g
	return nil
}
func (f *fakeRepo) UpdateGroup(_ context.Context, _ pgx.Tx, g *entity.ApprovalGroup) error {
	f.groups[g.ID] = g
	return nil
}
func (f *fakeRepo) GetGroup(_ context.Context, id uuid.UUID) (*entity.ApprovalGroup, error) {
	return f.groups[id], nil
}
func (f *fakeRepo) GetGroupByCode(_ context.Context, code string) (*entity.ApprovalGroup, error) {
	for _, g := range f.groups {
		if g.GroupCode == code {
			return g, nil
		}
	}
	return nil, nil
}
func (f *fakeRepo) ListGroups(_ context.Context, _ domain.GroupListFilter) ([]*entity.ApprovalGroup, error) {
	var out []*entity.ApprovalGroup
	for _, g := range f.groups {
		out = append(out, g)
	}
	return out, nil
}
func (f *fakeRepo) AddGroupMember(_ context.Context, _ pgx.Tx, m *entity.ApprovalGroupMember) error {
	f.groupMembers[m.ID] = m
	return nil
}
func (f *fakeRepo) UpdateGroupMember(_ context.Context, _ pgx.Tx, m *entity.ApprovalGroupMember) error {
	f.groupMembers[m.ID] = m
	return nil
}
func (f *fakeRepo) GetGroupMember(_ context.Context, id uuid.UUID) (*entity.ApprovalGroupMember, error) {
	return f.groupMembers[id], nil
}
func (f *fakeRepo) ListGroupMembers(_ context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	return f.membersByGroup(groupID, false), nil
}
func (f *fakeRepo) ListEligibleGroupMembers(_ context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	return f.membersByGroup(groupID, true), nil
}
func (f *fakeRepo) membersByGroup(groupID uuid.UUID, eligibleOnly bool) []*entity.ApprovalGroupMember {
	var out []*entity.ApprovalGroupMember
	for _, m := range f.groupMembers {
		if m.GroupID != groupID {
			continue
		}
		if eligibleOnly && !m.Eligible() {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PriorityOrder < out[j].PriorityOrder })
	return out
}

// ── Teams ──
func (f *fakeRepo) CreateTeam(_ context.Context, _ pgx.Tx, t *entity.ApprovalTeam) error {
	f.teams[t.ID] = t
	return nil
}
func (f *fakeRepo) UpdateTeam(_ context.Context, _ pgx.Tx, t *entity.ApprovalTeam) error {
	f.teams[t.ID] = t
	return nil
}
func (f *fakeRepo) GetTeam(_ context.Context, id uuid.UUID) (*entity.ApprovalTeam, error) {
	return f.teams[id], nil
}
func (f *fakeRepo) GetTeamByCode(_ context.Context, code string) (*entity.ApprovalTeam, error) {
	for _, t := range f.teams {
		if t.TeamCode == code {
			return t, nil
		}
	}
	return nil, nil
}
func (f *fakeRepo) ListTeams(_ context.Context) ([]*entity.ApprovalTeam, error) {
	var out []*entity.ApprovalTeam
	for _, t := range f.teams {
		out = append(out, t)
	}
	return out, nil
}
func (f *fakeRepo) AssignContract(_ context.Context, _ pgx.Tx, c *entity.ApprovalTeamContract) error {
	f.teamContracts[c.ID] = c
	return nil
}
func (f *fakeRepo) DeactivateContractAssignment(_ context.Context, _ pgx.Tx, contractID uuid.UUID) error {
	for _, c := range f.teamContracts {
		if c.ContractID == contractID {
			c.IsActive = false
		}
	}
	return nil
}
func (f *fakeRepo) GetActiveTeamForContract(_ context.Context, contractID uuid.UUID) (*entity.ApprovalTeam, error) {
	for _, c := range f.teamContracts {
		if c.ContractID == contractID && c.IsActive {
			return f.teams[c.TeamID], nil
		}
	}
	return nil, nil
}
func (f *fakeRepo) ListTeamContracts(_ context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamContract, error) {
	var out []*entity.ApprovalTeamContract
	for _, c := range f.teamContracts {
		if c.TeamID == teamID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (f *fakeRepo) AddTeamMember(_ context.Context, _ pgx.Tx, m *entity.ApprovalTeamMember) error {
	f.teamMembers[m.ID] = m
	return nil
}
func (f *fakeRepo) UpdateTeamMember(_ context.Context, _ pgx.Tx, m *entity.ApprovalTeamMember) error {
	f.teamMembers[m.ID] = m
	return nil
}
func (f *fakeRepo) RemoveTeamMember(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(f.teamMembers, id)
	return nil
}
func (f *fakeRepo) GetTeamMember(_ context.Context, id uuid.UUID) (*entity.ApprovalTeamMember, error) {
	return f.teamMembers[id], nil
}
func (f *fakeRepo) ListTeamMembers(_ context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error) {
	return f.teamMembersBy(teamID, false), nil
}
func (f *fakeRepo) ListEligibleTeamMembers(_ context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error) {
	return f.teamMembersBy(teamID, true), nil
}
func (f *fakeRepo) teamMembersBy(teamID uuid.UUID, eligibleOnly bool) []*entity.ApprovalTeamMember {
	var out []*entity.ApprovalTeamMember
	for _, m := range f.teamMembers {
		if m.TeamID != teamID {
			continue
		}
		if eligibleOnly && !m.Eligible() {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PriorityOrder < out[j].PriorityOrder })
	return out
}

// ── Process configs ──
func (f *fakeRepo) CreateConfig(_ context.Context, _ pgx.Tx, c *entity.ApprovalProcessConfig) error {
	f.configs[c.ID] = c
	return nil
}
func (f *fakeRepo) UpdateConfig(_ context.Context, _ pgx.Tx, c *entity.ApprovalProcessConfig) error {
	f.configs[c.ID] = c
	return nil
}
func (f *fakeRepo) SetConfigActive(_ context.Context, _ pgx.Tx, id uuid.UUID, active bool, _ *uuid.UUID) error {
	if c := f.configs[id]; c != nil {
		c.IsActive = active
	}
	return nil
}
func (f *fakeRepo) GetConfig(_ context.Context, id uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	return f.configs[id], nil
}
func (f *fakeRepo) GetConfigByCode(_ context.Context, code string) (*entity.ApprovalProcessConfig, error) {
	for _, c := range f.configs {
		if c.ProcessCode == code {
			return c, nil
		}
	}
	return nil, nil
}
func (f *fakeRepo) ListConfigs(_ context.Context, _ domain.ProcessListFilter) ([]*entity.ApprovalProcessConfig, error) {
	var out []*entity.ApprovalProcessConfig
	for _, c := range f.configs {
		out = append(out, c)
	}
	return out, nil
}
func (f *fakeRepo) ReplaceStages(_ context.Context, _ pgx.Tx, configID uuid.UUID, stages []entity.ApprovalProcessStage) error {
	if c := f.configs[configID]; c != nil {
		c.Stages = stages
	}
	return nil
}
func (f *fakeRepo) ListStages(_ context.Context, configID uuid.UUID) ([]entity.ApprovalProcessStage, error) {
	if c := f.configs[configID]; c != nil {
		return c.Stages, nil
	}
	return nil, nil
}
func (f *fakeRepo) Resolve(_ context.Context, pt vo.ProcessType, contractID *uuid.UUID, ct vo.ContractType, _ time.Time) (*entity.ApprovalProcessConfig, error) {
	var best *entity.ApprovalProcessConfig
	bestRank := 99
	for _, c := range f.configs {
		if !c.IsActive || c.ProcessType != pt {
			continue
		}
		rank := 99
		switch {
		case contractID != nil && c.ContractID != nil && *c.ContractID == *contractID:
			rank = 1
		case c.ContractID == nil && c.ContractType == ct:
			rank = 2
		case c.ContractID == nil && c.ContractType == vo.ContractTypeCompany:
			rank = 3
		}
		if rank < bestRank {
			bestRank = rank
			best = c
		}
	}
	return best, nil
}

// ── Requests ──
func (f *fakeRepo) NextRequestNumber(_ context.Context, _ pgx.Tx) (string, error) {
	f.seq++
	return fmt.Sprintf("APR-%06d", f.seq), nil
}
func (f *fakeRepo) CreateRequest(_ context.Context, _ pgx.Tx, r *entity.ApprovalRequest) error {
	if existing, _ := f.GetActiveBySubject(context.Background(), r.SubjectType, r.SubjectID); existing != nil {
		return domain.NewError(domain.ErrDuplicateActiveRequest, domain.ErrDuplicateActiveRequest.Error())
	}
	f.requests[r.ID] = copyRequest(r)
	return nil
}
func (f *fakeRepo) GetRequest(_ context.Context, id uuid.UUID) (*entity.ApprovalRequest, error) {
	return copyRequest(f.requests[id]), nil
}
func (f *fakeRepo) GetRequestForUpdate(_ context.Context, _ pgx.Tx, id uuid.UUID) (*entity.ApprovalRequest, error) {
	return copyRequest(f.requests[id]), nil
}
func (f *fakeRepo) UpdateRequestState(_ context.Context, _ pgx.Tx, r *entity.ApprovalRequest) error {
	f.requests[r.ID] = copyRequest(r)
	return nil
}
func (f *fakeRepo) ListRequests(_ context.Context, _ domain.RequestListFilter) ([]*entity.ApprovalRequest, int, error) {
	var out []*entity.ApprovalRequest
	for _, r := range f.requests {
		out = append(out, copyRequest(r))
	}
	return out, len(out), nil
}
func (f *fakeRepo) GetActiveBySubject(_ context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error) {
	for _, r := range f.requests {
		if r.SubjectType == st && r.SubjectID == subjectID && !r.Status.IsTerminal() {
			return copyRequest(r), nil
		}
	}
	return nil, nil
}
func (f *fakeRepo) GetLatestBySubject(_ context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error) {
	var latest *entity.ApprovalRequest
	for _, r := range f.requests {
		if r.SubjectType == st && r.SubjectID == subjectID {
			if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
				latest = r
			}
		}
	}
	return copyRequest(latest), nil
}

// ── Tasks ──
func (f *fakeRepo) CreateTask(_ context.Context, _ pgx.Tx, t *entity.ApprovalTask) error {
	f.tasks[t.ID] = copyTask(t)
	return nil
}
func (f *fakeRepo) GetTask(_ context.Context, id uuid.UUID) (*entity.ApprovalTask, error) {
	return copyTask(f.tasks[id]), nil
}
func (f *fakeRepo) GetTaskForUpdate(_ context.Context, _ pgx.Tx, id uuid.UUID) (*entity.ApprovalTask, error) {
	return copyTask(f.tasks[id]), nil
}
func (f *fakeRepo) UpdateTask(_ context.Context, _ pgx.Tx, t *entity.ApprovalTask) error {
	f.tasks[t.ID] = copyTask(t)
	return nil
}
func (f *fakeRepo) ListTasksByRequest(_ context.Context, requestID uuid.UUID) ([]*entity.ApprovalTask, error) {
	var out []*entity.ApprovalTask
	for _, t := range f.tasks {
		if t.ApprovalRequestID == requestID {
			out = append(out, copyTask(t))
		}
	}
	return out, nil
}
func (f *fakeRepo) ListPendingByStage(_ context.Context, _ pgx.Tx, requestID uuid.UUID, stage int) ([]*entity.ApprovalTask, error) {
	var out []*entity.ApprovalTask
	for _, t := range f.tasks {
		if t.ApprovalRequestID == requestID && t.StageNumber == stage && t.Status == vo.TaskStatusPending {
			out = append(out, copyTask(t))
		}
	}
	return out, nil
}
func (f *fakeRepo) CountApprovedInStage(_ context.Context, _ pgx.Tx, requestID uuid.UUID, stage int) (int, error) {
	n := 0
	for _, t := range f.tasks {
		if t.ApprovalRequestID == requestID && t.StageNumber == stage && t.Status == vo.TaskStatusApproved {
			n++
		}
	}
	return n, nil
}
func (f *fakeRepo) CancelPendingByRequest(_ context.Context, _ pgx.Tx, requestID uuid.UUID) error {
	for _, t := range f.tasks {
		if t.ApprovalRequestID == requestID && t.Status == vo.TaskStatusPending {
			t.Status = vo.TaskStatusCancelled
		}
	}
	return nil
}
func (f *fakeRepo) SkipOtherPendingInStage(_ context.Context, _ pgx.Tx, requestID uuid.UUID, stage int, except uuid.UUID) error {
	for _, t := range f.tasks {
		if t.ApprovalRequestID == requestID && t.StageNumber == stage && t.Status == vo.TaskStatusPending && t.ID != except {
			t.Status = vo.TaskStatusSkipped
		}
	}
	return nil
}
func (f *fakeRepo) Inbox(_ context.Context, fl domain.InboxFilter) ([]*domain.InboxItem, int, error) {
	status := fl.Status
	if status == "" {
		status = vo.TaskStatusPending
	}
	var out []*domain.InboxItem
	for _, t := range f.tasks {
		if t.AssignedUserID != nil && *t.AssignedUserID == fl.UserID && t.Status == status {
			r := f.requests[t.ApprovalRequestID]
			if r == nil {
				continue
			}
			out = append(out, &domain.InboxItem{Task: *copyTask(t), Request: *copyRequest(r)})
		}
	}
	return out, len(out), nil
}

// ── Events ──
func (f *fakeRepo) AppendEvent(_ context.Context, _ pgx.Tx, e *entity.ApprovalEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	cp := *e
	f.events = append(f.events, &cp)
	return nil
}
func (f *fakeRepo) ListEvents(_ context.Context, requestID uuid.UUID) ([]*entity.ApprovalEvent, error) {
	var out []*entity.ApprovalEvent
	for _, e := range f.events {
		if e.ApprovalRequestID == requestID {
			out = append(out, e)
		}
	}
	return out, nil
}

// ── Signatures ──
func (f *fakeRepo) CreateSignature(_ context.Context, _ pgx.Tx, s *entity.ApprovalSignatureRecord) error {
	cp := *s
	f.signatures = append(f.signatures, &cp)
	return nil
}
func (f *fakeRepo) ListSignatures(_ context.Context, requestID uuid.UUID) ([]*entity.ApprovalSignatureRecord, error) {
	var out []*entity.ApprovalSignatureRecord
	for _, s := range f.signatures {
		if s.ApprovalRequestID == requestID {
			out = append(out, s)
		}
	}
	return out, nil
}

// ── test helpers ──

// newTestRuntime builds a runtime service backed by the fake repo with a no-op
// tx runner (passes a nil pgx.Tx straight to the fake).
func newTestRuntime(repo *fakeRepo, opts ...func(*RuntimeDeps)) *ApprovalRuntimeService {
	d := RuntimeDeps{Repo: repo}
	for _, o := range opts {
		o(&d)
	}
	svc := NewApprovalRuntimeService(d)
	svc.runTxFn = func(ctx context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	return svc
}

// staticDelegate resolves a fixed delegate for a fixed principal.
type staticDelegate struct {
	principal uuid.UUID
	delegate  uuid.UUID
}

func (s staticDelegate) ResolveDelegate(_ context.Context, original uuid.UUID, _ *uuid.UUID, _ time.Time) (*domain.Delegate, error) {
	if original == s.principal {
		return &domain.Delegate{DelegateUserID: s.delegate}, nil
	}
	return nil, nil
}
