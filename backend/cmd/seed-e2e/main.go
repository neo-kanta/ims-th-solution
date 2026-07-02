// Command seed-e2e seeds deterministic IAM fixtures for the Playwright /
// backend E2E suite in tests/e2e and backend/tests/e2e. It is intentionally
// separate from cmd/seed: it must never run against a dev/prod database, and
// its fixtures (e2e_admin, e2e_manager, ...) must never mix with the
// zz_demo/* dev demo users.
//
// Run after `go run ./cmd/migrate up` and `go run ./cmd/seed` (the latter
// seeds the permission_code catalog these grants reference, plus the demo
// funds this command scopes data-rights against).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
)

// Fixed, deterministic UUIDs for E2E-only fixtures. Namespaced under
// e2e00000-... so they can never collide with dev seed IDs (a0000000-...,
// b0000000-...) or randomly generated production data.
const (
	userAdminID        = "e2e00000-0000-0000-0000-000000000001"
	userManagerID      = "e2e00000-0000-0000-0000-000000000002"
	userTraderID       = "e2e00000-0000-0000-0000-000000000003"
	userAuditorID      = "e2e00000-0000-0000-0000-000000000004"
	userDisabledID     = "e2e00000-0000-0000-0000-000000000005"
	userLockedID       = "e2e00000-0000-0000-0000-000000000006"
	userNoPermissionID = "e2e00000-0000-0000-0000-000000000007"

	groupAdminID   = "e2e00000-0000-0000-0000-0000000000a1"
	groupManagerID = "e2e00000-0000-0000-0000-0000000000a2"
	groupTraderID  = "e2e00000-0000-0000-0000-0000000000a3"
	groupAuditorID = "e2e00000-0000-0000-0000-0000000000a4"

	// Reused from database/seeds/zz_demo/01_demo_funds.sql (TH-GOV-LTF and
	// BBL-EQUITY). seed-e2e does not create its own funds; it only grants /
	// withholds data-rights against these two existing demo funds so the
	// data-permission scenario doesn't need to duplicate the investment
	// reference schema (fund_category, asset_class, etc. lookups).
	fundAID = "d0001000-0000-0000-0000-000000000001" // TH-GOV-LTF
	fundBID = "d0001000-0000-0000-0000-000000000002" // BBL-EQUITY
)

type e2eUser struct {
	id, username, displayName, email string
	active                           bool
	locked                           bool
}

var e2eUsers = []e2eUser{
	{userAdminID, "e2e_admin", "E2E Admin", "e2e_admin@ims-e2e.local", true, false},
	{userManagerID, "e2e_manager", "E2E Manager", "e2e_manager@ims-e2e.local", true, false},
	{userTraderID, "e2e_trader", "E2E Trader", "e2e_trader@ims-e2e.local", true, false},
	{userAuditorID, "e2e_auditor", "E2E Auditor", "e2e_auditor@ims-e2e.local", true, false},
	{userDisabledID, "e2e_disabled", "E2E Disabled", "e2e_disabled@ims-e2e.local", false, false},
	{userLockedID, "e2e_locked", "E2E Locked", "e2e_locked@ims-e2e.local", true, true},
	{userNoPermissionID, "e2e_no_permission", "E2E No Permission", "e2e_no_permission@ims-e2e.local", true, false},
}

type e2eGroup struct {
	id, name, userID string
	codes            []string
}

// Grants mirror real role analogs (see the E2E plan) but are declared as
// their own groups so a change to a real catalog role never silently
// reshapes what an E2E fixture can do.
var e2eGroups = []e2eGroup{
	{
		id: groupAdminID, name: "E2E Admin", userID: userAdminID,
		codes: []string{
			"IAM_USER_VIEW", "IAM_USER_CREATE", "IAM_USER_UPDATE", "IAM_USER_DEACTIVATE", "IAM_AUDIT_VIEW",
			"PERMISSIONS_VIEW", "PERMISSIONS_MANAGE",
			"permission.users.view", "permission.groups.view", "permission.roles.view",
			"permission.function_rights.view", "permission.data_rights.view",
			"permission.change_request.create", "permission.change_request.submit",
			"permission.change_request.review", "permission.change_request.approve",
			"permission.change_request.merge", "permission.change_request.reject",
			"permission.change_request.close", "permission.change_request.cancel",
			"permission.audit.view", "permission.audit.export",
			"INVESTMENT_FUND_VIEW",
		},
	},
	{
		id: groupManagerID, name: "E2E Manager", userID: userManagerID,
		codes: []string{"INVESTMENT_FUND_VIEW", "INVESTMENT_FUND_MANAGE"},
	},
	{
		id: groupTraderID, name: "E2E Trader", userID: userTraderID,
		codes: []string{"INVESTMENT_FUND_VIEW"},
	},
	{
		id: groupAuditorID, name: "E2E Auditor", userID: userAuditorID,
		codes: []string{"IAM_AUDIT_VIEW", "permission.audit.view", "INVESTMENT_FUND_VIEW"},
	},
	// e2e_no_permission intentionally has no group / no grants.
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	// Hard safety guard: refuse to run against anything that isn't clearly
	// the dedicated E2E database. This makes it physically impossible to
	// point this command at dev/prod by a stray env var.
	if !strings.EqualFold(cfg.Env, "test") || !strings.Contains(strings.ToLower(cfg.DBName), "e2e") {
		fmt.Fprintf(os.Stderr,
			"refusing to run: seed-e2e requires APP_ENV=test and a DB_NAME containing \"e2e\" (got APP_ENV=%q DB_NAME=%q)\n",
			cfg.Env, cfg.DBName)
		os.Exit(1)
	}

	password := os.Getenv("E2E_USER_PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "E2E_USER_PASSWORD is required")
		os.Exit(1)
	}
	passwordHash, err := valueobject.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hashing E2E_USER_PASSWORD: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := seedUsers(ctx, pool, passwordHash); err != nil {
		slog.Error("Failed to seed E2E users", "error", err)
		os.Exit(1)
	}
	if err := seedGroups(ctx, pool); err != nil {
		slog.Error("Failed to seed E2E groups", "error", err)
		os.Exit(1)
	}
	if err := seedDataRights(ctx, pool); err != nil {
		slog.Error("Failed to seed E2E data rights", "error", err)
		os.Exit(1)
	}

	slog.Info("E2E fixtures seeded successfully", "users", len(e2eUsers), "groups", len(e2eGroups))
}

