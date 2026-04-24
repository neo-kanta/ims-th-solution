package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// PostgresCheckRecordRepository implements domain.CheckRecordRepository.
// Append-only: no Update or Delete methods exposed.
type PostgresCheckRecordRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCheckRecordRepository creates a new instance.
func NewPostgresCheckRecordRepository(pool *pgxpool.Pool) *PostgresCheckRecordRepository {
	return &PostgresCheckRecordRepository{pool: pool}
}

// Create inserts a single check record.
func (r *PostgresCheckRecordRepository) Create(ctx context.Context, rec *entity.CheckRecord) error {
	return r.insert(ctx, r.pool, rec)
}

// CreateBatch inserts multiple check records in a single transaction.
func (r *PostgresCheckRecordRepository) CreateBatch(ctx context.Context, records []entity.CheckRecord) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i := range records {
		if err := r.insert(ctx, tx, &records[i]); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type dbExecer interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{ RowsAffected() int64 }, error)
}

// pgExecer wraps both pgxpool.Pool and pgx.Tx so we can reuse insert().
type pgExecer struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func (e *pgExecer) Exec(ctx context.Context, sql string, args ...interface{}) error {
	if e.tx != nil {
		_, err := e.tx.Exec(ctx, sql, args...)
		return err
	}
	_, err := e.pool.Exec(ctx, sql, args...)
	return err
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...interface{}) error
}

type poolExecer struct{ pool *pgxpool.Pool }

func (p *poolExecer) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}

type txExecer struct{ tx pgx.Tx }

func (t *txExecer) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := t.tx.Exec(ctx, sql, args...)
	return err
}

func (r *PostgresCheckRecordRepository) insert(ctx context.Context, db interface{}, rec *entity.CheckRecord) error {
	const query = `
		INSERT INTO compliance_check_records (
			id, check_group_id, timing, order_id, portfolio_id, contract_id, ticker,
			rule_type_id, rule_instance_id, rule_instance_version, parameter_snapshot,
			verdict, effective_severity, final_verdict, evidence, message,
			data_snapshot_hash, eval_duration_ms, checked_by, business_date, checked_at, created_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22
		)
	`
	args := []interface{}{
		rec.ID, rec.CheckGroupID, string(rec.Timing), rec.OrderID,
		rec.PortfolioID, rec.ContractID, rec.Ticker,
		rec.RuleTypeID, rec.RuleInstanceID, rec.RuleInstanceVersion,
		[]byte(rec.ParameterSnapshot),
		string(rec.Verdict), string(rec.EffectiveSeverity), string(rec.FinalVerdict),
		[]byte(rec.Evidence), rec.Message,
		rec.DataSnapshotHash, rec.EvalDurationMs, rec.CheckedBy,
		rec.BusinessDate, rec.CheckedAt, rec.CreatedAt,
	}

	var err error
	switch d := db.(type) {
	case pgx.Tx:
		_, err = d.Exec(ctx, query, args...)
	case *pgxpool.Pool:
		_, err = d.Exec(ctx, query, args...)
	default:
		return fmt.Errorf("unsupported db type %T", db)
	}
	if err != nil {
		return fmt.Errorf("inserting check record %s: %w", rec.ID, err)
	}
	return nil
}

// GetByID returns a check record by primary key.
func (r *PostgresCheckRecordRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CheckRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, check_group_id, timing, order_id, portfolio_id, contract_id, ticker,
		       rule_type_id, rule_instance_id, rule_instance_version, parameter_snapshot,
		       verdict, effective_severity, final_verdict, evidence, message,
		       data_snapshot_hash, eval_duration_ms, checked_by, business_date, checked_at, created_at
		FROM compliance_check_records WHERE id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("querying check record: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("querying check record: %w", err)
		}
		return nil, nil
	}
	rec, err := scanCheckRecord(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning check record: %w", err)
	}
	return rec, nil
}

// GetByGroupID returns all check records for a check group.
func (r *PostgresCheckRecordRepository) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]entity.CheckRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, check_group_id, timing, order_id, portfolio_id, contract_id, ticker,
		       rule_type_id, rule_instance_id, rule_instance_version, parameter_snapshot,
		       verdict, effective_severity, final_verdict, evidence, message,
		       data_snapshot_hash, eval_duration_ms, checked_by, business_date, checked_at, created_at
		FROM compliance_check_records
		WHERE check_group_id = $1
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("querying check records by group: %w", err)
	}
	defer rows.Close()
	return collectCheckRecords(rows)
}

