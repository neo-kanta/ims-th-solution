// Package approval wires the generic Approval Module: a reusable, auditable,
// permission-controlled approval engine that can approve any IMS business
// object (research report, investment decision, workflow operation, leave,
// delegation) via a polymorphic (subject_type, subject_id) reference.
//
// Cross-module integrations live behind pkg/contract interfaces — this module
// never imports another module's internal package. The IAM module is consumed
// via domain.PermissionPort; the audit module via domain.AuditPort; business
// modules submit through contract.ApprovalSubmitter and receive final-decision
// callbacks they register via RegisterSubjectCallback.
package approval

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/postgres"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/handler"
	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Compile-time assertion that the module satisfies the cross-module submitter,
// the read-side approval status provider, the batch-action actor, and the
// subject-approval canceller.
var _ contract.ApprovalSubmitter = (*Module)(nil)
var _ contract.ApprovalStatusProvider = (*Module)(nil)
var _ contract.ApprovalBatchActor = (*Module)(nil)
var _ contract.ApprovalCanceller = (*Module)(nil)

// Module owns the approval engine: persistence, services and HTTP transport.
type Module struct {
	pool           *pgxpool.Pool
	repo           *postgres.PostgresRepository
	configSvc      *service.ApprovalConfigService
	runtimeSvc     *service.ApprovalRuntimeService
	runtimeHandler *handler.RuntimeHandler
	configHandler  *handler.ConfigHandler
	perm           domain.PermissionPort
}

// NewModule constructs the approval module.
//
// Dependencies:
//   - pool:          Postgres pool used by every repository.
//   - perm:          IAM module — function-permission + contract data-scope checks.
//   - auditRecorder: shared audit recorder for cross-system audit events.
//   - notifier:      optional approval notifier; nil → NopNotifier (safe stub).
//   - delegate:      optional delegate resolver; nil → NopDelegateResolver.
//   - leave:         optional leave checker; nil → NopLeaveChecker.
//
// perm MUST NOT be nil in production wiring — backend permission enforcement is
// a non-negotiable control. nil is tolerated only in tests.
func NewModule(
	pool *pgxpool.Pool,
	perm domain.PermissionPort,
	auditRecorder auditdomain.Recorder,
	notifier domain.Notifier,
	delegate domain.DelegateResolver,
	leave domain.LeaveChecker,
) *Module {
	repo := postgres.NewPostgresRepository(pool)
	auditPort := adapter.NewAuditLoggerAdapter(auditRecorder)
	directory := adapter.NewPostgresUserDirectory(pool)

	configSvc := service.NewApprovalConfigService(pool, repo)
	runtimeSvc := service.NewApprovalRuntimeService(service.RuntimeDeps{
		Pool:      pool,
		Repo:      repo,
		Perms:     perm,
		Audit:     auditPort,
		Notifier:  notifier,
		Delegate:  delegate,
		Directory: directory,
		Leave:     leave,
	})

	return &Module{
		pool:           pool,
		repo:           repo,
		configSvc:      configSvc,
		runtimeSvc:     runtimeSvc,
		runtimeHandler: handler.NewRuntimeHandler(runtimeSvc),
		configHandler:  handler.NewConfigHandler(configSvc),
		perm:           perm,
	}
}

// RegisterRoutes mounts the approval routes onto an authenticated router.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.runtimeHandler == nil || m.perm == nil {
		return
	}
	// domain.PermissionPort's method set is a superset of
	// middleware.PermissionChecker, so it can gate routes directly.
	var pc middleware.PermissionChecker = m.perm
	transport.RegisterRoutes(r, m.runtimeHandler, m.configHandler, pc)
}

