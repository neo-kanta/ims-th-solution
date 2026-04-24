package handler

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// AuditHandler handles audit query endpoints.
type AuditHandler struct {
	listAuditQry *query.ListAuditEventsQuery
	recorder     auditdomain.Recorder
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(listAuditQry *query.ListAuditEventsQuery, recorder auditdomain.Recorder) *AuditHandler {
	if listAuditQry == nil {
		panic("audit handler requires list query")
	}
	if recorder == nil {
		panic("audit handler requires recorder")
	}
	return &AuditHandler{
		listAuditQry: listAuditQry,
		recorder:     recorder,
	}
}

// ListAuditEvents handles GET /admin/audit.
// @Summary List Audit Events
// @Description Query paginated audit events with filters
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
	filter, err := h.parseFilter(r, 50, 500)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

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

	h.recordAuditAccess(r, entity.AuditLogViewed, map[string]interface{}{
		"returned_count": len(events),
		"total":          total,
		"filters":        filterMetadata(filter),
	})

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
	filter, err := h.parseFilter(r, 10000, 10000)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	events, total, err := h.listAuditQry.ExecuteForExport(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, "failed to export audit events")
		return
	}

	payload, err := buildAuditCSV(events)
	if err != nil {
		slog.Error("failed to build audit csv export", "error", err)
		httputil.InternalError(w, "failed to export audit events")
		return
	}

	filename := fmt.Sprintf("audit_events_%s.csv", time.Now().UTC().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if _, err := w.Write(payload); err != nil {
		slog.Error("failed to write audit csv export", "error", err)
		return
	}

	h.recordAuditAccess(r, entity.AuditLogExported, map[string]interface{}{
		"exported_count": len(events),
		"total":          total,
		"filters":        filterMetadata(filter),
		"filename":       filename,
	})
}

func (h *AuditHandler) parseFilter(r *http.Request, defaultLimit, maxLimit int) (domain.AuditFilter, error) {
	filter := domain.AuditFilter{}

	if actorIDStr := r.URL.Query().Get("actor_id"); actorIDStr != "" {
		uid, err := uuid.Parse(actorIDStr)
		if err != nil {
			return domain.AuditFilter{}, fmt.Errorf("invalid actor_id")
		}
		filter.ActorID = &uid
	}
	filter.EventType = r.URL.Query().Get("event_type")
	filter.TargetType = r.URL.Query().Get("target_type")
	filter.TargetID = r.URL.Query().Get("target_id")

	if since := r.URL.Query().Get("since"); since != "" {
		parsed, err := parseAuditTime(since)
		if err != nil {
			return domain.AuditFilter{}, fmt.Errorf("invalid since")
		}
		filter.Since = &parsed
	}
	if until := r.URL.Query().Get("until"); until != "" {
		parsed, err := parseAuditTime(until)
		if err != nil {
			return domain.AuditFilter{}, fmt.Errorf("invalid until")
		}
		filter.Until = &parsed
	}

	if filter.Since != nil && filter.Until != nil && filter.Until.Before(*filter.Since) {
		return domain.AuditFilter{}, fmt.Errorf("until must be greater than or equal to since")
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return domain.AuditFilter{}, fmt.Errorf("invalid offset")
		}
		filter.Offset = offset
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return domain.AuditFilter{}, fmt.Errorf("invalid limit")
		}
		filter.Limit = limit
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultLimit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}

	return filter, nil
}

func (h *AuditHandler) recordAuditAccess(r *http.Request, eventType string, metadata map[string]interface{}) {
	actorID := currentActorID(r)
	h.recorder.Record(r.Context(), actorID, eventType, "audit", "iam_audit_events", clientIP(r), r.UserAgent(), metadata)
}

func currentActorID(r *http.Request) *uuid.UUID {
	claims := platformmw.GetUserClaims(r.Context())
	if claims == nil {
		return nil
	}
	actorID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil
	}
	return &actorID
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	return platformmw.GetClientIP(r)
}

func filterMetadata(filter domain.AuditFilter) map[string]interface{} {
	metadata := map[string]interface{}{
		"offset": filter.Offset,
		"limit":  filter.Limit,
	}
	if filter.ActorID != nil {
		metadata["actor_id"] = filter.ActorID.String()
	}
	if filter.EventType != "" {
		metadata["event_type"] = filter.EventType
	}
	if filter.TargetType != "" {
		metadata["target_type"] = filter.TargetType
	}
	if filter.TargetID != "" {
		metadata["target_id"] = filter.TargetID
	}
	if filter.Since != nil {
		metadata["since"] = filter.Since.UTC().Format(time.RFC3339)
	}
	if filter.Until != nil {
		metadata["until"] = filter.Until.UTC().Format(time.RFC3339)
	}
	return metadata
}

func parseAuditTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, errors.New("invalid time")
}

func buildAuditCSV(events []entity.AuditEvent) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return nil, fmt.Errorf("write utf-8 bom: %w", err)
	}

	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "actor_id", "event_type", "target_type", "target_id", "ip_address", "user_agent", "metadata", "created_at"}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, e := range events {
		actorID := ""
		if e.ActorID != nil {
			actorID = e.ActorID.String()
		}
		metaStr := ""
		if e.Metadata != nil {
			if b, err := json.Marshal(e.Metadata); err == nil {
				metaStr = string(b)
			}
		}

		if err := writer.Write([]string{
			sanitizeCSVCell(e.ID.String()),
			sanitizeCSVCell(actorID),
			sanitizeCSVCell(e.EventType),
			sanitizeCSVCell(e.TargetType),
			sanitizeCSVCell(e.TargetID),
			sanitizeCSVCell(e.IPAddress),
			sanitizeCSVCell(e.UserAgent),
			sanitizeCSVCell(metaStr),
			sanitizeCSVCell(e.CreatedAt.UTC().Format(time.RFC3339)),
		}); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv writer: %w", err)
	}

	return buf.Bytes(), nil
}

func sanitizeCSVCell(value string) string {
	if value == "" {
		return value
	}
	trimmed := strings.TrimLeft(value, " \t")
	if trimmed == "" {
		return value
	}

	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}
