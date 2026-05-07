package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ErrorEnvelope is the stable JSON shape every API error response uses.
// Frontend consumers MUST switch on ErrorCode; Message is for humans only.
//
// Shape locked at Phase 6 — adding a field is a non-breaking change, but
// renaming or removing one breaks every downstream consumer.
type ErrorEnvelope struct {
	ErrorCode string         `json:"error_code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id"`
}

// requestIDFrom retrieves the chi-injected request ID from the request
// context, falling back to the X-Request-Id response header (already set
// by chi's middleware) and finally the empty string. The empty case
// should never occur in production wiring; tests may exercise it.
func requestIDFrom(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := middleware.GetReqID(r.Context()); id != "" {
		return id
	}
	if r.Response != nil {
		if id := r.Response.Header.Get("X-Request-Id"); id != "" {
			return id
		}
	}
	return r.Header.Get("X-Request-Id")
}

// WriteDomainError serialises a typed domain error into the canonical
// envelope and writes it to w with the appropriate HTTP status. Any error
// reaching this function without the Coded interface is logged and
// surfaced as INTERNAL_ERROR with a generic message.
//
// On 5xx responses, the underlying error message is logged with the
// request id and never returned to the caller — leaking internals at
// the edge has been a recurring incident class in similar systems.
func WriteDomainError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	requestID := requestIDFrom(r)

	code, _ := errcode.CodeOf(err)
	status := errcode.StatusOf(err)
	details := errcode.DetailsOf(err)

	env := ErrorEnvelope{
		ErrorCode: code,
		Details:   details,
		RequestID: requestID,
	}
	if status >= 500 {
		slog.ErrorContext(
			r.Context(),
			"5xx response",
			"request_id", requestID,
			"path", r.URL.Path,
			"error", err.Error(),
		)
		env.ErrorCode = errcode.CodeInternal
		env.Message = "internal server error"
		env.Details = nil
	} else {
		env.Message = err.Error()
	}

	writeEnvelope(w, status, env)
}

// WriteEnvelope writes an arbitrary envelope at the given status. Useful
// for endpoints that want to construct the envelope directly (e.g.,
// health checks). Most callers should use WriteDomainError instead.
func WriteEnvelope(w http.ResponseWriter, r *http.Request, status int, code, message string, details map[string]any) {
	env := ErrorEnvelope{
		ErrorCode: code,
		Message:   message,
		Details:   details,
		RequestID: requestIDFrom(r),
	}
	writeEnvelope(w, status, env)
}

func writeEnvelope(w http.ResponseWriter, status int, env ErrorEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(env); err != nil {
		// At this point the headers are flushed; we cannot recover.
		slog.Error("error envelope encoding failed", "error", err)
	}
}
