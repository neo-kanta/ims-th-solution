package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// plannedTask is a resolved approver assignment for a stage, before persistence.
type plannedTask struct {
	AssignedUserID  uuid.UUID
	AssignedGroupID *uuid.UUID
	AssignedTeamID  *uuid.UUID
}

// stagePlan is the resolved set of tasks plus the count required to complete a stage.
type stagePlan struct {
	Stage         entity.ApprovalProcessStage
	RequiredCount int
	Tasks         []plannedTask
}

// resolver turns a process stage into concrete approver tasks, enforcing
// maker exclusion and skipping on-leave approvers. Delegation is resolved at
// action time (see runtime approve), not here.
type resolver struct {
	repo  domain.Repository
	leave domain.LeaveChecker
	nowFn func() time.Time
}

func newResolver(repo domain.Repository, leave domain.LeaveChecker, nowFn func() time.Time) *resolver {
	if leave == nil {
		leave = domain.NopLeaveChecker{}
	}
	if nowFn == nil {
		nowFn = nowUTC
	}
	return &resolver{repo: repo, leave: leave, nowFn: nowFn}
}

// available reports whether a candidate user can be assigned: not the maker and
// not on leave.
func (r *resolver) available(ctx context.Context, userID, submitterID uuid.UUID, asOf time.Time) (bool, error) {
	if userID == submitterID {
		return false, nil // maker-checker: never assign the submitter
	}
	onLeave, err := r.leave.IsOnLeave(ctx, userID, asOf)
	if err != nil {
		return false, err
	}
	return !onLeave, nil
}

// resolveStage resolves a single stage into a concrete plan of approver tasks.
func (r *resolver) resolveStage(
	ctx context.Context,
	stage entity.ApprovalProcessStage,
	submitterID uuid.UUID,
	contractID *uuid.UUID,
) (*stagePlan, error) {
	asOf := r.nowFn()
	plan := &stagePlan{Stage: stage, RequiredCount: 1}

	switch stage.ApproverMode {
	case vo.ApproverModeSingleUser:
		if stage.ApproverUserID == nil {
			return nil, domain.NewError(domain.ErrConfigNotFound, "stage is SINGLE_USER but has no configured approver")
		}
		if *stage.ApproverUserID == submitterID {
			return nil, domain.NewError(domain.ErrSelfApproval, "configured approver is the submitter; self-approval is not allowed")
		}
		plan.RequiredCount = 1
		plan.Tasks = []plannedTask{{AssignedUserID: *stage.ApproverUserID}}
		return plan, nil

	case vo.ApproverModeGroupPriority:
		if stage.ApprovalGroupID == nil {
			return nil, domain.NewError(domain.ErrConfigNotFound, "stage is GROUP_PRIORITY but has no configured group")
		}
		members, err := r.repo.ListEligibleGroupMembers(ctx, *stage.ApprovalGroupID)
		if err != nil {
			return nil, err
		}
		for _, m := range members { // ordered by priority
			ok, err := r.available(ctx, m.UserID, submitterID, asOf)
			if err != nil {
				return nil, err
			}
			if ok {
				plan.RequiredCount = 1
				plan.Tasks = []plannedTask{{AssignedUserID: m.UserID, AssignedGroupID: stage.ApprovalGroupID}}
				return plan, nil
			}
		}
		return nil, domain.NewError(domain.ErrConfigNotFound, "no eligible approver available in the configured group")

	case vo.ApproverModeGroupAny:
		if stage.ApprovalGroupID == nil {
			return nil, domain.NewError(domain.ErrConfigNotFound, "stage is GROUP_ANY but has no configured group")
		}
		members, err := r.repo.ListEligibleGroupMembers(ctx, *stage.ApprovalGroupID)
		if err != nil {
			return nil, err
		}
		var tasks []plannedTask
		for _, m := range members {
			ok, err := r.available(ctx, m.UserID, submitterID, asOf)
			if err != nil {
				return nil, err
			}
			if ok {
				tasks = append(tasks, plannedTask{AssignedUserID: m.UserID, AssignedGroupID: stage.ApprovalGroupID})
			}
		}
		if len(tasks) == 0 {
			return nil, domain.NewError(domain.ErrConfigNotFound, "no eligible approvers available in the configured group")
		}
		required := stage.RequiredApprovalCount
		if required < 1 {
			required = 1
		}
		if required > len(tasks) {
			required = len(tasks)
		}
		plan.RequiredCount = required
		plan.Tasks = tasks
		return plan, nil

	case vo.ApproverModeTeamMinimum:
		if contractID == nil {
			return nil, domain.NewError(domain.ErrConfigNotFound, "stage is TEAM_MINIMUM but the request has no contract")
		}
		team, err := r.repo.GetActiveTeamForContract(ctx, *contractID)
		if err != nil {
			return nil, err
		}
		if team == nil {
			return nil, domain.NewError(domain.ErrConfigNotFound, "no active approval team is assigned to this contract")
		}
		members, err := r.repo.ListEligibleTeamMembers(ctx, team.ID)
		if err != nil {
			return nil, err
		}
		var tasks []plannedTask
		for _, m := range members {
			ok, err := r.available(ctx, m.UserID, submitterID, asOf)
			if err != nil {
				return nil, err
			}
			if ok {
				tid := team.ID
				tasks = append(tasks, plannedTask{AssignedUserID: m.UserID, AssignedTeamID: &tid})
			}
		}
		if len(tasks) == 0 {
			return nil, domain.NewError(domain.ErrConfigNotFound, "no eligible reviewer agents available in the contract's approval team")
		}
		required := team.MinRequiredStamps
		if required < 1 {
			required = 1
		}
		if required > len(tasks) {
			required = len(tasks)
		}
		plan.RequiredCount = required
		plan.Tasks = tasks
		return plan, nil
	}

	return nil, domain.NewError(domain.ErrConfigNotFound, "unsupported approver mode")
}

// requiredCountForStage recomputes the completion threshold for an in-flight
// stage at action time (used to decide whether a stage is complete).
func (r *resolver) requiredCountForStage(
	ctx context.Context,
	stage entity.ApprovalProcessStage,
	contractID *uuid.UUID,
	taskCount int,
) (int, error) {
	switch stage.ApproverMode {
	case vo.ApproverModeSingleUser, vo.ApproverModeGroupPriority:
		return 1, nil
	case vo.ApproverModeGroupAny:
		req := stage.RequiredApprovalCount
		if req < 1 {
			req = 1
		}
		if req > taskCount && taskCount > 0 {
			req = taskCount
		}
		return req, nil
	case vo.ApproverModeTeamMinimum:
		if contractID == nil {
			return 1, nil
		}
		team, err := r.repo.GetActiveTeamForContract(ctx, *contractID)
		if err != nil {
			return 0, err
		}
		req := 1
		if team != nil {
			req = team.MinRequiredStamps
		}
		if req < 1 {
			req = 1
		}
		if req > taskCount && taskCount > 0 {
			req = taskCount
		}
		return req, nil
	}
	return 1, nil
}