// SubmitForApproval implements contract.ApprovalSubmitter so business modules
// can push a subject into the approval workflow without importing internals.
func (m *Module) SubmitForApproval(ctx context.Context, in contract.ApprovalSubmission) (*contract.ApprovalSubmissionResult, error) {
	if m == nil || m.runtimeSvc == nil {
		return nil, context.Canceled
	}
	req, err := m.runtimeSvc.SubmitApproval(ctx, service.SubmitInput{
		ProcessType:      vo.ProcessType(in.ProcessType),
		SubjectType:      vo.SubjectType(in.SubjectType),
		SubjectID:        in.SubjectID,
		SubjectTitle:     in.SubjectTitle,
		SubjectReference: in.SubjectReference,
		ContractType:     vo.ContractType(in.ContractType),
		ContractID:       in.ContractID,
		PortfolioID:      in.PortfolioID,
		SubmitterID:      in.SubmitterID,
	})
	if err != nil {
		return nil, err
	}
	return &contract.ApprovalSubmissionResult{
		RequestID:     req.ID,
		RequestNumber: req.RequestNumber,
		Status:        string(req.Status),
	}, nil
}

// GetApprovalStage implements contract.ApprovalStatusProvider. Returns the
// latest approval request's stage information for the given subject, or
// (nil, nil) when no request exists. Business modules use this to enrich
// their own read DTOs (e.g. the research-report "derived_review_stage" badge).
//
// The total stage count is taken from the request's resolved process config.
// When the request has no resolved config (older drafts, edge cases), Total
// degrades to the current stage number so the UI still renders sensibly.
// CurrentApprovers and PreviousApprovers are populated from task assignments
// for display in approval grids.
func (m *Module) GetApprovalStage(ctx context.Context, subjectType string, subjectID uuid.UUID) (*contract.ApprovalStageInfo, error) {
	if m == nil || m.runtimeSvc == nil {
		return nil, nil
	}
	st := vo.SubjectType(subjectType)
	if !vo.ValidSubjectType(st) {
		return nil, nil
	}
	req, err := m.runtimeSvc.GetSubjectApprovalStatus(ctx, st, subjectID, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, nil
	}
	total := req.CurrentStageNumber
	if req.ProcessConfigID != nil {
		cfg, cfgErr := m.repo.GetConfig(ctx, *req.ProcessConfigID)
		if cfgErr == nil && cfg != nil {
			n := len(cfg.Stages)
			if n > total {
				total = n
			}
		}
	}

	info := &contract.ApprovalStageInfo{
		RequestID:          req.ID,
		RequestNumber:      req.RequestNumber,
		Status:             string(req.Status),
		CurrentStageNumber: req.CurrentStageNumber,
		TotalStages:        total,
	}

	// Enrich with approver display names from task assignments.
	tasks, taskErr := m.repo.ListTasksByRequest(ctx, req.ID)
	if taskErr == nil {
		for _, t := range tasks {
			if t.AssignedUserID == nil {
				continue
			}
			ai := contract.ApproverInfo{UserID: *t.AssignedUserID, DisplayName: t.AssignedUserName}
			if t.StageNumber == req.CurrentStageNumber && string(t.Status) == "PENDING" {
				info.CurrentApprovers = append(info.CurrentApprovers, ai)
			} else if t.StageNumber == req.CurrentStageNumber-1 && string(t.Status) == "APPROVED" {
				info.PreviousApprovers = append(info.PreviousApprovers, ai)
			}
		}
	}

	return info, nil
}

// ApproveByRequest implements contract.ApprovalBatchActor. It resolves the
// actor's pending task on the given approval request and approves it.
func (m *Module) ApproveByRequest(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID, comment string) error {
	if m == nil || m.runtimeSvc == nil {
		return context.Canceled
	}
	return m.runtimeSvc.ApproveByRequest(ctx, requestID, actorID, comment)
}

// RejectByRequest implements contract.ApprovalBatchActor. It resolves the
// actor's pending task on the given approval request and rejects it.
func (m *Module) RejectByRequest(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID, reason string) error {
	if m == nil || m.runtimeSvc == nil {
		return context.Canceled
	}
	return m.runtimeSvc.RejectByRequest(ctx, requestID, actorID, reason)
}

