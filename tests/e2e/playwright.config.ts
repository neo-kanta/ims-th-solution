import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";
const isCI = Boolean(process.env.CI);

// Assumes the backend + frontend are already running against the dedicated
// ims_e2e database (see README.md → "IAM E2E Tests"): `make e2e-db-setup`
// then `make dev-backend` / `make dev-frontend` with infra/env/.env.e2e
// sourced. Playwright intentionally does not manage these processes itself —
// it mirrors the existing `cd tests/e2e && npx playwright test` Makefile
// convention, which assumes a running stack.
export default defineConfig({
  testDir: "./specs",
  globalSetup: "./support/globalSetup.ts",
  fullyParallel: true,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers: isCI ? 2 : undefined,
  reporter: isCI ? [["list"], ["html", { open: "never" }]] : "list",
  timeout: 30_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