// List returns check records with optional filtering.
func (r *PostgresCheckRecordRepository) List(
	ctx context.Context,
	filter domain.CheckRecordFilter,
) ([]entity.CheckRecord, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.OrderID != nil {
		conditions = append(conditions, fmt.Sprintf("order_id = $%d", argIdx))
		args = append(args, *filter.OrderID)
		argIdx++
	}
	if filter.PortfolioID != nil {
		conditions = append(conditions, fmt.Sprintf("portfolio_id = $%d", argIdx))
		args = append(args, *filter.PortfolioID)
		argIdx++
	}
	if filter.ContractID != nil {
		conditions = append(conditions, fmt.Sprintf("contract_id = $%d", argIdx))
		args = append(args, *filter.ContractID)
		argIdx++
	}
	if filter.Ticker != "" {
		conditions = append(conditions, fmt.Sprintf("ticker = $%d", argIdx))
		args = append(args, filter.Ticker)
		argIdx++
	}
	if filter.Timing != nil {
		conditions = append(conditions, fmt.Sprintf("timing = $%d", argIdx))
		args = append(args, string(*filter.Timing))
		argIdx++
	}
	if filter.Verdict != nil {
		conditions = append(conditions, fmt.Sprintf("final_verdict = $%d", argIdx))
		args = append(args, string(*filter.Verdict))
		argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("business_date >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("business_date <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	var total int64
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM compliance_check_records WHERE %s`, where),
		args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting check records: %w", err)
	}

	dataArgs := append(args, limit, filter.Offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, check_group_id, timing, order_id, portfolio_id, contract_id, ticker,
		       rule_type_id, rule_instance_id, rule_instance_version, parameter_snapshot,
		       verdict, effective_severity, final_verdict, evidence, message,
		       data_snapshot_hash, eval_duration_ms, checked_by, business_date, checked_at, created_at
		FROM compliance_check_records
		WHERE %s
		ORDER BY checked_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing check records: %w", err)
	}
	defer rows.Close()

	records, err := collectCheckRecords(rows)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// --- row scanners ---

type rowsScanner interface {
	Scan(dest ...interface{}) error
}

func scanCheckRecord(s rowsScanner) (*entity.CheckRecord, error) {
	var rec entity.CheckRecord
	var timingStr, verdictStr, severityStr, finalVerdictStr string
	var paramRaw, evidenceRaw []byte

	err := s.Scan(
		&rec.ID, &rec.CheckGroupID, &timingStr, &rec.OrderID,
		&rec.PortfolioID, &rec.ContractID, &rec.Ticker,
		&rec.RuleTypeID, &rec.RuleInstanceID, &rec.RuleInstanceVersion,
		&paramRaw,
		&verdictStr, &severityStr, &finalVerdictStr,
		&evidenceRaw, &rec.Message,
		&rec.DataSnapshotHash, &rec.EvalDurationMs, &rec.CheckedBy,
		&rec.BusinessDate, &rec.CheckedAt, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	rec.Timing = vo.CheckTiming(timingStr)
	rec.Verdict = vo.Verdict(verdictStr)
	rec.EffectiveSeverity = vo.Severity(severityStr)
	rec.FinalVerdict = vo.Verdict(finalVerdictStr)
	rec.ParameterSnapshot = json.RawMessage(paramRaw)
	rec.Evidence = json.RawMessage(evidenceRaw)
	rec.BusinessDate = rec.BusinessDate.UTC()
	rec.CheckedAt = rec.CheckedAt.UTC()
	rec.CreatedAt = rec.CreatedAt.UTC()
	return &rec, nil
}

func collectCheckRecords(rows pgx.Rows) ([]entity.CheckRecord, error) {
	var results []entity.CheckRecord
	for rows.Next() {
		rec, err := scanCheckRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning check record: %w", err)
		}
		results = append(results, *rec)
	}
	return results, rows.Err()
}

// ensure interface satisfaction at compile time
var _ domain.CheckRecordRepository = (*PostgresCheckRecordRepository)(nil)
