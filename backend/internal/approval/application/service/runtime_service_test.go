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

// 1. Submit creates a request + first stage task; group-priority picks the
//    highest-priority eligible member.
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
//        approval; other tasks are skipped.
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
//     completes the request and writes a signature per approval.
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

// 13. Delegated approval records the proxy/delegated markers and a DELEGATED
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
