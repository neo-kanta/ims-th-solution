package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// DecisionSubjectRef is the minimal projection of a decision used for
// authorisation checks in batch operations. It carries only the fields needed
// to verify access — never business details like amount, rationale, or status.
type DecisionSubjectRef struct {
	DecisionID        uuid.UUID
	DecisionNumber    string
	FundID            uuid.UUID
	ApprovalRequestID *uuid.UUID
}

// DecisionListFilter is a thin filter for the decision list endpoint. Empty
// fields are ignored — implementations should build the WHERE clause defensively.
type DecisionListFilter struct {
	FundID           *uuid.UUID
	PortfolioID      *uuid.UUID
	BusinessDate     *time.Time
	BusinessDateFrom *time.Time
	BusinessDateTo   *time.Time
	Status           *vo.DecisionLifecycleStatus
	InstrumentCode   string
	Search           string
	Page             int
	Limit            int

	// Extended filters for the batch-approval screen.
	DecisionNumber   string
	DecisionType     string
	ProcessType      string
	ProductType      string
	ResearchReportNo string
}

// DecisionRepository persists investment Decision aggregates.
// Implementations live in infrastructure/persistence/.
type DecisionRepository interface {
	// GetByID returns the decision or nil when missing.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Decision, error)

	// GetByDecisionNumber returns the decision matching the human-friendly
	// reference number (e.g. "DEC-20260615-0001"), or nil when missing.
	GetByDecisionNumber(ctx context.Context, decisionNumber string) (*entity.Decision, error)

	// FindDecisionSubjectRefByNumber returns the minimal authorisation
	// projection for a decision number: only id, fund_id, and
	// approval_request_id are populated — no amount, rationale, or status.
	// Returns nil when the decision number is not found.
	FindDecisionSubjectRefByNumber(ctx context.Context, decisionNumber string) (*DecisionSubjectRef, error)

	// Create persists a brand-new DRAFT decision row. The decision's ID must
	// already be set; DecisionNumber collisions return *ErrDecisionNumberConflict.
	Create(ctx context.Context, tx pgx.Tx, d *entity.Decision) error

	// Update applies a full row update. Status changes are allowed but the
	// application command layer is the only path that should call this with
	// a flipped Status.
	Update(ctx context.Context, tx pgx.Tx, d *entity.Decision) error

	// List returns decisions matching the filter (paginated).
	List(ctx context.Context, f DecisionListFilter) ([]*entity.Decision, int, error)

	// NextDecisionNumber returns the next monotonically-increasing decision
	// number string used for human-friendly references.
	NextDecisionNumber(ctx context.Context, tx pgx.Tx, businessDate time.Time) (string, error)

	// UpdateStatus atomically transitions the decision to the given status.
	// Implementations should use optimistic concurrency / compare-and-set so
	// that two concurrent submits cannot both land.
	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		to vo.DecisionStatus,
		checkGroupID uuid.UUID,
		updatedBy uuid.UUID,
		updatedAt time.Time,
	) error
}

// DecisionLineRepository persists the child lines for basket/rebalance/switch
// decisions. Implementations live in infrastructure/persistence/.
type DecisionLineRepository interface {
	// ListByDecision returns all lines for the given decision, ordered by line_number.
	ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]*entity.DecisionLine, error)

	// CreateLines inserts all lines for a decision inside a transaction.
	CreateLines(ctx context.Context, tx pgx.Tx, lines []*entity.DecisionLine) error

	// ReplaceLines deletes all existing lines for a decision and inserts the new
	// set inside a transaction. Used when a DRAFT basket decision is edited.
	ReplaceLines(ctx context.Context, tx pgx.Tx, decisionID uuid.UUID, lines []*entity.DecisionLine) error
}

// ExecutionRepository persists Execution rows.
type ExecutionRepository interface {
	Create(ctx context.Context, tx pgx.Tx, e *entity.Execution) error
	Update(ctx context.Context, tx pgx.Tx, e *entity.Execution) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]*entity.Execution, error)
	ListByFundDate(ctx context.Context, fundID uuid.UUID, businessDate time.Time) ([]*entity.Execution, error)
}

// TradeConfirmationRepository persists TradeConfirmation rows.
type TradeConfirmationRepository interface {
	Create(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error
	Update(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TradeConfirmation, error)
	ListByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.TradeConfirmation, error)
	ListByFundDate(ctx context.Context, fundID uuid.UUID, businessDate time.Time) ([]*entity.TradeConfirmation, error)

	// GetByBrokerReference returns the existing confirmation for a broker
	// reference string, or nil. Used by the batch importer to dedupe rows
	// against confirmations already persisted by earlier batches.
	GetByBrokerReference(ctx context.Context, brokerReference string) (*entity.TradeConfirmation, error)
}

// TradeConfirmationImportRepository persists the per-batch import header and
// per-row outcomes. Implementations live in infrastructure/persistence/.
type TradeConfirmationImportRepository interface {
	// CreateBatch inserts the header row in the supplied transaction.
	CreateBatch(ctx context.Context, tx pgx.Tx, b *entity.TradeConfirmationImportBatch) error

	// UpdateBatchSummary finalises the header counters and status when the
	// batch has been processed.
	UpdateBatchSummary(ctx context.Context, tx pgx.Tx, b *entity.TradeConfirmationImportBatch) error

	// CreateItem inserts a per-row outcome.
	CreateItem(ctx context.Context, tx pgx.Tx, i *entity.TradeConfirmationImportItem) error

	// GetBatch returns the header row.
	GetBatch(ctx context.Context, id uuid.UUID) (*entity.TradeConfirmationImportBatch, error)

	// ListBatchItems returns all per-row outcomes for a batch.
	ListBatchItems(ctx context.Context, batchID uuid.UUID) ([]*entity.TradeConfirmationImportItem, error)
}

// InvestmentProcessGuardRepository provides read models required to decide
// whether a user may execute one investment process step for a contract/date.
type InvestmentProcessGuardRepository interface {
	// GetActiveDaySetting resolves the effective workflow day setting for the
	// contract/date. Contract-specific settings should outrank GLOBAL settings.
	GetActiveDaySetting(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*entity.WorkflowDaySetting, error)

	// FindBlockingControlDecision returns an active rejection that blocks either
	// the whole day or the requested process step. Nil means no active block.
	FindBlockingControlDecision(
		ctx context.Context,
		contractID uuid.UUID,
		businessDate time.Time,
		processStep vo.ProcessStepKey,
	) (*entity.BlockingControlDecision, error)

	// FindUserProcessAssignment returns the best active assignment authorizing
	// userID for the process step. Nil means the user is not assigned.
	FindUserProcessAssignment(
		ctx context.Context,
		userID uuid.UUID,
		contractID uuid.UUID,
		businessDate time.Time,
		processStep vo.ProcessStepKey,
	) (*entity.ProcessAssignmentMatch, error)
}
