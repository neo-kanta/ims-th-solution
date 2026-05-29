package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// AppendEvent inserts an immutable approval event.
func (r *PostgresRepository) AppendEvent(ctx context.Context, tx pgx.Tx, e *entity.ApprovalEvent) error {
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__events
			(id, approval_request_id, event_type, stage_number, actor_user_id, delegated_from_user_id, comment, metadata_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.ApprovalRequestID, string(e.EventType), e.StageNumber, e.ActorUserID,
		e.DelegatedFromUserID, e.Comment, e.Metadata)
	if err != nil {
		return fmt.Errorf("appending approval event: %w", err)
	}
	return nil
}

// ListEvents returns the ordered, immutable timeline for a request.
func (r *PostgresRepository) ListEvents(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.approval_request_id, e.event_type, e.stage_number, e.actor_user_id,
		       e.delegated_from_user_id, e.comment, e.metadata_json, e.created_at,
		       COALESCE(u.display_name, u.username, '')
		FROM approval__events e
		LEFT JOIN iam_users u ON u.id = e.actor_user_id
		WHERE e.approval_request_id = $1
		ORDER BY e.created_at ASC, e.id ASC`, requestID)
	if err != nil {
		return nil, fmt.Errorf("listing approval events: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalEvent
	for rows.Next() {
		var e entity.ApprovalEvent
		var eventType string
		if err := rows.Scan(&e.ID, &e.ApprovalRequestID, &eventType, &e.StageNumber, &e.ActorUserID,
			&e.DelegatedFromUserID, &e.Comment, &e.Metadata, &e.CreatedAt, &e.ActorName); err != nil {
			return nil, fmt.Errorf("scanning approval event: %w", err)
		}
		e.EventType = vo.EventType(eventType)
		out = append(out, &e)
	}
	return out, rows.Err()
}
