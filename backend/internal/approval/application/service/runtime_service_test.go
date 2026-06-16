package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

func ctx() context.Context { return context.Background() }

// seedGroup creates an active group with APPROVED, active members at the given
// priorities (slice index order). Returns the group id and member user ids.
func seedGroup(repo *fakeRepo, code string, userIDs ...uuid.UUID) uuid.UUID {
	gid := uuid.New()
	repo.groups[gid] = &entity.ApprovalGroup{ID: gid, GroupCode: code, GroupName: code, IsActive: true}
	for i, uid := range userIDs {
		mid := uuid.New()
		repo.groupMembers[mid] = &entity.ApprovalGroupMember{
			ID: mid, GroupID: gid, UserID: uid, PriorityOrder: i + 1,
			MemberType: vo.GroupMemberTypeMember, Status: vo.GroupMemberStatusApproved, IsActive: true,
		}
	}
	return gid
}

func seedConfig(repo *fakeRepo, pt vo.ProcessType, stages ...entity.ApprovalProcessStage) *entity.ApprovalProcessConfig {
	cid := uuid.New()
	for i := range stages {
		stages[i].ID = uuid.New()
		stages[i].ProcessConfigID = cid
		if stages[i].RequiredApprovalCount < 1 {
			stages[i].RequiredApprovalCount = 1
		}
		if stages[i].RejectPolicy == "" {
			stages[i].RejectPolicy = "STOP"
		}
	}
	cfg := &entity.ApprovalProcessConfig{
		ID: cid, ProcessCode: string(pt), ProcessName: string(pt), ProcessType: pt,
		ContractType: vo.ContractTypeCompany, IsActive: true, EffectiveDate: time.Now().AddDate(0, 0, -1),
		Stages: stages,
	}
	repo.configs[cid] = cfg
	return cfg
}

func submitFixture(repo *fakeRepo, svc *ApprovalRuntimeService, submitter uuid.UUID) (*entity.ApprovalRequest, error) {
	return svc.SubmitApproval(ctx(), SubmitInput{
		ProcessType: vo.ProcessInvestmentAnalysisReport,
		SubjectType: vo.SubjectResearchReport,
		SubjectID:   uuid.New(),
		SubmitterID: submitter,
	})
}

func pendingTaskFor(repo *fakeRepo, requestID, user uuid.UUID) *entity.ApprovalTask {
	for _, t := range repo.tasks {
		if t.ApprovalRequestID == requestID && t.Status == vo.TaskStatusPending &&
			t.AssignedUserID != nil && *t.AssignedUserID == user {
			return t
		}
	}
	return nil
}

//  1. Submit creates a request + first stage task; group-priority picks the
//     highest-priority eligible member.
func TestSubmit_GroupPriority_PicksHighestEligible(t *testing.T) {
	repo := newFakeRepo()
	high, low, submitter := uuid.New(), uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", high, low)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupPriority, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if req.Status != vo.RequestStatusPendingApproval {
		t.Fatalf("status = %s, want PENDING_APPROVAL", req.Status)
	}
	if pendingTaskFor(repo, req.ID, high) == nil {
		t.Fatalf("expected a pending task for the highest-priority member")
	}
	if pendingTaskFor(repo, req.ID, low) != nil {
		t.Fatalf("group-priority must create exactly one task (highest priority)")
	}
}

// 3. Submit excludes the submitter from group-priority selection.
func TestSubmit_GroupPriority_ExcludesSubmitter(t *testing.T) {
	repo := newFakeRepo()
	submitter, other := uuid.New(), uuid.New()
	// submitter is highest priority but must be skipped (maker-checker).
	gid := seedGroup(repo, "REVIEWERS", submitter, other)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupPriority, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if pendingTaskFor(repo, req.ID, submitter) != nil {
		t.Fatalf("submitter must never be assigned an approval task")
	}
	if pendingTaskFor(repo, req.ID, other) == nil {
		t.Fatalf("expected the non-submitter member to be assigned")
	}
}

// 4 + 5. Group-any creates tasks for all eligible members and completes after one
//
//	approval; other tasks are skipped.
func TestSubmitAndApprove_GroupAny_CompletesAfterOne(t *testing.T) {
	repo := newFakeRepo()
	a, b, submitter := uuid.New(), uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", a, b)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid,
		RequiredApprovalCount: 1, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if pendingTaskFor(repo, req.ID, a) == nil || pendingTaskFor(repo, req.ID, b) == nil {
		t.Fatalf("group-any must create a task for every eligible member")
	}

	taskA := pendingTaskFor(repo, req.ID, a)
	out, err := svc.ApproveTask(ctx(), taskA.ID, a, "ok")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("status = %s, want APPROVED", out.Status)
	}
	// b's task must be skipped.
	if tb := repo.tasks[pendingTaskID(repo, req.ID, b)]; tb != nil && tb.Status == vo.TaskStatusPending {
		t.Fatalf("other group-any task should be skipped after completion")
	}
}

