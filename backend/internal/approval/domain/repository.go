package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// ─────────────────────────────────────────────────────────────────────────────
// Read models / filters
// ─────────────────────────────────────────────────────────────────────────────

// GroupListFilter filters approval group listings.
type GroupListFilter struct {
	ActiveOnly bool
	Search     string
}

// ProcessListFilter filters approval process configuration listings.
type ProcessListFilter struct {
	ProcessType vo.ProcessType
	ActiveOnly  bool
}

// RequestListFilter filters approval request listings.
type RequestListFilter struct {
	Status      vo.RequestStatus
	ProcessType vo.ProcessType
	SubmitterID *uuid.UUID
	ContractID  *uuid.UUID
	// ViewerID is the authenticated user making the list request. When non-nil
	// and ContractID is also set, the service enforces data permission for the
	// viewer on that contract before returning results.
	ViewerID uuid.UUID
	Page     int
	Limit    int
}

// InboxFilter filters the per-approver inbox.
type InboxFilter struct {
	UserID uuid.UUID
	Status vo.TaskStatus // empty → PENDING
	Page   int
	Limit  int
}

// InboxItem is a denormalised approver work item joined with its request.
type InboxItem struct {
	Task    entity.ApprovalTask
	Request entity.ApprovalRequest
}

// ─────────────────────────────────────────────────────────────────────────────
// Repository interfaces (implementations in infrastructure/postgres)
// ─────────────────────────────────────────────────────────────────────────────

// GroupRepository persists approval groups and their members.
type GroupRepository interface {
	CreateGroup(ctx context.Context, tx pgx.Tx, g *entity.ApprovalGroup) error
	UpdateGroup(ctx context.Context, tx pgx.Tx, g *entity.ApprovalGroup) error
	GetGroup(ctx context.Context, id uuid.UUID) (*entity.ApprovalGroup, error)
	GetGroupByCode(ctx context.Context, code string) (*entity.ApprovalGroup, error)
	ListGroups(ctx context.Context, f GroupListFilter) ([]*entity.ApprovalGroup, error)

	AddGroupMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalGroupMember) error
	UpdateGroupMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalGroupMember) error
	GetGroupMember(ctx context.Context, memberID uuid.UUID) (*entity.ApprovalGroupMember, error)
	ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error)
	// ListEligibleGroupMembers returns active + APPROVED members ordered by priority.
	ListEligibleGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error)
}

// TeamRepository persists approval teams, their contracts and members.
type TeamRepository interface {
	CreateTeam(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTeam) error
	UpdateTeam(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTeam) error
	GetTeam(ctx context.Context, id uuid.UUID) (*entity.ApprovalTeam, error)
	GetTeamByCode(ctx context.Context, code string) (*entity.ApprovalTeam, error)
	ListTeams(ctx context.Context) ([]*entity.ApprovalTeam, error)

	AssignContract(ctx context.Context, tx pgx.Tx, c *entity.ApprovalTeamContract) error
	DeactivateContractAssignment(ctx context.Context, tx pgx.Tx, contractID uuid.UUID) error
	GetActiveTeamForContract(ctx context.Context, contractID uuid.UUID) (*entity.ApprovalTeam, error)
	ListTeamContracts(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamContract, error)

	AddTeamMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalTeamMember) error
	UpdateTeamMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalTeamMember) error
	RemoveTeamMember(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) error
	GetTeamMember(ctx context.Context, memberID uuid.UUID) (*entity.ApprovalTeamMember, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error)
	ListEligibleTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error)
}

// ProcessRepository persists approval process configurations and stages.
type ProcessRepository interface {
	CreateConfig(ctx context.Context, tx pgx.Tx, c *entity.ApprovalProcessConfig) error
	UpdateConfig(ctx context.Context, tx pgx.Tx, c *entity.ApprovalProcessConfig) error
	SetConfigActive(ctx context.Context, tx pgx.Tx, id uuid.UUID, active bool, updatedBy *uuid.UUID) error
	GetConfig(ctx context.Context, id uuid.UUID) (*entity.ApprovalProcessConfig, error)
	GetConfigByCode(ctx context.Context, code string) (*entity.ApprovalProcessConfig, error)
	ListConfigs(ctx context.Context, f ProcessListFilter) ([]*entity.ApprovalProcessConfig, error)
	ReplaceStages(ctx context.Context, tx pgx.Tx, configID uuid.UUID, stages []entity.ApprovalProcessStage) error
	ListStages(ctx context.Context, configID uuid.UUID) ([]entity.ApprovalProcessStage, error)
	// Resolve returns the best-matching active config (with stages) for the
	// process type and contract scope, honouring effective_date <= asOf.
	Resolve(ctx context.Context, pt vo.ProcessType, contractID *uuid.UUID, ct vo.ContractType, asOf time.Time) (*entity.ApprovalProcessConfig, error)
}

