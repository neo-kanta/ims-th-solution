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

// PostgresRuleBindingRepository implements domain.RuleBindingRepository.
type PostgresRuleBindingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRuleBindingRepository creates a new instance.
func NewPostgresRuleBindingRepository(pool *pgxpool.Pool) *PostgresRuleBindingRepository {
	return &PostgresRuleBindingRepository{pool: pool}
}

// Create inserts a new rule binding.
func (r *PostgresRuleBindingRepository) Create(ctx context.Context, b *entity.RuleBinding) error {
	query := `
		INSERT INTO compliance_rule_bindings
			(id, rule_instance_id, rule_set_id, scope_type, scope_id,
			 severity, priority, effective_from, effective_to,
			 is_active, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`
	_, err := r.pool.Exec(ctx, query,
		b.ID, b.RuleInstanceID, b.RuleSetID,
		string(b.Scope.Type), b.Scope.ID,
		string(b.Severity), b.Priority,
		b.EffectiveWindow.ValidFrom, b.EffectiveWindow.ValidTo,
		b.IsActive, b.CreatedBy, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating rule binding: %w", err)
	}
	return nil
}

// GetByID returns a single binding.
func (r *PostgresRuleBindingRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleBinding, error) {
	query := `
		SELECT id, rule_instance_id, rule_set_id, scope_type, scope_id,
		       severity, priority, effective_from, effective_to,
		       is_active, created_by, created_at, updated_at
		FROM compliance_rule_bindings
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	b, err := scanBinding(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying binding: %w", err)
	}
	return b, nil
}

// List returns bindings with optional filtering.
func (r *PostgresRuleBindingRepository) List(
	ctx context.Context,
	filter domain.BindingFilter,
) ([]entity.RuleBinding, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.ScopeType != nil {
		conditions = append(conditions, fmt.Sprintf("scope_type = $%d", argIdx))
		args = append(args, string(*filter.ScopeType))
		argIdx++
	}
	if filter.ScopeID != nil {
		conditions = append(conditions, fmt.Sprintf("scope_id = $%d", argIdx))
		args = append(args, *filter.ScopeID)
		argIdx++
	}
	if filter.RuleInstanceID != nil {
		conditions = append(conditions, fmt.Sprintf("rule_instance_id = $%d", argIdx))
		args = append(args, *filter.RuleInstanceID)
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

	var total int64
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM compliance_rule_bindings WHERE %s`, where),
		args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting bindings: %w", err)
	}

	dataArgs := append(args, limit, filter.Offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, rule_instance_id, rule_set_id, scope_type, scope_id,
		       severity, priority, effective_from, effective_to,
		       is_active, created_by, created_at, updated_at
		FROM compliance_rule_bindings
		WHERE %s
		ORDER BY priority ASC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing bindings: %w", err)
	}
	defer rows.Close()

	var results []entity.RuleBinding
	for rows.Next() {
		b, err := scanBinding(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning binding: %w", err)
		}
		results = append(results, *b)
	}
	return results, total, rows.Err()
}

// Deactivate sets is_active = false for a binding.
func (r *PostgresRuleBindingRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE compliance_rule_bindings SET is_active = false, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deactivating binding: %w", err)
	}
	return nil
}

// ResolveApplicable is the hot-path query: finds all active, effective bindings
// applicable to any of the given scopes on the business date, then joins to the
// rule instance and its current version to produce fully resolved records.
//
// Scopes are evaluated with the most-specific-scope-wins logic (done in the engine,
// not here). This query simply returns ALL matching bindings; the engine sorts.
func (r *PostgresRuleBindingRepository) ResolveApplicable(
	ctx context.Context,
	scopes []vo.Scope,
	date time.Time,
) ([]domain.ResolvedBinding, error) {
	if len(scopes) == 0 {
		return nil, nil
	}

	// Build a dynamic OR clause for each scope.
	// GLOBAL scopes match where scope_id IS NULL.
	// Specific scopes match on (scope_type, scope_id) pair.
	scopeConditions := []string{}
	args := []interface{}{date, date, date, date}
	argIdx := 5

	for _, s := range scopes {
		if s.Type == vo.ScopeGlobal || s.ID == nil {
			scopeConditions = append(scopeConditions,
				fmt.Sprintf("(rb.scope_type = $%d AND rb.scope_id IS NULL)", argIdx))
			args = append(args, string(s.Type))
			argIdx++
		} else {
			scopeConditions = append(scopeConditions,
				fmt.Sprintf("(rb.scope_type = $%d AND rb.scope_id = $%d)", argIdx, argIdx+1))
			args = append(args, string(s.Type), *s.ID)
			argIdx += 2
		}
	}

	scopeWhere := strings.Join(scopeConditions, " OR ")

	query := fmt.Sprintf(`
		SELECT
			-- binding columns
			rb.id, rb.rule_instance_id, rb.rule_set_id, rb.scope_type, rb.scope_id,
			rb.severity, rb.priority, rb.effective_from, rb.effective_to,
			rb.is_active, rb.created_by, rb.created_at, rb.updated_at,
			-- rule instance columns
			ri.id, ri.rule_type_id, ri.name, ri.description, ri.current_version,
			ri.is_active, ri.effective_from, ri.effective_to, ri.created_by,
			ri.created_at, ri.updated_at,
			-- version columns
			riv.id, riv.rule_instance_id, riv.version_number, riv.parameters,
			riv.change_note, riv.created_by, riv.created_at
		FROM compliance_rule_bindings rb
		JOIN compliance_rule_instances ri
			ON rb.rule_instance_id = ri.id
		JOIN compliance_rule_instance_versions riv
			ON riv.rule_instance_id = ri.id
			AND riv.version_number = ri.current_version
		WHERE rb.is_active = true
		  AND ri.is_active = true
		  -- binding effective window
		  AND rb.effective_from  <= $1
		  AND (rb.effective_to   IS NULL OR rb.effective_to   >= $2)
		  -- instance effective window
		  AND ri.effective_from  <= $3
		  AND (ri.effective_to   IS NULL OR ri.effective_to   >= $4)
		  -- scope match
		  AND (%s)
	`, scopeWhere)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resolving applicable bindings: %w", err)
	}
	defer rows.Close()

	var results []domain.ResolvedBinding
	for rows.Next() {
		rb, err := scanResolvedBinding(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning resolved binding: %w", err)
		}
		results = append(results, rb)
	}
	return results, rows.Err()
}

