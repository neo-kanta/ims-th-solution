// Fixed, deterministic E2E fixtures seeded by `backend/cmd/seed-e2e`
// (see tests/e2e/scripts/setup-db.sh / `make e2e-db-setup`). These are
// dedicated to tests/e2e and backend/tests/e2e — never mixed with
// database/seeds/zz_demo dev demo users (ben/green/neo/admin).

const password = process.env.E2E_USER_PASSWORD;
if (!password) {
  throw new Error(
    "E2E_USER_PASSWORD is not set. Copy infra/env/.env.e2e.example to " +
      "infra/env/.env.e2e and source it before running the suite.",
  );
}

export const E2E_USER_PASSWORD = password;

export const e2eUsers = {
  admin: { username: "e2e_admin", password },
  manager: { username: "e2e_manager", password },
  trader: { username: "e2e_trader", password },
  auditor: { username: "e2e_auditor", password },
  disabled: { username: "e2e_disabled", password },
  locked: { username: "e2e_locked", password },
  noPermission: { username: "e2e_no_permission", password },
} as const;

export type E2EUserKey = keyof typeof e2eUsers;

// Reused from database/seeds/zz_demo/01_demo_funds.sql — seed-e2e grants
// e2e_trader access to Fund A only (see backend/cmd/seed-e2e/main.go).
export const e2eFunds = {
  fundA: { id: "d0001000-0000-0000-0000-000000000001", code: "TH-GOV-LTF" },
  fundB: { id: "d0001000-0000-0000-0000-000000000002", code: "BBL-EQUITY" },
} as const;
