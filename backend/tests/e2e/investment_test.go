//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment"
	marketdata "github.com/neo-kanta/ims-th-solution/backend/internal/market_data"
	referencedata "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// TestE2E_InvestmentOversellEnvelope is a real, executable end-to-end test
// that boots the full module graph (audit + iam + compliance + workflow +
// investment + market_data) against a live Postgres database, mounts the
// production chi router via httptest.NewServer, and asserts that the
// OVERSELL rejection path returns the canonical error envelope.
//
// Why this assertion is the load-bearing one for E2E: it exercises every
// layer the brief calls out — routing, auth middleware, permission
// gating, the typed-error to envelope mapping (D4), and the policy-layer
// oversell guard. A regression in any of those breaks this test.
//
// Scope of this test (intentional):
//   - Boots the full module graph in-process. Asserts no IMS_ALLOW_FAKE_*
//     is set so we exercise real implementations only.
//   - Logs in as the seeded admin user (per migration 20260301000005 +
//     20260301000007 fix-password) and uses its JWT.
//   - Seeds a fund / portfolio / instrument / opening BUY transaction
//     directly via SQL — fast and deterministic. The HTTP create-fund /
//     create-portfolio / create-instrument flows are exercised by
//     dedicated unit tests under application/command/.
//   - Asserts: HTTP 422; envelope keys {error_code, message, details,
//     request_id}; error_code == "OVERSELL"; request_id non-empty.
//
// Out of scope here (each tracked separately):
//   - Full lifecycle BUY/SELL/REVERSE/VALUATION assertions — covered by
//     the unit-test suite for projector and reverse_transaction handlers.
//   - Compliance pre-trade BLOCK envelope path — needs a configured rule
//     in compliance__rule_definitions; add when the compliance test
//     fixture lands.
//   - Permission gating with a non-admin user — needs user-creation flow
//     wired in test setup; add when iam admin test fixture lands.
//   - H7 immutability via raw SQL — addressed by the migration's RAISE
//     EXCEPTION triggers; covered by the migration's own contract tests
//     (not yet split out from the schema migration).
//
// Skips cleanly when IMS_TEST_DATABASE_DSN is unset or unreachable, or
// when migrations / seed haven't been applied.
func TestE2E_InvestmentOversellEnvelope(t *testing.T) {
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

	// A freshly migrated+seeded database has the bootstrap admin's
	// force_password_change flag set (by design, for real deployments) —
	// clear it here so this test is runnable against a clean env without a
	// manual first-login password reset step first.
	_, err = pool.Exec(ctx, `UPDATE iam_users SET force_password_change = false WHERE username = 'admin'`)
	require.NoError(t, err, "reset admin force_password_change for test")

	// Login admin → JWT.
	token := loginAdmin(t, server.URL)

	// Seed fund / portfolio / instrument / opening position directly via SQL.
	// The taxonomy reference tables (asset_classes, countries, fund_categories)
	// are assumed seeded by `make seed` before the test is run.
	fundID, portfolioID, instrumentID := seedInvestmentPrereqs(t, ctx, pool)
	defer cleanupInvestmentPrereqs(ctx, pool, fundID)

	// Seed a position of 10 shares so the OVERSELL guard has something to
	// compare against. We bypass HTTP to keep the test focused on the
	// rejection path; happy-path posting is exercised by command-layer tests.
	seedOpeningPosition(t, ctx, pool, portfolioID, instrumentID, 10)

	// Now attempt to SELL 1000 shares (we only hold 10) via HTTP. Expect
	// the canonical error envelope with HTTP 422 and error_code OVERSELL.
	envelope, status := postTransaction(t, server.URL, token, portfolioID, map[string]any{
		"transaction_type": "SELL",
		"side":             "SELL",
		"instrument_id":    instrumentID.String(),
		"quantity":         "1000",
		"price":            "100",
		"currency":         "THB",
		"business_date":    time.Now().UTC().Format("2006-01-02"),
	})

	require.Equal(t, http.StatusUnprocessableEntity, status, "OVERSELL must return 422; got %d / body %+v", status, envelope)
	require.NotEmpty(t, envelope, "envelope must be non-empty")
	require.Equal(t, "OVERSELL", envelope["error_code"], "error_code must be OVERSELL; got %v", envelope["error_code"])
	require.NotEmpty(t, envelope["message"], "message must be non-empty")
	require.NotEmpty(t, envelope["request_id"], "request_id must be non-empty")
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

// bootTestServer mirrors cmd/server/main.go's wiring without starting a
// real listener. Returns the httptest server and the underlying pool so the
// caller can clean up after.
func bootTestServer(t *testing.T, ctx context.Context, dsn string) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "config.Load")

	// Force memory rate-limit backend; tests should not depend on Redis.
	cfg.RateLimitBackend = "memory"
	// Override DSN to the test target (config.Load reads env at startup;
	// re-reading guarantees we hit the test DB even in shells where env
	// drifted).
	dsnHost, dsnPort, dsnDB, dsnUser, dsnPass := parseDSN(t, dsn)
	cfg.DBHost = dsnHost
	cfg.DBPort = dsnPort
	cfg.DBName = dsnDB
	cfg.DBUser = dsnUser
	cfg.DBPassword = dsnPass

	pool, err := database.NewPool(ctx, cfg)
	require.NoError(t, err, "database.NewPool")

	auditMod := audit.NewModule(pool)
	iamMod, err := iam.NewModule(pool, cfg, nil /* redis */, auditMod.Recorder(), auditMod)
	require.NoError(t, err, "iam.NewModule")

	complianceMod := compliance.NewModule(pool, iamMod)
	workflowMod := workflow.NewModule(pool, iamMod, complianceMod.ContractAdapter())
	investmentMod := investment.NewModule(
		pool,
		complianceMod.ContractAdapter(),
		workflowMod,
		iamMod,
		auditMod.Recorder(),
	)
	referenceDataMod := referencedata.NewModule(pool)
	marketDataMod := marketdata.NewModule(pool, cfg, nil /* redis */, referenceDataMod.Resolver())

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.NewCORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.SecureHeaders)
	r.Use(iamMod.GlobalRateLimitMiddleware())

	r.Route("/api/v1", func(r chi.Router) {
		iamMod.SetupRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(iamMod.AuthMiddleware())
			complianceMod.RegisterRoutes(r)
			workflowMod.RegisterRoutes(r)
			investmentMod.RegisterRoutes(r)
			marketDataMod.RegisterRoutes(r)
		})
	})

	// Mirrors cmd/server/main.go's /api/v2 mount (Portfolio V2 routes).
	r.Route("/api/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(iamMod.AuthMiddleware())
			investmentMod.RegisterRoutesV2(r)
		})
	})

	return httptest.NewServer(r), pool
}

