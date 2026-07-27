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
	DecisionID     uuid.UUID
	DecisionNumber string
	// FundID is nil for a decision on a fund-less portfolio — callers fall
	// back to PortfolioID as the data-permission scope key in that case.
	FundID            *uuid.UUID
	PortfolioID       uuid.UUID
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

	// AccessibleFundIDs is injected by the transport layer's data-permission
	// check (accessibleFundIDs in transport/handler/investment_handler.go),
	// mirroring PortfolioListFilter.AccessibleFundIDs and
	// FundListFilter.AccessibleFundIDs. nil means unrestricted (caller has
	// global/company-wide data scope); a non-nil empty slice means the
	// caller has zero fund access and the query must return no rows.
	// Internal callers that already operate inside an authorised scope (e.g.
	// the approval engine, or a Portfolio V2 handler that already resolved
	// and authorized a single portfolio) should leave this nil.
	AccessibleFundIDs []uuid.UUID
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

// ExecutionListFilter restricts ExecutionRepository.ListByPortfolio. The
// caller has already resolved and authorized a single portfolio (via
// resolvePortfolioByCode's hasFundAccess check), so this filter carries no
// separate AccessibleFundIDs field — portfolio scoping IS the access check.
type ExecutionListFilter struct {
	Status *vo.ExecutionStatus
	Page   int
	Limit  int
}

// ExecutionRepository persists Execution rows.
type ExecutionRepository interface {
	Create(ctx context.Context, tx pgx.Tx, e *entity.Execution) error
	Update(ctx context.Context, tx pgx.Tx, e *entity.Execution) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error)
	ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]*entity.Execution, error)
	ListByFundDate(ctx context.Context, fundID uuid.UUID, businessDate time.Time) ([]*entity.Execution, error)

	// ListByPortfolio returns paginated executions for a single portfolio,
	// optionally filtered by status. Backs the Portfolio V2
	// GET /portfolios/{portfolioCode}/executions endpoint.
	ListByPortfolio(ctx context.Context, portfolioID uuid.UUID, f ExecutionListFilter) ([]*entity.Execution, int, error)
}

// TradeConfirmationListFilter restricts
// TradeConfirmationRepository.ListByPortfolio. See ExecutionListFilter's doc
// comment for why no AccessibleFundIDs field is needed here.
type TradeConfirmationListFilter struct {
	Status *vo.TradeConfirmationStatus
	Page   int
	Limit  int
}

// TradeConfirmationRepository persists TradeConfirmation rows.
type TradeConfirmationRepository interface {
	Create(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error
	Update(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TradeConfirmation, error)
	ListByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.TradeConfirmation, error)
	ListByFundDate(ctx context.Context, fundID uuid.UUID, businessDate time.Time) ([]*entity.TradeConfirmation, error)

	// ListByPortfolio returns paginated trade confirmations for a single
	// portfolio, optionally filtered by status. Backs the Portfolio V2
	// GET /portfolios/{portfolioCode}/confirmations endpoint.
	ListByPortfolio(ctx context.Context, portfolioID uuid.UUID, f TradeConfirmationListFilter) ([]*entity.TradeConfirmation, int, error)

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

// CashRequestListFilter restricts PortfolioCashRequestRepository.ListByPortfolio.
// The caller has already resolved and authorized a single portfolio (via
// resolvePortfolioByCode's data-scope check), so portfolio scoping IS the
// access check — no separate AccessibleFundIDs field is needed.
type CashRequestListFilter struct {
	Status *vo.CashRequestStatus
	Page   int
	Limit  int
}

// PortfolioCashRequestRepository persists the mutable LIVE cash-movement
// approval requests. Unlike the append-only ledger this entity supports status
// transitions via Update. Implementations live in infrastructure/persistence/.
type PortfolioCashRequestRepository interface {
	// Create inserts a brand-new PENDING request in the supplied transaction.
	Create(ctx context.Context, tx pgx.Tx, r *entity.PortfolioCashRequest) error

	// GetByID returns the request or nil when missing (read via pool).
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PortfolioCashRequest, error)

	// GetByIdempotencyKey returns the existing request for a (portfolioID, key)
	// pair or nil when none exists (read via pool). Backs request-level
	// idempotency: a retry with the same key resolves to the original request
	// rather than creating a duplicate. The caller must not pass an empty key.
	GetByIdempotencyKey(ctx context.Context, portfolioID uuid.UUID, key string) (*entity.PortfolioCashRequest, error)

	// GetByFingerprint returns the earliest request for a (portfolioID,
	// fingerprint) pair or nil when none exists (read via pool). Reconciliation-
	// only: the fingerprint is NOT uniquely indexed (two distinct client keys may
	// carry the same payload), so implementations must return the first-submitted
	// match deterministically. The caller must not pass an empty fingerprint.
	GetByFingerprint(ctx context.Context, portfolioID uuid.UUID, fingerprint string) (*entity.PortfolioCashRequest, error)

	// GetForUpdate locks and returns the request row inside the supplied
	// transaction (SELECT ... FOR UPDATE). Used by the approval callback to
	// serialize duplicate materialization attempts. Returns nil when missing.
	GetForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.PortfolioCashRequest, error)

	// Update applies a full row update (status transitions, approval_request_id,
	// resulting_txn_id, decided_by/at). Bumps version. Runs in the supplied
	// transaction so the callback can materialize the ledger row and flip the
	// request status atomically.
	Update(ctx context.Context, tx pgx.Tx, r *entity.PortfolioCashRequest) error

	// ListByPortfolio returns paginated cash requests for a single portfolio,
	// optionally filtered by status, newest first.
	ListByPortfolio(ctx context.Context, portfolioID uuid.UUID, f CashRequestListFilter) ([]*entity.PortfolioCashRequest, int, error)
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
