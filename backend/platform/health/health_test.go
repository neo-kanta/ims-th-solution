package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

func decodeJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&m); err != nil {
		t.Fatalf("decode: %v body=%s", err, body)
	}
	return m
}

func TestLive_AlwaysReadyWhenNotShuttingDown(t *testing.T) {
	t.Parallel()
	svc := NewService(nil, 0)
	w := httptest.NewRecorder()
	svc.Live(w, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := decodeJSON(t, w.Body.String())
	if body["status"] != "alive" {
		t.Errorf("status field = %v, want alive", body["status"])
	}
}

func TestLive_WhenShuttingDownReturns503(t *testing.T) {
	t.Parallel()
	svc := NewService(nil, 0)
	svc.MarkShuttingDown()
	w := httptest.NewRecorder()
	svc.Live(w, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	body := decodeJSON(t, w.Body.String())
	if body["error_code"] != errcode.CodeInternal {
		t.Errorf("error_code = %v", body["error_code"])
	}
}

func TestReady_NoProbesNoPoolReturns200(t *testing.T) {
	t.Parallel()
	svc := NewService(nil, 0)
	w := httptest.NewRecorder()
	svc.Ready(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestReady_FailingProbeReturns503WithDetails(t *testing.T) {
	t.Parallel()
	svc := NewService(nil, 0)
	svc.Register(ProbeFunc{
		NameValue: "compliance",
		ReadyFunc: func(ctx context.Context) error { return errors.New("compliance backend down") },
	})
	svc.Register(ProbeFunc{
		NameValue: "workflow",
		ReadyFunc: func(ctx context.Context) error { return nil },
	})

	w := httptest.NewRecorder()
	svc.Ready(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	body := decodeJSON(t, w.Body.String())
	details, ok := body["details"].(map[string]any)
	if !ok {
		t.Fatalf("expected details map, got %T", body["details"])
	}
	if _, hasFail := details["compliance"]; !hasFail {
		t.Errorf("details should report compliance failure: %+v", details)
	}
	if _, hasOK := details["workflow"]; hasOK {
		t.Errorf("ok probes must not appear in details: %+v", details)
	}
}

func TestReady_ShuttingDownShortCircuits(t *testing.T) {
	t.Parallel()
	svc := NewService(nil, 0)
	svc.MarkShuttingDown()
	probeCalled := false
	svc.Register(ProbeFunc{
		NameValue: "x",
		ReadyFunc: func(ctx context.Context) error { probeCalled = true; return nil },
	})

	w := httptest.NewRecorder()
	svc.Ready(w, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	if probeCalled {
		t.Errorf("shutting-down readiness must not run probes")
	}
}

func TestProbeFunc_NilReadyFuncReturnsOK(t *testing.T) {
	t.Parallel()
	p := ProbeFunc{NameValue: "x"}
	if err := p.Ready(context.Background()); err != nil {
		t.Errorf("nil ReadyFunc must return nil; got %v", err)
	}
}
