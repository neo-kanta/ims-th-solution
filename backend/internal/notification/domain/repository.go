package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
)

// UserSummary is a lightweight snapshot of an IAM user, resolved from iam_users.
type UserSummary struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	Email       string
}

// ListFilter scopes a notification list query.
type ListFilter struct {
	RecipientUserID uuid.UUID
	UnreadOnly      bool
	Limit           int
	Offset          int
}

// Repository persists in-app notifications and provides IAM user lookups.
type Repository interface {
	Create(ctx context.Context, n *entity.Notification) error
	CreateIdempotent(ctx context.Context, n *entity.Notification) (created bool, err error)
	List(ctx context.Context, f ListFilter) ([]*entity.Notification, int, int, error) // items, total, unread
	MarkRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error
	MarkAllRead(ctx context.Context, recipientID uuid.UUID) (int, error)
	GetByIDForActor(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) (*entity.Notification, error)

	// RecentMatchExists reports whether a notification for the given recipient,
	// source ID and category was written within the recent window. Watchers
	// (e.g. the workflow stuck-day notifier) use this to dedupe periodic
	// emits so the recipient's inbox is not flooded.
	RecentMatchExists(ctx context.Context, recipientID uuid.UUID, sourceID uuid.UUID, category string, sinceHours int) (bool, error)

	// IAM user lookups used when creating email outbox rows.
	GetUserByID(ctx context.Context, id uuid.UUID) (UserSummary, error)
	GetUserByUsername(ctx context.Context, username string) (UserSummary, error)
}

// EmailOutboxFilter scopes an email outbox list query.
type EmailOutboxFilter struct {
	Status            string
	RecipientUsername string
	RecipientEmail    string
	EventType         string
	EventCategory     string
	BusinessType      string
	BusinessReference string
	CreatedFrom       *time.Time
	CreatedTo         *time.Time
	Limit             int
	Offset            int
}

// EmailOutboxRepository persists and manages the email outbox queue.
type EmailOutboxRepository interface {
	// Create inserts a new outbox row. Returns nil UUID when the idempotency_key
	// already exists (ON CONFLICT DO NOTHING) — the caller should treat this as
	// success rather than an error.
	Create(ctx context.Context, o *entity.EmailOutbox) (*uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.EmailOutbox, error)
	List(ctx context.Context, f EmailOutboxFilter) ([]*entity.EmailOutbox, int, error)

	// Worker operations.
	ClaimBatch(ctx context.Context, batchSize int, workerID string) ([]*entity.EmailOutbox, error)
	MarkSent(ctx context.Context, id uuid.UUID, providerMessageID string) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextDelay time.Duration) error
	Retry(ctx context.Context, id uuid.UUID) error
	RecoverStale(ctx context.Context, timeout time.Duration, nextDelay time.Duration) (int, error)
	HealthCounts(ctx context.Context) (pending, failed, dead int, err error)
}
