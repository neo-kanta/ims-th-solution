package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
)

// PostgresEmailOutboxRepository implements domain.EmailOutboxRepository.
type PostgresEmailOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEmailOutboxRepository(pool *pgxpool.Pool) *PostgresEmailOutboxRepository {
	return &PostgresEmailOutboxRepository{pool: pool}
}

// Create inserts a new outbox row using ON CONFLICT (idempotency_key) DO NOTHING.
// Returns nil UUID when the row already exists (idempotency conflict).
func (r *PostgresEmailOutboxRepository) Create(ctx context.Context, o *entity.EmailOutbox) (*uuid.UUID, error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	now := time.Now().UTC()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = now
	}
	if o.NextAttemptAt.IsZero() {
		o.NextAttemptAt = now
	}
	if o.Status == "" {
		o.Status = "PENDING"
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 5
	}

	var insertedID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notification__email_outbox (
			id, idempotency_key,
			notification_id, recipient_user_id,
			recipient_username, recipient_display_name, recipient_email,
			to_email, to_name,
			event_type, event_label, event_category, event_severity,
			business_type, business_label, business_id,
			business_reference, business_title,
			action_label, action_url,
			subject, body_text, body_html,
			status, attempts, max_attempts, next_attempt_at,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
			$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29
		)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id`,
		o.ID, o.IdempotencyKey,
		o.NotificationID, o.RecipientUserID,
		o.RecipientUsername, o.RecipientDisplayName, o.RecipientEmail,
		o.ToEmail, o.ToName,
		o.EventType, o.EventLabel, o.EventCategory, o.EventSeverity,
		nilStr(o.BusinessType), o.BusinessLabel, o.BusinessID,
		o.BusinessReference, o.BusinessTitle,
		o.ActionLabel, o.ActionURL,
		o.Subject, o.BodyText, o.BodyHTML,
		o.Status, o.Attempts, o.MaxAttempts, o.NextAttemptAt,
		o.CreatedAt, o.UpdatedAt,
	).Scan(&insertedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // idempotency conflict — already queued
	}
	if err != nil {
		return nil, fmt.Errorf("create email outbox: %w", err)
	}
	return &insertedID, nil
}

// GetByID fetches a single outbox row by its primary key.
func (r *PostgresEmailOutboxRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.EmailOutbox, error) {
	o := &entity.EmailOutbox{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, idempotency_key,
		       notification_id, recipient_user_id,
		       recipient_username, recipient_display_name, recipient_email,
		       to_email, to_name,
		       event_type, event_label, event_category, event_severity,
		       COALESCE(business_type,''), business_label, business_id,
		       business_reference, business_title,
		       action_label, action_url,
		       subject, body_text, COALESCE(body_html,''),
		       status, attempts, max_attempts, next_attempt_at,
		       locked_at, COALESCE(locked_by,''),
		       sent_at, COALESCE(last_error,''), COALESCE(provider_message_id,''),
		       created_at, updated_at
		FROM notification__email_outbox WHERE id = $1`, id,
	).Scan(
		&o.ID, &o.IdempotencyKey,
		&o.NotificationID, &o.RecipientUserID,
		&o.RecipientUsername, &o.RecipientDisplayName, &o.RecipientEmail,
		&o.ToEmail, &o.ToName,
		&o.EventType, &o.EventLabel, &o.EventCategory, &o.EventSeverity,
		&o.BusinessType, &o.BusinessLabel, &o.BusinessID,
		&o.BusinessReference, &o.BusinessTitle,
		&o.ActionLabel, &o.ActionURL,
		&o.Subject, &o.BodyText, &o.BodyHTML,
		&o.Status, &o.Attempts, &o.MaxAttempts, &o.NextAttemptAt,
		&o.LockedAt, &o.LockedBy,
		&o.SentAt, &o.LastError, &o.ProviderMessageID,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get email outbox by id: %w", err)
	}
	return o, nil
}