// --- row scanners ---

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanBinding(s scanner) (*entity.RuleBinding, error) {
	var b entity.RuleBinding
	var scopeType string
	var severityStr string
	var validTo *time.Time
	var createdBy *uuid.UUID

	err := s.Scan(
		&b.ID, &b.RuleInstanceID, &b.RuleSetID,
		&scopeType, &b.Scope.ID,
		&severityStr, &b.Priority,
		&b.EffectiveWindow.ValidFrom, &validTo,
		&b.IsActive, &createdBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Scope.Type = vo.ScopeType(scopeType)
	b.Severity = vo.Severity(severityStr)
	b.EffectiveWindow.ValidTo = validTo
	if createdBy != nil {
		b.CreatedBy = *createdBy
	}
	b.EffectiveWindow.ValidFrom = b.EffectiveWindow.ValidFrom.UTC()
	if b.EffectiveWindow.ValidTo != nil {
		t := b.EffectiveWindow.ValidTo.UTC()
		b.EffectiveWindow.ValidTo = &t
	}
	return &b, nil
}

func scanResolvedBinding(s scanner) (domain.ResolvedBinding, error) {
	var rb domain.ResolvedBinding
	b := &rb.Binding
	ri := &rb.RuleInstance
	riv := &rb.CurrentVersion

	var bScopeType, bSeverity string
	var bValidTo *time.Time
	var bCreatedBy *uuid.UUID

	var riEffectiveTo *time.Time
	var riCreatedBy *uuid.UUID

	var rivParamsRaw []byte
	var rivCreatedBy *uuid.UUID

	err := s.Scan(
		// binding
		&b.ID, &b.RuleInstanceID, &b.RuleSetID,
		&bScopeType, &b.Scope.ID,
		&bSeverity, &b.Priority,
		&b.EffectiveWindow.ValidFrom, &bValidTo,
		&b.IsActive, &bCreatedBy, &b.CreatedAt, &b.UpdatedAt,
		// rule instance
		&ri.ID, &ri.RuleTypeID, &ri.Name, &ri.Description,
		&ri.CurrentVersion, &ri.IsActive,
		&ri.EffectiveWindow.ValidFrom, &riEffectiveTo,
		&riCreatedBy, &ri.CreatedAt, &ri.UpdatedAt,
		// version
		&riv.ID, &riv.RuleInstanceID, &riv.VersionNumber,
		&rivParamsRaw, &riv.ChangeReason, &rivCreatedBy, &riv.CreatedAt,
	)
	if err != nil {
		return domain.ResolvedBinding{}, err
	}

	b.Scope.Type = vo.ScopeType(bScopeType)
	b.Severity = vo.Severity(bSeverity)
	b.EffectiveWindow.ValidTo = bValidTo
	if bCreatedBy != nil {
		b.CreatedBy = *bCreatedBy
	}
	b.EffectiveWindow.ValidFrom = b.EffectiveWindow.ValidFrom.UTC()
	if b.EffectiveWindow.ValidTo != nil {
		t := b.EffectiveWindow.ValidTo.UTC()
		b.EffectiveWindow.ValidTo = &t
	}

	ri.EffectiveWindow.ValidTo = riEffectiveTo
	if riCreatedBy != nil {
		ri.CreatedBy = *riCreatedBy
	}
	ri.EffectiveWindow.ValidFrom = ri.EffectiveWindow.ValidFrom.UTC()

	riv.Parameters = json.RawMessage(rivParamsRaw)
	if rivCreatedBy != nil {
		riv.CreatedBy = *rivCreatedBy
	}
	riv.CreatedAt = riv.CreatedAt.UTC()

	return rb, nil
}

// ensure interface satisfaction at compile time
var _ domain.RuleBindingRepository = (*PostgresRuleBindingRepository)(nil)
