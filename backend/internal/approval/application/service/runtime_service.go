package service

import (
	"context"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// ApprovalRuntimeService drives the approval lifecycle: submit, approve, reject,
// withdraw, cancel, plus the inbox/detail/timeline reads. All state transitions
// run inside a DB transaction with a row lock on the request to make
// maker-checker enforcement and stage advancement atomic.
type ApprovalRuntimeService struct {
	pool      *pgxpool.Pool
	repo      domain.Repository
	resolver  *resolver
	perms     domain.PermissionPort
	audit     domain.AuditPort
	notifier  domain.Notifier
	delegate  domain.DelegateResolver
	directory domain.UserDirectory
	now       func() time.Time
	runTxFn   func(context.Context, func(pgx.Tx) error) error

	subjectSyncs       map[vo.SubjectType]domain.SubjectSync
	subjectValidators  map[vo.SubjectType]domain.SubjectValidator
	subjectAccessPorts map[vo.SubjectType]domain.ApprovalSubjectAccessPort
}

// RuntimeDeps bundles the runtime service dependencies.
type RuntimeDeps struct {
	Pool      *pgxpool.Pool
	Repo      domain.Repository
	Perms     domain.PermissionPort
	Audit     domain.AuditPort
	Notifier  domain.Notifier
	Delegate  domain.DelegateResolver
	Directory domain.UserDirectory
	Leave     domain.LeaveChecker
}

// NewApprovalRuntimeService wires the runtime service with safe defaults.
func NewApprovalRuntimeService(d RuntimeDeps) *ApprovalRuntimeService {
	if d.Audit == nil {
		d.Audit = domain.NopAudit{}
	}
	if d.Notifier == nil {
		d.Notifier = domain.NopNotifier{}
	}
	if d.Delegate == nil {
		d.Delegate = domain.NopDelegateResolver{}
	}
	if d.Leave == nil {
		d.Leave = domain.NopLeaveChecker{}
	}
	svc := &ApprovalRuntimeService{
		pool:               d.Pool,
		repo:               d.Repo,
		resolver:           newResolver(d.Repo, d.Leave, nowUTC),
		perms:              d.Perms,
		audit:              d.Audit,
		notifier:           d.Notifier,
		delegate:           d.Delegate,
		directory:          d.Directory,
		now:                nowUTC,
		subjectSyncs:       map[vo.SubjectType]domain.SubjectSync{},
		subjectValidators:  map[vo.SubjectType]domain.SubjectValidator{},
		subjectAccessPorts: map[vo.SubjectType]domain.ApprovalSubjectAccessPort{},
	}
	svc.runTxFn = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return withTransaction(ctx, svc.pool, fn)
	}
	return svc
}

// RegisterSubjectSync registers a business-object callback for a subject type.
// Called at wire-up after both modules are constructed (avoids a circular dep).
func (s *ApprovalRuntimeService) RegisterSubjectSync(st vo.SubjectType, sync domain.SubjectSync) {
	if sync != nil {
		s.subjectSyncs[st] = sync
	}
}

// RegisterSubjectValidator registers a business-object state guard for a
// subject type. The guard is called inside the approve/reject transaction;
// returning a non-nil error aborts the action with a 409 Conflict.
func (s *ApprovalRuntimeService) RegisterSubjectValidator(st vo.SubjectType, v domain.SubjectValidator) {
	if v != nil {
		s.subjectValidators[st] = v
	}
}

// RegisterSubjectAccessPort registers the per-subject-type authorisation gate.
// All read and write operations call this port before disclosing any request
// data. When no port is registered for a subject type the engine fails closed
// (returns ErrForbidden). Call at wire-up after both modules are constructed.
func (s *ApprovalRuntimeService) RegisterSubjectAccessPort(st vo.SubjectType, port domain.ApprovalSubjectAccessPort) {
	if port != nil {
		s.subjectAccessPorts[st] = port
	}
}

// CancelApprovalBySubject cancels the active approval request for a subject,
// if one exists. Returns nil when no active request is found (safe no-op).
// Used by business modules to terminate an in-flight approval when the subject
// itself is cancelled or withdrawn.
func (s *ApprovalRuntimeService) CancelApprovalBySubject(ctx context.Context, subjectType vo.SubjectType, subjectID uuid.UUID, actorID uuid.UUID) error {
	if !vo.ValidSubjectType(subjectType) {
		return domain.Validation("invalid subject_type: " + string(subjectType))
	}
	req, err := s.repo.GetActiveBySubject(ctx, subjectType, subjectID)
	if err != nil {
		return err
	}
	if req == nil {
		return nil // no active request — nothing to cancel
	}
	_, err = s.CancelRequest(ctx, req.ID, actorID)
	return err
}

func (s *ApprovalRuntimeService) runTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return s.runTxFn(ctx, fn)
}

// ─────────────────────────────────────────────────────────────────────────────
// Submit
// ─────────────────────────────────────────────────────────────────────────────

