package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// This file holds transport-layer helpers shared across every handler in
// the investment module. They were originally defined in
// investment_handler.go, but ResearchReportHandler (and any future split
// handlers) need them too — moving them here makes the dependency
// explicit and stops the helper definitions from drifting alongside a
// single handler's lifecycle.

// actorID extracts the authenticated user's UUID from the request context.
// Returns (uuid.Nil, false) when the request is anonymous or the subject
// claim is malformed.
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

// parseUUIDParam extracts a UUID-shaped URL parameter (e.g. {id}).
func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	v := chi.URLParam(r, key)
	return uuid.Parse(v)
}

// parseUUID parses a UUID string, returning a typed error on failure.
func parseUUID(v string) (uuid.UUID, error) {
	return uuid.Parse(v)
}

// parseDate parses a YYYY-MM-DD date string. Empty input is treated as an
// error so callers wanting nullable dates should use parseDateOpt instead.
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	return time.Parse("2006-01-02", s)
}

// parseDateOpt parses an optional YYYY-MM-DD date string. Empty input
// returns (nil, nil) — the caller decides whether nil is allowed.
func parseDateOpt(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := parseDate(s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// parseDecimalOpt parses an optional decimal string. Empty input returns
// (nil, nil); a malformed value bubbles up.
func parseDecimalOpt(s string) (*decimal.Decimal, error) {
	if s == "" {
		return nil, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// parseDecimalRequiredOrZero parses a decimal that the caller treats as a
// required field but defaults to zero when blank.
func parseDecimalRequiredOrZero(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(s)
}

// paginationParams reads page/limit from the URL query and clamps them.
// Defaults: page=1, limit=50, hard ceiling limit=200.
func paginationParams(r *http.Request) (page, limit int) {
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
