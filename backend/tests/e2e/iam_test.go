//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

// TestE2E_IAM_AuthAndAuthorization proves, directly against the real HTTP
// API (no frontend involved), that IAM enforces authentication and
// authorization end to end:
//   - unauthenticated requests to admin endpoints get 401
//   - an authenticated user without IAM permission gets 403
//   - e2e_admin succeeds against admin endpoints (200)
//   - e2e_disabled / e2e_locked cannot log in (401, no user enumeration)
//   - wrong password is rejected (401)
//   - data-permission scoping is enforced on a business endpoint: e2e_trader
//     can read Fund A but not Fund B
//   - a real POST /auth/logout revokes the session (its refresh token can no
//     longer be used)
//
// Requires backend/cmd/seed-e2e fixtures to already be present — run
// `make e2e-db-setup` first. Skips cleanly when IMS_TEST_DATABASE_DSN or
// E2E_USER_PASSWORD is unset, or the database is unreachable, mirroring
// investment_test.go's escape valves.
func TestE2E_IAM_AuthAndAuthorization(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e test")
	}

	dsn := os.Getenv("IMS_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5437 user=ims_app password=ims_dev_password dbname=ims_e2e sslmode=disable"
	}

	password := os.Getenv("E2E_USER_PASSWORD")
	if password == "" {
		t.Skip("E2E_USER_PASSWORD not set; run via `make test-e2e-backend` with infra/env/.env.e2e sourced")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	probe, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Postgres E2E database unavailable: %v", err)
	}
	probe.Close(ctx)

	server, pool := bootTestServer(t, ctx, dsn)
	defer server.Close()
	defer pool.Close()

	t.Run("unauthenticated request to admin endpoint returns 401", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/admin/users")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("e2e_no_permission is authenticated but forbidden from admin endpoint", func(t *testing.T) {
		token, status := loginE2E(t, server.URL, "e2e_no_permission", password)
		require.Equal(t, http.StatusOK, status, "e2e_no_permission must be able to log in")
		require.Equal(t, http.StatusForbidden, getWithToken(t, server.URL+"/api/v1/admin/users", token))
	})

	t.Run("e2e_admin succeeds against admin endpoints", func(t *testing.T) {
		token, status := loginE2E(t, server.URL, "e2e_admin", password)
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, http.StatusOK, getWithToken(t, server.URL+"/api/v1/admin/users", token))
		require.Equal(t, http.StatusOK, getWithToken(t, server.URL+"/api/v1/admin/audit", token))
	})

	t.Run("e2e_disabled cannot log in", func(t *testing.T) {
		_, status := loginE2E(t, server.URL, "e2e_disabled", password)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("e2e_locked cannot log in", func(t *testing.T) {
		_, status := loginE2E(t, server.URL, "e2e_locked", password)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		_, status := loginE2E(t, server.URL, "e2e_admin", "definitely-wrong-password")
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("e2e_trader has fund A but not fund B", func(t *testing.T) {
		token, status := loginE2E(t, server.URL, "e2e_trader", password)
		require.Equal(t, http.StatusOK, status)
		const fundA = "d0001000-0000-0000-0000-000000000001" // TH-GOV-LTF
		const fundB = "d0001000-0000-0000-0000-000000000002" // BBL-EQUITY
		require.Equal(t, http.StatusOK, getWithToken(t, server.URL+"/api/v1/investment/funds/"+fundA, token),
			"e2e_trader must have data-permission access to fund A")
		require.Equal(t, http.StatusForbidden, getWithToken(t, server.URL+"/api/v1/investment/funds/"+fundB, token),
			"e2e_trader must NOT have data-permission access to fund B")
	})

	t.Run("logout revokes the session's refresh token", func(t *testing.T) {
		accessToken, refreshToken := loginE2EFull(t, server.URL, "e2e_manager", password)
		require.Equal(t, http.StatusOK, getWithToken(t, server.URL+"/api/v1/auth/me", accessToken))

		logoutBody, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/logout", bytes.NewReader(logoutBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusNoContent, resp.StatusCode, "logout should succeed")

		// Concrete proof of server-side revocation: the refresh token that
		// belonged to the logged-out session must no longer be usable.
		refreshBody, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
		refreshResp, err := http.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(refreshBody))
		require.NoError(t, err)
		defer refreshResp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, refreshResp.StatusCode,
			"refreshing with a revoked session's refresh token must fail")
	})
}

// loginE2E logs in and returns (access_token, http_status). Does not fail
// the test on non-200 — callers assert the status themselves since some
// scenarios (disabled/locked/wrong password) expect a rejection.
func loginE2E(t *testing.T, baseURL, username, password string) (string, int) {
	t.Helper()
	accessToken, _, status := loginE2ERaw(t, baseURL, username, password)
	return accessToken, status
}

// loginE2EFull logs in a user expected to succeed and returns
// (access_token, refresh_token).
func loginE2EFull(t *testing.T, baseURL, username, password string) (string, string) {
	t.Helper()
	accessToken, refreshToken, status := loginE2ERaw(t, baseURL, username, password)
	require.Equal(t, http.StatusOK, status, "login for %s must succeed", username)
	return accessToken, refreshToken
}

func loginE2ERaw(t *testing.T, baseURL, username, password string) (accessToken, refreshToken string, status int) {
	t.Helper()
	buf, err := json.Marshal(map[string]string{"username": username, "password": password})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", "", resp.StatusCode
	}

	var envelope struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope), "decode login response: %s", string(raw))
	return envelope.Data.AccessToken, envelope.Data.RefreshToken, resp.StatusCode
}

func getWithToken(t *testing.T, url, token string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}
