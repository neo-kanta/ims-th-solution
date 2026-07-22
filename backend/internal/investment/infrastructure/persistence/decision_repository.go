package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresDecisionRepository implements domain.DecisionRepository.
type PostgresDecisionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresDecisionRepository wires the repo.
func NewPostgresDecisionRepository(pool *pgxpool.Pool) *PostgresDecisionRepository {
	return &PostgresDecisionRepository{pool: pool}
}

const decisionSelect = `
	SELECT id, decision_number, fund_id, portfolio_id,
	       instrument_id, instrument_code, business_date,
	       side, quantity, amount, limit_price, currency, exchange,
	       research_report_id, research_report_no, rationale,
	       status, approval_request_id, approval_status, compliance_check_group_id,
	       submitter_user_id, submitted_at,
	       cancelled_at, cancelled_by, cancellation_reason,
	       ready_for_execution_at,
	       created_at, created_by, updated_at, updated_by,
	       COALESCE(decision_type, 'SINGLE_ORDER'),
	       COALESCE(process_type, 'INVESTMENT_DECISION'),
	       COALESCE(product_type, 'MUTUAL_FUND'),
	       COALESCE(strategy_code, ''),
	       COALESCE(amendment_no, 0),
	       compliance_release_approval_request_id
	FROM investment__decisions`

