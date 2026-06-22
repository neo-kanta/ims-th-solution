package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// EmailOutboxHandler serves admin/demo email outbox endpoints.
type EmailOutboxHandler struct {
	emailSvc *service.EmailOutboxService
	cfg      *config.AppConfig
}

func NewEmailOutboxHandler(emailSvc *service.EmailOutboxService, cfg *config.AppConfig) *EmailOutboxHandler {
	return &EmailOutboxHandler{emailSvc: emailSvc, cfg: cfg}
}

// --- response types ---

type outboxItemResponse struct {
	OutboxID           uuid.UUID           `json:"outbox_id"`
	NotificationID     *uuid.UUID          `json:"notification_id,omitempty"`
	Recipient          userSummaryResponse `json:"recipient"`
	ToEmail            string              `json:"to_email"`
	ToName             string              `json:"to_name"`
	Event              eventResponse       `json:"event"`
	Context            contextResponse     `json:"context"`
	Subject            string              `json:"subject"`
	Status             string              `json:"status"`
	Attempts           int                 `json:"attempts"`
	MaxAttempts        int                 `json:"max_attempts"`
	NextAttemptAt      *string             `json:"next_attempt_at,omitempty"`
	LastError          string              `json:"last_error,omitempty"`
	ProviderMessageID  string              `json:"provider_message_id,omitempty"`
	CreatedAt          string              `json:"created_at"`
	SentAt             *string             `json:"sent_at,omitempty"`
}

type outboxDetailResponse struct {
	outboxItemResponse
	BodyText  string  `json:"body_text"`
	BodyHTML  string  `json:"body_html,omitempty"`
	UpdatedAt string  `json:"updated_at"`
}

type outboxListResponse struct {
	Items []outboxItemResponse `json:"items"`
	Total int                  `json:"total"`
}

type retryResponse struct {
	OutboxID          uuid.UUID `json:"outbox_id"`
	RecipientUsername string    `json:"recipient_username"`
	EventType         string    `json:"event_type"`
	BusinessReference string    `json:"business_reference"`
	Status            string    `json:"status"`
	Attempts          int       `json:"attempts"`
	MaxAttempts       int       `json:"max_attempts"`
	NextAttemptAt     *string   `json:"next_attempt_at,omitempty"`
}

