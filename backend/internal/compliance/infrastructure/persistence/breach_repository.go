package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// PostgresBreachRepository implements domain.BreachRepository.
type PostgresBreachRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresBreachRepository creates a new instance.
func NewPostgresBreachRepository(pool *pgxpool.Pool) *PostgresBreachRepository {
	return &PostgresBreachRepository{pool: pool}
}

// Create inserts a new breach record.
func (r *PostgresBreachRepository) Create(ctx context.Context, b *entity.Breach) error {
	evidenceRaw, err := json.Marshal(b.Evidence)
	if err != nil {
		return fmt.Errorf("marshalling breach evidence: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO compliance_breaches (
			id, check_record_id, check_group_id, portfolio_id, contract_id,
			rule_type_id, rule_instance_id, severity, verdict, status,
			evidence, message, business_date, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`,
		b.ID, b.CheckRecordID, b.CheckGroupID, b.PortfolioID, b.ContractID,
		b.RuleTypeID, b.RuleInstanceID,
		string(b.Severity), string(b.Verdict), string(b.Status),
		evidenceRaw, b.Message, b.BusinessDate, b.CreatedAt, b.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating breach: %w", err)
	}
	return nil
}

// GetByID returns a breach by primary key.
func (r *PostgresBreachRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Breach, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, check_record_id, check_group_id, portfolio_id, contract_id,
		       rule_type_id, rule_instance_id, severity, verdict, status,
		       evidence, message, business_date, created_at, resolved_at, resolved_by
		FROM compliance_breaches
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("querying breach: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return scanBreach(rows)
}

// List returns breaches with optional filtering.
func (r *PostgresBreachRepository) List(
	ctx context.Context,
	filter domain.BreachFilter,
) ([]entity.Breach, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

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
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.RuleTypeID != "" {
		conditions = append(conditions, fmt.Sprintf("rule_type_id = $%d", argIdx))
		args = append(args, filter.RuleTypeID)
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
		fmt.Sprintf(`SELECT COUNT(*) FROM compliance_breaches WHERE %s`, where),
		args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting breaches: %w", err)
	}

	dataArgs := append(args, limit, filter.Offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, check_record_id, check_group_id, portfolio_id, contract_id,
		       rule_type_id, rule_instance_id, severity, verdict, status,
		       evidence, message, business_date, created_at, resolved_at, resolved_by
		FROM compliance_breaches
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing breaches: %w", err)
	}
	defer rows.Close()

	var results []entity.Breach
	for rows.Next() {
		b, err := scanBreach(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning breach: %w", err)
		}
		results = append(results, *b)
	}
	return results, total, rows.Err()
}

// UpdateStatus transitions a breach to a new status (e.g. RESOLVED or OVERRIDDEN).
func (r *PostgresBreachRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status entity.BreachStatus,
	resolvedBy *uuid.UUID,
	resolvedAt *time.Time,
) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE compliance_breaches
		SET status = $2, resolved_by = $3, resolved_at = $4, updated_at = NOW()
		WHERE id = $1
	`, id, string(status), resolvedBy, resolvedAt)
	if err != nil {
		return fmt.Errorf("updating breach status: %w", err)
	}
	return nil
}

// --- row scanner ---

func scanBreach(s pgx.Rows) (*entity.Breach, error) {
	var b entity.Breach
	var severityStr, verdictStr, statusStr string
	var evidenceRaw []byte

	err := s.Scan(
		&b.ID, &b.CheckRecordID, &b.CheckGroupID,
		&b.PortfolioID, &b.ContractID,
		&b.RuleTypeID, &b.RuleInstanceID,
		&severityStr, &verdictStr, &statusStr,
		&evidenceRaw, &b.Message,
		&b.BusinessDate, &b.CreatedAt, &b.ResolvedAt, &b.ResolvedBy,
	)
	if err != nil {
		return nil, err
	}
	b.Severity = vo.Severity(severityStr)
	b.Verdict = vo.Verdict(verdictStr)
	b.Status = entity.BreachStatus(statusStr)
	if err := json.Unmarshal(evidenceRaw, &b.Evidence); err != nil {
		// Don't fail on evidence parse errors — store as-is
		b.Evidence = vo.Evidence{}
	}
	b.BusinessDate = b.BusinessDate.UTC()
	b.CreatedAt = b.CreatedAt.UTC()
	if b.ResolvedAt != nil {
		t := b.ResolvedAt.UTC()
		b.ResolvedAt = &t
	}
	return &b, nil
}

// ensure interface satisfaction at compile time
var _ domain.BreachRepository = (*PostgresBreachRepository)(nil)