// RequestRepository persists approval request aggregates.
type RequestRepository interface {
	NextRequestNumber(ctx context.Context, tx pgx.Tx) (string, error)
	CreateRequest(ctx context.Context, tx pgx.Tx, r *entity.ApprovalRequest) error
	GetRequest(ctx context.Context, id uuid.UUID) (*entity.ApprovalRequest, error)
	// GetRequestForUpdate locks the request row inside the supplied transaction.
	GetRequestForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ApprovalRequest, error)
	UpdateRequestState(ctx context.Context, tx pgx.Tx, r *entity.ApprovalRequest) error
	ListRequests(ctx context.Context, f RequestListFilter) ([]*entity.ApprovalRequest, int, error)
	GetActiveBySubject(ctx context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error)
	GetLatestBySubject(ctx context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error)
}

// TaskRepository persists approval tasks.
type TaskRepository interface {
	CreateTask(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTask) error
	GetTask(ctx context.Context, id uuid.UUID) (*entity.ApprovalTask, error)
	GetTaskForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ApprovalTask, error)
	UpdateTask(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTask) error
	ListTasksByRequest(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalTask, error)
	ListPendingByStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int) ([]*entity.ApprovalTask, error)
	CountApprovedInStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int) (int, error)
	CancelPendingByRequest(ctx context.Context, tx pgx.Tx, requestID uuid.UUID) error
	SkipOtherPendingInStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int, exceptTaskID uuid.UUID) error
	Inbox(ctx context.Context, f InboxFilter) ([]*InboxItem, int, error)
	// FindPendingTaskForActor returns the single PENDING task assigned to
	// actorID on the given request, or nil if none exists.
	FindPendingTaskForActor(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*entity.ApprovalTask, error)
}

// EventRepository persists the immutable approval timeline.
type EventRepository interface {
	AppendEvent(ctx context.Context, tx pgx.Tx, e *entity.ApprovalEvent) error
	ListEvents(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalEvent, error)
}

// SignatureRepository persists approval signature/stamp records.
type SignatureRepository interface {
	CreateSignature(ctx context.Context, tx pgx.Tx, s *entity.ApprovalSignatureRecord) error
	ListSignatures(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalSignatureRecord, error)
}

// SyncFailureFilter filters the operator-facing sync-failure listing.
type SyncFailureFilter struct {
	Status vo.SyncFailureStatus // empty means all statuses
	Page   int
	Limit  int
}

// SyncFailureRepository persists durable subject-sync replay records (see
// entity.SyncFailure).
type SyncFailureRepository interface {
	// CreateSyncFailure inserts a new PENDING record for a just-failed
	// subject-sync callback. Runs inside the caller's transaction so it is
	// never lost even though the approval decision itself already committed
	// (the caller opens a short-lived tx solely for this insert).
	CreateSyncFailure(ctx context.Context, tx pgx.Tx, f *entity.SyncFailure) error
	GetSyncFailure(ctx context.Context, id uuid.UUID) (*entity.SyncFailure, error)
	// GetSyncFailureForUpdate row-locks the record for a retry attempt.
	GetSyncFailureForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.SyncFailure, error)
	ListSyncFailures(ctx context.Context, f SyncFailureFilter) ([]*entity.SyncFailure, int, error)
	// UpdateSyncFailure persists the outcome of a retry attempt (attempt count,
	// last error, status, resolved-by/at).
	UpdateSyncFailure(ctx context.Context, tx pgx.Tx, f *entity.SyncFailure) error
}

// Repository aggregates every approval persistence concern. The Postgres
// implementation satisfies all of them, which keeps module wiring and test
// fakes to a single type.
type Repository interface {
	GroupRepository
	TeamRepository
	ProcessRepository
	RequestRepository
	TaskRepository
	EventRepository
	SignatureRepository
	SyncFailureRepository
}
