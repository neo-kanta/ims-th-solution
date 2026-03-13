package types

// Pagination holds pagination parameters for list queries.
type Pagination struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"-"`
}

// PagedResult wraps a result set with pagination metadata.
type PagedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// NewPagedResult creates a PagedResult with computed total pages.
func NewPagedResult[T any](items []T, total int64, page, limit int) PagedResult[T] {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return PagedResult[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
