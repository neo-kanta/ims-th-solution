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
			actor_id, actor_type, actor_username,
			reason, metadata, occurred_at, request_id
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14
		)`,
		t.ID, t.WorkflowDayID, t.ContractID, t.BusinessDate,
		string(t.FromState), string(t.ToState), string(t.Action),
		t.ActorID, string(t.ActorType), t.ActorUsername,
		t.Reason, metaBytes, t.OccurredAt, t.RequestID,
	)
	if err != nil {
		return fmt.Errorf("appending transition log: %w", err)
	}
	return nil
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
			actor_id, actor_type, actor_username,
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

func scanTransition(rows pgx.Rows) (*entity.WorkflowTransition, error) {
	var t entity.WorkflowTransition
	var fromStr, toStr, actionStr, actorTypeStr string
	var metaBytes []byte

	err := rows.Scan(
		&t.ID, &t.WorkflowDayID, &t.ContractID, &t.BusinessDate,
		&fromStr, &toStr, &actionStr,
		&t.ActorID, &actorTypeStr, &t.ActorUsername,
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
			t.Metadata = map[string]any{} // degrade gracefully
		}
	} else {
		t.Metadata = map[string]any{}
	}

	return &t, nil
}

// compile-time interface check
var _ domain.TransitionLogRepository = (*PostgresTransitionLogRepository)(nil)
