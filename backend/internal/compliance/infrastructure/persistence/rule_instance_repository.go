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
)

// PostgresRuleInstanceRepository implements domain.RuleInstanceRepository.
type PostgresRuleInstanceRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRuleInstanceRepository creates a new instance.
func NewPostgresRuleInstanceRepository(pool *pgxpool.Pool) *PostgresRuleInstanceRepository {
	return &PostgresRuleInstanceRepository{pool: pool}
}

// Create inserts a new rule instance.
func (r *PostgresRuleInstanceRepository) Create(ctx context.Context, inst *entity.RuleInstance) error {
	query := `
		INSERT INTO compliance_rule_instances
			(id, rule_type_id, name, description, current_version, is_active,
			 effective_from, effective_to, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		inst.ID, inst.RuleTypeID, inst.Name, inst.Description, inst.CurrentVersion,
		inst.IsActive, inst.EffectiveWindow.ValidFrom, inst.EffectiveWindow.ValidTo,
		inst.CreatedBy, inst.CreatedAt, inst.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating rule instance: %w", err)
	}
	return nil
}

// GetByID returns a rule instance by primary key.
func (r *PostgresRuleInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleInstance, error) {
	query := `
		SELECT id, rule_type_id, name, description, current_version, is_active,
		       effective_from, effective_to, created_by, created_at, updated_at
		FROM compliance_rule_instances
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	inst, err := scanRuleInstance(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying rule instance: %w", err)
	}
	return inst, nil
}

// List returns rule instances with optional filtering.
func (r *PostgresRuleInstanceRepository) List(
	ctx context.Context,
	filter domain.RuleInstanceFilter,
) ([]entity.RuleInstance, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.RuleTypeID != nil {
		conditions = append(conditions, fmt.Sprintf("rule_type_id = $%d", argIdx))
		args = append(args, *filter.RuleTypeID)
		argIdx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *filter.IsActive)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM compliance_rule_instances WHERE %s`, where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting rule instances: %w", err)
	}

	dataArgs := append(args, limit, filter.Offset)
	listQuery := fmt.Sprintf(`
		SELECT id, rule_type_id, name, description, current_version, is_active,
		       effective_from, effective_to, created_by, created_at, updated_at
		FROM compliance_rule_instances
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, listQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing rule instances: %w", err)
	}
	defer rows.Close()

	var results []entity.RuleInstance
	for rows.Next() {
		inst, err := scanRuleInstance(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning rule instance: %w", err)
		}
		results = append(results, *inst)
	}
	return results, total, rows.Err()
}

// UpdateActive toggles the is_active flag.
func (r *PostgresRuleInstanceRepository) UpdateActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE compliance_rule_instances SET is_active = $2, updated_at = NOW() WHERE id = $1`,
		id, isActive,
	)
	if err != nil {
		return fmt.Errorf("updating rule instance active: %w", err)
	}
	return nil
}

// CreateVersion inserts an immutable version record and bumps current_version on the instance.
func (r *PostgresRuleInstanceRepository) CreateVersion(ctx context.Context, v *entity.RuleInstanceVersion) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO compliance_rule_instance_versions
			(id, rule_instance_id, version_number, parameters, change_note, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, v.ID, v.RuleInstanceID, v.VersionNumber, []byte(v.Parameters), v.ChangeReason, v.CreatedBy, v.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting rule version: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE compliance_rule_instances
		SET current_version = $2, updated_at = NOW()
		WHERE id = $1
	`, v.RuleInstanceID, v.VersionNumber)
	if err != nil {
		return fmt.Errorf("updating current version: %w", err)
	}

	return tx.Commit(ctx)
}