// parseDSN extracts host/port/db/user/password from a libpq keyword DSN.
// Supports the keyword form used by integrationDSN() and the project's env
// files; URL-style DSNs are not supported.
func parseDSN(t *testing.T, dsn string) (host, port, db, user, pass string) {
	t.Helper()
	host, port, db, user, pass = "localhost", "5432", "ims_dev", "ims_app", ""
	for _, kv := range splitFields(dsn) {
		k, v := splitKV(kv)
		switch k {
		case "host":
			host = v
		case "port":
			port = v
		case "dbname":
			db = v
		case "user":
			user = v
		case "password":
			pass = v
		}
	}
	return
}

func splitFields(s string) []string {
	out := []string{}
	cur := ""
	for _, ch := range s {
		if ch == ' ' || ch == '\t' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(ch)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func splitKV(s string) (string, string) {
	for i, ch := range s {
		if ch == '=' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

// loginAdmin authenticates against the seeded admin user (admin/admin123).
// Refuses to return a restricted/MFA token — the seeded admin must be in a
// usable state for the test to continue.
func loginAdmin(t *testing.T, baseURL string) string {
	t.Helper()
	body := map[string]string{"username": "admin", "password": "admin123"}
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(buf))
	require.NoError(t, err, "login POST")
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "login expected 200, got %d body=%s", resp.StatusCode, string(raw))

	// Login responses are wrapped in the standard {"data": ...} envelope
	// (httputil.OK) — decode through it rather than expecting these fields
	// at the top level.
	var envelope struct {
		Data struct {
			AccessToken         string `json:"access_token"`
			ForcePasswordChange bool   `json:"force_password_change"`
			MFARequired         bool   `json:"mfa_required"`
			RestrictedSession   bool   `json:"restricted_session"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope), "decode login response: %s", string(raw))
	got := envelope.Data
	require.False(t, got.ForcePasswordChange, "admin force_password_change is true; reset it before running E2E")
	require.False(t, got.MFARequired, "admin MFA required; disable for test environment")
	require.False(t, got.RestrictedSession, "admin returned restricted_session token")
	require.NotEmpty(t, got.AccessToken, "no access_token in login response")
	return got.AccessToken
}

// seedInvestmentPrereqs writes a fund + portfolio + instrument directly into
// the DB. Picks taxonomy IDs from the seeded reference tables; if any are
// missing the test fails (it's a pre-condition violation, not an E2E
// regression).
func seedInvestmentPrereqs(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (fundID, portfolioID, instrumentID uuid.UUID) {
	t.Helper()
	const adminID = "a0000000-0000-0000-0000-000000000001"

	var fundCategoryID, assetClassID, assetSubtypeID, countryID uuid.UUID

	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM investment__fund_categories LIMIT 1`,
	).Scan(&fundCategoryID), "no fund_categories seeded; run make seed")
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM investment__asset_classes WHERE code = 'EQUITY' LIMIT 1`,
	).Scan(&assetClassID), "no EQUITY asset_class seeded; run make seed")
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM investment__asset_subtypes WHERE asset_class_id = $1 LIMIT 1`, assetClassID,
	).Scan(&assetSubtypeID), "no asset_subtype under EQUITY")
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT id FROM investment__countries WHERE iso_code = 'TH' LIMIT 1`,
	).Scan(&countryID), "no TH country seeded")

	fundID = uuid.New()
	portfolioID = uuid.New()
	instrumentID = uuid.New()
	suffix := uuid.NewString()[:8]
	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO investment__funds (
			id, code, name, fund_category_id, base_currency,
			inception_date, has_units, status, version,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, 'THB', $5, false, 'ACTIVE', 1, NOW(), NOW(), $6)`,
		fundID, "E2E-"+suffix, "E2E Fund "+suffix, fundCategoryID, now, adminID,
	)
	require.NoError(t, err, "insert fund")

	_, err = pool.Exec(ctx, `
		INSERT INTO investment__portfolios (
			id, fund_id, code, name, base_currency, valuation_currency,
			inception_date, has_units, status, tax_lot_method, version,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, 'THB', 'THB', $5, false, 'ACTIVE', 'AVERAGE', 1, NOW(), NOW(), $6)`,
		portfolioID, fundID, "E2E-P-"+suffix, "E2E Portfolio "+suffix, now, adminID,
	)
	require.NoError(t, err, "insert portfolio")

	_, err = pool.Exec(ctx, `
		INSERT INTO investment__instruments (
			id, primary_ticker, name, asset_class_id, asset_subtype_id,
			currency, country_id, is_tradable, status,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, 'THB', $6, true, 'ACTIVE', NOW(), NOW(), $7)`,
		instrumentID, "E2E"+suffix, "E2E Instrument "+suffix, assetClassID, assetSubtypeID, countryID, adminID,
	)
	require.NoError(t, err, "insert instrument")

	// Grant the admin user access to this fund (data permissions). The Admin
	// group from the seed migration only has function permissions; data
	// permissions are per-user-per-contract.
	_, err = pool.Exec(ctx, `
		INSERT INTO permissions_data_rights (user_id, contract_id, is_granted, granted_by)
		VALUES ($1, $2::TEXT, true, $1)
		ON CONFLICT (user_id, contract_id) DO UPDATE SET is_granted = true`,
		adminID, fundID,
	)
	require.NoError(t, err, "grant admin data access to fund")

	return fundID, portfolioID, instrumentID
}

