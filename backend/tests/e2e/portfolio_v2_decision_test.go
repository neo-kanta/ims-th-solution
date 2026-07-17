//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestE2E_PortfolioV2_DecisionOwnership exercises Milestone 5
// (docs/handoff/portfolio-v2-claude-implementation-prompt.md) end to end
// against a live Postgres database: create a DRAFT decision through the
// portfolio-code V2 route and prove the server-derived fund_id/contract_id
// are correct (never client-supplied — the V2 request body has no such
// field), plus the ownership-consistency rule: a decision belonging to one
// portfolio must 404 when accessed through a different portfolio's code.
//
// Submit/execution/confirmation lifecycle transitions are intentionally out
// of scope here — Submit routes through the same IRG pre-trade compliance
// gate TestE2E_PortfolioV2_LedgerWrites already found blocks a fresh
// portfolio's first order (see that test's CASH_IN comment). Create does
// not call compliance at all (see application/command/decision_lifecycle.go
// Create — the gate only runs on Submit), so Create is the clean, real
// assertion point for the derive logic.
func TestE2E_PortfolioV2_DecisionOwnership(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e test")
	}

	dsn := os.Getenv("IMS_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5437 user=ims_app password=ims_dev_password dbname=ims_dev sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	probe, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}
	probe.Close(ctx)

	if v := os.Getenv("IMS_ALLOW_FAKE_WORKFLOW_STATE"); v == "true" {
		t.Fatalf("IMS_ALLOW_FAKE_WORKFLOW_STATE=true; E2E test refuses fake bindings")
	}

	server, pool := bootTestServer(t, ctx, dsn)
	defer server.Close()
	defer pool.Close()

	_, err = pool.Exec(ctx, `UPDATE iam_users SET force_password_change = false WHERE username = 'admin'`)
	require.NoError(t, err, "reset admin force_password_change for test")

	token := loginAdmin(t, server.URL)

	// The seeded dev DB's Admin group has no INVESTMENT_DECISION_* function
	// rights at all (pre-existing seed gap — same limitation would block a
	// V1 decision create through HTTP, not something this migration
	// introduced). Grant them here so the test can prove the V2 create/get
	// path end to end; idempotent and harmless to leave in a dev DB.
	grantAdminFunctionPermissions(t, ctx, pool,
		"INVESTMENT_DECISION_VIEW", "INVESTMENT_DECISION_MANAGE",
		"INVESTMENT_DECISION_SUBMIT", "INVESTMENT_DECISION_CANCEL",
	)

	// Two independent portfolios (under two independent funds) so the
	// ownership check has something real to discriminate against.
	fundA, portfolioA, instrumentA := seedInvestmentPrereqs(t, ctx, pool)
	defer cleanupInvestmentPrereqs(ctx, pool, fundA)
	fundB, portfolioB, _ := seedInvestmentPrereqs(t, ctx, pool)
	defer cleanupInvestmentPrereqs(ctx, pool, fundB)

	var codeA, codeB string
	require.NoError(t, pool.QueryRow(ctx, `SELECT code FROM investment__portfolios WHERE id = $1`, portfolioA).Scan(&codeA))
	require.NoError(t, pool.QueryRow(ctx, `SELECT code FROM investment__portfolios WHERE id = $1`, portfolioB).Scan(&codeB))

	businessDate := time.Now().UTC().Format("2006-01-02")
	body := map[string]any{
		"instrument_id":   instrumentA.String(),
		"instrument_code": "PTT",
		"business_date":   businessDate,
		"side":            "BUY",
		"quantity":        "100",
		"currency":        "THB",
		"rationale":       "e2e ownership test",
	}

	// 1. Create against portfolio A — must derive fund_id = contract_id =
	// fundA from the resolved portfolio.
	createEnvelope, createStatus := postV2(t, server.URL, token, codeA, "decisions", body)
	require.Equal(t, http.StatusCreated, createStatus, "create decision expected 201; body=%+v", createEnvelope)
	data, ok := createEnvelope["data"].(map[string]any)
	require.True(t, ok, "create response missing data envelope: %+v", createEnvelope)
	decisionID, _ := data["id"].(string)
	require.NotEmpty(t, decisionID, "created decision must have an id")

	var storedFundID, storedContractID, storedPortfolioID string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT fund_id, contract_id, portfolio_id FROM investment__decisions WHERE id = $1`, decisionID,
	).Scan(&storedFundID, &storedContractID, &storedPortfolioID))
	require.Equal(t, fundA.String(), storedFundID, "fund_id must be derived from the resolved portfolio")
	require.Equal(t, fundA.String(), storedContractID, "contract_id must equal fund_id (Milestone 1 CHECK constraint)")
	require.Equal(t, portfolioA.String(), storedPortfolioID, "portfolio_id must be the resolved portfolio")

	// 2. The same decision accessed through portfolio B's code must 404 —
	// ownership consistency, proven against real routing/DB, not stubs.
	_, crossStatus := getV2(t, server.URL, token, codeB, "decisions/"+decisionID)
	require.Equal(t, http.StatusNotFound, crossStatus, "decision from portfolio A must 404 under portfolio B's code")

	// 3. The same decision accessed through its own portfolio's code
	// succeeds.
	_, ownStatus := getV2(t, server.URL, token, codeA, "decisions/"+decisionID)
	require.Equal(t, http.StatusOK, ownStatus, "decision must be visible under its own portfolio's code")

	// 4. Unknown portfolio code -> 404 on create.
	_, unknownStatus := postV2(t, server.URL, token, "NOT-A-REAL-CODE", "decisions", body)
	require.Equal(t, http.StatusNotFound, unknownStatus, "unknown portfolio code must return 404")
}

// grantAdminFunctionPermissions grants the given permission codes to the
// seeded Admin group (id b0000000-0000-0000-0000-000000000001). Used only
// to work around the seeded dev DB missing these grants; not part of the
// Portfolio V2 migration itself.
func grantAdminFunctionPermissions(t *testing.T, ctx context.Context, pool *pgxpool.Pool, codes ...string) {
	t.Helper()
	const adminGroupID = "b0000000-0000-0000-0000-000000000001"
	const adminUserID = "a0000000-0000-0000-0000-000000000001"
	for _, code := range codes {
		_, err := pool.Exec(ctx, `
			INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
			VALUES ($1, $2, true, $3)
			ON CONFLICT (group_id, permission_code) DO UPDATE SET is_granted = true`,
			adminGroupID, code, adminUserID,
		)
		require.NoError(t, err, "grant %s to Admin group", code)
	}
}

func getV2(t *testing.T, baseURL, token, portfolioCode, subpath string) (map[string]any, int) {
	t.Helper()
	url := fmt.Sprintf("%s/api/v2/portfolios/%s/%s", baseURL, portfolioCode, subpath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out, resp.StatusCode
}
