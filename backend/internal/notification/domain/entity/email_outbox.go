package entity

import (
	"time"

	"github.com/google/uuid"
)

// EmailOutbox is one durable email delivery record in the outbox queue.
type EmailOutbox struct {
	ID             uuid.UUID
	IdempotencyKey string

	NotificationID  *uuid.UUID
	RecipientUserID *uuid.UUID

	RecipientUsername    string
	RecipientDisplayName string
	RecipientEmail       string

	ToEmail string
	ToName  string

	EventType     string
	EventLabel    string
	EventCategory string
	EventSeverity string

	BusinessType      string
	BusinessLabel     string
	BusinessID        *uuid.UUID
	BusinessReference string
	BusinessTitle     string

	ActionLabel string
	ActionURL   string

	Subject  string
	BodyText string
	BodyHTML string

	Status        string
	Attempts      int
	MaxAttempts   int
	NextAttemptAt time.Time

	LockedAt *time.Time
	LockedBy string

	SentAt            *time.Time
	LastError         string
	ProviderMessageID string

	CreatedAt time.Time
	UpdatedAt time.Time
}
