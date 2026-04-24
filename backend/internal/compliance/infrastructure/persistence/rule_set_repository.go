package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// PostgresRuleSetRepository implements domain.RuleSetRepository.
type PostgresRuleSetRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRuleSetRepository creates a new instance.
func NewPostgresRuleSetRepository(pool *pgxpool.Pool) *PostgresRuleSetRepository {
	return &PostgresRuleSetRepository{pool: pool}
}

// Create inserts a rule set and its members in a single transaction.
func (r *PostgresRuleSetRepository) Create(ctx context.Context, rs *entity.RuleSet) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO compliance_rule_sets (id, name, description, is_active, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, rs.ID, rs.Name, rs.Description, rs.IsActive, rs.CreatedBy, rs.CreatedAt, rs.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting rule set: %w", err)
	}

	for _, m := range rs.Members {
		_, err = tx.Exec(ctx, `
			INSERT INTO compliance_rule_set_members (id, rule_set_id, rule_instance_id, added_by, added_at)
			VALUES (gen_random_uuid(), $1, $2, $3, NOW())
			ON CONFLICT (rule_set_id, rule_instance_id) DO NOTHING
		`, rs.ID, m.RuleInstanceID, rs.CreatedBy)
		if err != nil {
			return fmt.Errorf("inserting rule set member %s: %w", m.RuleInstanceID, err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID returns a rule set with its members.
func (r *PostgresRuleSetRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleSet, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM compliance_rule_sets WHERE id = $1
	`, id)

	var rs entity.RuleSet
	var createdBy *uuid.UUID
	err := row.Scan(&rs.ID, &rs.Name, &rs.Description, &rs.IsActive, &createdBy, &rs.CreatedAt, &rs.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying rule set: %w", err)
	}
	if createdBy != nil {
		rs.CreatedBy = *createdBy
	}

	members, err := r.loadMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	rs.Members = members
	return &rs, nil
}

// List returns rule sets with pagination.
func (r *PostgresRuleSetRepository) List(ctx context.Context, offset, limit int) ([]entity.RuleSet, int64, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_rule_sets`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting rule sets: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM compliance_rule_sets
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing rule sets: %w", err)
	}
	defer rows.Close()

	var results []entity.RuleSet
	for rows.Next() {
		var rs entity.RuleSet
		var createdBy *uuid.UUID
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.Description, &rs.IsActive,
			&createdBy, &rs.CreatedAt, &rs.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning rule set: %w", err)
		}
		if createdBy != nil {
			rs.CreatedBy = *createdBy
		}
		results = append(results, rs)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load members for each set.
	for i := range results {
		members, err := r.loadMembers(ctx, results[i].ID)
		if err != nil {
			return nil, 0, err
		}
		results[i].Members = members
	}
	return results, total, nil
}

func (r *PostgresRuleSetRepository) loadMembers(ctx context.Context, setID uuid.UUID) ([]entity.RuleSetMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rsm.rule_instance_id
		FROM compliance_rule_set_members rsm
		WHERE rsm.rule_set_id = $1
		ORDER BY rsm.added_at ASC
	`, setID)
	if err != nil {
		return nil, fmt.Errorf("loading rule set members: %w", err)
	}
	defer rows.Close()

	var members []entity.RuleSetMember
	for rows.Next() {
		var m entity.RuleSetMember
		if err := rows.Scan(&m.RuleInstanceID); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		// Default severity/priority — overridden by the binding layer.
		m.Severity = vo.SeverityBlock
		members = append(members, m)
	}
	return members, rows.Err()
}

// ensure interface satisfaction at compile time
var _ domain.RuleSetRepository = (*PostgresRuleSetRepository)(nil)