// SubmitInput is the payload for creating an approval request.
type SubmitInput struct {
	ProcessType      vo.ProcessType
	SubjectType      vo.SubjectType
	SubjectID        uuid.UUID
	SubjectTitle     string
	SubjectReference string
	ContractType     vo.ContractType
	ContractID       *uuid.UUID
	PortfolioID      *uuid.UUID
	SubmitterID      uuid.UUID
}

// SubmitApproval resolves the process config, creates the request and the
// first stage's tasks atomically, then notifies the initial approvers.
func (s *ApprovalRuntimeService) SubmitApproval(ctx context.Context, in SubmitInput) (*entity.ApprovalRequest, error) {
	if !vo.ValidProcessType(in.ProcessType) {
		return nil, domain.Validation("invalid process_type")
	}
	if !vo.ValidSubjectType(in.SubjectType) {
		return nil, domain.Validation("invalid subject_type")
	}
	if in.SubjectID == uuid.Nil {
		return nil, domain.Validation("subject_id is required")
	}
	if in.SubmitterID == uuid.Nil {
		return nil, domain.Validation("submitter_id is required")
	}

	// Authorise the submitter via the subject access port.
	if err := s.checkSubjectSubmit(ctx, in.SubmitterID, in.SubjectType, in.SubjectID); err != nil {
		return nil, err
	}

	// Reject duplicate active requests for the same subject.
	if existing, err := s.repo.GetActiveBySubject(ctx, in.SubjectType, in.SubjectID); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, domain.NewError(domain.ErrDuplicateActiveRequest, domain.ErrDuplicateActiveRequest.Error())
	}

	// Resolve config by scope.
	ct := in.ContractType
	if ct == "" {
		if in.ContractID != nil {
			ct = vo.ContractTypeFund
		} else {
			ct = vo.ContractTypeCompany
		}
	}
	cfg, err := s.repo.Resolve(ctx, in.ProcessType, in.ContractID, ct, s.now())
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, domain.NewError(domain.ErrConfigNotFound, "no active approval process is configured for this process type / contract")
	}
	entryStage, ok := entryStage(cfg.Stages)
	if !ok {
		return nil, domain.NewError(domain.ErrConfigNotFound, "approval process has no stages configured")
	}

	now := s.now()
	req := &entity.ApprovalRequest{
		ID:                 uuid.New(),
		ProcessType:        in.ProcessType,
		ProcessConfigID:    &cfg.ID,
		SubjectType:        in.SubjectType,
		SubjectID:          in.SubjectID,
		SubjectTitle:       strings.TrimSpace(in.SubjectTitle),
		SubjectReference:   strings.TrimSpace(in.SubjectReference),
		ContractID:         in.ContractID,
		PortfolioID:        in.PortfolioID,
		SubmitterID:        in.SubmitterID,
		SubmittedAt:        &now,
		CurrentStageNumber: entryStage.StageNumber,
		Status:             vo.RequestStatusPendingApproval,
	}

	var createdTasks []*entity.ApprovalTask
	err = s.runTx(ctx, func(tx pgx.Tx) error {
		num, err := s.repo.NextRequestNumber(ctx, tx)
		if err != nil {
			return err
		}
		req.RequestNumber = num
		if err := s.repo.CreateRequest(ctx, tx, req); err != nil {
			return err
		}
		if err := s.appendEvent(ctx, tx, req.ID, vo.EventSubmitted, &entryStage.StageNumber, &in.SubmitterID, nil, "", map[string]any{
			"process_type": string(in.ProcessType),
			"subject_type": string(in.SubjectType),
			"subject_id":   in.SubjectID.String(),
		}); err != nil {
			return err
		}
		tasks, err := s.createStageTasks(ctx, tx, req, entryStage)
		if err != nil {
			return err
		}
		createdTasks = tasks
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &in.SubmitterID, "APPROVAL_SUBMITTED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
		"request_number": req.RequestNumber,
		"process_type":   string(in.ProcessType),
		"subject_type":   string(in.SubjectType),
		"subject_id":     in.SubjectID.String(),
	})
	for _, t := range createdTasks {
		s.notifier.NotifyApprovalTaskCreated(ctx, req, t)
	}
	return s.repo.GetRequest(ctx, req.ID)
}

// ─────────────────────────────────────────────────────────────────────────────
// Approve / Reject
// ─────────────────────────────────────────────────────────────────────────────

type actionOutcome struct {
	completed     bool
	rejected      bool
	advancedTasks []*entity.ApprovalTask
}

