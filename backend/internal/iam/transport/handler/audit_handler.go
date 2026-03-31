package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// AuditHandler handles audit query endpoints.
type AuditHandler struct {
	listAuditQry *query.ListAuditEventsQuery
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(listAuditQry *query.ListAuditEventsQuery) *AuditHandler {
	return &AuditHandler{listAuditQry: listAuditQry}
}

// ListAuditEvents handles GET /admin/audit.
// @Summary List Audit Events
// @Description Query paginated IAM audit events with filters
// @Tags Audit
// @Security BearerAuth
// @Produce json
// @Param actor_id query string false "Actor UUID"
// @Param event_type query string false "Event type (e.g., LOGIN_SUCCESS)"
// @Param target_type query string false "Target type (e.g., user, session)"
// @Param target_id query string false "Target ID"
// @Param since query string false "Since datetime (ISO 8601)"
// @Param until query string false "Until datetime (ISO 8601)"
// @Param offset query int false "Offset (default 0)"
// @Param limit query int false "Limit (default 50, max 500)"
// @Success 200 {object} response.AuditListResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/audit [get]
func (h *AuditHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)

	events, total, err := h.listAuditQry.Execute(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, "failed to query audit events")
		return
	}

	result := response.AuditListResponse{
		Events: make([]response.AuditEventResponse, 0, len(events)),
		Total:  total,
		Offset: filter.Offset,
		Limit:  filter.Limit,
	}
	for _, e := range events {
		var actorID *string
		if e.ActorID != nil {
			s := e.ActorID.String()
			actorID = &s
		}
		result.Events = append(result.Events, response.AuditEventResponse{
			ID:         e.ID.String(),
			ActorID:    actorID,
			EventType:  e.EventType,
			TargetType: e.TargetType,
			TargetID:   e.TargetID,
			IPAddress:  e.IPAddress,
			UserAgent:  e.UserAgent,
			Metadata:   e.Metadata,
			CreatedAt:  e.CreatedAt,
		})
	}

	httputil.OK(w, result)
}

// ExportAuditCSV handles GET /admin/audit/export.
// @Summary Export Audit Events as CSV
// @Description Export audit events matching filter criteria as CSV download
// @Tags Audit
// @Security BearerAuth
// @Produce text/csv
// @Param actor_id query string false "Actor UUID"
// @Param event_type query string false "Event type"
// @Param since query string false "Since datetime"
// @Param until query string false "Until datetime"
// @Success 200 {file} csv
// @Router /admin/audit/export [get]
func (h *AuditHandler) ExportAuditCSV(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)
	filter.Limit = 10000 // max export size
	filter.Offset = 0

	events, _, err := h.listAuditQry.Execute(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, "failed to export audit events")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=audit_events.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header row
	_ = writer.Write([]string{"id", "actor_id", "event_type", "target_type", "target_id", "ip_address", "user_agent", "created_at"})

	for _, e := range events {
		actorID := ""
		if e.ActorID != nil {
			actorID = e.ActorID.String()
		}
		_ = writer.Write([]string{
			e.ID.String(),
			actorID,
			e.EventType,
			e.TargetType,
			e.TargetID,
			e.IPAddress,
			e.UserAgent,
			e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
}

func (h *AuditHandler) parseFilter(r *http.Request) domain.AuditFilter {
	filter := domain.AuditFilter{}

	if actorIDStr := r.URL.Query().Get("actor_id"); actorIDStr != "" {
		if uid, err := uuid.Parse(actorIDStr); err == nil {
			filter.ActorID = &uid
		}
	}
	filter.EventType = r.URL.Query().Get("event_type")
	filter.TargetType = r.URL.Query().Get("target_type")
	filter.TargetID = r.URL.Query().Get("target_id")

	if since := r.URL.Query().Get("since"); since != "" {
		filter.Since = &since
	}
	if until := r.URL.Query().Get("until"); until != "" {
		filter.Until = &until
	}

	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	return filter
}