// GetCurrentVersion returns the active parameter version for an instance.
func (r *PostgresRuleInstanceRepository) GetCurrentVersion(
	ctx context.Context, instanceID uuid.UUID,
) (*entity.RuleInstanceVersion, error) {
	query := `
		SELECT riv.id, riv.rule_instance_id, riv.version_number, riv.parameters,
		       riv.change_note, riv.created_by, riv.created_at
		FROM compliance_rule_instance_versions riv
		JOIN compliance_rule_instances ri ON ri.id = riv.rule_instance_id
		WHERE riv.rule_instance_id = $1 AND riv.version_number = ri.current_version
	`
	row := r.pool.QueryRow(ctx, query, instanceID)
	v, err := scanVersion(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting current version: %w", err)
	}
	return v, nil
}

// GetCurrentVersions batch-loads the current version for many instances in a
// single query. Instances without a current version are absent from the map.
func (r *PostgresRuleInstanceRepository) GetCurrentVersions(
	ctx context.Context, instanceIDs []uuid.UUID,
) (map[uuid.UUID]*entity.RuleInstanceVersion, error) {
	out := make(map[uuid.UUID]*entity.RuleInstanceVersion, len(instanceIDs))
	if len(instanceIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT riv.id, riv.rule_instance_id, riv.version_number, riv.parameters,
		       riv.change_note, riv.created_by, riv.created_at
		FROM compliance_rule_instance_versions riv
		JOIN compliance_rule_instances ri ON ri.id = riv.rule_instance_id
		WHERE riv.rule_instance_id = ANY($1) AND riv.version_number = ri.current_version
	`, instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("getting current versions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning current version: %w", err)
		}
		out[v.RuleInstanceID] = v
	}
	return out, rows.Err()
}

// ListVersions returns all versions for an instance in ascending order.
func (r *PostgresRuleInstanceRepository) ListVersions(
	ctx context.Context, instanceID uuid.UUID,
) ([]entity.RuleInstanceVersion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, rule_instance_id, version_number, parameters, change_note, created_by, created_at
		FROM compliance_rule_instance_versions
		WHERE rule_instance_id = $1
		ORDER BY version_number ASC
	`, instanceID)
	if err != nil {
		return nil, fmt.Errorf("listing versions: %w", err)
	}
	defer rows.Close()

	var results []entity.RuleInstanceVersion
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning version: %w", err)
		}
		results = append(results, *v)
	}
	return results, rows.Err()
}

// --- row scanners ---

type ruleInstanceScanner interface {
	Scan(dest ...interface{}) error
}

func scanRuleInstance(s ruleInstanceScanner) (*entity.RuleInstance, error) {
	var inst entity.RuleInstance
	var validTo *time.Time
	var createdBy *uuid.UUID
	err := s.Scan(
		&inst.ID, &inst.RuleTypeID, &inst.Name, &inst.Description,
		&inst.CurrentVersion, &inst.IsActive,
		&inst.EffectiveWindow.ValidFrom, &validTo,
		&createdBy, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	inst.EffectiveWindow.ValidTo = validTo
	if createdBy != nil {
		inst.CreatedBy = *createdBy
	}
	// Normalise to UTC
	inst.EffectiveWindow.ValidFrom = inst.EffectiveWindow.ValidFrom.UTC()
	if inst.EffectiveWindow.ValidTo != nil {
		t := inst.EffectiveWindow.ValidTo.UTC()
		inst.EffectiveWindow.ValidTo = &t
	}
	return &inst, nil
}

func scanVersion(s ruleInstanceScanner) (*entity.RuleInstanceVersion, error) {
	var v entity.RuleInstanceVersion
	var paramsRaw []byte
	var createdBy *uuid.UUID
	err := s.Scan(
		&v.ID, &v.RuleInstanceID, &v.VersionNumber,
		&paramsRaw, &v.ChangeReason, &createdBy, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	v.Parameters = json.RawMessage(paramsRaw)
	if createdBy != nil {
		v.CreatedBy = *createdBy
	}
	v.CreatedAt = v.CreatedAt.UTC()
	return &v, nil
}

// ensure interface satisfaction at compile time
var _ domain.RuleInstanceRepository = (*PostgresRuleInstanceRepository)(nil)
