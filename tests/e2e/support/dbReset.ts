import { execFile } from "node:child_process";
import { promisify } from "node:util";
import path from "node:path";
import { fileURLToPath } from "node:url";

const execFileAsync = promisify(execFile);
const backendDir = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../backend",
);

/**
 * Re-runs backend/cmd/seed-e2e, which idempotently resets the 7 fixed E2E
 * fixtures (e2e_admin, e2e_manager, ...) to their baseline state. It does
 * NOT touch scratch users/records that individual specs create themselves
 * (account-management.spec.ts, group-role-permission.spec.ts) — those are
 * uniquely named per run and cheap to leave behind in the throwaway ims_e2e
 * database.
 *
 * Safe to call multiple times. Requires the same env vars as
 * `make e2e-db-setup` (DB_*, APP_ENV=test, E2E_USER_PASSWORD) to already be
 * present in the process environment.
 */
export async function resetE2EFixtures(): Promise<void> {
  await execFileAsync("go", ["run", "./cmd/seed-e2e"], {
    cwd: backendDir,
    env: process.env,
  });
}
