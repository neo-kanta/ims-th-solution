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

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// pgUniqueViolation is the Postgres SQLSTATE code for a unique-constraint
// violation. We map it to ErrResearchReportNoAlreadyExists so the HTTP
// layer can return 409 instead of 500 when two concurrent creates race
// past the in-memory pre-check.
const pgUniqueViolation = "23505"

// PostgresResearchReportRepository implements domain.ResearchReportRepository.
type PostgresResearchReportRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresResearchReportRepository wires the repo.
func NewPostgresResearchReportRepository(pool *pgxpool.Pool) *PostgresResearchReportRepository {
	return &PostgresResearchReportRepository{pool: pool}
}

const researchReportSelect = `
	SELECT id, report_no, report_date, effective_date,
	       owner_user_id, author_user_id, applicable_contract_id,
	       instrument_type, instrument_code, instrument_name, market, currency,
	       recommendation, report_title,
	       company_overview, company_outlook, esg_comment, financial_status, investment_analysis,
	       rejection_reason, post_submission_note,
	       report_status, review_status,
	       invalidated_at, invalidated_by, invalidation_reason,
	       created_at, created_by, updated_at, updated_by, deleted_at
	FROM investment__research_reports`

// Create inserts a research report row inside the supplied transaction.
func (r *PostgresResearchReportRepository) Create(ctx context.Context, tx pgx.Tx, e *entity.ResearchReport) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__research_reports (
			id, report_no, report_date, effective_date,
			owner_user_id, author_user_id, applicable_contract_id,
			instrument_type, instrument_code, instrument_name, market, currency,
			recommendation, report_title,
			company_overview, company_outlook, esg_comment, financial_status, investment_analysis,
			rejection_reason, post_submission_note,
			report_status, review_status,
			created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14,
			$15, $16, $17, $18, $19,
			$20, $21,
			$22, $23,
			$24, $25, $26, $27
		)`,
		e.ID, e.ReportNo, e.ReportDate, e.EffectiveDate,
		e.OwnerUserID, e.AuthorUserID, e.ApplicableContractID,
		e.InstrumentType, e.InstrumentCode, e.InstrumentName, e.Market, e.Currency,
		string(e.Recommendation), e.ReportTitle,
		e.CompanyOverview, e.CompanyOutlook, e.ESGComment, e.FinancialStatus, e.InvestmentAnalysis,
		e.RejectionReason, e.PostSubmissionNote,
		string(e.ReportStatus), string(e.ReviewStatus),
		e.CreatedAt, e.CreatedBy, e.UpdatedAt, e.UpdatedBy,
	)
	if err != nil {
		// Translate the unique-violation race into a typed domain error so
		// the HTTP layer returns 409 instead of 500. Other DB errors keep
		// the generic wrap.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return &domain.ErrResearchReportNoAlreadyExists{ReportNo: e.ReportNo}
		}
		return fmt.Errorf("inserting research report: %w", err)
	}
	return nil
}

// GetByID returns the live report by primary key, alive only.
func (r *PostgresResearchReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ResearchReport, error) {
	row := r.pool.QueryRow(ctx, researchReportSelect+" WHERE id = $1 AND deleted_at IS NULL", id)
	return scanResearchReport(row)
}

// GetByReportNo returns the live report by its business report_no, alive only.
func (r *PostgresResearchReportRepository) GetByReportNo(ctx context.Context, reportNo string) (*entity.ResearchReport, error) {
	row := r.pool.QueryRow(ctx, researchReportSelect+" WHERE report_no = $1 AND deleted_at IS NULL", reportNo)
	return scanResearchReport(row)
}

// List paginates reports with optional filters.
func (r *PostgresResearchReportRepository) List(ctx context.Context, filter domain.ResearchReportListFilter) ([]*entity.ResearchReport, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1

	if filter.ReportStatus != nil {
		conds = append(conds, fmt.Sprintf("report_status = $%d", idx))
		args = append(args, string(*filter.ReportStatus))
		idx++
	}
	if filter.ReviewStatus != nil {
		conds = append(conds, fmt.Sprintf("review_status = $%d", idx))
		args = append(args, string(*filter.ReviewStatus))
		idx++
	}
	if filter.Recommendation != nil {
		conds = append(conds, fmt.Sprintf("recommendation = $%d", idx))
		args = append(args, string(*filter.Recommendation))
		idx++
	}
	if code := strings.TrimSpace(filter.InstrumentCode); code != "" {
		conds = append(conds, fmt.Sprintf("instrument_code = $%d", idx))
		args = append(args, code)
		idx++
	}
	if filter.OwnerUserID != nil {
		conds = append(conds, fmt.Sprintf("owner_user_id = $%d", idx))
		args = append(args, *filter.OwnerUserID)
		idx++
	}
	if filter.ReportDateFrom != nil {
		conds = append(conds, fmt.Sprintf("report_date >= $%d", idx))
		args = append(args, *filter.ReportDateFrom)
		idx++
	}
	if filter.ReportDateTo != nil {
		conds = append(conds, fmt.Sprintf("report_date <= $%d", idx))
		args = append(args, *filter.ReportDateTo)
		idx++
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		conds = append(conds,
			fmt.Sprintf("(report_no ILIKE $%d OR instrument_code ILIKE $%d OR report_title ILIKE $%d)", idx, idx, idx),
		)
		args = append(args, "%"+search+"%")
		idx++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__research_reports "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting research reports: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * limit
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx,
		researchReportSelect+" "+where+
			fmt.Sprintf(" ORDER BY report_date DESC, report_no DESC LIMIT $%d OFFSET $%d", idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing research reports: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.ResearchReport, 0, limit)
	for rows.Next() {
		e, err := scanResearchReport(rows)
		if err != nil {
			return nil, 0, err
		}
		if e != nil {
			out = append(out, e)
		}
	}
	return out, total, rows.Err()
}

// Update overwrites the mutable columns. No optimistic version lock during
// PoC — the lifecycle guards in the domain entity prevent dangerous edits.
func (r *PostgresResearchReportRepository) Update(ctx context.Context, tx pgx.Tx, e *entity.ResearchReport) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__research_reports
		   SET report_date           = $2,
		       effective_date        = $3,
		       owner_user_id         = $4,
		       author_user_id        = $5,
		       applicable_contract_id = $6,
		       instrument_type       = $7,
		       instrument_code       = $8,
		       instrument_name       = $9,
		       market                = $10,
		       currency              = $11,
		       recommendation        = $12,
		       report_title          = $13,
		       company_overview      = $14,
		       company_outlook       = $15,
		       esg_comment           = $16,
		       financial_status      = $17,
		       investment_analysis   = $18,
		       rejection_reason      = $19,
		       post_submission_note  = $20,
		       report_status         = $21,
		       review_status         = $22,
		       updated_at            = $23,
		       updated_by            = $24
		 WHERE id = $1 AND deleted_at IS NULL`,
		e.ID,
		e.ReportDate, e.EffectiveDate,
		e.OwnerUserID, e.AuthorUserID, e.ApplicableContractID,
		e.InstrumentType, e.InstrumentCode, e.InstrumentName, e.Market, e.Currency,
		string(e.Recommendation), e.ReportTitle,
		e.CompanyOverview, e.CompanyOutlook, e.ESGComment, e.FinancialStatus, e.InvestmentAnalysis,
		e.RejectionReason, e.PostSubmissionNote,
		string(e.ReportStatus), string(e.ReviewStatus),
		e.UpdatedAt, e.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("updating research report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrResearchReportNotFound{ReportID: e.ID.String()}
	}
	return nil
}