// List returns a filtered, paged list of outbox rows.
func (r *PostgresEmailOutboxRepository) List(ctx context.Context, f domain.EmailOutboxFilter) ([]*entity.EmailOutbox, int, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	args := []any{}
	where := " WHERE 1=1"
	addArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if f.Status != "" {
		where += " AND status = " + addArg(f.Status)
	}
	if f.RecipientUsername != "" {
		where += " AND recipient_username = " + addArg(f.RecipientUsername)
	}
	if f.RecipientEmail != "" {
		where += " AND to_email = " + addArg(f.RecipientEmail)
	}
	if f.EventType != "" {
		where += " AND event_type = " + addArg(f.EventType)
	}
	if f.EventCategory != "" {
		where += " AND event_category = " + addArg(f.EventCategory)
	}
	if f.BusinessType != "" {
		where += " AND business_type = " + addArg(f.BusinessType)
	}
	if f.BusinessReference != "" {
		where += " AND business_reference = " + addArg(f.BusinessReference)
	}
	if f.CreatedFrom != nil {
		where += " AND created_at >= " + addArg(*f.CreatedFrom)
	}
	if f.CreatedTo != nil {
		where += " AND created_at <= " + addArg(*f.CreatedTo)
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification__email_outbox`+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count email outbox: %w", err)
	}

	q := `SELECT id, idempotency_key,
		       notification_id, recipient_user_id,
		       recipient_username, recipient_display_name, recipient_email,
		       to_email, to_name,
		       event_type, event_label, event_category, event_severity,
		       COALESCE(business_type,''), business_label, business_id,
		       business_reference, business_title,
		       action_label, action_url,
		       subject, status, attempts, max_attempts, next_attempt_at,
		       sent_at, COALESCE(last_error,''), COALESCE(provider_message_id,''),
		       created_at, updated_at
		FROM notification__email_outbox` + where +
		` ORDER BY created_at DESC LIMIT ` + addArg(limit) + ` OFFSET ` + addArg(offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list email outbox: %w", err)
	}
	defer rows.Close()

	var items []*entity.EmailOutbox
	for rows.Next() {
		o := &entity.EmailOutbox{}
		if err := rows.Scan(
			&o.ID, &o.IdempotencyKey,
			&o.NotificationID, &o.RecipientUserID,
			&o.RecipientUsername, &o.RecipientDisplayName, &o.RecipientEmail,
			&o.ToEmail, &o.ToName,
			&o.EventType, &o.EventLabel, &o.EventCategory, &o.EventSeverity,
			&o.BusinessType, &o.BusinessLabel, &o.BusinessID,
			&o.BusinessReference, &o.BusinessTitle,
			&o.ActionLabel, &o.ActionURL,
			&o.Subject, &o.Status, &o.Attempts, &o.MaxAttempts, &o.NextAttemptAt,
			&o.SentAt, &o.LastError, &o.ProviderMessageID,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, o)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return items, total, nil
}

// ClaimBatch atomically locks up to batchSize due rows for sending.
func (r *PostgresEmailOutboxRepository) ClaimBatch(ctx context.Context, batchSize int, workerID string) ([]*entity.EmailOutbox, error) {
	rows, err := r.pool.Query(ctx, `
		WITH due AS (
			SELECT id
			FROM notification__email_outbox
			WHERE status IN ('PENDING', 'FAILED')
			  AND next_attempt_at <= NOW()
			  AND attempts < max_attempts
			ORDER BY next_attempt_at ASC, created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notification__email_outbox e
		SET status     = 'SENDING',
		    locked_at  = NOW(),
		    locked_by  = $2,
		    updated_at = NOW()
		FROM due
		WHERE e.id = due.id
		RETURNING e.id, e.idempotency_key,
		          e.notification_id, e.recipient_user_id,
		          e.recipient_username, e.recipient_display_name, e.recipient_email,
		          e.to_email, e.to_name,
		          e.event_type, e.event_label, e.event_category, e.event_severity,
		          COALESCE(e.business_type,''), e.business_label, e.business_id,
		          e.business_reference, e.business_title,
		          e.action_label, e.action_url,
		          e.subject, e.body_text, COALESCE(e.body_html,''),
		          e.status, e.attempts, e.max_attempts, e.next_attempt_at,
		          e.locked_at, COALESCE(e.locked_by,''),
		          e.sent_at, COALESCE(e.last_error,''), COALESCE(e.provider_message_id,''),
		          e.created_at, e.updated_at`,
		batchSize, workerID,
	)
	if err != nil {
		return nil, fmt.Errorf("claim outbox batch: %w", err)
	}
	defer rows.Close()

	var items []*entity.EmailOutbox
	for rows.Next() {
		o := &entity.EmailOutbox{}
		if err := rows.Scan(
			&o.ID, &o.IdempotencyKey,
			&o.NotificationID, &o.RecipientUserID,
			&o.RecipientUsername, &o.RecipientDisplayName, &o.RecipientEmail,
			&o.ToEmail, &o.ToName,
			&o.EventType, &o.EventLabel, &o.EventCategory, &o.EventSeverity,
			&o.BusinessType, &o.BusinessLabel, &o.BusinessID,
			&o.BusinessReference, &o.BusinessTitle,
			&o.ActionLabel, &o.ActionURL,
			&o.Subject, &o.BodyText, &o.BodyHTML,
			&o.Status, &o.Attempts, &o.MaxAttempts, &o.NextAttemptAt,
			&o.LockedAt, &o.LockedBy,
			&o.SentAt, &o.LastError, &o.ProviderMessageID,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	return items, rows.Err()
}

func (r *PostgresEmailOutboxRepository) MarkSent(ctx context.Context, id uuid.UUID, providerMessageID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notification__email_outbox
		SET status             = 'SENT',
		    sent_at            = NOW(),
		    last_error         = NULL,
		    provider_message_id = $2,
		    locked_at          = NULL,
		    locked_by          = NULL,
		    updated_at         = NOW()
		WHERE id = $1`, id, providerMessageID)
	if err != nil {
		return fmt.Errorf("mark outbox sent: %w", err)
	}
	return nil
}

func (r *PostgresEmailOutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextDelay time.Duration) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notification__email_outbox
		SET status = CASE
		        WHEN attempts + 1 >= max_attempts THEN 'DEAD'
		        ELSE 'FAILED'
		    END,
		    attempts = attempts + 1,
		    next_attempt_at = CASE
		        WHEN attempts + 1 >= max_attempts THEN next_attempt_at
		        ELSE NOW() + make_interval(secs => $2)
		    END,
		    last_error  = $3,
		    locked_at   = NULL,
		    locked_by   = NULL,
		    updated_at  = NOW()
		WHERE id = $1`,
		id, nextDelay.Seconds(), errMsg)
	if err != nil {
		return fmt.Errorf("mark outbox failed: %w", err)
	}
	return nil
}

func (r *PostgresEmailOutboxRepository) Retry(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE notification__email_outbox
		SET status         = 'PENDING',
		    next_attempt_at = NOW(),
		    locked_at       = NULL,
		    locked_by       = NULL,
		    last_error      = NULL,
		    updated_at      = NOW()
		WHERE id = $1 AND status IN ('FAILED', 'DEAD')`, id)
	if err != nil {
		return fmt.Errorf("retry outbox: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrOutboxNotRetryable
	}
	return nil
}