func seedUsers(ctx context.Context, pool *pgxpool.Pool, passwordHash string) error {
	farFuture := time.Now().UTC().AddDate(100, 0, 0)

	for _, u := range e2eUsers {
		var lockedUntil *time.Time
		if u.locked {
			lockedUntil = &farFuture
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO iam_users (
				id, username, display_name, email, password_hash,
				is_active, force_password_change, failed_login_attempts, locked_until,
				version, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, false, 0, $7, 1, NOW(), NOW())
			ON CONFLICT (username) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				email = EXCLUDED.email,
				password_hash = EXCLUDED.password_hash,
				is_active = EXCLUDED.is_active,
				force_password_change = false,
				failed_login_attempts = 0,
				locked_until = EXCLUDED.locked_until,
				deleted_at = NULL,
				updated_at = NOW()`,
			u.id, u.username, u.displayName, u.email, passwordHash, u.active, lockedUntil,
		)
		if err != nil {
			return fmt.Errorf("seeding user %s: %w", u.username, err)
		}
	}
	return nil
}

func seedGroups(ctx context.Context, pool *pgxpool.Pool) error {
	for _, g := range e2eGroups {
		if _, err := pool.Exec(ctx, `
			INSERT INTO permissions_groups (id, name, description, is_active, created_at, updated_at, created_by, updated_by)
			VALUES ($1, $2, $3, true, NOW(), NOW(), $4, $4)
			ON CONFLICT (name) DO UPDATE SET
				is_active = true,
				deleted_at = NULL,
				updated_at = NOW()`,
			g.id, g.name, "E2E fixture group (tests/e2e) — do not assign real users to this group", userAdminID,
		); err != nil {
			return fmt.Errorf("seeding group %s: %w", g.name, err)
		}

		if _, err := pool.Exec(ctx, `
			INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_at, assigned_by)
			VALUES ($1, $2, NOW(), $1)
			ON CONFLICT (user_id, group_id) DO NOTHING`,
			g.userID, g.id,
		); err != nil {
			return fmt.Errorf("assigning %s to group %s: %w", g.userID, g.name, err)
		}

		for _, code := range g.codes {
			// permissions_function_rights.permission_code has an FK to the
			// legacy permissions_function_definitions catalog, which enforces
			// UPPER(code) — dot-notation codes (e.g. permission.users.view)
			// live in the newer, fine-grained permission_function_rights
			// table instead (subject_type/subject_id, not group_id). Route
			// each code to whichever table actually accepts it; both feed
			// PermissionsFetcher's effective-permissions UNION query.
			if code == strings.ToUpper(code) {
				if _, err := pool.Exec(ctx, `
					INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_at, created_by)
					VALUES ($1, $2, true, NOW(), $3)
					ON CONFLICT (group_id, permission_code) DO UPDATE SET is_granted = true`,
					g.id, code, userAdminID,
				); err != nil {
					return fmt.Errorf("granting %s to group %s: %w", code, g.name, err)
				}
				continue
			}

			if _, err := pool.Exec(ctx, `
				INSERT INTO permission_function_rights (
					subject_type, subject_id, permission_code,
					can_view, can_search, can_add, can_edit, can_delete,
					can_approve, can_revoke_approval, can_export, can_configure,
					created_at, updated_at, approved_at, approved_by
				) VALUES ('GROUP', $1, $2, true, true, true, true, true, true, true, true, true, NOW(), NOW(), NOW(), $3)
				ON CONFLICT (subject_type, subject_id, permission_code) DO UPDATE SET
					can_view = true, can_search = true, can_add = true, can_edit = true, can_delete = true,
					can_approve = true, can_revoke_approval = true, can_export = true, can_configure = true,
					updated_at = NOW()`,
				g.id, code, userAdminID,
			); err != nil {
				return fmt.Errorf("granting %s to group %s: %w", code, g.name, err)
			}
		}
	}
	return nil
}

func seedDataRights(ctx context.Context, pool *pgxpool.Pool) error {
	// e2e_trader: Fund A granted.
	_, err := pool.Exec(ctx, `
		INSERT INTO permissions_data_rights (user_id, contract_id, is_granted, granted_at, granted_by)
		VALUES ($1, $2, true, NOW(), $3)
		ON CONFLICT (user_id, contract_id) DO UPDATE SET
			is_granted = true,
			granted_at = NOW(),
			granted_by = EXCLUDED.granted_by`,
		userTraderID, fundAID, userAdminID,
	)
	if err != nil {
		return fmt.Errorf("granting e2e_trader access to fund A: %w", err)
	}

	// e2e_trader: Fund B must stay ungranted. Delete defensively so re-runs
	// stay deterministic even if a prior local run granted it by hand.
	_, err = pool.Exec(ctx, `
		DELETE FROM permissions_data_rights WHERE user_id = $1 AND contract_id = $2`,
		userTraderID, fundBID,
	)
	if err != nil {
		return fmt.Errorf("ensuring e2e_trader has no access to fund B: %w", err)
	}

	return nil
}
