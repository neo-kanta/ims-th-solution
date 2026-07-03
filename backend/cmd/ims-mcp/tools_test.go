package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestResolveFundID(t *testing.T) {
	// Setup a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/investment/funds" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Error("missing authorization header")
		}

		// Return mock funds list
		resp := map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "d0001000-0000-0000-0000-000000000002", "code": "BBL-EQUITY", "name": "Large-cap Momentum — Equity"},
				{"id": "d0001000-0000-0000-0000-000000000003", "code": "GLOBAL-TECH", "name": "Global Tech Thematic"},
				{"id": "d0001000-0000-0000-0000-000000000004", "code": "KTB-BALANCED", "name": "Balanced — Quarterly Rebalance Q2"},
			},
			"total": 3,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	deps := &toolDeps{ims: newIMSClient(server.URL)}

	// Helper function to create standard request with injected token metadata
	newRequestWithMeta := func() *mcp.CallToolRequest {
		req := &mcp.CallToolRequest{}
		req.Params = &mcp.CallToolParamsRaw{
			Meta: mcp.Meta{
				authMetaKey: "mock-token",
			},
		}
		return req
	}

	ctx := context.Background()

	// 1. UUID passes through directly
	uuidInput := "d0001000-0000-0000-0000-000000000002"
	resolved, err := deps.resolveFundID(ctx, newRequestWithMeta(), uuidInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != uuidInput {
		t.Errorf("expected %q, got %q", uuidInput, resolved)
	}

	// 2. Exact match on Code
	resolved, err = deps.resolveFundID(ctx, newRequestWithMeta(), "BBL-EQUITY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "d0001000-0000-0000-0000-000000000002" {
		t.Errorf("expected UUID 02, got %q", resolved)
	}

	// 3. Exact match on Name
	resolved, err = deps.resolveFundID(ctx, newRequestWithMeta(), "Balanced — Quarterly Rebalance Q2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "d0001000-0000-0000-0000-000000000004" {
		t.Errorf("expected UUID 04, got %q", resolved)
	}

	// 4. Substring match (input contains Name)
	resolved, err = deps.resolveFundID(ctx, newRequestWithMeta(), "Balanced — Quarterly Rebalance Q2 (KTB-BALANCED) — THB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "d0001000-0000-0000-0000-000000000004" {
		t.Errorf("expected UUID 04, got %q", resolved)
	}

	// 5. Substring match (Name contains input)
	resolved, err = deps.resolveFundID(ctx, newRequestWithMeta(), "Balanced")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "d0001000-0000-0000-0000-000000000004" {
		t.Errorf("expected UUID 04, got %q", resolved)
	}

	// 6. No match
	_, err = deps.resolveFundID(ctx, newRequestWithMeta(), "NON-EXISTENT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
