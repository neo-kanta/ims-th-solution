package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

type testErr struct {
	code    string
	msg     string
	status  int
	details map[string]any
}

func (e *testErr) Error() string                { return e.msg }
func (e *testErr) ErrorCode() string            { return e.code }
func (e *testErr) ErrorDetails() map[string]any { return e.details }
func (e *testErr) HTTPStatus() int              { return e.status }

func newRequest() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	ctx := context.WithValue(r.Context(), middleware.RequestIDKey, "test-req-1")
	return r.WithContext(ctx)
}

func decodeEnvelope(t *testing.T, body string) ErrorEnvelope {
	t.Helper()
	var env ErrorEnvelope
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, body)
	}
	return env
}

func TestWriteDomainError_FullCodedError(t *testing.T) {
	t.Parallel()
	err := &testErr{code: errcode.CodeOversell, msg: "no inventory", details: map[string]any{"available": 50}}

	w := httptest.NewRecorder()
	WriteDomainError(w, newRequest(), err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
	env := decodeEnvelope(t, w.Body.String())
	if env.ErrorCode != errcode.CodeOversell {
		t.Errorf("error_code = %q, want OVERSELL", env.ErrorCode)
	}
	if env.Message != "no inventory" {
		t.Errorf("message = %q, want %q", env.Message, "no inventory")
	}
	if env.Details["available"] == nil {
		t.Errorf("details lost: %+v", env.Details)
	}
	if env.RequestID != "test-req-1" {
		t.Errorf("request_id = %q, want test-req-1", env.RequestID)
	}
}

func TestWriteDomainError_RespectsHTTPStatusOverride(t *testing.T) {
	t.Parallel()
	err := &testErr{code: errcode.CodeOversell, msg: "x", status: http.StatusTeapot}
	w := httptest.NewRecorder()
	WriteDomainError(w, newRequest(), err)
	if w.Code != http.StatusTeapot {
		t.Errorf("status = %d, want 418", w.Code)
	}
}

func TestWriteDomainError_5xxScrubsMessageAndDetails(t *testing.T) {
	t.Parallel()
	plain := errors.New("database connection refused with credentials xyz")
	w := httptest.NewRecorder()
	WriteDomainError(w, newRequest(), plain)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	env := decodeEnvelope(t, w.Body.String())
	if env.ErrorCode != errcode.CodeInternal {
		t.Errorf("error_code = %q, want INTERNAL_ERROR", env.ErrorCode)
	}
	if env.Message == "" || strings.Contains(env.Message, "credentials") {
		t.Errorf("5xx message must not leak internals; got %q", env.Message)
	}
	if env.Details != nil {
		t.Errorf("5xx must drop details, got %+v", env.Details)
	}
	if env.RequestID != "test-req-1" {
		t.Errorf("request_id missing on 5xx envelope")
	}
}

func TestWriteDomainError_NilIsNoop(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	WriteDomainError(w, newRequest(), nil)
	if w.Code != http.StatusOK {
		t.Errorf("nil error must not flush a response; got status %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("nil error must not write body; got %q", w.Body.String())
	}
}

func TestWriteDomainError_NonCodedFallsBackToInternal(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	WriteDomainError(w, newRequest(), errors.New("anything"))
	env := decodeEnvelope(t, w.Body.String())
	if env.ErrorCode != errcode.CodeInternal {
		t.Errorf("error_code = %q, want INTERNAL_ERROR", env.ErrorCode)
	}
}

func TestWriteEnvelope_EmitsAllFields(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	WriteEnvelope(w, newRequest(), http.StatusServiceUnavailable, errcode.CodeInternal, "not ready", map[string]any{"db": "down"})
	env := decodeEnvelope(t, w.Body.String())
	if env.ErrorCode != errcode.CodeInternal || env.Message != "not ready" || env.Details["db"] != "down" {
		t.Errorf("envelope fields not preserved: %+v", env)
	}
}