// seedOpeningPosition installs a position row with the given quantity at a
// fixed average cost so the OVERSELL guard has something to compare against.
// Bypassing the projector keeps this test focused on the SELL rejection path.
func seedOpeningPosition(t *testing.T, ctx context.Context, pool *pgxpool.Pool, portfolioID, instrumentID uuid.UUID, qty int) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO investment__portfolio_positions (
			id, portfolio_id, instrument_id, quantity, average_cost, cost_basis,
			version, updated_at
		) VALUES ($1, $2, $3, $4, 100, $4 * 100, 1, NOW())
		ON CONFLICT (portfolio_id, instrument_id) DO UPDATE
			SET quantity = EXCLUDED.quantity,
			    average_cost = EXCLUDED.average_cost,
			    cost_basis = EXCLUDED.cost_basis,
			    version = investment__portfolio_positions.version + 1,
			    updated_at = NOW()`,
		uuid.New(), portfolioID, instrumentID, qty,
	)
	require.NoError(t, err, "seed opening position")
}

// cleanupInvestmentPrereqs best-effort tears down the test fund / portfolio
// / instrument so re-runs don't accumulate. Errors are non-fatal — the test
// has already passed or failed by this point.
func cleanupInvestmentPrereqs(ctx context.Context, pool *pgxpool.Pool, fundID uuid.UUID) {
	// Drop child rows first; FKs would block parent removal. Portfolio
	// transactions and cash movements must go before portfolios too —
	// otherwise ON DELETE RESTRICT leaves the portfolio delete a silent
	// no-op on any run that posted ledger activity (e.g.
	// TestE2E_PortfolioV2_LedgerWrites).
	stmts := []string{
		`DELETE FROM investment__portfolio_positions WHERE portfolio_id IN (
			SELECT id FROM investment__portfolios WHERE fund_id = $1)`,
		`DELETE FROM investment__cash_balances WHERE portfolio_id IN (
			SELECT id FROM investment__portfolios WHERE fund_id = $1)`,
		`DELETE FROM investment__cash_movements WHERE portfolio_id IN (
			SELECT id FROM investment__portfolios WHERE fund_id = $1)`,
		`DELETE FROM investment__portfolio_transactions WHERE portfolio_id IN (
			SELECT id FROM investment__portfolios WHERE fund_id = $1)`,
		// Decision/execution/confirmation rows are ON DELETE RESTRICT on
		// portfolio_id/decision_id/execution_id, so they must go before the
		// portfolio delete — otherwise any test that created a decision leaves
		// the portfolio and fund rows behind on every run.
		`DELETE FROM investment__trade_confirmations WHERE decision_id IN (
			SELECT id FROM investment__decisions WHERE portfolio_id IN (
				SELECT id FROM investment__portfolios WHERE fund_id = $1))`,
		`DELETE FROM investment__executions WHERE decision_id IN (
			SELECT id FROM investment__decisions WHERE portfolio_id IN (
				SELECT id FROM investment__portfolios WHERE fund_id = $1))`,
		`DELETE FROM investment__decisions WHERE portfolio_id IN (
			SELECT id FROM investment__portfolios WHERE fund_id = $1)`,
		`DELETE FROM investment__portfolios WHERE fund_id = $1`,
		`DELETE FROM investment__funds WHERE id = $1`,
	}
	for _, s := range stmts {
		_, _ = pool.Exec(ctx, s, fundID)
	}
}

// postTransaction calls POST /investment/portfolios/{id}/transactions with
// the supplied body and bearer token. Returns the parsed JSON envelope (or
// success body) plus the HTTP status code.
func postTransaction(t *testing.T, baseURL, token string, portfolioID uuid.UUID, body map[string]any) (map[string]any, int) {
	t.Helper()
	buf, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/api/v1/investment/portfolios/%s/transactions", baseURL, portfolioID)
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
