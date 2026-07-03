package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
)

// PostgresAlertEventRepository implements repository.AlertEventRepository.
type PostgresAlertEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAlertEventRepository(pool *pgxpool.Pool) *PostgresAlertEventRepository {
	return &PostgresAlertEventRepository{pool: pool}
}

const alertSelect = `
	SELECT id, watchlist_item_id, threshold_rule_id,
	       scope_type, owner_user_id, portfolio_id, security_id, created_by_user_id,
	       direction, previous_state, current_state,
	       observed_price, threshold_value, currency, quote_provider,
	       observed_at, evaluated_at,
	       stale, stale_reason,
	       idempotency_key, notification_status, notification_id, notification_error,
	       acknowledged_by, acknowledged_at, acknowledgement_note,
	       created_at
	FROM watchlist_alert_events`

func scanAlert(row pgx.Row) (*entity.AlertEvent, error) {
	var a entity.AlertEvent
	var scopeType, direction, prevState, curState, notifStatus string
	var observedPrice, thresholdValue decimal.Decimal
	err := row.Scan(
		&a.ID, &a.WatchlistItemID, &a.ThresholdRuleID,
		&scopeType, &a.OwnerUserID, &a.PortfolioID, &a.SecurityID, &a.CreatedByUserID,
		&direction, &prevState, &curState,
		&observedPrice, &thresholdValue, &a.Currency, &a.QuoteProvider,
		&a.ObservedAt, &a.EvaluatedAt,
		&a.Stale, &a.StaleReason,
		&a.IdempotencyKey, &notifStatus, &a.NotificationID, &a.NotificationError,
		&a.AcknowledgedBy, &a.AcknowledgedAt, &a.AcknowledgementNote,
		&a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	a.ScopeType = entity.ScopeType(scopeType)
	a.Direction = entity.Direction(direction)
	a.PreviousState = entity.RuleState(prevState)
	a.CurrentState = entity.RuleState(curState)
	a.ObservedPrice = observedPrice
	a.ThresholdValue = thresholdValue
	a.NotificationStatus = entity.NotificationStatus(notifStatus)
	return &a, nil
}

func (r *PostgresAlertEventRepository) Insert(ctx context.Context, tx pgx.Tx, event *entity.AlertEvent) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO watchlist_alert_events (
			id, watchlist_item_id, threshold_rule_id,
			scope_type, owner_user_id, portfolio_id, security_id, created_by_user_id,
			direction, previous_state, current_state,
			observed_price, threshold_value, currency, quote_provider,
			observed_at, evaluated_at,
			stale, stale_reason,
			idempotency_key, notification_status,
			created_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15,
			$16, $17,
			$18, $19,
			$20, $21,
			$22
		)`,
		event.ID, event.WatchlistItemID, event.ThresholdRuleID,
		string(event.ScopeType), event.OwnerUserID, event.PortfolioID, event.SecurityID, event.CreatedByUserID,
		string(event.Direction), string(event.PreviousState), string(event.CurrentState),
		event.ObservedPrice, event.ThresholdValue, event.Currency, event.QuoteProvider,
		event.ObservedAt, event.EvaluatedAt,
		event.Stale, event.StaleReason,
		event.IdempotencyKey, string(event.NotificationStatus),
		event.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlertIdempotencyConflict
		}
		return fmt.Errorf("inserting alert event: %w", err)
	}
	return nil
}

func (r *PostgresAlertEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AlertEvent, error) {
	row := r.pool.QueryRow(ctx, alertSelect+` WHERE id = $1`, id)
	return scanAlert(row)
}

func (r *PostgresAlertEventRepository) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status entity.NotificationStatus, notifErr *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE watchlist_alert_events SET notification_status = $2, notification_error = $3
		WHERE id = $1`,
		id, string(status), notifErr,
	)
	return err
}