func (r *PostgresEmailOutboxRepository) RecoverStale(ctx context.Context, timeout time.Duration, nextDelay time.Duration) (int, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE notification__email_outbox
		SET status = CASE
		        WHEN attempts + 1 >= max_attempts THEN 'DEAD'
		        ELSE 'FAILED'
		    END,
		    attempts = attempts + 1,
		    next_attempt_at = CASE
		        WHEN attempts + 1 >= max_attempts THEN next_attempt_at
		        ELSE NOW() + make_interval(secs => $2)
		    END,
		    last_error  = COALESCE(last_error, 'email worker crashed or timed out while sending'),
		    locked_at   = NULL,
		    locked_by   = NULL,
		    updated_at  = NOW()
		WHERE status = 'SENDING'
		  AND locked_at < NOW() - make_interval(secs => $1)`,
		timeout.Seconds(), nextDelay.Seconds())
	if err != nil {
		return 0, fmt.Errorf("recover stale outbox: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *PostgresEmailOutboxRepository) HealthCounts(ctx context.Context) (pending, failed, dead int, err error) {
	rows, err := r.pool.Query(ctx, `
		SELECT status, COUNT(*)
		FROM notification__email_outbox
		WHERE status IN ('PENDING', 'FAILED', 'DEAD')
		GROUP BY status`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("health counts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return 0, 0, 0, err
		}
		switch status {
		case "PENDING":
			pending = count
		case "FAILED":
			failed = count
		case "DEAD":
			dead = count
		}
	}
	return pending, failed, dead, rows.Err()
}

// ErrOutboxNotRetryable is returned when the outbox row is not in a retryable state.
var ErrOutboxNotRetryable = errors.New("email outbox: row not in FAILED or DEAD status")