// ApproveTask records an approval on a task and advances the request.
func (s *ApprovalRuntimeService) ApproveTask(ctx context.Context, taskID, actorID uuid.UUID, comment string) (*entity.ApprovalRequest, error) {
	var outcome actionOutcome
	var req *entity.ApprovalRequest

	err := s.runTx(ctx, func(tx pgx.Tx) error {
		task, r, stage, err := s.loadActionContext(ctx, tx, taskID)
		if err != nil {
			return err
		}
		req = r

		actedFor, delegated, err := s.authorizeAction(ctx, task, r, actorID, domain.ApprovalActionApprove)
		if err != nil {
			return err
		}

		// Guard against the subject being cancelled/invalidated between submit and approve.
		if v, ok := s.subjectValidators[r.SubjectType]; ok {
			if verr := v.ValidateSubjectApprovable(ctx, r.SubjectID); verr != nil {
				return domain.NewError(domain.ErrConflict, "subject is no longer approvable: "+verr.Error())
			}
		}

		now := s.now()
		task.Status = vo.TaskStatusApproved
		task.ActedBy = &actorID
		task.ActedAt = &now
		task.ActionComment = strings.TrimSpace(comment)
		task.IsDelegatedAction = delegated
		if delegated {
			task.DelegatedFromUserID = &actedFor
		}
		if err := s.repo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}

		if delegated {
			if err := s.appendEvent(ctx, tx, r.ID, vo.EventDelegated, &task.StageNumber, &actorID, &actedFor, task.ActionComment, nil); err != nil {
				return err
			}
		}
		if err := s.appendEvent(ctx, tx, r.ID, vo.EventApproved, &task.StageNumber, &actorID, delegatedPtr(delegated, actedFor), task.ActionComment, nil); err != nil {
			return err
		}
		if err := s.writeSignature(ctx, tx, r, task.StageNumber, actorID, delegated, actedFor); err != nil {
			return err
		}

		// Stage completion check.
		approved, err := s.repo.CountApprovedInStage(ctx, tx, r.ID, task.StageNumber)
		if err != nil {
			return err
		}
		// total tasks created in this stage (pending + approved + skipped) — use
		// approved + pending to size GROUP_ANY/TEAM thresholds sensibly.
		pending, err := s.repo.ListPendingByStage(ctx, tx, r.ID, task.StageNumber)
		if err != nil {
			return err
		}
		required, err := s.resolver.requiredCountForStage(ctx, stage, r.ContractID, approved+len(pending))
		if err != nil {
			return err
		}
		if approved < required {
			return nil // stage still in progress
		}

		// Stage complete: skip remaining pending tasks.
		if err := s.repo.SkipOtherPendingInStage(ctx, tx, r.ID, task.StageNumber, task.ID); err != nil {
			return err
		}
		if err := s.appendEvent(ctx, tx, r.ID, vo.EventStageCompleted, &task.StageNumber, &actorID, nil, "", nil); err != nil {
			return err
		}

		cfg, err := s.repo.GetConfig(ctx, derefUUID(r.ProcessConfigID))
		if err != nil {
			return err
		}
		next, hasNext := nextStage(cfgStages(cfg), task.StageNumber)
		if hasNext && !stage.IsFinalStage {
			r.CurrentStageNumber = next.StageNumber
			if err := s.repo.UpdateRequestState(ctx, tx, r); err != nil {
				return err
			}
			tasks, err := s.createStageTasks(ctx, tx, r, next)
			if err != nil {
				return err
			}
			outcome.advancedTasks = tasks
			return nil
		}

		// Final: request approved.
		r.Status = vo.RequestStatusApproved
		r.FinalDecisionBy = &actorID
		r.FinalDecisionAt = &now
		if err := s.repo.UpdateRequestState(ctx, tx, r); err != nil {
			return err
		}
		if err := s.appendEvent(ctx, tx, r.ID, vo.EventRequestCompleted, nil, &actorID, nil, "", map[string]any{"final_status": "APPROVED"}); err != nil {
			return err
		}
		outcome.completed = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, "APPROVAL_TASK_APPROVED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
		"request_number": req.RequestNumber,
		"task_id":        taskID.String(),
	})
	s.runPostAction(ctx, req.ID, outcome)
	return s.repo.GetRequest(ctx, req.ID)
}

// RejectTask records a rejection; one rejection stops the whole request.
func (s *ApprovalRuntimeService) RejectTask(ctx context.Context, taskID, actorID uuid.UUID, reason string) (*entity.ApprovalRequest, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, domain.Validation("a rejection reason is required")
	}
	var req *entity.ApprovalRequest

	err := s.runTx(ctx, func(tx pgx.Tx) error {
		task, r, _, err := s.loadActionContext(ctx, tx, taskID)
		if err != nil {
			return err
		}
		req = r
		actedFor, delegated, err := s.authorizeAction(ctx, task, r, actorID, domain.ApprovalActionReject)
		if err != nil {
			return err
		}

		// Guard against the subject being cancelled/invalidated between submit and reject.
		if v, ok := s.subjectValidators[r.SubjectType]; ok {
			if verr := v.ValidateSubjectApprovable(ctx, r.SubjectID); verr != nil {
				return domain.NewError(domain.ErrConflict, "subject is no longer approvable: "+verr.Error())
			}
		}

		now := s.now()
		task.Status = vo.TaskStatusRejected
		task.ActedBy = &actorID
		task.ActedAt = &now
		task.ActionComment = reason
		task.IsDelegatedAction = delegated
		if delegated {
			task.DelegatedFromUserID = &actedFor
		}
		if err := s.repo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}
		if delegated {
			if err := s.appendEvent(ctx, tx, r.ID, vo.EventDelegated, &task.StageNumber, &actorID, &actedFor, reason, nil); err != nil {
				return err
			}
		}
		if err := s.appendEvent(ctx, tx, r.ID, vo.EventRejected, &task.StageNumber, &actorID, delegatedPtr(delegated, actedFor), reason, nil); err != nil {
			return err
		}
		if err := s.repo.CancelPendingByRequest(ctx, tx, r.ID); err != nil {
			return err
		}
		r.Status = vo.RequestStatusRejected
		r.FinalDecisionBy = &actorID
		r.FinalDecisionAt = &now
		r.RejectionReason = reason
		if err := s.repo.UpdateRequestState(ctx, tx, r); err != nil {
			return err
		}
		return s.appendEvent(ctx, tx, r.ID, vo.EventRequestCompleted, nil, &actorID, nil, "", map[string]any{"final_status": "REJECTED"})
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &actorID, "APPROVAL_TASK_REJECTED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
		"request_number": req.RequestNumber,
		"task_id":        taskID.String(),
		"reason":         reason,
	})
	s.runPostAction(ctx, req.ID, actionOutcome{rejected: true})
	return s.repo.GetRequest(ctx, req.ID)
}

