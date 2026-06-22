package entity

import (
	"time"

	"github.com/google/uuid"
)

// Notification is a single in-app notification record for one recipient.
type Notification struct {
	ID              uuid.UUID
	RecipientUserID uuid.UUID
	Category        string
	Title           string
	Body            string
	Link            string
	SourceModule    string
	SourceType      string
	SourceID        *uuid.UUID
	IsRead          bool
	ReadAt          *time.Time
	CreatedAt       time.Time

	// Event and context fields (added via migration 20260619000001).
	EventType         string
	IdempotencyKey    string
	BusinessType      string
	BusinessLabel     string
	BusinessID        *uuid.UUID
	BusinessReference string
	BusinessTitle     string
	ActionLabel       string
	ActionURL         string

	// Populated by JOIN on iam_users — not stored in the notifications table.
	RecipientUsername    string
	RecipientDisplayName string
	RecipientEmail       string
}
