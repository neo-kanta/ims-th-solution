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

type PostgresNotificationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationRepository(pool *pgxpool.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{pool: pool}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	eventType := n.EventType
	if eventType == "" {
		eventType = n.Category
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notification__notifications (
			id, recipient_user_id, category, title, body, link,
			source_module, source_type, source_id,
			is_read, read_at, created_at,
			event_type, idempotency_key,
			business_type, business_label, business_id,
			business_reference, business_title,
			action_label, action_url
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
		          $13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		n.ID, n.RecipientUserID, n.Category, n.Title, n.Body, n.Link,
		n.SourceModule, n.SourceType, n.SourceID,
		n.IsRead, n.ReadAt, n.CreatedAt,
		eventType, nilStr(n.IdempotencyKey),
		nilStr(n.BusinessType), n.BusinessLabel, n.BusinessID,
		n.BusinessReference, n.BusinessTitle,
		n.ActionLabel, n.ActionURL,
	)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

// CreateIdempotent inserts only when the idempotency_key is absent.
// Returns created=false (no error) when the key already exists.
func (r *PostgresNotificationRepository) CreateIdempotent(ctx context.Context, n *entity.Notification) (bool, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	eventType := n.EventType
	if eventType == "" {
		eventType = n.Category
	}
	var insertedID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notification__notifications (
			id, recipient_user_id, category, title, body, link,
			source_module, source_type, source_id,
			is_read, read_at, created_at,
			event_type, idempotency_key,
			business_type, business_label, business_id,
			business_reference, business_title,
			action_label, action_url
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
		          $13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (idempotency_key)
		WHERE idempotency_key IS NOT NULL AND idempotency_key <> ''
		DO NOTHING
		RETURNING id`,
		n.ID, n.RecipientUserID, n.Category, n.Title, n.Body, n.Link,
		n.SourceModule, n.SourceType, n.SourceID,
		n.IsRead, n.ReadAt, n.CreatedAt,
		eventType, n.IdempotencyKey,
		nilStr(n.BusinessType), n.BusinessLabel, n.BusinessID,
		n.BusinessReference, n.BusinessTitle,
		n.ActionLabel, n.ActionURL,
	).Scan(&insertedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil // already existed
	}
	if err != nil {
		return false, fmt.Errorf("insert notification idempotent: %w", err)
	}
	n.ID = insertedID
	return true, nil
}

func (r *PostgresNotificationRepository) List(ctx context.Context, f domain.ListFilter) ([]*entity.Notification, int, int, error) {
	if f.RecipientUserID == uuid.Nil {
		return nil, 0, 0, fmt.Errorf("recipient_user_id is required")
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	var total, unread int
	if err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE TRUE),
			COUNT(*) FILTER (WHERE is_read = FALSE)
		FROM notification__notifications
		WHERE recipient_user_id = $1`, f.RecipientUserID).Scan(&total, &unread); err != nil {
		return nil, 0, 0, fmt.Errorf("count notifications: %w", err)
	}

	q := `
		SELECT n.id, n.recipient_user_id, n.category, n.title, n.body, n.link,
		       n.source_module, n.source_type, n.source_id,
		       n.is_read, n.read_at, n.created_at,
		       COALESCE(n.event_type, n.category) AS event_type,
		       COALESCE(n.idempotency_key, '') AS idempotency_key,
		       COALESCE(n.business_type, '') AS business_type,
		       COALESCE(n.business_label, '') AS business_label,
		       n.business_id,
		       COALESCE(n.business_reference, '') AS business_reference,
		       COALESCE(n.business_title, '') AS business_title,
		       COALESCE(n.action_label, '') AS action_label,
		       COALESCE(n.action_url, '') AS action_url,
		       COALESCE(u.username, '') AS recipient_username,
		       COALESCE(u.display_name, '') AS recipient_display_name,
		       COALESCE(u.email, '') AS recipient_email
		FROM notification__notifications n
		LEFT JOIN iam_users u ON u.id = n.recipient_user_id
		WHERE n.recipient_user_id = $1`
	args := []any{f.RecipientUserID}
	if f.UnreadOnly {
		q += ` AND n.is_read = FALSE`
	}
	q += ` ORDER BY n.created_at DESC LIMIT $2 OFFSET $3`
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items := []*entity.Notification{}
	for rows.Next() {
		n := &entity.Notification{}
		var readAt *time.Time
		var sourceID, businessID *uuid.UUID
		if err := rows.Scan(
			&n.ID, &n.RecipientUserID, &n.Category, &n.Title, &n.Body, &n.Link,
			&n.SourceModule, &n.SourceType, &sourceID,
			&n.IsRead, &readAt, &n.CreatedAt,
			&n.EventType, &n.IdempotencyKey,
			&n.BusinessType, &n.BusinessLabel, &businessID,
			&n.BusinessReference, &n.BusinessTitle,
			&n.ActionLabel, &n.ActionURL,
			&n.RecipientUsername, &n.RecipientDisplayName, &n.RecipientEmail,
		); err != nil {
			return nil, 0, 0, err
		}
		n.ReadAt = readAt
		n.SourceID = sourceID
		n.BusinessID = businessID
		items = append(items, n)
	}
	if rows.Err() != nil {
		return nil, 0, 0, rows.Err()
	}
	return items, total, unread, nil
}

func (r *PostgresNotificationRepository) MarkRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE notification__notifications
		   SET is_read = TRUE, read_at = $3
		 WHERE id = $1 AND recipient_user_id = $2 AND is_read = FALSE`,
		id, recipientID, now)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (r *PostgresNotificationRepository) MarkAllRead(ctx context.Context, recipientID uuid.UUID) (int, error) {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE notification__notifications
		   SET is_read = TRUE, read_at = $2
		 WHERE recipient_user_id = $1 AND is_read = FALSE`,
		recipientID, now)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications read: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *PostgresNotificationRepository) GetByIDForActor(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) (*entity.Notification, error) {
	n := &entity.Notification{}
	var readAt *time.Time
	var sourceID, businessID *uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT n.id, n.recipient_user_id, n.category, n.title, n.body, n.link,
		       n.source_module, n.source_type, n.source_id,
		       n.is_read, n.read_at, n.created_at,
		       COALESCE(n.event_type, n.category) AS event_type,
		       COALESCE(n.idempotency_key, '') AS idempotency_key,
		       COALESCE(n.business_type, '') AS business_type,
		       COALESCE(n.business_label, '') AS business_label,
		       n.business_id,
		       COALESCE(n.business_reference, '') AS business_reference,
		       COALESCE(n.business_title, '') AS business_title,
		       COALESCE(n.action_label, '') AS action_label,
		       COALESCE(n.action_url, '') AS action_url,
		       COALESCE(u.username, '') AS recipient_username,
		       COALESCE(u.display_name, '') AS recipient_display_name,
		       COALESCE(u.email, '') AS recipient_email
		FROM notification__notifications n
		LEFT JOIN iam_users u ON u.id = n.recipient_user_id
		WHERE n.id = $1 AND n.recipient_user_id = $2`,
		id, recipientID,
	).Scan(
		&n.ID, &n.RecipientUserID, &n.Category, &n.Title, &n.Body, &n.Link,
		&n.SourceModule, &n.SourceType, &sourceID,
		&n.IsRead, &readAt, &n.CreatedAt,
		&n.EventType, &n.IdempotencyKey,
		&n.BusinessType, &n.BusinessLabel, &businessID,
		&n.BusinessReference, &n.BusinessTitle,
		&n.ActionLabel, &n.ActionURL,
		&n.RecipientUsername, &n.RecipientDisplayName, &n.RecipientEmail,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}
	n.ReadAt = readAt
	n.SourceID = sourceID
	n.BusinessID = businessID
	return n, nil
}

// RecentMatchExists returns true if a notification with the same recipient,
// source ID and category was created within sinceHours hours.
func (r *PostgresNotificationRepository) RecentMatchExists(
	ctx context.Context,
	recipientID uuid.UUID,
	sourceID uuid.UUID,
	category string,
	sinceHours int,
) (bool, error) {
	if recipientID == uuid.Nil || sourceID == uuid.Nil || category == "" {
		return false, nil
	}
	if sinceHours <= 0 {
		sinceHours = 24
	}
	cutoff := time.Now().UTC().Add(-time.Duration(sinceHours) * time.Hour)
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM notification__notifications
			 WHERE recipient_user_id = $1
			   AND source_id         = $2
			   AND category          = $3
			   AND created_at        >= $4
		)`,
		recipientID, sourceID, category, cutoff,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query notification existence: %w", err)
	}
	return exists, nil
}

// GetUserByID resolves a user summary from iam_users by ID.
func (r *PostgresNotificationRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domain.UserSummary, error) {
	var u domain.UserSummary
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, display_name, email
		FROM iam_users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserSummary{}, nil
	}
	if err != nil {
		return domain.UserSummary{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetUserByUsername resolves a user summary from iam_users by username.
func (r *PostgresNotificationRepository) GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error) {
	var u domain.UserSummary
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, display_name, email
		FROM iam_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserSummary{}, nil
	}
	if err != nil {
		return domain.UserSummary{}, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// ErrNoRowsAffected is returned by MarkRead when no row was updated.
var ErrNoRowsAffected = errors.New("notification: no row updated")

// nilStr converts an empty string to nil for nullable DB columns.
func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
