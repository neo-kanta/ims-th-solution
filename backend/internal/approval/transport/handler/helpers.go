// Package handler holds the approval module's thin HTTP handlers: parse → call
// service → format. Business rules live in the application/service layer.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// actorID extracts the authenticated user's UUID from the request context.
func actorID(r *http.Request) (uuid.UUID, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// parseUUIDParam extracts a UUID-shaped URL parameter.
func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}

// pathParam returns a raw URL path parameter.
func pathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// decodeJSON decodes a JSON request body into v.
func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

// optUUID parses an optional UUID string ("" → nil).
func optUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// optDate parses an optional YYYY-MM-DD date ("" → nil).
func optDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// boolOr returns *p when set, else def.
func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// pagination reads page/limit from the query string with clamping.
func pagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return page, limit
}

// writeError maps approval domain errors to HTTP status codes uniformly.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, domain.ErrValidation):
		httputil.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		httputil.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrConfigNotFound):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.Is(err, domain.ErrForbidden),
		errors.Is(err, domain.ErrSelfApproval),
		errors.Is(err, domain.ErrNotAssigned):
		httputil.Forbidden(w, err.Error())
	case errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrDuplicateActiveRequest),
		errors.Is(err, domain.ErrTaskNotPending),
		errors.Is(err, domain.ErrStaleTask),
		errors.Is(err, domain.ErrRequestNotActionable):
		httputil.Conflict(w, err.Error())
	default:
		httputil.InternalError(w, "an unexpected error occurred")
	}
}