// WithdrawRequest lets the submitter withdraw an active request.
func (s *ApprovalRuntimeService) WithdrawRequest(ctx context.Context, requestID, actorID uuid.UUID) (*entity.ApprovalRequest, error) {
	return s.terminate(ctx, requestID, actorID, vo.RequestStatusWithdrawn, vo.EventWithdrawn, true)
}

// CancelRequest cancels an active request (privileged action).
func (s *ApprovalRuntimeService) CancelRequest(ctx context.Context, requestID, actorID uuid.UUID) (*entity.ApprovalRequest, error) {
	return s.terminate(ctx, requestID, actorID, vo.RequestStatusCancelled, vo.EventCancelled, false)
}

// RevokeRequest revokes a previously approved request, returning the subject to
// an un-approved state. The approval history is preserved. The subject sync
// callback is invoked with Approved=false so the business module can reopen
// the document for correction.
func (s *ApprovalRuntimeService) RevokeRequest(ctx context.Context, requestID, actorID uuid.UUID, reason string) (*entity.ApprovalRequest, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, domain.Validation("a revocation reason is required")
	}
	var req *entity.ApprovalRequest
	err := s.runTx(ctx, func(tx pgx.Tx) error {
		r, err := s.repo.GetRequestForUpdate(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if r == nil {
			return domain.NotFound("approval request not found")
		}
		if err := s.checkSubjectAct(ctx, actorID, r.SubjectType, r.SubjectID, domain.ApprovalActionRevoke); err != nil {
			return err
		}
		if r.Status != vo.RequestStatusApproved {
			return domain.NewError(domain.ErrConflict, "only an APPROVED request can be revoked")
		}
		now := s.now()
		r.Status = vo.RequestStatusRevoked
		r.FinalDecisionBy = &actorID
		r.FinalDecisionAt = &now
		r.RejectionReason = reason
		if err := s.repo.UpdateRequestState(ctx, tx, r); err != nil {
			return err
		}
		req = r
		return s.appendEvent(ctx, tx, r.ID, vo.EventRevoked, nil, &actorID, nil, reason, nil)
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &actorID, "APPROVAL_REVOKED", "APPROVAL_REQUEST", requestID.String(), map[string]any{
		"request_number": req.RequestNumber,
		"reason":         reason,
	})
	// Notify the business module that the subject is no longer approved.
	if sync, ok := s.subjectSyncs[req.SubjectType]; ok {
		if serr := sync.OnRejected(ctx, req.SubjectType, req.SubjectID, req.ID, reason); serr != nil {
			slog.Error("approval sync callback failed after revoke",
				"request_id", req.ID.String(),
				"subject_type", string(req.SubjectType),
				"subject_id", req.SubjectID.String(),
				"error", serr,
			)
			s.audit.Record(ctx, nil, "APPROVAL_REVOKE_SYNC_FAILED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
				"subject_type":        string(req.SubjectType),
				"subject_id":          req.SubjectID.String(),
				"approval_request_id": req.ID.String(),
				"outcome":             "REVOKED",
				"reason":              reason,
				"error":               serr.Error(),
				"retryable":           true,
			})
		}
	}
	return s.repo.GetRequest(ctx, requestID)
}