func pendingTaskID(repo *fakeRepo, requestID, user uuid.UUID) uuid.UUID {
	for _, t := range repo.tasks {
		if t.ApprovalRequestID == requestID && t.AssignedUserID != nil && *t.AssignedUserID == user {
			return t.ID
		}
	}
	return uuid.Nil
}

// 6. Maker cannot approve their own request.
func TestApprove_MakerCannotApproveOwn(t *testing.T) {
	repo := newFakeRepo()
	submitter, approver := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	task := pendingTaskFor(repo, req.ID, approver)

	// Submitter tries to approve the approver's task → self-approval forbidden.
	_, err := svc.ApproveTask(ctx(), task.ID, submitter, "")
	if !errors.Is(err, domain.ErrSelfApproval) {
		t.Fatalf("err = %v, want ErrSelfApproval", err)
	}
}

// 6b. Two-stage flow: approving stage 1 advances to stage 2; final approval
//
//	completes the request and writes a signature per approval.
func TestApprove_TwoStage_AdvancesThenCompletes(t *testing.T) {
	repo := newFakeRepo()
	s1u, s2u, submitter := uuid.New(), uuid.New(), uuid.New()
	g1 := seedGroup(repo, "STAGE1", s1u)
	g2 := seedGroup(repo, "STAGE2", s2u)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport,
		entity.ApprovalProcessStage{StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &g1},
		entity.ApprovalProcessStage{StageNumber: 2, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &g2, IsFinalStage: true},
	)
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)

	t1 := pendingTaskFor(repo, req.ID, s1u)
	out, err := svc.ApproveTask(ctx(), t1.ID, s1u, "stage1 ok")
	if err != nil {
		t.Fatalf("approve stage1: %v", err)
	}
	if out.Status != vo.RequestStatusPendingApproval || out.CurrentStageNumber != 2 {
		t.Fatalf("after stage1: status=%s stage=%d, want PENDING_APPROVAL stage 2", out.Status, out.CurrentStageNumber)
	}
	t2 := pendingTaskFor(repo, req.ID, s2u)
	if t2 == nil {
		t.Fatalf("stage 2 task should be created")
	}
	out, err = svc.ApproveTask(ctx(), t2.ID, s2u, "stage2 ok")
	if err != nil {
		t.Fatalf("approve stage2: %v", err)
	}
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("final status = %s, want APPROVED", out.Status)
	}
	sigs, _ := repo.ListSignatures(ctx(), req.ID)
	if len(sigs) != 2 {
		t.Fatalf("expected 2 signature records (one per approval), got %d", len(sigs))
	}
	// REQUEST_COMPLETED event must exist.
	events, _ := repo.ListEvents(ctx(), req.ID)
	if !hasEvent(events, vo.EventRequestCompleted) {
		t.Fatalf("expected a REQUEST_COMPLETED timeline event")
	}
}

// 8. Reject stops the request and cancels remaining pending tasks.
func TestReject_StopsAndCancelsPending(t *testing.T) {
	repo := newFakeRepo()
	a, b, submitter := uuid.New(), uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", a, b)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid,
		RequiredApprovalCount: 2, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	ta := pendingTaskFor(repo, req.ID, a)

	out, err := svc.RejectTask(ctx(), ta.ID, a, "not acceptable")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if out.Status != vo.RequestStatusRejected {
		t.Fatalf("status = %s, want REJECTED", out.Status)
	}
	if out.RejectionReason == "" {
		t.Fatalf("rejection reason should be stored")
	}
	if tb := pendingTaskFor(repo, req.ID, b); tb != nil {
		t.Fatalf("remaining pending task should be cancelled after reject")
	}
}

// Reject requires a reason.
func TestReject_RequiresReason(t *testing.T) {
	repo := newFakeRepo()
	a, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", a)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	ta := pendingTaskFor(repo, req.ID, a)
	if _, err := svc.RejectTask(ctx(), ta.ID, a, "   "); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
}

// 9. Duplicate active submit for the same subject is prevented.
func TestSubmit_DuplicateActiveSubjectPrevented(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	subject := uuid.New()
	in := SubmitInput{ProcessType: vo.ProcessInvestmentAnalysisReport, SubjectType: vo.SubjectResearchReport, SubjectID: subject, SubmitterID: submitter}
	if _, err := svc.SubmitApproval(ctx(), in); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if _, err := svc.SubmitApproval(ctx(), in); !errors.Is(err, domain.ErrDuplicateActiveRequest) {
		t.Fatalf("err = %v, want ErrDuplicateActiveRequest", err)
	}
}

