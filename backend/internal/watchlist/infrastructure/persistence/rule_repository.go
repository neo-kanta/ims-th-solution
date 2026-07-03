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

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
)

// PostgresThresholdRuleRepository implements repository.ThresholdRuleRepository.
type PostgresThresholdRuleRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresThresholdRuleRepository(pool *pgxpool.Pool) *PostgresThresholdRuleRepository {
	return &PostgresThresholdRuleRepository{pool: pool}
}

const ruleSelect = `
	SELECT id, watchlist_item_id, metric_type, direction,
	       threshold_value, currency, cooldown_minutes, status, last_state,
	       last_observed_price, last_observed_at, last_evaluated_at,
	       last_state_changed_at, last_alerted_at, last_quote_stale, last_stale_reason,
	       created_by, updated_by, deleted_by,
	       created_at, updated_at, deleted_at
	FROM watchlist_threshold_rules`

func scanRule(row pgx.Row) (*entity.ThresholdRule, error) {
	var r entity.ThresholdRule
	var metricType, direction, status, lastState string
	var thresholdValue decimal.Decimal
	var lastObservedPrice *decimal.Decimal
	err := row.Scan(
		&r.ID, &r.WatchlistItemID, &metricType, &direction,
		&thresholdValue, &r.Currency, &r.CooldownMinutes, &status, &lastState,
		&lastObservedPrice, &r.LastObservedAt, &r.LastEvaluatedAt,
		&r.LastStateChangedAt, &r.LastAlertedAt, &r.LastQuoteStale, &r.LastStaleReason,
		&r.CreatedBy, &r.UpdatedBy, &r.DeletedBy,
		&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	r.MetricType = entity.MetricType(metricType)
	r.Direction = entity.Direction(direction)
	r.ThresholdValue = thresholdValue
	r.LastObservedPrice = lastObservedPrice
	r.Status = entity.RuleStatus(status)
	r.LastState = entity.RuleState(lastState)
	return &r, nil
}

func (r *PostgresThresholdRuleRepository) Create(ctx context.Context, tx pgx.Tx, rule *entity.ThresholdRule) error {
	now := time.Now().UTC()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	_, err := tx.Exec(ctx, `
		INSERT INTO watchlist_threshold_rules (
			id, watchlist_item_id, metric_type, direction,
			threshold_value, currency, cooldown_minutes, status, last_state,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13
		)`,
		rule.ID, rule.WatchlistItemID, string(rule.MetricType), string(rule.Direction),
		rule.ThresholdValue, rule.Currency, rule.CooldownMinutes, string(rule.Status), string(rule.LastState),
		rule.CreatedBy, rule.UpdatedBy, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateThresholdRule
		}
		return fmt.Errorf("creating threshold rule: %w", err)
	}
	return nil
}

func (r *PostgresThresholdRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ThresholdRule, error) {
	row := r.pool.QueryRow(ctx, ruleSelect+` WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanRule(row)
}

func (r *PostgresThresholdRuleRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ThresholdRule, error) {
	row := tx.QueryRow(ctx, ruleSelect+` WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, id)
	return scanRule(row)
}

func (r *PostgresThresholdRuleRepository) ListByItemID(ctx context.Context, itemID uuid.UUID) ([]*entity.ThresholdRule, error) {
	rows, err := r.pool.Query(ctx,
		ruleSelect+` WHERE watchlist_item_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC`,
		itemID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing rules for item: %w", err)
	}
	defer rows.Close()
	return collectRules(rows)
}

func (r *PostgresThresholdRuleRepository) Update(ctx context.Context, tx pgx.Tx, rule *entity.ThresholdRule) error {
	rule.UpdatedAt = time.Now().UTC()
	_, err := tx.Exec(ctx, `
		UPDATE watchlist_threshold_rules SET
			direction = $2, threshold_value = $3, currency = $4,
			cooldown_minutes = $5, status = $6,
			last_state = $7,
			updated_by = $8, updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL`,
		rule.ID, string(rule.Direction), rule.ThresholdValue, rule.Currency,
		rule.CooldownMinutes, string(rule.Status), string(rule.LastState),
		rule.UpdatedBy, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating threshold rule: %w", err)
	}
	return nil
}