func (s *ApprovalRuntimeService) terminate(ctx context.Context, requestID, actorID uuid.UUID, status vo.RequestStatus, event vo.EventType, submitterOnly bool) (*entity.ApprovalRequest, error) {
	err := s.runTx(ctx, func(tx pgx.Tx) error {
		r, err := s.repo.GetRequestForUpdate(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if r == nil {
			return domain.NotFound("approval request not found")
		}
		terminateAction := domain.ApprovalActionCancel
		if submitterOnly {
			terminateAction = domain.ApprovalActionWithdraw
		}
		if err := s.checkSubjectAct(ctx, actorID, r.SubjectType, r.SubjectID, terminateAction); err != nil {
			return err
		}
		if r.Status.IsTerminal() {
			return domain.NewError(domain.ErrConflict, "request is already finalised")
		}
		if submitterOnly && r.SubmitterID != actorID {
			return domain.Forbidden("only the submitter can withdraw this request")
		}
		now := s.now()
		r.Status = status
		r.FinalDecisionBy = &actorID
		r.FinalDecisionAt = &now
		if err := s.repo.CancelPendingByRequest(ctx, tx, r.ID); err != nil {
			return err
		}
		if err := s.repo.UpdateRequestState(ctx, tx, r); err != nil {
			return err
		}
		return s.appendEvent(ctx, tx, r.ID, event, nil, &actorID, nil, "", nil)
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &actorID, "APPROVAL_"+string(event), "APPROVAL_REQUEST", requestID.String(), nil)
	return s.repo.GetRequest(ctx, requestID)
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────────────────

// loadActionContext locks the request + task and validates the request is
// actionable, the task is pending, and the task belongs to the current stage.
func (s *ApprovalRuntimeService) loadActionContext(ctx context.Context, tx pgx.Tx, taskID uuid.UUID) (*entity.ApprovalTask, *entity.ApprovalRequest, entity.ApprovalProcessStage, error) {
	var zero entity.ApprovalProcessStage
	probe, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, nil, zero, err
	}
	if probe == nil {
		return nil, nil, zero, domain.NotFound("approval task not found")
	}
	// Lock the request first to serialise concurrent actions on its stages.
	r, err := s.repo.GetRequestForUpdate(ctx, tx, probe.ApprovalRequestID)
	if err != nil {
		return nil, nil, zero, err
	}
	if r == nil {
		return nil, nil, zero, domain.NotFound("approval request not found")
	}
	task, err := s.repo.GetTaskForUpdate(ctx, tx, taskID)
	if err != nil {
		return nil, nil, zero, err
	}
	if task == nil {
		return nil, nil, zero, domain.NotFound("approval task not found")
	}
	if !r.Status.IsActionable() {
		return nil, nil, zero, domain.NewError(domain.ErrRequestNotActionable, domain.ErrRequestNotActionable.Error())
	}
	if task.Status != vo.TaskStatusPending {
		return nil, nil, zero, domain.NewError(domain.ErrTaskNotPending, domain.ErrTaskNotPending.Error())
	}
	if task.StageNumber != r.CurrentStageNumber {
		return nil, nil, zero, domain.NewError(domain.ErrStaleTask, domain.ErrStaleTask.Error())
	}
	cfg, err := s.repo.GetConfig(ctx, derefUUID(r.ProcessConfigID))
	if err != nil {
		return nil, nil, zero, err
	}
	stage, ok := stageByNumber(cfgStages(cfg), task.StageNumber)
	if !ok {
		return nil, nil, zero, domain.NewError(domain.ErrConfigNotFound, "stage configuration not found")
	}
	return task, r, stage, nil
}

// authorizeAction enforces maker-checker, subject access, and assignment/delegation rules.
// Returns (actedForUserID, isDelegated, error).
func (s *ApprovalRuntimeService) authorizeAction(ctx context.Context, task *entity.ApprovalTask, r *entity.ApprovalRequest, actorID uuid.UUID, action domain.ApprovalAction) (uuid.UUID, bool, error) {
	// Maker-checker: the submitter can never approve/reject their own request.
	if actorID == r.SubmitterID {
		return uuid.Nil, false, domain.NewError(domain.ErrSelfApproval, domain.ErrSelfApproval.Error())
	}
	// Subject access port enforces object-level authorisation generically.
	if err := s.checkSubjectAct(ctx, actorID, r.SubjectType, r.SubjectID, action); err != nil {
		return uuid.Nil, false, err
	}

	assignee := uuid.Nil
	if task.AssignedUserID != nil {
		assignee = *task.AssignedUserID
	}
	if assignee == actorID {
		return assignee, false, nil // normal, direct approval
	}

	// Delegation path: the actor may be a valid delegate for the assignee.
	if assignee != uuid.Nil {
		del, err := s.delegate.ResolveDelegate(ctx, assignee, r.ContractID, s.now())
		if err != nil {
			return uuid.Nil, false, err
		}
		if del != nil && del.DelegateUserID == actorID && actorID != r.SubmitterID {
			return assignee, true, nil
		}
	}
	return uuid.Nil, false, domain.NewError(domain.ErrNotAssigned, domain.ErrNotAssigned.Error())
}

// checkSubjectView authorises a read for the given actor on a subject.
// Zero actorID is always rejected. Missing port → ErrForbidden (fail closed).
func (s *ApprovalRuntimeService) checkSubjectView(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID) error {
	if actorID == uuid.Nil {
		return domain.Validation("actor_id is required")
	}
	port, ok := s.subjectAccessPorts[st]
	if !ok {
		return domain.Forbidden("no subject access port configured for subject type: " + string(st))
	}
	return port.CanViewApprovalSubject(ctx, actorID, st, subjectID)
}

// checkSubjectSubmit authorises a submit for the given actor on a subject.
func (s *ApprovalRuntimeService) checkSubjectSubmit(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID) error {
	if actorID == uuid.Nil {
		return domain.Validation("actor_id is required")
	}
	port, ok := s.subjectAccessPorts[st]
	if !ok {
		return domain.Forbidden("no subject access port configured for subject type: " + string(st))
	}
	return port.CanSubmitApprovalSubject(ctx, actorID, st, subjectID)
}

// checkSubjectAct authorises a write action (approve/reject/revoke/cancel/withdraw).
func (s *ApprovalRuntimeService) checkSubjectAct(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID, action domain.ApprovalAction) error {
	if actorID == uuid.Nil {
		return domain.Validation("actor_id is required")
	}
	port, ok := s.subjectAccessPorts[st]
	if !ok {
		return domain.Forbidden("no subject access port configured for subject type: " + string(st))
	}
	return port.CanActOnApprovalSubject(ctx, actorID, st, subjectID, action)
}

// createStageTasks resolves and persists the tasks for a stage and appends a
// single TASK_CREATED timeline event.
func (s *ApprovalRuntimeService) createStageTasks(ctx context.Context, tx pgx.Tx, r *entity.ApprovalRequest, stage entity.ApprovalProcessStage) ([]*entity.ApprovalTask, error) {
	plan, err := s.resolver.resolveStage(ctx, stage, r.SubmitterID, r.ContractID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	var created []*entity.ApprovalTask
	assignees := make([]string, 0, len(plan.Tasks))
	for _, pt := range plan.Tasks {
		uid := pt.AssignedUserID
		t := &entity.ApprovalTask{
			ID:                uuid.New(),
			ApprovalRequestID: r.ID,
			StageNumber:       stage.StageNumber,
			AssignedUserID:    &uid,
			AssignedGroupID:   pt.AssignedGroupID,
			AssignedTeamID:    pt.AssignedTeamID,
			Status:            vo.TaskStatusPending,
			CreatedAt:         now,
		}
		if err := s.repo.CreateTask(ctx, tx, t); err != nil {
			return nil, err
		}
		created = append(created, t)
		assignees = append(assignees, uid.String())
	}
	stageNum := stage.StageNumber
	if err := s.appendEvent(ctx, tx, r.ID, vo.EventTaskCreated, &stageNum, nil, nil, "", map[string]any{
		"stage_number":   stage.StageNumber,
		"stage_name":     stage.StageName,
		"approver_mode":  string(stage.ApproverMode),
		"required_count": plan.RequiredCount,
		"assignee_count": len(plan.Tasks),
		"assignees":      assignees,
	}); err != nil {
		return nil, err
	}
	return created, nil
}

// writeSignature creates a signature/stamp record for an approval action.
func (s *ApprovalRuntimeService) writeSignature(ctx context.Context, tx pgx.Tx, r *entity.ApprovalRequest, stage int, signer uuid.UUID, delegated bool, principal uuid.UUID) error {
	display := signer.String()
	if s.directory != nil {
		if info, err := s.directory.GetUser(ctx, signer); err == nil && info != nil && info.DisplayName != "" {
			display = info.DisplayName
		}
	}
	label := vo.SignatureLabelNormal
	var proxyFor *uuid.UUID
	if delegated {
		label = vo.SignatureLabelDelegated
		p := principal
		proxyFor = &p
	}
	sig := &entity.ApprovalSignatureRecord{
		ID:                uuid.New(),
		ApprovalRequestID: r.ID,
		StageNumber:       stage,
		SignerUserID:      signer,
		SignerDisplayName: display,
		SignedAt:          s.now(),
		IsProxySignature:  delegated,
		ProxyForUserID:    proxyFor,
		SignatureLabel:    label,
	}
	return s.repo.CreateSignature(ctx, tx, sig)
}

func (s *ApprovalRuntimeService) appendEvent(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, et vo.EventType, stage *int, actor, delegatedFrom *uuid.UUID, comment string, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	return s.repo.AppendEvent(ctx, tx, &entity.ApprovalEvent{
		ID:                  uuid.New(),
		ApprovalRequestID:   requestID,
		EventType:           et,
		StageNumber:         stage,
		ActorUserID:         actor,
		DelegatedFromUserID: delegatedFrom,
		Comment:             comment,
		Metadata:            metadata,
	})
}

// runPostAction runs best-effort, post-commit side effects (notifications +
// business-object sync). The approval is already committed; errors from the
// sync callback cannot roll it back. They are logged and recorded in the audit
// trail so operators can detect and replay failed syncs.
func (s *ApprovalRuntimeService) runPostAction(ctx context.Context, requestID uuid.UUID, outcome actionOutcome) {
	req, err := s.repo.GetRequest(ctx, requestID)
	if err != nil || req == nil {
		return
	}
	switch {
	case outcome.completed:
		s.notifier.NotifyApprovalCompleted(ctx, req)
		if sync, ok := s.subjectSyncs[req.SubjectType]; ok {
			if serr := sync.OnApproved(ctx, req.SubjectType, req.SubjectID, req.ID); serr != nil {
				slog.Error("approval sync callback failed after approve",
					"request_id", req.ID.String(),
					"subject_type", string(req.SubjectType),
					"subject_id", req.SubjectID.String(),
					"error", serr,
				)
				s.audit.Record(ctx, nil, "APPROVAL_SYNC_FAILED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
					"subject_type":        string(req.SubjectType),
					"subject_id":          req.SubjectID.String(),
					"approval_request_id": req.ID.String(),
					"outcome":             "APPROVED",
					"error":               serr.Error(),
					"retryable":           true,
				})
			}
		}
	case outcome.rejected:
		s.notifier.NotifyApprovalRejected(ctx, req)
		if sync, ok := s.subjectSyncs[req.SubjectType]; ok {
			if serr := sync.OnRejected(ctx, req.SubjectType, req.SubjectID, req.ID, req.RejectionReason); serr != nil {
				slog.Error("approval sync callback failed after reject",
					"request_id", req.ID.String(),
					"subject_type", string(req.SubjectType),
					"subject_id", req.SubjectID.String(),
					"error", serr,
				)
				s.audit.Record(ctx, nil, "APPROVAL_SYNC_FAILED", "APPROVAL_REQUEST", req.ID.String(), map[string]any{
					"subject_type":        string(req.SubjectType),
					"subject_id":          req.SubjectID.String(),
					"approval_request_id": req.ID.String(),
					"outcome":             "REJECTED",
					"error":               serr.Error(),
					"retryable":           true,
				})
			}
		}
	default:
		for _, t := range outcome.advancedTasks {
			s.notifier.NotifyApprovalTaskCreated(ctx, req, t)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Reads
// ─────────────────────────────────────────────────────────────────────────────

// RequestDetail bundles a request with its tasks, timeline and signatures.
type RequestDetail struct {
	Request        *entity.ApprovalRequest
	Tasks          []*entity.ApprovalTask
	Events         []*entity.ApprovalEvent
	Signatures     []*entity.ApprovalSignatureRecord
	ViewerTask     *entity.ApprovalTask // the viewer's actionable pending task, if any
	AllowedActions []string             // business actions the current viewer may take
}

// GetMyInbox returns the actor's pending (or filtered) approval work items,
// filtered to only those subjects the actor can access via the subject access port.
func (s *ApprovalRuntimeService) GetMyInbox(ctx context.Context, f domain.InboxFilter) ([]*domain.InboxItem, int, error) {
	if f.UserID == uuid.Nil {
		return nil, 0, domain.Validation("user_id is required")
	}
	items, _, err := s.repo.Inbox(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	var permitted []*domain.InboxItem
	for _, item := range items {
		if s.checkSubjectView(ctx, f.UserID, item.Request.SubjectType, item.Request.SubjectID) == nil {
			permitted = append(permitted, item)
		}
	}
	return permitted, len(permitted), nil
}

// ListRequests lists requests with filters. When ViewerID is set, results are
// post-filtered per subject so the viewer cannot see subjects they cannot access.
// When ViewerID is uuid.Nil the full unfiltered list is returned (system/admin use).
func (s *ApprovalRuntimeService) ListRequests(ctx context.Context, f domain.RequestListFilter) ([]*entity.ApprovalRequest, int, error) {
	all, _, err := s.repo.ListRequests(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if f.ViewerID == uuid.Nil {
		return all, len(all), nil
	}
	var permitted []*entity.ApprovalRequest
	for _, r := range all {
		if s.checkSubjectView(ctx, f.ViewerID, r.SubjectType, r.SubjectID) == nil {
			permitted = append(permitted, r)
		}
	}
	return permitted, len(permitted), nil
}

// GetApprovalRequest returns the full request detail for a viewer.
func (s *ApprovalRuntimeService) GetApprovalRequest(ctx context.Context, requestID, viewerID uuid.UUID) (*RequestDetail, error) {
	req, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, domain.NotFound("approval request not found")
	}
	if err := s.checkSubjectView(ctx, viewerID, req.SubjectType, req.SubjectID); err != nil {
		return nil, err
	}
	tasks, err := s.repo.ListTasksByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	events, err := s.repo.ListEvents(ctx, requestID)
	if err != nil {
		return nil, err
	}
	sigs, err := s.repo.ListSignatures(ctx, requestID)
	if err != nil {
		return nil, err
	}
	detail := &RequestDetail{Request: req, Tasks: tasks, Events: events, Signatures: sigs}
	for _, t := range tasks {
		if t.Status != vo.TaskStatusPending || t.StageNumber != req.CurrentStageNumber || t.AssignedUserID == nil {
			continue
		}
		if *t.AssignedUserID == viewerID {
			detail.ViewerTask = t
			break
		}
		// Also set ViewerTask if the viewer is a valid delegate for the task's assignee,
		// mirroring the delegation check in authorizeAction so buttons appear for delegates.
		if viewerID != req.SubmitterID {
			del, _ := s.delegate.ResolveDelegate(ctx, *t.AssignedUserID, req.ContractID, s.now())
			if del != nil && del.DelegateUserID == viewerID {
				detail.ViewerTask = t
				break
			}
		}
	}
	detail.AllowedActions = computeAllowedActions(req, detail.ViewerTask, viewerID)
	return detail, nil
}

// computeAllowedActions returns the business actions the given viewer can take
// on a request. Function-permission enforcement remains at the route middleware;
// this set is about per-request business eligibility (maker-checker, ownership,
// status machine). Frontend renders buttons based on this list.
func computeAllowedActions(req *entity.ApprovalRequest, viewerTask *entity.ApprovalTask, viewerID uuid.UUID) []string {
	var actions []string
	if req.Status == vo.RequestStatusPendingApproval && viewerTask != nil {
		actions = append(actions, "approve", "reject")
	}
	if !req.Status.IsTerminal() && req.SubmitterID == viewerID {
		actions = append(actions, "withdraw")
	}
	if !req.Status.IsTerminal() {
		actions = append(actions, "cancel")
	}
	if req.Status == vo.RequestStatusApproved {
		actions = append(actions, "revoke")
	}
	return actions
}

// GetApprovalTimeline returns the immutable ordered events of a request.
// viewerID must be non-zero; the viewer must have subject view access.
func (s *ApprovalRuntimeService) GetApprovalTimeline(ctx context.Context, requestID, viewerID uuid.UUID) ([]*entity.ApprovalEvent, error) {
	req, err := s.repo.GetRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, domain.NotFound("approval request not found")
	}
	if err := s.checkSubjectView(ctx, viewerID, req.SubjectType, req.SubjectID); err != nil {
		return nil, err
	}
	return s.repo.ListEvents(ctx, requestID)
}

// GetSubjectApprovalStatus returns the latest request for a subject, or nil.
// Pass uuid.Nil as viewerID to skip the access check — for internal cross-module
// enrichment calls (e.g. GetApprovalStage badge population). All user-facing
// calls must supply a real viewerID so the port is enforced.
func (s *ApprovalRuntimeService) GetSubjectApprovalStatus(ctx context.Context, st vo.SubjectType, subjectID, viewerID uuid.UUID) (*entity.ApprovalRequest, error) {
	if !vo.ValidSubjectType(st) {
		return nil, domain.Validation("invalid subject_type")
	}
	req, err := s.repo.GetLatestBySubject(ctx, st, subjectID)
	if err != nil {
		return nil, err
	}
	if req != nil && viewerID != uuid.Nil {
		if err := s.checkSubjectView(ctx, viewerID, req.SubjectType, req.SubjectID); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ApproveByRequest resolves the actor's pending task on the given approval
// request and approves it. Used by investment batch-approve endpoints that
// supply request IDs rather than task IDs.
func (s *ApprovalRuntimeService) ApproveByRequest(ctx context.Context, requestID, actorID uuid.UUID, comment string) error {
	task, err := s.repo.FindPendingTaskForActor(ctx, requestID, actorID)
	if err != nil {
		return err
	}
	if task == nil {
		return domain.Forbidden("no pending approval task assigned to you for this request")
	}
	_, err = s.ApproveTask(ctx, task.ID, actorID, comment)
	return err
}

// RejectByRequest resolves the actor's pending task on the given approval
// request and rejects it. Used by investment batch-reject endpoints.
func (s *ApprovalRuntimeService) RejectByRequest(ctx context.Context, requestID, actorID uuid.UUID, reason string) error {
	task, err := s.repo.FindPendingTaskForActor(ctx, requestID, actorID)
	if err != nil {
		return err
	}
	if task == nil {
		return domain.Forbidden("no pending approval task assigned to you for this request")
	}
	_, err = s.RejectTask(ctx, task.ID, actorID, reason)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// stage helpers
// ─────────────────────────────────────────────────────────────────────────────

func cfgStages(cfg *entity.ApprovalProcessConfig) []entity.ApprovalProcessStage {
	if cfg == nil {
		return nil
	}
	return cfg.Stages
}

func sortedStages(stages []entity.ApprovalProcessStage) []entity.ApprovalProcessStage {
	out := make([]entity.ApprovalProcessStage, len(stages))
	copy(out, stages)
	sort.Slice(out, func(i, j int) bool { return out[i].StageNumber < out[j].StageNumber })
	return out
}

func entryStage(stages []entity.ApprovalProcessStage) (entity.ApprovalProcessStage, bool) {
	s := sortedStages(stages)
	if len(s) == 0 {
		return entity.ApprovalProcessStage{}, false
	}
	return s[0], true
}

func stageByNumber(stages []entity.ApprovalProcessStage, n int) (entity.ApprovalProcessStage, bool) {
	for _, st := range stages {
		if st.StageNumber == n {
			return st, true
		}
	}
	return entity.ApprovalProcessStage{}, false
}

func nextStage(stages []entity.ApprovalProcessStage, current int) (entity.ApprovalProcessStage, bool) {
	for _, st := range sortedStages(stages) {
		if st.StageNumber > current {
			return st, true
		}
	}
	return entity.ApprovalProcessStage{}, false
}

func derefUUID(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}

func delegatedPtr(delegated bool, principal uuid.UUID) *uuid.UUID {
	if delegated {
		p := principal
		return &p
	}
	return nil
}