// RegisterSubjectCallback registers a business-object final-decision callback
// for a subject type. Called at wire-up after both modules are constructed,
// avoiding a circular construction dependency.
func (m *Module) RegisterSubjectCallback(subjectType string, cb contract.ApprovalSubjectCallback) {
	if m == nil || m.runtimeSvc == nil || cb == nil {
		return
	}
	m.runtimeSvc.RegisterSubjectSync(vo.SubjectType(subjectType), &subjectCallbackAdapter{cb: cb})
}

// RegisterSubjectValidator registers a business-object state guard for a
// subject type. The guard is called inside the approve/reject transaction to
// verify the subject is still approvable, preventing approvals of documents
// that were cancelled or invalidated concurrently.
func (m *Module) RegisterSubjectValidator(subjectType string, v contract.ApprovalSubjectValidator) {
	if m == nil || m.runtimeSvc == nil || v == nil {
		return
	}
	m.runtimeSvc.RegisterSubjectValidator(vo.SubjectType(subjectType), v)
}

// CancelApprovalBySubject implements contract.ApprovalCanceller. Cancels the
// active approval request for the given subject, if one exists. Returns nil
// (no-op) when no active request is found. Called by business modules when
// the subject itself is cancelled or withdrawn so the two states stay in sync.
func (m *Module) CancelApprovalBySubject(ctx context.Context, subjectType string, subjectID uuid.UUID, actorID uuid.UUID) error {
	if m == nil || m.runtimeSvc == nil {
		return nil
	}
	return m.runtimeSvc.CancelApprovalBySubject(ctx, vo.SubjectType(subjectType), subjectID, actorID)
}

// RegisterSubjectAccessPort registers the per-subject-type access port so the
// approval engine can authorise reads and writes without inspecting business-
// specific keys. Call at wire-up (cmd/server/main.go) after both modules are
// constructed to avoid a circular construction dependency.
func (m *Module) RegisterSubjectAccessPort(subjectType string, accessor contract.SubjectAccessor) {
	if m == nil || m.runtimeSvc == nil || accessor == nil {
		return
	}
	m.runtimeSvc.RegisterSubjectAccessPort(vo.SubjectType(subjectType), &subjectAccessAdapter{accessor: accessor})
}

// subjectAccessAdapter bridges contract.SubjectAccessor (pkg/contract) to the
// internal domain.ApprovalSubjectAccessPort.
type subjectAccessAdapter struct {
	accessor contract.SubjectAccessor
}

func (a *subjectAccessAdapter) CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID) error {
	return a.accessor.CanViewApprovalSubject(ctx, actorID, string(st), subjectID)
}

func (a *subjectAccessAdapter) CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID) error {
	return a.accessor.CanSubmitApprovalSubject(ctx, actorID, string(st), subjectID)
}

func (a *subjectAccessAdapter) CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, st vo.SubjectType, subjectID uuid.UUID, action domain.ApprovalAction) error {
	return a.accessor.CanActOnApprovalSubject(ctx, actorID, string(st), subjectID, string(action))
}

// subjectCallbackAdapter bridges contract.ApprovalSubjectCallback to the
// internal domain.SubjectSync port.
type subjectCallbackAdapter struct {
	cb contract.ApprovalSubjectCallback
}

func (a *subjectCallbackAdapter) OnApproved(ctx context.Context, st vo.SubjectType, subjectID, requestID uuid.UUID) error {
	return a.cb.OnApprovalDecision(ctx, contract.ApprovalDecision{
		SubjectType: string(st), SubjectID: subjectID, RequestID: requestID, Approved: true,
	})
}

func (a *subjectCallbackAdapter) OnRejected(ctx context.Context, st vo.SubjectType, subjectID, requestID uuid.UUID, reason string) error {
	return a.cb.OnApprovalDecision(ctx, contract.ApprovalDecision{
		SubjectType: string(st), SubjectID: subjectID, RequestID: requestID, Approved: false, Reason: reason,
	})
}