func (r *PostgresThresholdRuleRepository) UpdateState(ctx context.Context, tx pgx.Tx, upd repository.RuleStateUpdate) error {
	upd.LastEvaluatedAt = upd.LastEvaluatedAt.UTC()
	_, err := tx.Exec(ctx, `
		UPDATE watchlist_threshold_rules SET
			last_state = $2,
			last_observed_price = $3, last_observed_at = $4,
			last_evaluated_at = $5, last_state_changed_at = $6,
			last_alerted_at = $7,
			last_quote_stale = $8, last_stale_reason = $9,
			updated_by = $10,
			updated_at = $5
		WHERE id = $1`,
		upd.ID, string(upd.LastState),
		upd.LastObservedPrice, upd.LastObservedAt,
		upd.LastEvaluatedAt, upd.LastStateChangedAt,
		upd.LastAlertedAt,
		upd.LastQuoteStale, upd.LastStaleReason,
		upd.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("updating rule state: %w", err)
	}
	return nil
}

func (r *PostgresThresholdRuleRepository) SoftDisable(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE watchlist_threshold_rules SET
			status = 'DISABLED', deleted_by = $2, deleted_at = $3, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`,
		id, deletedBy, now,
	)
	if err != nil {
		return fmt.Errorf("disabling threshold rule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRuleDisabled
	}
	return nil
}

func (r *PostgresThresholdRuleRepository) ListForEvaluation(ctx context.Context, filter repository.EvaluatorRuleFilter) ([]*entity.ThresholdRule, error) {
	conds := []string{
		"r.deleted_at IS NULL",
		"r.status = 'ENABLED'",
		"i.deleted_at IS NULL",
		"i.status = 'ACTIVE'",
	}
	args := []interface{}{}
	idx := 1

	if filter.ScopeType != nil {
		conds = append(conds, fmt.Sprintf("i.scope_type = $%d", idx))
		args = append(args, string(*filter.ScopeType))
		idx++
	}
	if filter.PortfolioID != nil {
		conds = append(conds, fmt.Sprintf("i.portfolio_id = $%d", idx))
		args = append(args, *filter.PortfolioID)
		idx++
	}
	if filter.SecurityID != nil {
		conds = append(conds, fmt.Sprintf("i.security_id = $%d", idx))
		args = append(args, *filter.SecurityID)
		idx++
	}
	if filter.ItemID != nil {
		conds = append(conds, fmt.Sprintf("r.watchlist_item_id = $%d", idx))
		args = append(args, *filter.ItemID)
		idx++
	}
	if filter.RuleID != nil {
		conds = append(conds, fmt.Sprintf("r.id = $%d", idx))
		args = append(args, *filter.RuleID)
		idx++
	}

	where := "WHERE " + strings.Join(conds, " AND ")
	q := `SELECT r.id, r.watchlist_item_id, r.metric_type, r.direction,
	       r.threshold_value, r.currency, r.cooldown_minutes, r.status, r.last_state,
	       r.last_observed_price, r.last_observed_at, r.last_evaluated_at,
	       r.last_state_changed_at, r.last_alerted_at, r.last_quote_stale, r.last_stale_reason,
	       r.created_by, r.updated_by, r.deleted_by,
	       r.created_at, r.updated_at, r.deleted_at
	FROM watchlist_threshold_rules r
	JOIN watchlist_items i ON i.id = r.watchlist_item_id ` + where + ` ORDER BY r.id ASC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing rules for evaluation: %w", err)
	}
	defer rows.Close()
	return collectRules(rows)
}

func collectRules(rows pgx.Rows) ([]*entity.ThresholdRule, error) {
	var rules []*entity.ThresholdRule
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	return rules, rows.Err()
}
