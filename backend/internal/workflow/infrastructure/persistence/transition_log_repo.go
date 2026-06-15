package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// PostgresTransitionLogRepository implements domain.TransitionLogRepository.
// This table is append-only: no UPDATE or DELETE methods exist, by design.
type PostgresTransitionLogRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTransitionLogRepository creates the repository.
func NewPostgresTransitionLogRepository(pool *pgxpool.Pool) *PostgresTransitionLogRepository {
	return &PostgresTransitionLogRepository{pool: pool}
}

// Append inserts one immutable transition record within an existing transaction.
func (r *PostgresTransitionLogRepository) Append(
	ctx context.Context,
	tx pgx.Tx,
	t *entity.WorkflowTransition,
) error {
	metaBytes, err := json.Marshal(t.Metadata)
	if err != nil {
		return fmt.Errorf("marshalling transition metadata: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workflow__transition_log (
			id, workflow_day_id, contract_id, business_date,
			from_state, to_state, action,
			actor_id, actor_type, actor_username, actor_account_code, is_admin_override,
			reason, metadata, occurred_at, request_id
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16
		)`,
		t.ID, t.WorkflowDayID, t.ContractID, t.BusinessDate,
		string(t.FromState), string(t.ToState), string(t.Action),
		t.ActorID, string(t.ActorType), t.ActorUsername, t.ActorAccountCode, t.IsAdminOverride,
		t.Reason, metaBytes, t.OccurredAt, t.RequestID,
	)
	if err != nil {
		return fmt.Errorf("appending transition log: %w", err)
	}
	return nil
}

// ListByBusinessDate returns all transitions for a business date in chronological order.
func (r *PostgresTransitionLogRepository) ListByBusinessDate(
	ctx context.Context,
	businessDate time.Time,
) ([]*entity.WorkflowTransition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			from_state, to_state, action,
			actor_id, actor_type, actor_username, actor_account_code, is_admin_override,
			reason, metadata, occurred_at, request_id
		FROM workflow__transition_log
		WHERE business_date = $1
		ORDER BY occurred_at ASC`,
		businessDate,
	)
	if err != nil {
		return nil, fmt.Errorf("querying transition log: %w", err)
	}
	defer rows.Close()

	var results []*entity.WorkflowTransition
	for rows.Next() {
		t, err := scanTransition(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// ListByContractDate returns all transitions for a contract+date in chronological order.
func (r *PostgresTransitionLogRepository) ListByContractDate(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) ([]*entity.WorkflowTransition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			from_state, to_state, action,
			actor_id, actor_type, actor_username, actor_account_code, is_admin_override,
			reason, metadata, occurred_at, request_id
		FROM workflow__transition_log
		WHERE contract_id = $1 AND business_date = $2
		ORDER BY occurred_at ASC`,
		contractID, businessDate,
	)
	if err != nil {
		return nil, fmt.Errorf("querying transition log: %w", err)
	}
	defer rows.Close()

	var results []*entity.WorkflowTransition
	for rows.Next() {
		t, err := scanTransition(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// ListByBusinessDatePaginated returns a page of transitions plus the total count.
func (r *PostgresTransitionLogRepository) ListByBusinessDatePaginated(
	ctx context.Context,
	req domain.TransitionLogPageRequest,
) (*domain.TransitionLogPageResult, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow__transition_log WHERE business_date = $1`,
		req.BusinessDate,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting transition log: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			from_state, to_state, action,
			actor_id, actor_type, actor_username, actor_account_code, is_admin_override,
			reason, metadata, occurred_at, request_id
		FROM workflow__transition_log
		WHERE business_date = $1
		ORDER BY occurred_at ASC
		LIMIT $2 OFFSET $3`,
		req.BusinessDate, req.PageSize, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("querying transition log page: %w", err)
	}
	defer rows.Close()

	var transitions []*entity.WorkflowTransition
	for rows.Next() {
		t, err := scanTransition(rows)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.TransitionLogPageResult{
		Transitions: transitions,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// ListByContractDatePaginated returns a page of transitions plus the total count.
func (r *PostgresTransitionLogRepository) ListByContractDatePaginated(
	ctx context.Context,
	req domain.TransitionLogPageRequest,
) (*domain.TransitionLogPageResult, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM workflow__transition_log WHERE contract_id = $1 AND business_date = $2`,
		req.ContractID, req.BusinessDate,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting transition log: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			from_state, to_state, action,
			actor_id, actor_type, actor_username, actor_account_code, is_admin_override,
			reason, metadata, occurred_at, request_id
		FROM workflow__transition_log
		WHERE contract_id = $1 AND business_date = $2
		ORDER BY occurred_at ASC
		LIMIT $3 OFFSET $4`,
		req.ContractID, req.BusinessDate, req.PageSize, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("querying transition log page: %w", err)
	}
	defer rows.Close()

	var transitions []*entity.WorkflowTransition
	for rows.Next() {
		t, err := scanTransition(rows)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.TransitionLogPageResult{
		Transitions: transitions,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func scanTransition(rows pgx.Rows) (*entity.WorkflowTransition, error) {
	var t entity.WorkflowTransition
	var fromStr, toStr, actionStr, actorTypeStr string
	var metaBytes []byte

	err := rows.Scan(
		&t.ID, &t.WorkflowDayID, &t.ContractID, &t.BusinessDate,
		&fromStr, &toStr, &actionStr,
		&t.ActorID, &actorTypeStr, &t.ActorUsername, &t.ActorAccountCode, &t.IsAdminOverride,
		&t.Reason, &metaBytes, &t.OccurredAt, &t.RequestID,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning transition: %w", err)
	}

	t.FromState = vo.WorkflowState(fromStr)
	t.ToState = vo.WorkflowState(toStr)
	t.Action = vo.WorkflowAction(actionStr)
	t.ActorType = vo.ActorType(actorTypeStr)
	t.OccurredAt = t.OccurredAt.UTC()

	if len(metaBytes) > 0 {
		if err := json.Unmarshal(metaBytes, &t.Metadata); err != nil {
			t.Metadata = map[string]any{}
		}
	} else {
		t.Metadata = map[string]any{}
	}

	return &t, nil
}

// compile-time interface check
var _ domain.TransitionLogRepository = (*PostgresTransitionLogRepository)(nil)
