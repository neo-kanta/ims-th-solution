package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository defines persistence operations for the permission-management
// module. SQL belongs in infrastructure/persistence; application services own
// state transitions and transaction boundaries.
type Repository interface {
	NextRequestNo(ctx context.Context) (string, error)
	ListChangeRequests(ctx context.Context, filter ChangeRequestFilter) ([]ChangeRequest, int, error)
	GetChangeRequest(ctx context.Context, id uuid.UUID) (*ChangeRequest, error)
	CreateChangeRequest(ctx context.Context, tx pgx.Tx, request *ChangeRequest) error
	UpdateChangeRequest(ctx context.Context, tx pgx.Tx, request *ChangeRequest) error
	LockChangeRequest(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*ChangeRequest, error)

	ListItems(ctx context.Context, requestID uuid.UUID) ([]ChangeItem, error)
	AddItem(ctx context.Context, tx pgx.Tx, item *ChangeItem) error
	UpdateItem(ctx context.Context, tx pgx.Tx, item *ChangeItem) error
	DeleteItem(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, itemID uuid.UUID) error
	CountItems(ctx context.Context, requestID uuid.UUID) (int, error)

	FindApprovalSetting(ctx context.Context, requestType string, riskLevel string) (*ApprovalSetting, error)
	ListApprovalSettings(ctx context.Context) ([]ApprovalSetting, error)
	GetApprovalSetting(ctx context.Context, id uuid.UUID) (*ApprovalSetting, error)
	CreateRequestStep(ctx context.Context, tx pgx.Tx, step *ApprovalStep) error
	CreateStepApprover(ctx context.Context, tx pgx.Tx, approver *StepApprover) error
	DeleteRequestApprovalState(ctx context.Context, tx pgx.Tx, requestID uuid.UUID) error
	ListApprovalSteps(ctx context.Context, requestID uuid.UUID) ([]ApprovalStep, error)
	GetApprovalStep(ctx context.Context, requestID uuid.UUID, stepID uuid.UUID) (*ApprovalStep, error)
	GetCurrentPendingStep(ctx context.Context, requestID uuid.UUID) (*ApprovalStep, error)
	UpdateApprovalStep(ctx context.Context, tx pgx.Tx, step *ApprovalStep) error
	UpdateStepApproverDecision(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stepID uuid.UUID, actorID uuid.UUID, roleCode string, status string, comment string) error
	StartNextApprovalStep(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, completedStepNo int) (bool, error)

	ListComments(ctx context.Context, requestID uuid.UUID) ([]Comment, error)
	AddComment(ctx context.Context, tx pgx.Tx, comment *Comment) error
	UpdateComment(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, commentID uuid.UUID, actorID uuid.UUID, text string) error
	DeleteComment(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, commentID uuid.UUID, actorID uuid.UUID) error

	ReplaceChecks(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, checks []Check) error
	ListChecks(ctx context.Context, requestID uuid.UUID) ([]Check, error)
	HasBlockingChecks(ctx context.Context, requestID uuid.UUID) (bool, error)

	ListLabels(ctx context.Context) ([]Label, error)
	UpsertLabel(ctx context.Context, tx pgx.Tx, label *Label) error
	AddRequestLabel(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, labelCode string, actorID uuid.UUID) error
	RemoveRequestLabel(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, labelID uuid.UUID) error
	ListRequestLabels(ctx context.Context, requestID uuid.UUID) ([]Label, error)

	AppendEvent(ctx context.Context, tx pgx.Tx, event WorkflowEvent) error
	ListEvents(ctx context.Context, requestID uuid.UUID) ([]WorkflowEvent, error)
	AppendAuditLog(ctx context.Context, tx pgx.Tx, log AuditLog) error

	ListRoles(ctx context.Context) ([]Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoleByCode(ctx context.Context, code string) (*Role, error)
	GetUserApprovedRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)

	ListUsers(ctx context.Context, search string, page int, limit int) ([]UserSummary, int, error)
	GetUser(ctx context.Context, id uuid.UUID) (*UserSummary, error)
	ListGroups(ctx context.Context, search string, page int, limit int) ([]GroupSummary, int, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*GroupSummary, error)

	ListFunctionDefinitions(ctx context.Context) ([]FunctionDefinition, error)
	FunctionDefinitionExists(ctx context.Context, code string) (bool, error)
	ListFunctionRights(ctx context.Context) ([]FunctionRight, error)
	UpsertFunctionRight(ctx context.Context, tx pgx.Tx, right FunctionRight) error
	ListDataRights(ctx context.Context) ([]DataRight, error)
	UpsertDataRight(ctx context.Context, tx pgx.Tx, right DataRight) error
	InsertRoleAssignment(ctx context.Context, tx pgx.Tx, userID uuid.UUID, roleID uuid.UUID, actorID uuid.UUID) error
	InsertGroupMembership(ctx context.Context, tx pgx.Tx, userID uuid.UUID, groupID uuid.UUID, actorID uuid.UUID) error
	DeleteGroupMembership(ctx context.Context, tx pgx.Tx, userID uuid.UUID, groupID uuid.UUID) error

	EffectivePermissions(ctx context.Context, userID uuid.UUID) (*EffectivePermissions, error)
	ListAuditLogs(ctx context.Context, page int, limit int) ([]AuditLog, int, error)
	ListNotificationSettings(ctx context.Context) ([]NotificationSetting, error)
	UpdateNotificationSetting(ctx context.Context, tx pgx.Tx, setting NotificationSetting) error
}
