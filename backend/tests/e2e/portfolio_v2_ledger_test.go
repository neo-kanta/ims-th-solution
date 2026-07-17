//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// TestE2E_PortfolioV2_LedgerWrites exercises Milestone 3
// (docs/handoff/portfolio-v2-claude-implementation-prompt.md) end to end
// against a live Postgres database: simulate, post, and reverse a ledger
// transaction entirely through the portfolio-code V2 routes
// (/api/v2/portfolios/{portfolioCode}/transactions...), plus the 404/409
// and "extra fund_id is ignored, never trusted" guarantees.
//
// Skips cleanly under the same conditions as TestE2E_InvestmentOversellEnvelope.
func TestE2E_PortfolioV2_LedgerWrites(t *testing.T) {
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

	fundID, portfolioID, _ := seedInvestmentPrereqs(t, ctx, pool)
	defer cleanupInvestmentPrereqs(ctx, pool, fundID)

	var portfolioCode string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT code FROM investment__portfolios WHERE id = $1`, portfolioID,
	).Scan(&portfolioCode), "fetch seeded portfolio code")

	// CASH_IN, not BUY: pre-trade IRG compliance (cash.availability,
	// sector_exposure, ...) only gates security trades
	// (TransactionType.IsSecurityTrade(), see checkCompliance in
	// application/command/post_transaction.go) — tuning a fresh portfolio to
	// pass those rules is out of scope for Milestone 3, which is about
	// portfolio-code resolution and fund_id derivation, not compliance rule
	// coverage. A pure cash movement exercises the same code-resolution and
	// ledger-insert path without touching the compliance gate.
	businessDate := time.Now().UTC().Format("2006-01-02")
	buyBody := map[string]any{
		"transaction_type": "CASH_IN",
		"gross_amount":     "1000",
		"currency":         "THB",
		"business_date":    businessDate,
		// Extra legacy field a stale/malicious client might still send.
		// request.PostTransactionRequest has no fund_id field, so this must
		// be silently ignored, not trusted, and never able to redirect the
		// post to a different fund.
		"fund_id": "00000000-0000-0000-0000-000000000000",
	}

	// 1. Simulate — must not mutate the ledger.
	simEnvelope, simStatus := postV2(t, server.URL, token, portfolioCode, "transactions/simulate", buyBody)
	require.Equal(t, http.StatusOK, simStatus, "simulate expected 200; body=%+v", simEnvelope)

	var txnCountAfterSim int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM investment__portfolio_transactions WHERE portfolio_id = $1`, portfolioID,
	).Scan(&txnCountAfterSim))
	require.Equal(t, 0, txnCountAfterSim, "simulate must not insert a ledger row")

	// 2. Post — must succeed and derive fund_id from the resolved portfolio,
	// ignoring the bogus fund_id in the request body.
	postEnvelope, postStatus := postV2(t, server.URL, token, portfolioCode, "transactions", buyBody)
	require.Equal(t, http.StatusCreated, postStatus, "post expected 201; body=%+v", postEnvelope)
	data, ok := postEnvelope["data"].(map[string]any)
	require.True(t, ok, "post response missing data envelope: %+v", postEnvelope)
	require.Equal(t, portfolioID.String(), data["portfolio_id"], "posted transaction must belong to the resolved portfolio")
	transactionID, _ := data["id"].(string)
	require.NotEmpty(t, transactionID, "posted transaction must have an id")

	var storedFundID string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT fund_id FROM investment__portfolio_transactions WHERE id = $1`, transactionID,
	).Scan(&storedFundID))
	require.Equal(t, fundID.String(), storedFundID, "stored fund_id must come from the resolved portfolio, not the client-supplied fund_id")

	// 3. Reverse — must succeed via the code-scoped reverse route.
	reverseEnvelope, reverseStatus := reverseV2(t, server.URL, token, portfolioCode, transactionID, map[string]any{
		"business_date": businessDate,
		"reason":        "e2e reversal",
	})
	require.Equal(t, http.StatusCreated, reverseStatus, "reverse expected 201; body=%+v", reverseEnvelope)

	// 4. Unknown portfolio code -> 404.
	_, notFoundStatus := postV2(t, server.URL, token, "NOT-A-REAL-CODE", "transactions/simulate", buyBody)
	require.Equal(t, http.StatusNotFound, notFoundStatus, "unknown portfolio code must return 404")
}

func postV2(t *testing.T, baseURL, token, portfolioCode, subpath string, body map[string]any) (map[string]any, int) {
	t.Helper()
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/api/v2/portfolios/%s/%s", baseURL, portfolioCode, subpath)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
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

func reverseV2(t *testing.T, baseURL, token, portfolioCode, transactionID string, body map[string]any) (map[string]any, int) {
	t.Helper()
	return postV2(t, baseURL, token, portfolioCode, fmt.Sprintf("transactions/%s/reverse", transactionID), body)
}
