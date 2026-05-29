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

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/postgres"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Compile-time assertion that the module satisfies the cross-module submitter.
var _ contract.ApprovalSubmitter = (*Module)(nil)

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

// RegisterSubjectCallback registers a business-object final-decision callback
// for a subject type. Called at wire-up after both modules are constructed,
// avoiding a circular construction dependency.
func (m *Module) RegisterSubjectCallback(subjectType string, cb contract.ApprovalSubjectCallback) {
	if m == nil || m.runtimeSvc == nil || cb == nil {
		return
	}
	m.runtimeSvc.RegisterSubjectSync(vo.SubjectType(subjectType), &subjectCallbackAdapter{cb: cb})
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