func (r *PostgresDecisionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Decision, error) {
	row := r.pool.QueryRow(ctx, decisionSelect+` WHERE id = $1`, id)
	d, err := scanDecision(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

func (r *PostgresDecisionRepository) GetByDecisionNumber(ctx context.Context, decisionNumber string) (*entity.Decision, error) {
	row := r.pool.QueryRow(ctx, decisionSelect+` WHERE decision_number = $1`, decisionNumber)
	d, err := scanDecision(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

func (r *PostgresDecisionRepository) FindDecisionSubjectRefByNumber(ctx context.Context, decisionNumber string) (*domain.DecisionSubjectRef, error) {
	var ref domain.DecisionSubjectRef
	err := r.pool.QueryRow(ctx,
		`SELECT id, decision_number, fund_id, approval_request_id
		   FROM investment__decisions
		  WHERE decision_number = $1`,
		decisionNumber,
	).Scan(&ref.DecisionID, &ref.DecisionNumber, &ref.FundID, &ref.ApprovalRequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ref, nil
}

func (r *PostgresDecisionRepository) Create(ctx context.Context, tx pgx.Tx, d *entity.Decision) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__decisions (
			id, decision_number, fund_id, portfolio_id,
			instrument_id, instrument_code, business_date,
			side, quantity, amount, limit_price, currency, exchange,
			research_report_id, research_report_no, rationale,
			status, approval_request_id, approval_status, compliance_check_group_id,
			submitter_user_id, submitted_at,
			cancelled_at, cancelled_by, cancellation_reason,
			ready_for_execution_at,
			created_at, created_by, updated_at, updated_by,
			decision_type, process_type, product_type, strategy_code, amendment_no,
			compliance_release_approval_request_id
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16,
			$17, $18, $19, $20,
			$21, $22,
			$23, $24, $25,
			$26,
			$27, $28, $29, $30,
			$31, $32, $33, $34, $35,
			$36
		)`,
		d.ID, d.DecisionNumber, d.FundID, d.PortfolioID,
		d.InstrumentID, nullableStr(d.InstrumentCode), d.BusinessDate,
		nullableOrderSide(d.Side), d.Quantity, d.Amount, d.LimitPrice, d.Currency, d.Exchange,
		d.ResearchReportID, d.ResearchReportNo, d.Rationale,
		string(d.Status), d.ApprovalRequestID, d.ApprovalStatus, d.ComplianceCheckGroupID,
		d.SubmitterUserID, d.SubmittedAt,
		d.CancelledAt, d.CancelledBy, d.CancellationReason,
		d.ReadyForExecutionAt,
		d.CreatedAt, d.CreatedBy, d.UpdatedAt, d.UpdatedBy,
		string(d.DecisionType), string(d.ProcessType), string(d.ProductType),
		nullableStr(d.StrategyCode), d.AmendmentNo,
		d.ComplianceReleaseApprovalRequestID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return &domain.ErrDecisionNumberConflict{DecisionNumber: d.DecisionNumber}
		}
		return fmt.Errorf("insert decision: %w", err)
	}
	return nil
}

func (r *PostgresDecisionRepository) Update(ctx context.Context, tx pgx.Tx, d *entity.Decision) error {
	_, err := tx.Exec(ctx, `
		UPDATE investment__decisions SET
			fund_id = $2, portfolio_id = $3,
			instrument_id = $4, instrument_code = $5, business_date = $6,
			side = $7, quantity = $8, amount = $9, limit_price = $10,
			currency = $11, exchange = $12,
			research_report_id = $13, research_report_no = $14, rationale = $15,
			status = $16, approval_request_id = $17, approval_status = $18,
			compliance_check_group_id = $19,
			submitter_user_id = $20, submitted_at = $21,
			cancelled_at = $22, cancelled_by = $23, cancellation_reason = $24,
			ready_for_execution_at = $25,
			updated_at = $26, updated_by = $27,
			decision_type = $28, process_type = $29, product_type = $30,
			strategy_code = $31, amendment_no = $32,
			compliance_release_approval_request_id = $33
		WHERE id = $1`,
		d.ID, d.FundID, d.PortfolioID,
		d.InstrumentID, nullableStr(d.InstrumentCode), d.BusinessDate,
		nullableOrderSide(d.Side), d.Quantity, d.Amount, d.LimitPrice, d.Currency, d.Exchange,
		d.ResearchReportID, d.ResearchReportNo, d.Rationale,
		string(d.Status), d.ApprovalRequestID, d.ApprovalStatus,
		d.ComplianceCheckGroupID,
		d.SubmitterUserID, d.SubmittedAt,
		d.CancelledAt, d.CancelledBy, d.CancellationReason,
		d.ReadyForExecutionAt,
		d.UpdatedAt, d.UpdatedBy,
		string(d.DecisionType), string(d.ProcessType), string(d.ProductType),
		nullableStr(d.StrategyCode), d.AmendmentNo,
		d.ComplianceReleaseApprovalRequestID,
	)
	if err != nil {
		return fmt.Errorf("update decision: %w", err)
	}
	return nil
}

func (r *PostgresDecisionRepository) List(ctx context.Context, f domain.DecisionListFilter) ([]*entity.Decision, int, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	where := []string{"TRUE"}
	args := []any{}
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}
	if f.FundID != nil {
		add("fund_id = $%d", *f.FundID)
	}
	if f.PortfolioID != nil {
		add("portfolio_id = $%d", *f.PortfolioID)
	}
	if f.BusinessDate != nil {
		add("business_date = $%d", *f.BusinessDate)
	}
	if f.BusinessDateFrom != nil {
		add("business_date >= $%d", *f.BusinessDateFrom)
	}
	if f.BusinessDateTo != nil {
		add("business_date <= $%d", *f.BusinessDateTo)
	}
	if f.Status != nil {
		add("status = $%d", string(*f.Status))
	}
	if s := strings.TrimSpace(f.InstrumentCode); s != "" {
		add("instrument_code = $%d", strings.ToUpper(s))
	}
	if s := strings.TrimSpace(f.DecisionNumber); s != "" {
		add("decision_number = $%d", s)
	}
	if s := strings.TrimSpace(f.DecisionType); s != "" {
		add("decision_type = $%d", s)
	}
	if s := strings.TrimSpace(f.ProcessType); s != "" {
		add("process_type = $%d", s)
	}
	if s := strings.TrimSpace(f.ProductType); s != "" {
		add("product_type = $%d", s)
	}
	if s := strings.TrimSpace(f.ResearchReportNo); s != "" {
		add("research_report_no = $%d", s)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		args = append(args, "%"+s+"%")
		where = append(where, fmt.Sprintf(
			"(decision_number ILIKE $%d OR instrument_code ILIKE $%d OR research_report_no ILIKE $%d)",
			len(args), len(args), len(args)))
	}
	// Data permission: AccessibleFundIDs nil means unrestricted; a non-nil
	// empty slice means the caller has zero fund access, so short-circuit to
	// an empty result without querying (mirrors fund_repository.go and
	// portfolio_repository.go's List implementations).
	if f.AccessibleFundIDs != nil {
		if len(f.AccessibleFundIDs) == 0 {
			return []*entity.Decision{}, 0, nil
		}
		add("fund_id = ANY($%d)", f.AccessibleFundIDs)
	}
	whereClause := strings.Join(where, " AND ")

	var total int
	countQ := "SELECT COUNT(*) FROM investment__decisions WHERE " + whereClause
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count decisions: %w", err)
	}

	args = append(args, limit, offset)
	q := decisionSelect + ` WHERE ` + whereClause +
		fmt.Sprintf(` ORDER BY business_date DESC, created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list decisions: %w", err)
	}
	defer rows.Close()

	items := []*entity.Decision{}
	for rows.Next() {
		d, err := scanDecision(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return items, total, nil
}

func (r *PostgresDecisionRepository) NextDecisionNumber(ctx context.Context, tx pgx.Tx, businessDate time.Time) (string, error) {
	var n int
	prefix := "DEC-" + businessDate.Format("20060102") + "-"
	row := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__decisions
		WHERE decision_number LIKE $1`, prefix+"%")
	if err := row.Scan(&n); err != nil {
		return "", fmt.Errorf("next decision number: %w", err)
	}
	return fmt.Sprintf("%s%04d", prefix, n+1), nil
}

// UpdateStatus preserves the legacy narrow DecisionStatus API used by the
// pre-trade submit pipeline. Lifecycle-rich status flips go through Update.
func (r *PostgresDecisionRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	to vo.DecisionStatus,
	checkGroupID uuid.UUID,
	updatedBy uuid.UUID,
	updatedAt time.Time,
) error {
	var checkArg any
	if checkGroupID != uuid.Nil {
		checkArg = checkGroupID
	}
	statusStr := mapNarrowToLifecycle(to)
	_, err := r.pool.Exec(ctx, `
		UPDATE investment__decisions
		   SET status = $2,
		       compliance_check_group_id = COALESCE($3::uuid, compliance_check_group_id),
		       updated_at = $4,
		       updated_by = $5
		 WHERE id = $1`,
		id, statusStr, checkArg, updatedAt, updatedBy,
	)
	if err != nil {
		return fmt.Errorf("update decision status: %w", err)
	}
	return nil
}

func mapNarrowToLifecycle(s vo.DecisionStatus) string {
	switch s {
	case vo.DecisionStatusDraft:
		return string(vo.DecisionLifecycleDraft)
	case vo.DecisionStatusSubmitted:
		return string(vo.DecisionLifecyclePendingApproval)
	case vo.DecisionStatusBlocked:
		return string(vo.DecisionLifecycleRejected)
	case vo.DecisionStatusExecuted:
		return string(vo.DecisionLifecycleExecuted)
	case vo.DecisionStatusCancelled:
		return string(vo.DecisionLifecycleCancelled)
	}
	return string(s)
}

// scanDecision is shared by GetByID, GetByDecisionNumber, and List.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanDecision(row rowScanner) (*entity.Decision, error) {
	d := &entity.Decision{}
	var (
		sidePtr                            *string
		instrumentCode                     *string
		status                             string
		instrument                         *uuid.UUID
		quantity, amount, limitPrice       *decimal.Decimal
		researchReportID                   *uuid.UUID
		researchReportNo                   string
		approvalRequestID                  *uuid.UUID
		complianceReleaseApprovalRequestID *uuid.UUID
		approvalStatus                     string
		checkGroupID                       *uuid.UUID
		submittedAt                        *time.Time
		cancelledAt                        *time.Time
		cancelledBy                        *uuid.UUID
		cancellationReason                 string
		readyForExecutionAt                *time.Time
		exchange                           string
		rationale                          string
		decisionType                       string
		processType                        string
		productType                        string
		strategyCode                       string
		amendmentNo                        int
	)
	if err := row.Scan(
		&d.ID, &d.DecisionNumber, &d.FundID, &d.PortfolioID,
		&instrument, &instrumentCode, &d.BusinessDate,
		&sidePtr, &quantity, &amount, &limitPrice, &d.Currency, &exchange,
		&researchReportID, &researchReportNo, &rationale,
		&status, &approvalRequestID, &approvalStatus, &checkGroupID,
		&d.SubmitterUserID, &submittedAt,
		&cancelledAt, &cancelledBy, &cancellationReason,
		&readyForExecutionAt,
		&d.CreatedAt, &d.CreatedBy, &d.UpdatedAt, &d.UpdatedBy,
		&decisionType, &processType, &productType, &strategyCode, &amendmentNo,
		&complianceReleaseApprovalRequestID,
	); err != nil {
		return nil, err
	}
	if sidePtr != nil {
		d.Side = vo.OrderSide(*sidePtr)
	}
	if instrumentCode != nil {
		d.InstrumentCode = *instrumentCode
	}
	d.Status = vo.DecisionLifecycleStatus(status)
	d.DecisionType = vo.DecisionType(decisionType)
	d.ProcessType = vo.DecisionProcessType(processType)
	d.ProductType = vo.DecisionProductType(productType)
	d.StrategyCode = strategyCode
	d.AmendmentNo = amendmentNo
	d.InstrumentID = instrument
	d.Quantity = quantity
	d.Amount = amount
	d.LimitPrice = limitPrice
	d.ResearchReportID = researchReportID
	d.ResearchReportNo = researchReportNo
	d.ApprovalRequestID = approvalRequestID
	d.ComplianceReleaseApprovalRequestID = complianceReleaseApprovalRequestID
	d.ApprovalStatus = approvalStatus
	d.ComplianceCheckGroupID = checkGroupID
	d.SubmittedAt = submittedAt
	d.CancelledAt = cancelledAt
	d.CancelledBy = cancelledBy
	d.CancellationReason = cancellationReason
	d.ReadyForExecutionAt = readyForExecutionAt
	d.Exchange = exchange
	d.Rationale = rationale
	return d, nil
}

// nullableStr converts an empty string to nil for nullable VARCHAR columns.
func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullableOrderSide converts an empty OrderSide to nil for nullable side columns.
func nullableOrderSide(s vo.OrderSide) *string {
	if s == "" {
		return nil
	}
	v := string(s)
	return &v
}

var _ domain.DecisionRepository = (*PostgresDecisionRepository)(nil)
