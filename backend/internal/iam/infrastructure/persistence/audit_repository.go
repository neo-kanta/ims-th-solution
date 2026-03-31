package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// PostgresAuditRepository implements domain.AuditRepository.
type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAuditRepository creates a new audit repository.
func NewPostgresAuditRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

// Record inserts an immutable audit event.
func (r *PostgresAuditRepository) Record(ctx context.Context, event *entity.AuditEvent) error {
	query := `
		INSERT INTO iam_audit_events (id, actor_id, event_type, target_type, target_id,
		                              ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6::inet, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		event.ID, event.ActorID, event.EventType, event.TargetType, event.TargetID,
		nullableString(event.IPAddress), event.UserAgent, event.Metadata,
	)
	if err != nil {
		return fmt.Errorf("recording audit event: %w", err)
	}
	return nil
}

// List returns paginated audit events matching the filter.
func (r *PostgresAuditRepository) List(ctx context.Context, filter domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	where := "1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.ActorID != nil {
		where += fmt.Sprintf(" AND actor_id = $%d", argIdx)
		args = append(args, *filter.ActorID)
		argIdx++
	}
	if filter.EventType != "" {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, filter.EventType)
		argIdx++
	}
	if filter.TargetType != "" {
		where += fmt.Sprintf(" AND target_type = $%d", argIdx)
		args = append(args, filter.TargetType)
		argIdx++
	}
	if filter.TargetID != "" {
		where += fmt.Sprintf(" AND target_id = $%d", argIdx)
		args = append(args, filter.TargetID)
		argIdx++
	}
	if filter.Since != nil {
		where += fmt.Sprintf(" AND created_at >= $%d::timestamptz", argIdx)
		args = append(args, *filter.Since)
		argIdx++
	}
	if filter.Until != nil {
		where += fmt.Sprintf(" AND created_at <= $%d::timestamptz", argIdx)
		args = append(args, *filter.Until)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM iam_audit_events WHERE %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit events: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, actor_id, event_type, target_type, target_id,
		       host(ip_address), user_agent, metadata, created_at
		FROM iam_audit_events
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit events: %w", err)
	}
	defer rows.Close()

	var events []entity.AuditEvent
	for rows.Next() {
		var e entity.AuditEvent
		var ipAddr *string
		if err := rows.Scan(
			&e.ID, &e.ActorID, &e.EventType, &e.TargetType, &e.TargetID,
			&ipAddr, &e.UserAgent, &e.Metadata, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning audit event: %w", err)
		}
		if ipAddr != nil {
			e.IPAddress = *ipAddr
		}
		events = append(events, e)
	}
	if events == nil {
		events = []entity.AuditEvent{}
	}

	return events, total, nil
}
