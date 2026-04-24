package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
)

// TestWriteOverrideError verifies that each override-flow error type maps to
// the correct HTTP status code. This is the HTTP-contract surface of the
// atomic-override fix: genuine server-side failures must NOT be misclassified
// as client errors.
func TestWriteOverrideError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "invalid request → 400",
			err:        &domain.ErrInvalidOverrideRequest{Field: "reason"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "breach not found → 404",
			err:        &domain.ErrBreachNotFound{BreachID: "abc"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "breach not open → 409",
			err:        &domain.ErrBreachNotOpen{BreachID: "abc", CurrentStatus: "RESOLVED"},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "override already exists → 409",
			err:        &domain.ErrOverrideAlreadyExists{BreachID: "abc"},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "unexpected internal error → 500 (not 400)",
			err:        errors.New("connection reset by peer"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped internal error → 500 (not 400)",
			err:        errors.Join(errors.New("begin tx"), errors.New("io timeout")),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			writeOverrideError(rec, tc.err)
			if rec.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body=%s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

// TestWriteOverrideError_InternalDoesNotLeakDetail ensures the 500 branch
// returns a generic message rather than forwarding internal error text that
// could reveal infrastructure shape to unauthenticated observers.
func TestWriteOverrideError_InternalDoesNotLeakDetail(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	writeOverrideError(rec, errors.New("pgx: connection to host 10.0.2.5:5437 refused"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if contains(body, "10.0.2.5") || contains(body, "5437") || contains(body, "pgx") {
		t.Errorf("response body leaks internal detail: %s", body)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

// tiny substring search (stdlib strings.Contains would be identical; kept
// local to avoid an import for one call).
func indexOf(s, substr string) int {
	n, m := len(s), len(substr)
	if m == 0 {
		return 0
	}
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == substr {
			return i
		}
	}
	return -1
}