func (r *PostgresAlertEventRepository) Acknowledge(ctx context.Context, tx pgx.Tx, id uuid.UUID, by uuid.UUID, note *string) error {
	tag, err := tx.Exec(ctx, `
		UPDATE watchlist_alert_events SET
			acknowledged_by = $2, acknowledged_at = NOW(), acknowledgement_note = $3
		WHERE id = $1 AND acknowledged_at IS NULL`,
		id, by, note,
	)
	if err != nil {
		return fmt.Errorf("acknowledging alert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either not found or already acknowledged; distinguish by re-reading.
		existing, err2 := r.GetByID(ctx, id)
		if err2 != nil {
			return domain.ErrAlertAlreadyAcknowledged
		}
		if existing == nil {
			return domain.ErrAlertNotFound
		}
		return domain.ErrAlertAlreadyAcknowledged
	}
	return nil
}

func (r *PostgresAlertEventRepository) List(ctx context.Context, filter repository.AlertEventFilter) ([]*entity.AlertEvent, int, error) {
	conds := []string{}
	args := []interface{}{}
	idx := 1

	if filter.UnionPersonalOwnerID != nil {
		// Unscoped union: personal alerts owned by actor OR portfolio alerts accessible via fund IDs.
		if filter.UnionPortfolioAll {
			conds = append(conds, fmt.Sprintf(
				"(scope_type = 'PERSONAL' AND owner_user_id = $%d OR scope_type = 'PORTFOLIO')", idx,
			))
			args = append(args, *filter.UnionPersonalOwnerID)
			idx++
		} else if len(filter.FundIDs) > 0 {
			conds = append(conds, fmt.Sprintf(
				"(scope_type = 'PERSONAL' AND owner_user_id = $%d OR scope_type = 'PORTFOLIO' AND EXISTS (SELECT 1 FROM investment__portfolios p WHERE p.id = watchlist_alert_events.portfolio_id AND p.fund_id = ANY($%d)))",
				idx, idx+1,
			))
			args = append(args, *filter.UnionPersonalOwnerID, filter.FundIDs)
			idx += 2
		} else {
			conds = append(conds, fmt.Sprintf("scope_type = 'PERSONAL' AND owner_user_id = $%d", idx))
			args = append(args, *filter.UnionPersonalOwnerID)
			idx++
		}
	} else {
		if filter.OwnerUserID != nil {
			conds = append(conds, fmt.Sprintf("owner_user_id = $%d", idx))
			args = append(args, *filter.OwnerUserID)
			idx++
		}
		if len(filter.PortfolioIDs) > 0 {
			conds = append(conds, fmt.Sprintf("portfolio_id = ANY($%d)", idx))
			args = append(args, filter.PortfolioIDs)
			idx++
		}
		if len(filter.FundIDs) > 0 {
			conds = append(conds, fmt.Sprintf(
				"EXISTS (SELECT 1 FROM investment__portfolios p WHERE p.id = watchlist_alert_events.portfolio_id AND p.fund_id = ANY($%d))", idx,
			))
			args = append(args, filter.FundIDs)
			idx++
		}
		if filter.ScopeType != nil {
			conds = append(conds, fmt.Sprintf("scope_type = $%d", idx))
			args = append(args, string(*filter.ScopeType))
			idx++
		}
	}
	if filter.SecurityID != nil {
		conds = append(conds, fmt.Sprintf("security_id = $%d", idx))
		args = append(args, *filter.SecurityID)
		idx++
	}
	if filter.RuleID != nil {
		conds = append(conds, fmt.Sprintf("threshold_rule_id = $%d", idx))
		args = append(args, *filter.RuleID)
		idx++
	}
	if filter.Acknowledged != nil {
		if *filter.Acknowledged {
			conds = append(conds, "acknowledged_at IS NOT NULL")
		} else {
			conds = append(conds, "acknowledged_at IS NULL")
		}
	}
	if filter.CreatedFrom != nil {
		conds = append(conds, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *filter.CreatedFrom)
		idx++
	}
	if filter.CreatedTo != nil {
		conds = append(conds, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *filter.CreatedTo)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM watchlist_alert_events "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting alert events: %w", err)
	}
	if total == 0 {
		return []*entity.AlertEvent{}, 0, nil
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	q := alertSelect + " " + where +
		" ORDER BY created_at DESC, id DESC" +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing alert events: %w", err)
	}
	defer rows.Close()

	var alerts []*entity.AlertEvent
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, 0, err
		}
		if alert != nil {
			alerts = append(alerts, alert)
		}
	}
	return alerts, total, rows.Err()
}
