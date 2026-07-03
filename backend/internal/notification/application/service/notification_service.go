// Package service holds the application layer for the notification module.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/enum"
)

// CreateInput is the writer-side payload used by business module adapters to
// record a notification.
type CreateInput struct {
	RecipientUserID uuid.UUID
	Category        string
	Title           string
	Body            string
	Link            string
	SourceModule    string
	SourceType      string
	SourceID        *uuid.UUID

	// New event and context fields.
	EventType         string
	IdempotencyKey    string
	BusinessType      string
	BusinessLabel     string
	BusinessID        *uuid.UUID
	BusinessReference string
	BusinessTitle     string
	ActionLabel       string
	ActionURL         string
}

// Service exposes the in-app notification baseline and email outbox integration.
type Service struct {
	repo        domain.Repository
	emailRepo   domain.EmailOutboxRepository // optional; nil when email disabled
	templateSvc *EmailTemplateService
	emailCfg    EmailConfig
}

// EmailConfig holds the email-related flags the service needs.
type EmailConfig struct {
	Enabled        bool
	MaxAttempts    int
	AllowedDomains []string
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// WithEmail attaches the email outbox repository and configuration.
func (s *Service) WithEmail(emailRepo domain.EmailOutboxRepository, templateSvc *EmailTemplateService, cfg EmailConfig) {
	s.emailRepo = emailRepo
	s.templateSvc = templateSvc
	s.emailCfg = cfg
}

// Create stores a notification. When an idempotency_key is set, uses
// ON CONFLICT to avoid duplicates. On idempotency conflict, returns nil
// (not an error) — the originating business flow already succeeded.
func (s *Service) Create(ctx context.Context, in CreateInput) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if in.RecipientUserID == uuid.Nil || in.Title == "" {
		return nil // silently drop invalid; the originating flow already succeeded
	}

	eventType := in.EventType
	if eventType == "" {
		eventType = in.Category
	}

	n := &entity.Notification{
		RecipientUserID:   in.RecipientUserID,
		Category:          in.Category,
		Title:             in.Title,
		Body:              in.Body,
		Link:              in.Link,
		SourceModule:      in.SourceModule,
		SourceType:        in.SourceType,
		SourceID:          in.SourceID,
		EventType:         eventType,
		IdempotencyKey:    in.IdempotencyKey,
		BusinessType:      in.BusinessType,
		BusinessLabel:     in.BusinessLabel,
		BusinessID:        in.BusinessID,
		BusinessReference: in.BusinessReference,
		BusinessTitle:     in.BusinessTitle,
		ActionLabel:       in.ActionLabel,
		ActionURL:         in.ActionURL,
	}

	var notificationID *uuid.UUID
	if in.IdempotencyKey != "" {
		created, err := s.repo.CreateIdempotent(ctx, n)
		if err != nil {
			return fmt.Errorf("create notification: %w", err)
		}
		if !created {
			return nil // already exists — email is also already queued
		}
		notificationID = &n.ID
	} else {
		if err := s.repo.Create(ctx, n); err != nil {
			return fmt.Errorf("create notification: %w", err)
		}
		notificationID = &n.ID
	}

	s.createEmailOutbox(ctx, in, eventType, notificationID)
	return nil
}

