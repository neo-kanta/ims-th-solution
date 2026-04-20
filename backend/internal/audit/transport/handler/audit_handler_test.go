package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type auditHandlerTestRepo struct {
	listEvents []entity.AuditEvent
	recorded   []*entity.AuditEvent
}

type failingResponseWriter struct {
	header http.Header
}

func (w *failingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(int) {}

func (r *auditHandlerTestRepo) Record(_ context.Context, event *entity.AuditEvent) error {
	r.recorded = append(r.recorded, event)
	return nil
}

func (r *auditHandlerTestRepo) List(context.Context, domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	events := make([]entity.AuditEvent, len(r.listEvents))
	copy(events, r.listEvents)
	return events, len(events), nil
}

func TestListAuditEventsRejectsInvalidActorID(t *testing.T) {
	repo := &auditHandlerTestRepo{}
	handler := NewAuditHandler(query.NewListAuditEventsQuery(repo), service.NewRecorder(repo))

	req := httptest.NewRequest(http.MethodGet, "/admin/audit?actor_id=not-a-uuid", nil)
	rr := httptest.NewRecorder()

	handler.ListAuditEvents(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "invalid actor_id")
	require.Empty(t, repo.recorded)
}

func TestExportAuditCSVSanitizesFormulaCellsAndSelfAudits(t *testing.T) {
	repo := &auditHandlerTestRepo{
		listEvents: []entity.AuditEvent{
			{
				ID:         uuid.New(),
				EventType:  entity.AuditLoginFailure,
				TargetType: "user",
				TargetID:   "=SUM(1,1)",
				IPAddress:  "127.0.0.1",
				UserAgent:  "+cmd",
				Metadata: map[string]interface{}{
					"note": "@danger",
				},
				CreatedAt: time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := NewAuditHandler(query.NewListAuditEventsQuery(repo), service.NewRecorder(repo))

	actorID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/audit/export", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req = req.WithContext(context.WithValue(req.Context(), platformmw.UserContextKey, &platformmw.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: actorID.String()},
	}))
	rr := httptest.NewRecorder()

	handler.ExportAuditCSV(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	require.True(t, strings.Contains(body, "'=SUM(1,1)"))
	require.True(t, strings.Contains(body, "'+cmd"))
	require.Contains(t, body, "@danger")
	require.Len(t, repo.recorded, 1)
	require.Equal(t, entity.AuditLogExported, repo.recorded[0].EventType)
	require.Equal(t, actorID, *repo.recorded[0].ActorID)
}

func TestExportAuditCSVDoesNotSelfAuditWhenResponseWriteFails(t *testing.T) {
	repo := &auditHandlerTestRepo{
		listEvents: []entity.AuditEvent{
			{
				ID:        uuid.New(),
				EventType: entity.AuditLoginFailure,
				CreatedAt: time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := NewAuditHandler(query.NewListAuditEventsQuery(repo), service.NewRecorder(repo))

	req := httptest.NewRequest(http.MethodGet, "/admin/audit/export", nil)
	writer := &failingResponseWriter{}

	handler.ExportAuditCSV(writer, req)

	require.Empty(t, repo.recorded)
}

func TestNewAuditHandlerPanicsWithoutRecorder(t *testing.T) {
	repo := &auditHandlerTestRepo{}

	require.PanicsWithValue(t, "audit handler requires recorder", func() {
		NewAuditHandler(query.NewListAuditEventsQuery(repo), nil)
	})
}
