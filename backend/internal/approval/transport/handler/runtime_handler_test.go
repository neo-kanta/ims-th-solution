package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// P0-1 regression: ListRequests must return 401 when no actor context is set.
// Previously actor, _ := actorID(r) silently defaulted to uuid.Nil, which the
// service treats as a system bypass returning all requests unfiltered.

func TestListRequests_NoActor_Returns401(t *testing.T) {
	// Handler with nil service — the 401 guard fires before svc is touched.
	h := NewRuntimeHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/approvals/requests", nil)
	// No UserClaims injected into context → actorID returns ok=false
	rr := httptest.NewRecorder()
	h.ListRequests(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ListRequests without actor = %d, want %d (Unauthorized)",
			rr.Code, http.StatusUnauthorized)
	}
}