// 10. Missing config returns a clean domain error.
func TestSubmit_MissingConfig(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestRuntime(repo)
	_, err := submitFixture(repo, svc, uuid.New())
	if !errors.Is(err, domain.ErrConfigNotFound) {
		t.Fatalf("err = %v, want ErrConfigNotFound", err)
	}
}

// 11. A user who is neither the assignee nor a valid delegate cannot approve.
func TestApprove_UnauthorizedUser(t *testing.T) {
	repo := newFakeRepo()
	approver, intruder, submitter := uuid.New(), uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	task := pendingTaskFor(repo, req.ID, approver)
	if _, err := svc.ApproveTask(ctx(), task.ID, intruder, ""); !errors.Is(err, domain.ErrNotAssigned) {
		t.Fatalf("err = %v, want ErrNotAssigned", err)
	}
}

// 12. Approving an already-finalised request's task is rejected (stale/duplicate).
func TestApprove_AfterCompletedRejected(t *testing.T) {
	repo := newFakeRepo()
	a, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", a)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	ta := pendingTaskFor(repo, req.ID, a)
	if _, err := svc.ApproveTask(ctx(), ta.ID, a, ""); err != nil {
		t.Fatalf("approve: %v", err)
	}
	// Re-approving the now-approved task must fail.
	if _, err := svc.ApproveTask(ctx(), ta.ID, a, ""); err == nil {
		t.Fatalf("expected error re-approving a finalised task")
	}
}

//  13. Delegated approval records the proxy/delegated markers and a DELEGATED
//     signature when the delegate resolver returns a valid delegate.
func TestApprove_DelegatedRecordsProxy(t *testing.T) {
	repo := newFakeRepo()
	principal, delegate, submitter := uuid.New(), uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", principal)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo, func(d *RuntimeDeps) {
		d.Delegate = staticDelegate{principal: principal, delegate: delegate}
	})
	req, _ := submitFixture(repo, svc, submitter)
	task := pendingTaskFor(repo, req.ID, principal)

	out, err := svc.ApproveTask(ctx(), task.ID, delegate, "acting as agent")
	if err != nil {
		t.Fatalf("delegated approve: %v", err)
	}
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("status = %s, want APPROVED", out.Status)
	}
	updated := repo.tasks[task.ID]
	if !updated.IsDelegatedAction || updated.DelegatedFromUserID == nil || *updated.DelegatedFromUserID != principal {
		t.Fatalf("task must record delegated action and principal")
	}
	sigs, _ := repo.ListSignatures(ctx(), req.ID)
	if len(sigs) != 1 || sigs[0].SignatureLabel != vo.SignatureLabelDelegated || !sigs[0].IsProxySignature {
		t.Fatalf("expected one DELEGATED proxy signature, got %+v", sigs)
	}
}

func hasEvent(events []*entity.ApprovalEvent, et vo.EventType) bool {
	for _, e := range events {
		if e.EventType == et {
			return true
		}
	}
	return false
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// 14. computeAllowedActions returns the correct action set based on request
// status and viewer assignment, keeping button visibility driven from backend.
func TestAllowedActions_ComputedByStatus(t *testing.T) {
	assignee := uuid.New()
	submitter := uuid.New()
	reqID := uuid.New()

	req := &entity.ApprovalRequest{ID: reqID, SubmitterID: submitter, Status: vo.RequestStatusPendingApproval}
	task := &entity.ApprovalTask{ID: uuid.New(), ApprovalRequestID: reqID, StageNumber: 1, Status: vo.TaskStatusPending, AssignedUserID: &assignee}

	// Assignee sees approve+reject but not withdraw (they are not the submitter).
	actions := computeAllowedActions(req, task, assignee)
	if !contains(actions, "approve") || !contains(actions, "reject") {
		t.Fatalf("assignee must see approve+reject, got %v", actions)
	}
	if contains(actions, "withdraw") {
		t.Fatalf("non-submitter assignee must not see withdraw, got %v", actions)
	}

	// Submitter with no task sees withdraw+cancel.
	actions = computeAllowedActions(req, nil, submitter)
	if !contains(actions, "withdraw") {
		t.Fatalf("submitter must see withdraw, got %v", actions)
	}
	if contains(actions, "approve") || contains(actions, "reject") {
		t.Fatalf("submitter without task must not see approve/reject, got %v", actions)
	}

	// APPROVED status grants revoke to all viewers; terminal state removes withdraw/cancel.
	req.Status = vo.RequestStatusApproved
	actions = computeAllowedActions(req, nil, assignee)
	if !contains(actions, "revoke") {
		t.Fatalf("APPROVED request must expose revoke action, got %v", actions)
	}
	if contains(actions, "withdraw") || contains(actions, "cancel") {
		t.Fatalf("terminal APPROVED must not expose withdraw/cancel, got %v", actions)
	}
}

// 15. Revoking an APPROVED request transitions it to REVOKED and records a
// REVOKED timeline event; the audit trail is preserved.
func TestRevoke_ApprovedRequest_FlipsToRevoked(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)
	task := pendingTaskFor(repo, req.ID, approver)

	approved, err := svc.ApproveTask(ctx(), task.ID, approver, "looks good")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != vo.RequestStatusApproved {
		t.Fatalf("precondition: status = %s, want APPROVED", approved.Status)
	}

	revoked, err := svc.RevokeRequest(ctx(), req.ID, approver, "compliance review found an issue")
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if revoked.Status != vo.RequestStatusRevoked {
		t.Fatalf("status = %s, want REVOKED", revoked.Status)
	}
	if revoked.RejectionReason == "" {
		t.Fatalf("revocation reason must be stored")
	}
	events, _ := repo.ListEvents(ctx(), req.ID)
	if !hasEvent(events, vo.EventRevoked) {
		t.Fatalf("REVOKED event must be appended to timeline")
	}
	// Original approval events must still be present (immutable history).
	if !hasEvent(events, vo.EventRequestCompleted) {
		t.Fatalf("REQUEST_COMPLETED event must be preserved after revocation")
	}
}

