package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/enum"
)

// TestEmailInput is the request payload for the test email endpoint.
type TestEmailInput struct {
	ToUsername    string
	ToEmail       string
	Subject       string
	Body          string
	AllowRawEmail bool
	AllowedDomains []string
	MaxAttempts   int
}

// TestEmailResult is returned by SendTestEmailIntent.
type TestEmailResult struct {
	OutboxID  uuid.UUID
	Recipient domain.UserSummary
}

// EmailOutboxService handles admin/demo email outbox operations.
type EmailOutboxService struct {
	outboxRepo  domain.EmailOutboxRepository
	userRepo    domain.Repository
	templateSvc *EmailTemplateService
}

func NewEmailOutboxService(outboxRepo domain.EmailOutboxRepository, userRepo domain.Repository, templateSvc *EmailTemplateService) *EmailOutboxService {
	return &EmailOutboxService{
		outboxRepo:  outboxRepo,
		userRepo:    userRepo,
		templateSvc: templateSvc,
	}
}

// SendTestEmailIntent creates a test email outbox row. It does not send immediately;
// the background worker picks it up on its next tick.
func (s *EmailOutboxService) SendTestEmailIntent(ctx context.Context, in TestEmailInput) (*TestEmailResult, error) {
	if in.Subject == "" || (in.ToUsername == "" && in.ToEmail == "") {
		return nil, fmt.Errorf("subject and at least one of to_username or to_email are required")
	}

	var user domain.UserSummary
	var toEmail, toName string

	if in.ToUsername != "" {
		u, err := s.userRepo.GetUserByUsername(ctx, in.ToUsername)
		if err != nil {
			return nil, fmt.Errorf("resolve username: %w", err)
		}
		if u.Email == "" {
			return nil, fmt.Errorf("user %q not found", in.ToUsername)
		}
		user = u
		toEmail = u.Email
		toName = u.DisplayName
	} else {
		// Raw email path — only allowed when explicitly enabled.
		if !in.AllowRawEmail {
			return nil, fmt.Errorf("raw to_email is not enabled; set NOTIFICATION_EMAIL_ALLOW_RAW_EMAIL=true or supply to_username")
		}
		if !domainAllowed(in.ToEmail, in.AllowedDomains) {
			return nil, fmt.Errorf("email domain not in NOTIFICATION_EMAIL_ALLOWED_DOMAINS")
		}
		toEmail = in.ToEmail
	}

	def := valueobject.LookupEvent(enum.NotificationEventEmailTest)
	subject, bodyText, bodyHTML := s.templateSvc.Render(string(enum.NotificationEventEmailTest), EmailTemplateData{
		RecipientName: toName,
		CustomSubject: in.Subject,
		CustomBody:    in.Body,
	})

	maxAttempts := in.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	requestID := uuid.New()
	idempotencyKey := fmt.Sprintf("EMAIL_TEST:%s:%s", requestID, toEmail)

	var recipientUserID *uuid.UUID
	if user.ID != uuid.Nil {
		id := user.ID
		recipientUserID = &id
	}

	o := &entity.EmailOutbox{
		IdempotencyKey:       idempotencyKey,
		RecipientUserID:      recipientUserID,
		RecipientUsername:    user.Username,
		RecipientDisplayName: user.DisplayName,
		RecipientEmail:       user.Email,
		ToEmail:              toEmail,
		ToName:               toName,
		EventType:            string(enum.NotificationEventEmailTest),
		EventLabel:           def.Label,
		EventCategory:        def.Category,
		EventSeverity:        def.Severity,
		BusinessType:         "EMAIL_TEST",
		BusinessTitle:        in.Subject,
		Subject:              subject,
		BodyText:             bodyText,
		BodyHTML:             bodyHTML,
		Status:               "PENDING",
		MaxAttempts:          maxAttempts,
	}

	insertedID, err := s.outboxRepo.Create(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("create test email outbox: %w", err)
	}
	if insertedID == nil {
		// Idempotency conflict — return existing request ID as placeholder.
		insertedID = &requestID
	}

	return &TestEmailResult{
		OutboxID:  *insertedID,
		Recipient: user,
	}, nil
}

// GetOutbox returns a filtered, paged list of outbox rows.
func (s *EmailOutboxService) GetOutbox(ctx context.Context, f domain.EmailOutboxFilter) ([]*entity.EmailOutbox, int, error) {
	return s.outboxRepo.List(ctx, f)
}

// GetOutboxByID returns a single outbox row by ID.
func (s *EmailOutboxService) GetOutboxByID(ctx context.Context, id uuid.UUID) (*entity.EmailOutbox, error) {
	return s.outboxRepo.GetByID(ctx, id)
}

// RetryOutbox requeues a failed or dead outbox row.
func (s *EmailOutboxService) RetryOutbox(ctx context.Context, id uuid.UUID) (*entity.EmailOutbox, error) {
	if err := s.outboxRepo.Retry(ctx, id); err != nil {
		return nil, err
	}
	return s.outboxRepo.GetByID(ctx, id)
}

// HealthCounts returns the pending, failed, and dead queue counts.
func (s *EmailOutboxService) HealthCounts(ctx context.Context) (pending, failed, dead int, err error) {
	return s.outboxRepo.HealthCounts(ctx)
}

// domainAllowed returns true when the email address's domain is in the allowed list.
// Empty allowed list permits all domains (development-only default).
func domainAllowed(email string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := strings.ToLower(email[at+1:])
	for _, d := range allowed {
		if strings.ToLower(strings.TrimSpace(d)) == domain {
			return true
		}
	}
	return false
}