type testEmailRequest struct {
	ToUsername string `json:"to_username"`
	ToEmail    string `json:"to_email"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
}

type testEmailResponse struct {
	OutboxID  uuid.UUID           `json:"outbox_id"`
	Recipient userSummaryResponse `json:"recipient"`
	Event     eventResponse       `json:"event"`
	Status    string              `json:"status"`
}

type healthResponse struct {
	Enabled              bool     `json:"enabled"`
	WorkerEnabled        bool     `json:"worker_enabled"`
	SMTPHost             string   `json:"smtp_host"`
	SMTPPort             int      `json:"smtp_port"`
	SMTPTLSMode          string   `json:"smtp_tls_mode"`
	FromAddress          string   `json:"from_address"`
	FromName             string   `json:"from_name"`
	SendRealEmail        bool     `json:"send_real_email"`
	TestEndpointEnabled  bool     `json:"test_endpoint_enabled"`
	WorkerInterval       string   `json:"worker_interval"`
	WorkerBatchSize      int      `json:"worker_batch_size"`
	StaleSendingTimeout  string   `json:"stale_sending_timeout"`
	RetryPolicy          []string `json:"retry_policy"`
	PendingCount         int      `json:"pending_count"`
	FailedCount          int      `json:"failed_count"`
	DeadCount            int      `json:"dead_count"`
}

// ListOutbox handles GET /notifications/email-outbox
// @Summary List email outbox
// @Description Returns paged email delivery records for admin/operator use.
// @Tags Notification - Email Admin
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (PENDING, SENDING, SENT, FAILED, DEAD)"
// @Param recipient_username query string false "Filter by recipient username"
// @Param recipient_email query string false "Filter by recipient email"
// @Param event_type query string false "Filter by event type"
// @Param event_category query string false "Filter by event category"
// @Param business_type query string false "Filter by business type"
// @Param business_reference query string false "Filter by business reference"
// @Param created_from query string false "Filter created_at >= (RFC3339)"
// @Param created_to query string false "Filter created_at <= (RFC3339)"
// @Param limit query int false "Page size (default 50)"
// @Param offset query int false "Offset for paging"
// @Success 200 {object} outboxListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/email-outbox [get]
func (h *EmailOutboxHandler) ListOutbox(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	f := domain.EmailOutboxFilter{
		Status:            q.Get("status"),
		RecipientUsername: q.Get("recipient_username"),
		RecipientEmail:    q.Get("recipient_email"),
		EventType:         q.Get("event_type"),
		EventCategory:     q.Get("event_category"),
		BusinessType:      q.Get("business_type"),
		BusinessReference: q.Get("business_reference"),
		Limit:             limit,
		Offset:            offset,
	}
	if s := q.Get("created_from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			f.CreatedFrom = &t
		}
	}
	if s := q.Get("created_to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			f.CreatedTo = &t
		}
	}

	items, total, err := h.emailSvc.GetOutbox(r.Context(), f)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]outboxItemResponse, 0, len(items))
	for _, o := range items {
		out = append(out, toOutboxItem(o))
	}
	httputil.OK(w, outboxListResponse{Items: out, Total: total})
}

// GetOutboxDetail handles GET /notifications/email-outbox/{outbox_id}
// @Summary Get email outbox detail
// @Description Returns one email outbox record with full body for troubleshooting.
// @Tags Notification - Email Admin
// @Security BearerAuth
// @Produce json
// @Param outbox_id path string true "Outbox record UUID"
// @Success 200 {object} outboxDetailResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/email-outbox/{outbox_id} [get]
func (h *EmailOutboxHandler) GetOutboxDetail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "outbox_id"))
	if err != nil {
		httputil.BadRequest(w, "invalid outbox_id")
		return
	}
	o, err := h.emailSvc.GetOutboxByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if o == nil {
		httputil.NotFound(w, "outbox record not found")
		return
	}
	httputil.OK(w, toOutboxDetail(o))
}

// RetryOutbox handles POST /notifications/email-outbox/{outbox_id}/retry
// @Summary Retry failed email
// @Description Moves a FAILED or DEAD outbox row back to PENDING for another send attempt.
// @Tags Notification - Email Admin
// @Security BearerAuth
// @Produce json
// @Param outbox_id path string true "Outbox record UUID"
// @Success 200 {object} retryResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/email-outbox/{outbox_id}/retry [post]
func (h *EmailOutboxHandler) RetryOutbox(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "outbox_id"))
	if err != nil {
		httputil.BadRequest(w, "invalid outbox_id")
		return
	}
	o, err := h.emailSvc.RetryOutbox(r.Context(), id)
	if err != nil {
		if err.Error() == "email outbox: row not in FAILED or DEAD status" {
			httputil.UnprocessableEntity(w, "outbox row is not in a retryable state (must be FAILED or DEAD)")
			return
		}
		httputil.InternalError(w, err.Error())
		return
	}
	if o == nil {
		httputil.NotFound(w, "outbox record not found")
		return
	}
	var nextAttemptAt *string
	if !o.NextAttemptAt.IsZero() {
		s := o.NextAttemptAt.UTC().Format(time.RFC3339)
		nextAttemptAt = &s
	}
	httputil.OK(w, retryResponse{
		OutboxID:          o.ID,
		RecipientUsername: o.RecipientUsername,
		EventType:         o.EventType,
		BusinessReference: o.BusinessReference,
		Status:            o.Status,
		Attempts:          o.Attempts,
		MaxAttempts:       o.MaxAttempts,
		NextAttemptAt:     nextAttemptAt,
	})
}

// SendTestEmail handles POST /notifications/email/test
// @Summary Send test email
// @Description Creates a test email outbox row for SMTP verification. The background worker delivers it.
// @Tags Notification - Email Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body testEmailRequest true "Test email request"
// @Success 202 {object} testEmailResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/email/test [post]
func (h *EmailOutboxHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.EmailTestEndpointEnabled {
		httputil.NotFound(w, "test email endpoint is not enabled")
		return
	}

	var req testEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if req.Subject == "" {
		httputil.BadRequest(w, "subject is required")
		return
	}
	if req.ToUsername == "" && req.ToEmail == "" {
		httputil.BadRequest(w, "to_username or to_email is required")
		return
	}

	result, err := h.emailSvc.SendTestEmailIntent(r.Context(), service.TestEmailInput{
		ToUsername:     req.ToUsername,
		ToEmail:        req.ToEmail,
		Subject:        req.Subject,
		Body:           req.Body,
		AllowRawEmail:  h.cfg.EmailAllowRawEmail,
		AllowedDomains: h.cfg.EmailAllowedDomains,
		MaxAttempts:    h.cfg.EmailMaxAttempts,
	})
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	def := valueobject.LookupEventByString("EMAIL_TEST")
	var recipID *uuid.UUID
	if result.Recipient.ID != uuid.Nil {
		id := result.Recipient.ID
		recipID = &id
	}
	httputil.Created(w, testEmailResponse{
		OutboxID: result.OutboxID,
		Recipient: userSummaryResponse{
			ID:          recipID,
			Username:    result.Recipient.Username,
			DisplayName: result.Recipient.DisplayName,
			Email:       result.Recipient.Email,
		},
		Event: eventResponse{
			Type:     string(def.Type),
			Label:    def.Label,
			Category: def.Category,
			Severity: def.Severity,
		},
		Status: "QUEUED",
	})
}

// EmailHealth handles GET /notifications/email/health
// @Summary Email health
// @Description Returns email configuration and queue health without exposing secrets.
// @Tags Notification - Email Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} healthResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/email/health [get]
func (h *EmailOutboxHandler) EmailHealth(w http.ResponseWriter, r *http.Request) {
	pending, failed, dead, err := h.emailSvc.HealthCounts(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	retryPolicy := make([]string, len(h.cfg.EmailRetryPolicy))
	for i, d := range h.cfg.EmailRetryPolicy {
		retryPolicy[i] = d.String()
	}

	httputil.OK(w, healthResponse{
		Enabled:             h.cfg.EmailEnabled,
		WorkerEnabled:       h.cfg.EmailWorkerEnabled,
		SMTPHost:            h.cfg.SMTPHost,
		SMTPPort:            h.cfg.SMTPPort,
		SMTPTLSMode:         h.cfg.SMTPTLSMode,
		FromAddress:         h.cfg.SMTPFromAddress,
		FromName:            h.cfg.SMTPFromName,
		SendRealEmail:       h.cfg.EmailSendRealEmail,
		TestEndpointEnabled: h.cfg.EmailTestEndpointEnabled,
		WorkerInterval:      h.cfg.EmailWorkerInterval.String(),
		WorkerBatchSize:     h.cfg.EmailWorkerBatchSize,
		StaleSendingTimeout: h.cfg.EmailStaleSendingTimeout.String(),
		RetryPolicy:         retryPolicy,
		PendingCount:        pending,
		FailedCount:         failed,
		DeadCount:           dead,
	})
}

// --- helpers ---

func toOutboxItem(o *entity.EmailOutbox) outboxItemResponse {
	def := valueobject.LookupEventByString(o.EventType)

	var recipID *uuid.UUID
	if o.RecipientUserID != nil {
		recipID = o.RecipientUserID
	}

	var nextAttemptAt *string
	if !o.NextAttemptAt.IsZero() {
		s := o.NextAttemptAt.UTC().Format(time.RFC3339)
		nextAttemptAt = &s
	}
	var sentAt *string
	if o.SentAt != nil {
		s := o.SentAt.UTC().Format(time.RFC3339)
		sentAt = &s
	}

	return outboxItemResponse{
		OutboxID:       o.ID,
		NotificationID: o.NotificationID,
		Recipient: userSummaryResponse{
			ID:          recipID,
			Username:    o.RecipientUsername,
			DisplayName: o.RecipientDisplayName,
			Email:       o.RecipientEmail,
		},
		ToEmail: o.ToEmail,
		ToName:  o.ToName,
		Event: eventResponse{
			Type:     string(def.Type),
			Label:    def.Label,
			Category: def.Category,
			Severity: def.Severity,
		},
		Context: contextResponse{
			BusinessType:      o.BusinessType,
			BusinessLabel:     o.BusinessLabel,
			BusinessID:        o.BusinessID,
			BusinessReference: o.BusinessReference,
			BusinessTitle:     o.BusinessTitle,
		},
		Subject:           o.Subject,
		Status:            o.Status,
		Attempts:          o.Attempts,
		MaxAttempts:       o.MaxAttempts,
		NextAttemptAt:     nextAttemptAt,
		LastError:         o.LastError,
		ProviderMessageID: o.ProviderMessageID,
		CreatedAt:         o.CreatedAt.UTC().Format(time.RFC3339),
		SentAt:            sentAt,
	}
}

func toOutboxDetail(o *entity.EmailOutbox) outboxDetailResponse {
	return outboxDetailResponse{
		outboxItemResponse: toOutboxItem(o),
		BodyText:           o.BodyText,
		BodyHTML:           o.BodyHTML,
		UpdatedAt:          o.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