// 16. Revoking a non-APPROVED request (e.g. PENDING_APPROVAL) returns ErrConflict.
func TestRevoke_NotApproved_Rejected(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, _ := submitFixture(repo, svc, submitter)

	_, err := svc.RevokeRequest(ctx(), req.ID, approver, "mistaken revoke")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

// 17. Actor denied by the subject access port cannot approve.
// The request is seeded directly so this test isolates the approve-time access
// check from the submit-time access check.
func TestApprove_SubjectAccessDenied(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	cfg := seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})

	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-000001",
		ProcessType: vo.ProcessInvestmentAnalysisReport, SubjectType: vo.SubjectResearchReport,
		SubjectID: uuid.New(), SubmitterID: submitter,
		ProcessConfigID: &cfg.ID,
		Status: vo.RequestStatusPendingApproval, CurrentStageNumber: 1,
	}
	taskID := uuid.New()
	repo.tasks[taskID] = &entity.ApprovalTask{
		ID: taskID, ApprovalRequestID: reqID, StageNumber: 1,
		AssignedUserID: &approver, Status: vo.TaskStatusPending,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.ApproveTask(ctx(), taskID, approver, "")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden (subject access denied)", err)
	}
}

// 18. TEAM_MINIMUM mode: stage completes only after the minimum stamp count is
// reached; a single approval is insufficient.
func TestApprove_TeamMinimum_MinimumStampCount(t *testing.T) {
	repo := newFakeRepo()
	m1, m2, m3, submitter := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contractID := uuid.New()

	// Create team with MinRequiredStamps = 2.
	teamID := uuid.New()
	repo.teams[teamID] = &entity.ApprovalTeam{
		ID: teamID, TeamCode: "FUND_TEAM", TeamName: "Fund Team",
		MinRequiredStamps: 2, IsActive: true,
	}
	// Associate team with contract.
	tcID := uuid.New()
	repo.teamContracts[tcID] = &entity.ApprovalTeamContract{
		ID: tcID, TeamID: teamID, ContractID: contractID, IsActive: true,
	}
	// Three eligible reviewer-agent members; m3 is the submitter so will be excluded.
	for i, uid := range []uuid.UUID{m1, m2, m3} {
		mid := uuid.New()
		repo.teamMembers[mid] = &entity.ApprovalTeamMember{
			ID: mid, TeamID: teamID, UserID: uid,
			MemberType: vo.TeamMemberReviewerAgent, PriorityOrder: i + 1, IsActive: true,
		}
	}

	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeTeamMinimum, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)

	req, err := svc.SubmitApproval(ctx(), SubmitInput{
		ProcessType:  vo.ProcessInvestmentAnalysisReport,
		SubjectType:  vo.SubjectResearchReport,
		SubjectID:    uuid.New(),
		SubmitterID:  submitter,
		ContractID:   &contractID,
		ContractType: vo.ContractTypeCompany,
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Three tasks created (submitter is not a team member, so no exclusion needed here).
	task1 := pendingTaskFor(repo, req.ID, m1)
	task2 := pendingTaskFor(repo, req.ID, m2)
	if task1 == nil || task2 == nil {
		t.Fatalf("expected pending tasks for m1 and m2")
	}

	// First approval: stage must still be in progress.
	out, err := svc.ApproveTask(ctx(), task1.ID, m1, "first stamp")
	if err != nil {
		t.Fatalf("first approval: %v", err)
	}
	if out.Status != vo.RequestStatusPendingApproval {
		t.Fatalf("after 1 approval (need 2): status = %s, want PENDING_APPROVAL", out.Status)
	}

	// Second approval: minimum reached, stage and request complete.
	out, err = svc.ApproveTask(ctx(), task2.ID, m2, "second stamp")
	if err != nil {
		t.Fatalf("second approval: %v", err)
	}
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("after 2 approvals (minimum met): status = %s, want APPROVED", out.Status)
	}
	sigs, _ := repo.ListSignatures(ctx(), req.ID)
	if len(sigs) != 2 {
		t.Fatalf("expected 2 signature records (one per stamp), got %d", len(sigs))
	}
}

