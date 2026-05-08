package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// recordingExecutor is a minimal in-memory catalogExecutor that captures
// every Exec call so a test can assert what would have been written.
type recordingExecutor struct {
	calls   []recordedExec
	failOn  string
	failErr error
}

type recordedExec struct {
	sql  string
	args []any
}

func (r *recordingExecutor) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	r.calls = append(r.calls, recordedExec{sql: sql, args: args})
	if r.failOn != "" && len(args) > 0 {
		if code, ok := args[0].(string); ok && code == r.failOn {
			return pgconn.CommandTag{}, r.failErr
		}
	}
	return pgconn.CommandTag{}, nil
}

type fakeCatalog struct {
	module string
	defs   []contract.PermissionDefinition
}

func (f fakeCatalog) Module() string                                   { return f.module }
func (f fakeCatalog) Permissions() []contract.PermissionDefinition     { return f.defs }

func TestAggregateCatalogs_Sorted(t *testing.T) {
	t.Parallel()
	catalogs := []contract.PermissionCatalog{
		fakeCatalog{module: "alpha", defs: []contract.PermissionDefinition{
			{Code: "ALPHA_TWO", Name: "Alpha Two", Description: "second"},
			{Code: "ALPHA_ONE", Name: "Alpha One", Description: "first"},
		}},
		fakeCatalog{module: "beta", defs: []contract.PermissionDefinition{
			{Code: "BETA_X", Name: "Beta X"},
		}},
	}

	entries, err := aggregateCatalogs(catalogs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	wantOrder := []string{"ALPHA_ONE", "ALPHA_TWO", "BETA_X"}
	for i, want := range wantOrder {
		if entries[i].Code != want {
			t.Errorf("entries[%d].Code = %q, want %q", i, entries[i].Code, want)
		}
	}
	if entries[0].Module != "alpha" || entries[2].Module != "beta" {
		t.Errorf("module attribution wrong: %+v", entries)
	}
}

func TestAggregateCatalogs_DuplicateAcrossModulesIsError(t *testing.T) {
	t.Parallel()
	catalogs := []contract.PermissionCatalog{
		fakeCatalog{module: "alpha", defs: []contract.PermissionDefinition{
			{Code: "SHARED", Name: "Shared"},
		}},
		fakeCatalog{module: "beta", defs: []contract.PermissionDefinition{
			{Code: "SHARED", Name: "Shared"},
		}},
	}

	_, err := aggregateCatalogs(catalogs)
	if err == nil {
		t.Fatalf("expected duplicate-code error, got nil")
	}
	if !strings.Contains(err.Error(), `permission code "SHARED"`) {
		t.Errorf("error message should cite the duplicate code: %v", err)
	}
}

func TestAggregateCatalogs_DuplicateWithinModuleIsLastWins(t *testing.T) {
	t.Parallel()
	catalogs := []contract.PermissionCatalog{
		fakeCatalog{module: "alpha", defs: []contract.PermissionDefinition{
			{Code: "DUP", Name: "First", Description: "first"},
			{Code: "DUP", Name: "Second", Description: "second"},
		}},
	}

	entries, err := aggregateCatalogs(catalogs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "Second" {
		t.Errorf("expected last-wins within module, got %q", entries[0].Name)
	}
}

func TestAggregateCatalogs_RejectsEmptyFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		catalog contract.PermissionCatalog
		want    string
	}{
		{
			name:    "empty module",
			catalog: fakeCatalog{module: "", defs: []contract.PermissionDefinition{{Code: "X", Name: "X"}}},
			want:    "empty module name",
		},
		{
			name:    "empty code",
			catalog: fakeCatalog{module: "m", defs: []contract.PermissionDefinition{{Code: "", Name: "X"}}},
			want:    "empty code",
		},
		{
			name:    "empty name",
			catalog: fakeCatalog{module: "m", defs: []contract.PermissionDefinition{{Code: "X", Name: ""}}},
			want:    "empty name",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := aggregateCatalogs([]contract.PermissionCatalog{tc.catalog})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestUpsertPermissionCatalog_WritesEveryCode(t *testing.T) {
	t.Parallel()
	catalogs := defaultPermissionCatalogs()
	exec := &recordingExecutor{}

	if err := upsertPermissionCatalog(context.Background(), exec, catalogs); err != nil {
		t.Fatalf("upsert returned error: %v", err)
	}

	expected, err := aggregateCatalogs(catalogs)
	if err != nil {
		t.Fatalf("aggregate error: %v", err)
	}
	if len(exec.calls) != len(expected) {
		t.Fatalf("expected %d Exec calls, got %d", len(expected), len(exec.calls))
	}

	for i, entry := range expected {
		call := exec.calls[i]
		if !strings.Contains(call.sql, "INSERT INTO permissions_function_definitions") {
			t.Errorf("call %d SQL missing INSERT, got %q", i, call.sql)
		}
		if len(call.args) != 4 {
			t.Fatalf("call %d expected 4 args, got %d", i, len(call.args))
		}
		if got, want := call.args[0].(string), entry.Code; got != want {
			t.Errorf("call %d code: got %q, want %q", i, got, want)
		}
		if got, want := call.args[1].(string), entry.Module; got != want {
			t.Errorf("call %d module: got %q, want %q", i, got, want)
		}
	}
}

func TestUpsertPermissionCatalog_PropagatesExecError(t *testing.T) {
	t.Parallel()
	catalogs := defaultPermissionCatalogs()
	if len(catalogs) == 0 {
		t.Skip("no catalogs registered")
	}

	// Pick a known code so the failure target is deterministic.
	target := "WORKFLOW_VIEW"
	wantErr := errors.New("boom")
	exec := &recordingExecutor{failOn: target, failErr: wantErr}

	err := upsertPermissionCatalog(context.Background(), exec, catalogs)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("error chain should contain underlying exec error: %v", err)
	}
	if !strings.Contains(err.Error(), target) {
		t.Errorf("error should cite the failing code %q: %v", target, err)
	}
}

// TestDefaultCatalogs_CoverPhase1Codes ensures every Phase 1+ code declared
// by an active module resolves through the seeder. Codes from the legacy
// admin seed (INVESTMENT_VIEW etc.) are intentionally absent from the
// catalog: the catalog migration backfills them with module='legacy' so the
// FK still applies, but no module Provider claims them. Updating the admin
// seed to grant the richer Phase 1 codes is tracked as a follow-up.
func TestDefaultCatalogs_CoverPhase1Codes(t *testing.T) {
	t.Parallel()
	expected := []string{
		"IAM_USER_VIEW",
		"IAM_USER_CREATE",
		"IAM_USER_UPDATE",
		"IAM_USER_DEACTIVATE",
		"IAM_AUDIT_VIEW",
		"WORKFLOW_VIEW",
		"WORKFLOW_EXECUTE",
		"INVESTMENT_FUND_VIEW",
		"INVESTMENT_FUND_MANAGE",
		"INVESTMENT_PORTFOLIO_VIEW",
		"INVESTMENT_PORTFOLIO_MANAGE",
		"INVESTMENT_INSTRUMENT_VIEW",
		"INVESTMENT_INSTRUMENT_MANAGE",
		"INVESTMENT_REFERENCE_VIEW",
		"INVESTMENT_LEDGER_VIEW",
		"INVESTMENT_LEDGER_POST",
		"INVESTMENT_LEDGER_FORCE_POST",
		"INVESTMENT_LEDGER_REVERSE",
		"INVESTMENT_VALUATION_VIEW",
		"INVESTMENT_VALUATION_RUN",
		"INVESTMENT_PRICE_POST",
		"LEAVE_VIEW",
		"LEAVE_MANAGE",
		"APPROVAL_VIEW",
		"APPROVAL_CONFIG",
		"PERMISSIONS_VIEW",
		"PERMISSIONS_MANAGE",
		"NOTIFICATION_CONFIG",
		"AUDIT_VIEW",
	}

	catalogs := defaultPermissionCatalogs()
	entries, err := aggregateCatalogs(catalogs)
	if err != nil {
		t.Fatalf("aggregate error: %v", err)
	}

	known := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		known[e.Code] = struct{}{}
	}

	var missing []string
	for _, code := range expected {
		if _, ok := known[code]; !ok {
			missing = append(missing, code)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Phase 1 codes missing from module catalogs: %v", missing)
	}
}
