package httputil

import (
	"net/http"
	"strconv"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Pagination holds parsed pagination parameters from query string.
type Pagination struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"-"`
}

// PaginatedResponse wraps paginated data with metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// ParsePagination extracts page and limit from query parameters with defaults and bounds.
func ParsePagination(r *http.Request) Pagination {
	page := parseIntParam(r, "page", DefaultPage)
	limit := parseIntParam(r, "limit", DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