// ── P0 Fixes #4/5: stale-subject guard blocks approve/reject ─────────────────

// errSubjectCancelled is the error returned by a validator that simulates a
// subject being cancelled before the approval was recorded.
var errSubjectCancelled = errors.New("subject has been cancelled")

type alwaysInvalidValidator struct{}

func (alwaysInvalidValidator) ValidateSubjectApprovable(context.Context, uuid.UUID) error {
	return errSubjectCancelled
}

// TestApprove_StaleSubject_Blocked verifies that approving a task is rejected
// when the business object is no longer in an approvable state (e.g. the
// research report was cancelled or invalidated after submission).
func TestApprove_StaleSubject_Blocked(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	svc.RegisterSubjectValidator(vo.SubjectResearchReport, alwaysInvalidValidator{})

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	task := pendingTaskFor(repo, req.ID, approver)

	_, err = svc.ApproveTask(ctx(), task.ID, approver, "")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict (subject no longer approvable)", err)
	}
}

// TestReject_StaleSubject_Blocked verifies the same guard fires on reject.
func TestReject_StaleSubject_Blocked(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	svc.RegisterSubjectValidator(vo.SubjectResearchReport, alwaysInvalidValidator{})

	req, _ := submitFixture(repo, svc, submitter)
	task := pendingTaskFor(repo, req.ID, approver)

	_, err := svc.RejectTask(ctx(), task.ID, approver, "reject reason")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict (subject no longer approvable)", err)
	}
}

// ── P0 Fix #6: callback failures are not silently swallowed ──────────────────

// recordingAudit captures audit events emitted by the service. Used here to
// verify that callback failures are written to the audit trail.
type recordingAudit struct {
	events []string // event type strings
}

func (a *recordingAudit) Record(_ context.Context, _ *uuid.UUID, eventType, _, _ string, _ map[string]any) {
	a.events = append(a.events, eventType)
}

// failingSync always returns an error from the final-decision callbacks.
type failingSync struct{}

func (failingSync) OnApproved(_ context.Context, _ vo.SubjectType, _ uuid.UUID, _ uuid.UUID) error {
	return errors.New("downstream system unavailable")
}
func (failingSync) OnRejected(_ context.Context, _ vo.SubjectType, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return errors.New("downstream system unavailable")
}

// TestRunPostAction_CallbackFailure_AuditRecorded verifies that when the
// SubjectSync callback returns an error after a final approval, the error is
// recorded in the audit trail rather than being silently discarded.
func TestRunPostAction_CallbackFailure_AuditRecorded(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	audit := &recordingAudit{}
	svc := newTestRuntime(repo, func(d *RuntimeDeps) {
		d.Audit = audit
	})
	svc.RegisterSubjectSync(vo.SubjectResearchReport, failingSync{})

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	task := pendingTaskFor(repo, req.ID, approver)

	out, err := svc.ApproveTask(ctx(), task.ID, approver, "")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	// The approval must still be committed even when the callback fails.
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("request status = %s, want APPROVED", out.Status)
	}
	// The audit trail must record the callback failure.
	for _, ev := range audit.events {
		if ev == "APPROVAL_SYNC_FAILED" {
			return
		}
	}
	t.Fatalf("expected APPROVAL_SYNC_FAILED audit event; got %v", audit.events)
}

// ── Subject access port: read and write enforcement ───────────────────────────