// createEmailOutbox enqueues an email outbox row when email is enabled and the
// event type requires email delivery. Errors are logged and swallowed to keep
// the in-app notification path unblocked.
func (s *Service) createEmailOutbox(ctx context.Context, in CreateInput, eventType string, notificationID *uuid.UUID) {
	if !s.emailCfg.Enabled || s.emailRepo == nil || s.templateSvc == nil {
		return
	}
	def := valueobject.LookupEventByString(eventType)
	if def.Type == enum.NotificationEventEmailTest {
		// Test event has its own code path via SendTestEmailIntent.
		return
	}

	// Resolve recipient from IAM.
	user, err := s.repo.GetUserByID(ctx, in.RecipientUserID)
	if err != nil || user.Email == "" {
		slog.Warn("email outbox: could not resolve recipient email", "user_id", in.RecipientUserID, "err", err)
		return
	}

	actionLabel := in.ActionLabel
	if actionLabel == "" {
		actionLabel = def.DefaultActionLabel
	}

	subject, bodyText, bodyHTML := s.templateSvc.Render(eventType, EmailTemplateData{
		RecipientName:     user.DisplayName,
		BusinessReference: in.BusinessReference,
		BusinessTitle:     in.BusinessTitle,
		BusinessType:      in.BusinessType,
		ActionURL:         in.ActionURL,
		CustomSubject:     "",
		CustomBody:        "",
	})

	maxAttempts := s.emailCfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	o := &entity.EmailOutbox{
		IdempotencyKey:       "email:" + in.IdempotencyKey, // prefix to keep namespaces separate
		NotificationID:       notificationID,
		RecipientUserID:      &in.RecipientUserID,
		RecipientUsername:    user.Username,
		RecipientDisplayName: user.DisplayName,
		RecipientEmail:       user.Email,
		ToEmail:              user.Email,
		ToName:               user.DisplayName,
		EventType:            eventType,
		EventLabel:           def.Label,
		EventCategory:        def.Category,
		EventSeverity:        def.Severity,
		BusinessType:         in.BusinessType,
		BusinessLabel:        in.BusinessLabel,
		BusinessID:           in.BusinessID,
		BusinessReference:    in.BusinessReference,
		BusinessTitle:        in.BusinessTitle,
		ActionLabel:          actionLabel,
		ActionURL:            in.ActionURL,
		Subject:              subject,
		BodyText:             bodyText,
		BodyHTML:             bodyHTML,
		Status:               "PENDING",
		MaxAttempts:          maxAttempts,
	}
	// When no idempotency_key is set on the in-app notification, generate one
	// for the outbox using the notification ID.
	if in.IdempotencyKey == "" && notificationID != nil {
		o.IdempotencyKey = "email:notif:" + notificationID.String()
	}

	if _, err := s.emailRepo.Create(ctx, o); err != nil {
		slog.Warn("email outbox: failed to enqueue", "event_type", eventType, "err", err)
	}
}

// CreateIfAbsent stores a notification only when no notification with the
// same (recipient, source_id, category) has been written in the recent window.
// Used by periodic watchers — e.g. the workflow stuck-day notifier — to avoid
// flooding the recipient's inbox when the underlying condition persists across
// multiple scheduler ticks.
//
// SourceID is required; otherwise CreateIfAbsent falls back to a plain Create.
func (s *Service) CreateIfAbsent(ctx context.Context, in CreateInput, dedupeHours int) error {
	if s == nil || s.repo == nil {
		return nil
	}
	if in.IdempotencyKey != "" {
		// Use idempotency-key-based deduplication when a key is provided.
		return s.Create(ctx, in)
	}
	if in.SourceID == nil || *in.SourceID == uuid.Nil {
		return s.Create(ctx, in)
	}
	category := in.Category
	if category == "" {
		category = in.EventType
	}
	exists, err := s.repo.RecentMatchExists(ctx, in.RecipientUserID, *in.SourceID, category, dedupeHours)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.Create(ctx, in)
}

// List returns the recipient's notifications + total + unread counts.
func (s *Service) List(ctx context.Context, recipientID uuid.UUID, unreadOnly bool, limit, offset int) ([]*entity.Notification, int, int, error) {
	if s == nil || s.repo == nil {
		return nil, 0, 0, nil
	}
	return s.repo.List(ctx, domain.ListFilter{
		RecipientUserID: recipientID, UnreadOnly: unreadOnly, Limit: limit, Offset: offset,
	})
}

// MarkRead marks a single notification read (only when owned by recipient).
func (s *Service) MarkRead(ctx context.Context, id, recipientID uuid.UUID) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.MarkRead(ctx, id, recipientID)
}

// MarkAllRead marks every unread notification read for the recipient and
// returns the user summary for the response.
func (s *Service) MarkAllRead(ctx context.Context, recipientID uuid.UUID) (int, domain.UserSummary, error) {
	if s == nil || s.repo == nil {
		return 0, domain.UserSummary{}, nil
	}
	n, err := s.repo.MarkAllRead(ctx, recipientID)
	if err != nil {
		return 0, domain.UserSummary{}, err
	}
	user, _ := s.repo.GetUserByID(ctx, recipientID)
	return n, user, nil
}

// GetUserByID resolves a user summary from iam_users.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (domain.UserSummary, error) {
	return s.repo.GetUserByID(ctx, id)
}

// GetUserByUsername resolves a user summary from iam_users.
func (s *Service) GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

// GetByIDForActor fetches a notification by ID, scoped to the owning recipient.
func (s *Service) GetByIDForActor(ctx context.Context, id, recipientID uuid.UUID) (*entity.Notification, error) {
	return s.repo.GetByIDForActor(ctx, id, recipientID)
}