// Invalidate atomically transitions the report to the INVALIDATED terminal
// state. It updates report_status, review_status, invalidated_at,
// invalidated_by, invalidation_reason, updated_at, updated_by in a single
// statement so the DB CHECK chk_inv_research_invalidation_coherent always
// sees a coherent row.
//
// The WHERE clause refuses to invalidate a row that is already INVALIDATED
// or has been soft-deleted; callers should still pre-check with the entity
// guard CanInvalidate so a typed lifecycle error reaches the API caller
// instead of a generic "not found".
func (r *PostgresResearchReportRepository) Invalidate(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	actorID uuid.UUID,
	reason string,
	at time.Time,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__research_reports
		   SET report_status        = 'INVALIDATED',
		       review_status        = 'INVALIDATED',
		       invalidated_at       = $2,
		       invalidated_by       = $3,
		       invalidation_reason  = $4,
		       updated_at           = $2,
		       updated_by           = $3
		 WHERE id = $1
		   AND deleted_at IS NULL
		   AND report_status <> 'INVALIDATED'
		   AND review_status <> 'INVALIDATED'`,
		id, at, actorID, reason,
	)
	if err != nil {
		return fmt.Errorf("invalidating research report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	return nil
}

// SoftDelete stamps deleted_at and updated_at fields.
func (r *PostgresResearchReportRepository) SoftDelete(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	deletedBy uuid.UUID,
	deletedAt time.Time,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__research_reports
		   SET deleted_at = $2,
		       updated_at = $2,
		       updated_by = $3
		 WHERE id = $1 AND deleted_at IS NULL`,
		id, deletedAt, deletedBy,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting research report: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	return nil
}

func scanResearchReport(s scanner) (*entity.ResearchReport, error) {
	var (
		e          entity.ResearchReport
		recStr     string
		repStatStr string
		revStatStr string
	)
	err := s.Scan(
		&e.ID, &e.ReportNo, &e.ReportDate, &e.EffectiveDate,
		&e.OwnerUserID, &e.AuthorUserID, &e.ApplicableContractID,
		&e.InstrumentType, &e.InstrumentCode, &e.InstrumentName, &e.Market, &e.Currency,
		&recStr, &e.ReportTitle,
		&e.CompanyOverview, &e.CompanyOutlook, &e.ESGComment, &e.FinancialStatus, &e.InvestmentAnalysis,
		&e.RejectionReason, &e.PostSubmissionNote,
		&repStatStr, &revStatStr,
		&e.InvalidatedAt, &e.InvalidatedBy, &e.InvalidationReason,
		&e.CreatedAt, &e.CreatedBy, &e.UpdatedAt, &e.UpdatedBy, &e.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning research report: %w", err)
	}
	e.Recommendation = vo.Recommendation(recStr)
	e.ReportStatus = vo.ReportStatus(repStatStr)
	e.ReviewStatus = vo.ReviewStatus(revStatStr)
	return &e, nil
}

var _ domain.ResearchReportRepository = (*PostgresResearchReportRepository)(nil)