// TestGetApprovalRequest_SubjectAccessDenied verifies that GetApprovalRequest
// checks the subject access port before disclosing the request detail.
func TestGetApprovalRequest_SubjectAccessDenied(t *testing.T) {
	repo := newFakeRepo()
	subjectID := uuid.New()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-000001",
		ProcessType: vo.ProcessInvestmentAnalysisReport, SubjectType: vo.SubjectResearchReport,
		SubjectID: subjectID, SubmitterID: uuid.New(),
		Status: vo.RequestStatusPendingApproval, CurrentStageNumber: 1,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.GetApprovalRequest(ctx(), reqID, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

// TestListRequests_AccessPortFiltersInaccessibleSubjects verifies that
// ListRequests filters out subjects the viewer cannot access via the port.
func TestListRequests_AccessPortFiltersInaccessibleSubjects(t *testing.T) {
	repo := newFakeRepo()
	subjectID := uuid.New()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: subjectID,
		Status: vo.RequestStatusPendingApproval,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	results, _, err := svc.ListRequests(ctx(), domain.RequestListFilter{ViewerID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results (subject denied), got %d", len(results))
	}
}

// TestGetSubjectApprovalStatus_AccessDenied verifies that GetSubjectApprovalStatus
// checks the subject access port when viewerID is non-zero.
func TestGetSubjectApprovalStatus_AccessDenied(t *testing.T) {
	repo := newFakeRepo()
	subjectID := uuid.New()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-000001",
		ProcessType: vo.ProcessInvestmentAnalysisReport, SubjectType: vo.SubjectResearchReport,
		SubjectID: subjectID, SubmitterID: uuid.New(),
		Status: vo.RequestStatusPendingApproval,
		CurrentStageNumber: 1, CreatedAt: time.Now(),
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.GetSubjectApprovalStatus(ctx(), vo.SubjectResearchReport, subjectID, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

// TestGetSubjectApprovalStatus_InternalBypass verifies that viewerID=uuid.Nil
// skips the access port check (used by GetApprovalStage for badge enrichment).
func TestGetSubjectApprovalStatus_InternalBypass(t *testing.T) {
	repo := newFakeRepo()
	subjectID := uuid.New()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: subjectID,
		Status: vo.RequestStatusPendingApproval, CreatedAt: time.Now(),
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	// uuid.Nil bypass must succeed even with the denying port registered.
	req, err := svc.GetSubjectApprovalStatus(ctx(), vo.SubjectResearchReport, subjectID, uuid.Nil)
	if err != nil {
		t.Fatalf("internal bypass: %v", err)
	}
	if req == nil {
		t.Fatalf("expected request, got nil")
	}
}

// TestSubmitApproval_SubjectAccessDenied verifies that SubmitApproval checks
// the subject access port before creating a request.
func TestSubmitApproval_SubjectAccessDenied(t *testing.T) {
	repo := newFakeRepo()
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, IsFinalStage: true,
	})

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.SubmitApproval(ctx(), SubmitInput{
		ProcessType: vo.ProcessInvestmentAnalysisReport,
		SubjectType: vo.SubjectResearchReport,
		SubjectID:   uuid.New(),
		SubmitterID: uuid.New(),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

// TestRevokeRequest_SubjectAccessDenied verifies that RevokeRequest checks the
// subject access port before revoking.
func TestRevokeRequest_SubjectAccessDenied(t *testing.T) {
	repo := newFakeRepo()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-000001",
		ProcessType: vo.ProcessInvestmentAnalysisReport, SubjectType: vo.SubjectResearchReport,
		SubjectID: uuid.New(), SubmitterID: uuid.New(),
		Status: vo.RequestStatusApproved,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.RevokeRequest(ctx(), reqID, uuid.New(), "revoke reason")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

// ── New tests: cross-subject filtering, timeline, missing port, zero actor ───

// TestListRequests_SubjectPortFiltersInaccessible verifies that when a ViewerID
// is set, ListRequests post-filters results to only subjects the viewer can see.
func TestListRequests_SubjectPortFiltersInaccessible(t *testing.T) {
	repo := newFakeRepo()

	allowedSubjectID := uuid.New()
	deniedSubjectID := uuid.New()

	reqAllowed := uuid.New()
	repo.requests[reqAllowed] = &entity.ApprovalRequest{
		ID: reqAllowed, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: allowedSubjectID,
		Status: vo.RequestStatusPendingApproval,
	}
	reqDenied := uuid.New()
	repo.requests[reqDenied] = &entity.ApprovalRequest{
		ID: reqDenied, RequestNumber: "APR-002",
		SubjectType: vo.SubjectResearchReport, SubjectID: deniedSubjectID,
		Status: vo.RequestStatusPendingApproval,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, selectiveSubjectPort{
		allowed: map[uuid.UUID]bool{allowedSubjectID: true},
	})

	results, total, err := svc.ListRequests(ctx(), domain.RequestListFilter{ViewerID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(results) != 1 {
		t.Fatalf("expected 1 permitted result, got %d (total=%d)", len(results), total)
	}
	if results[0].ID != reqAllowed {
		t.Fatalf("expected the allowed request, got %v", results[0].ID)
	}
}

// TestGetMyInbox_SubjectPortFiltersInaccessible verifies that GetMyInbox filters
// inbox items to only subjects the actor can access via the subject access port.
func TestGetMyInbox_SubjectPortFiltersInaccessible(t *testing.T) {
	repo := newFakeRepo()

	actor := uuid.New()
	allowedSubjectID := uuid.New()
	deniedSubjectID := uuid.New()

	for i, subjectID := range []uuid.UUID{allowedSubjectID, deniedSubjectID} {
		reqID := uuid.New()
		taskID := uuid.New()
		sid := subjectID
		repo.requests[reqID] = &entity.ApprovalRequest{
			ID: reqID, RequestNumber: "APR-" + string(rune('A'+i)),
			SubjectType: vo.SubjectResearchReport, SubjectID: sid,
			Status: vo.RequestStatusPendingApproval,
		}
		repo.tasks[taskID] = &entity.ApprovalTask{
			ID: taskID, ApprovalRequestID: reqID,
			AssignedUserID: &actor, Status: vo.TaskStatusPending,
		}
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, selectiveSubjectPort{
		allowed: map[uuid.UUID]bool{allowedSubjectID: true},
	})

	items, total, err := svc.GetMyInbox(ctx(), domain.InboxFilter{UserID: actor})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 inbox item, got %d (total=%d)", len(items), total)
	}
	if items[0].Request.SubjectID != allowedSubjectID {
		t.Fatalf("expected item for allowed subject, got %v", items[0].Request.SubjectID)
	}
}

// TestGetApprovalTimeline_SubjectAccessDenied verifies that GetApprovalTimeline
// enforces subject access before returning events.
func TestGetApprovalTimeline_SubjectAccessDenied(t *testing.T) {
	repo := newFakeRepo()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: uuid.New(),
		Status: vo.RequestStatusPendingApproval,
	}

	svc := newTestRuntime(repo)
	svc.RegisterSubjectAccessPort(vo.SubjectResearchReport, denyingSubjectPort{})

	_, err := svc.GetApprovalTimeline(ctx(), reqID, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

// TestRevokeRequest_SyncCallbackFailure_AuditRecorded verifies that a failing
// sync callback after a revoke is surfaced in the audit trail rather than
// silently discarded.
func TestRevokeRequest_SyncCallbackFailure_AuditRecorded(t *testing.T) {
	repo := newFakeRepo()
	contractID := uuid.New()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: uuid.New(),
		ContractID: &contractID, Status: vo.RequestStatusApproved,
	}

	audit := &recordingAudit{}
	svc := newTestRuntime(repo, func(d *RuntimeDeps) {
		d.Audit = audit
	})
	svc.RegisterSubjectSync(vo.SubjectResearchReport, failingSync{})

	_, err := svc.RevokeRequest(ctx(), reqID, uuid.New(), "revoke reason")
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	for _, ev := range audit.events {
		if ev == "APPROVAL_REVOKE_SYNC_FAILED" {
			return
		}
	}
	t.Fatalf("expected APPROVAL_REVOKE_SYNC_FAILED audit event; got %v", audit.events)
}

// TestCancelApprovalBySubject_InvalidSubjectType_ReturnsValidationError verifies
// that CancelApprovalBySubject rejects an unknown subject type with a validation
// error rather than silently treating it as a no-op.
func TestCancelApprovalBySubject_InvalidSubjectType_ReturnsValidationError(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestRuntime(repo)

	err := svc.CancelApprovalBySubject(ctx(), vo.SubjectType("INVALID_TYPE"), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation for invalid subject type", err)
	}
}

// ── Subject access port: missing port, zero actor, generic subject ────────────

// TestSubjectAccessPort_MissingPort_FailsClosed verifies that when no port is
// registered for a subject type, the engine rejects reads with ErrForbidden
// rather than silently disclosing the request (fail-closed design).
func TestSubjectAccessPort_MissingPort_FailsClosed(t *testing.T) {
	repo := newFakeRepo()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: uuid.New(),
		Status: vo.RequestStatusPendingApproval,
	}

	svc := newBareTestRuntime(repo) // no ports registered

	_, err := svc.GetApprovalRequest(ctx(), reqID, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("missing port must return ErrForbidden, got %v", err)
	}
}

// TestSubjectAccessPort_MissingPort_ListFiltersAll verifies that when no port
// is registered, ListRequests with a ViewerID returns no items (fail closed).
func TestSubjectAccessPort_MissingPort_ListFiltersAll(t *testing.T) {
	repo := newFakeRepo()
	reqID := uuid.New()
	repo.requests[reqID] = &entity.ApprovalRequest{
		ID: reqID, RequestNumber: "APR-001",
		SubjectType: vo.SubjectResearchReport, SubjectID: uuid.New(),
		Status: vo.RequestStatusPendingApproval,
	}

	svc := newBareTestRuntime(repo)

	results, _, err := svc.ListRequests(ctx(), domain.RequestListFilter{ViewerID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("missing port must filter all results (fail closed), got %d", len(results))
	}
}

// TestZeroActorID_Rejected verifies that a zero actorID is rejected before any
// subject access port is consulted; no data is disclosed.
func TestZeroActorID_Rejected(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)
	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	task := pendingTaskFor(repo, req.ID, approver)

	_, err = svc.ApproveTask(ctx(), task.ID, uuid.Nil, "")
	if err == nil {
		t.Fatalf("expected error for zero actorID, got nil")
	}
}

// TestSubmit_GenericSubject_NoContractID verifies that subjects without a
// contract_id can be submitted and approved successfully. The approval module
// must not require a contract_id for any part of its workflow.
func TestSubmit_GenericSubject_NoContractID(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	svc := newTestRuntime(repo)

	req, err := svc.SubmitApproval(ctx(), SubmitInput{
		ProcessType: vo.ProcessInvestmentAnalysisReport,
		SubjectType: vo.SubjectResearchReport,
		SubjectID:   uuid.New(),
		SubmitterID: submitter,
		// ContractID intentionally omitted — generic subject
	})
	if err != nil {
		t.Fatalf("generic subject submit: %v", err)
	}
	if req.ContractID != nil {
		t.Fatalf("ContractID should be nil for generic subject, got %v", req.ContractID)
	}

	task := pendingTaskFor(repo, req.ID, approver)
	out, err := svc.ApproveTask(ctx(), task.ID, approver, "looks good")
	if err != nil {
		t.Fatalf("generic subject approve: %v", err)
	}
	if out.Status != vo.RequestStatusApproved {
		t.Fatalf("status = %s, want APPROVED", out.Status)
	}
}

// auditEventRecord captures a single audit call for metadata inspection.
type auditEventRecord struct {
	eventType string
	metadata  map[string]any
}

// metadataRecordingAudit captures full audit payloads for field-level assertions.
type metadataRecordingAudit struct {
	events []auditEventRecord
}

func (a *metadataRecordingAudit) Record(_ context.Context, _ *uuid.UUID, eventType, _, _ string, meta map[string]any) {
	a.events = append(a.events, auditEventRecord{eventType: eventType, metadata: meta})
}

// TestRunPostAction_CallbackFailure_AuditHasRequiredFields verifies that the
// APPROVAL_SYNC_FAILED audit record contains the compliance-required fields
// approval_request_id and retryable so operators can replay failed syncs.
func TestRunPostAction_CallbackFailure_AuditHasRequiredFields(t *testing.T) {
	repo := newFakeRepo()
	approver, submitter := uuid.New(), uuid.New()
	gid := seedGroup(repo, "REVIEWERS", approver)
	seedConfig(repo, vo.ProcessInvestmentAnalysisReport, entity.ApprovalProcessStage{
		StageNumber: 1, ApproverMode: vo.ApproverModeGroupAny, ApprovalGroupID: &gid, IsFinalStage: true,
	})
	audit := &metadataRecordingAudit{}
	svc := newTestRuntime(repo, func(d *RuntimeDeps) {
		d.Audit = audit
	})
	svc.RegisterSubjectSync(vo.SubjectResearchReport, failingSync{})

	req, err := submitFixture(repo, svc, submitter)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	task := pendingTaskFor(repo, req.ID, approver)
	if _, err := svc.ApproveTask(ctx(), task.ID, approver, ""); err != nil {
		t.Fatalf("approve: %v", err)
	}

	var syncFailed *auditEventRecord
	for i := range audit.events {
		if audit.events[i].eventType == "APPROVAL_SYNC_FAILED" {
			syncFailed = &audit.events[i]
			break
		}
	}
	if syncFailed == nil {
		t.Fatalf("expected APPROVAL_SYNC_FAILED audit event; got %v", audit.events)
	}
	if _, ok := syncFailed.metadata["approval_request_id"]; !ok {
		t.Error("APPROVAL_SYNC_FAILED metadata missing 'approval_request_id'")
	}
	if v, ok := syncFailed.metadata["retryable"]; !ok || v != true {
		t.Errorf("APPROVAL_SYNC_FAILED metadata 'retryable' = %v, want true", v)
	}
}

