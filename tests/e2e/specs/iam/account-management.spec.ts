import { test, expect } from "@playwright/test";
import { loginViaApi } from "../../support/auth";
import { apiLogin, withToken } from "../../support/apiClient";
import { e2eUsers } from "../../support/testUsers";

// Scenario E — Account management. Runs entirely against a scratch user
// created fresh in each test (unique username per run) so this file never
// touches the 7 fixed E2E fixtures and stays safe to run repeatedly without
// a database reset between runs.
//
// KNOWN GAP (documented, not faked — see tests/e2e/COVERAGE.md): the brief
// expects a newly created account to stay unusable until activation/approval.
// Current backend behavior (backend/internal/iam/application/command/admin_user.go)
// creates the account immediately active; only ForcePasswordChange=true is
// set. This spec asserts the REAL behavior rather than an activation gate
// that doesn't exist.
test.describe("IAM — Account management", () => {
  // See route-protection.spec.ts's "KNOWN GAP" comment: a hard navigation to
  // a permission-gated page bounces even an authorized user to /403, because
  // authStore.restoreSession() (which populates `permissions` from the auth
  // cookie) is client-only and doesn't run during Nuxt's SSR pass. Landing
  // on "/" first and clicking through avoids it — and matches how an admin
  // actually gets to this page in the running app.
  async function goToAdministrationSettings(page: import("@playwright/test").Page) {
    await page.goto("/");
    await page
      .locator('a.nav-link[href="/administration/settings"]')
      .evaluate((el) => (el as HTMLElement).click());
    await expect(page).toHaveURL("/administration/settings");
    await page.getByTestId("iam-settings-nav-users").click();
  }

  // disable/enable/lock/unlock/reset-password all route through a shared
  // confirmation modal (SettingsControlCenter.vue's useSettingsConfirm) —
  // the action button click only opens it; nothing is sent to the backend
  // until this confirms it.
  async function confirmAction(page: import("@playwright/test").Page) {
    await page.getByTestId("iam-confirm-dialog-confirm").click();
  }

  function scratchUser() {
    const suffix = Date.now().toString(36);
    return {
      username: `e2e_scratch_${suffix}`,
      displayName: `E2E Scratch ${suffix}`,
      email: `e2e_scratch_${suffix}@ims-e2e.local`,
      password: "E2eScratch#Pass1234",
    };
  }

  test("admin can create an account via the UI; it is usable immediately (documented gap) with a forced password change", async ({
    page,
    context,
    request,
  }) => {
    const user = scratchUser();
    await loginViaApi(context, e2eUsers.admin.username, e2eUsers.admin.password);

    await goToAdministrationSettings(page);
    await page.locator("#settings-create-username").fill(user.username);
    await page.locator("#settings-create-display-name").fill(user.displayName);
    await page.locator("#settings-create-email").fill(user.email);
    await page.locator("#settings-create-password").fill(user.password);
    await page.getByTestId("iam-create-user-submit").click();

    // The new account appears in the directory once created.
    // The search box only filters on form submit (no auto-search on input).
    await page.getByTestId("iam-user-search").fill(user.username);
    await page.getByTestId("iam-user-search").press("Enter");
    await expect(page.getByTestId(`iam-user-row-${user.username}`)).toBeVisible();

    // Real current behavior: immediately usable, no activation gate.
    const login = await apiLogin(request, user.username, user.password);
    expect(login.status).toBe(200);
  });

  test("admin can disable and re-enable an account via the UI; disabled account cannot log in", async ({
    page,
    context,
    request,
  }) => {
    const user = scratchUser();
    await loginViaApi(context, e2eUsers.admin.username, e2eUsers.admin.password);

    await goToAdministrationSettings(page);
    await page.locator("#settings-create-username").fill(user.username);
    await page.locator("#settings-create-display-name").fill(user.displayName);
    await page.locator("#settings-create-email").fill(user.email);
    await page.locator("#settings-create-password").fill(user.password);
    await page.getByTestId("iam-create-user-submit").click();
    // The search box only filters on form submit (no auto-search on input).
    await page.getByTestId("iam-user-search").fill(user.username);
    await page.getByTestId("iam-user-search").press("Enter");
    await page.getByTestId(`iam-user-row-${user.username}`).click();

    // Grab a live token for this user *before* disabling it, to prove the
    // per-request re-check (middleware.Auth checks IsUserActive on every
    // call, not just at login) — a still-valid JWT must stop working too.
    const preDisableLogin = await apiLogin(request, user.username, user.password);
    expect(preDisableLogin.status).toBe(200);
    const staleToken = withToken(request, preDisableLogin.accessToken!);
    expect((await staleToken.get("/auth/me")).status()).toBe(200);

    await page.getByTestId("iam-user-disable").click();
    await confirmAction(page);
    await expect(page.getByTestId("iam-user-enable")).toBeEnabled();

    // Backend: disabled user cannot start a new session.
    expect((await apiLogin(request, user.username, user.password)).status).toBe(401);
    // Backend: the already-issued token from before the disable is now rejected too.
    expect((await staleToken.get("/auth/me")).status()).toBe(403);

    await page.getByTestId("iam-user-enable").click();
    await confirmAction(page);
    await expect(page.getByTestId("iam-user-disable")).toBeEnabled();
    expect((await apiLogin(request, user.username, user.password)).status).toBe(200);
  });

  test("admin can lock and unlock an account via the UI", async ({ page, context, request }) => {
    const user = scratchUser();
    await loginViaApi(context, e2eUsers.admin.username, e2eUsers.admin.password);

    await goToAdministrationSettings(page);
    await page.locator("#settings-create-username").fill(user.username);
    await page.locator("#settings-create-display-name").fill(user.displayName);
    await page.locator("#settings-create-email").fill(user.email);
    await page.locator("#settings-create-password").fill(user.password);
    await page.getByTestId("iam-create-user-submit").click();
    // The search box only filters on form submit (no auto-search on input).
    await page.getByTestId("iam-user-search").fill(user.username);
    await page.getByTestId("iam-user-search").press("Enter");
    await page.getByTestId(`iam-user-row-${user.username}`).click();

    await page.getByTestId("iam-user-lock").click();
    await confirmAction(page);
    await expect(page.getByTestId("iam-user-unlock")).toBeEnabled();
    expect((await apiLogin(request, user.username, user.password)).status).toBe(401);

    await page.getByTestId("iam-user-unlock").click();
    await confirmAction(page);
    await expect(page.getByTestId("iam-user-lock")).toBeEnabled();
    expect((await apiLogin(request, user.username, user.password)).status).toBe(200);
  });

  test("admin can reset a user's password via the UI", async ({ page, context, request }) => {
    const user = scratchUser();
    await loginViaApi(context, e2eUsers.admin.username, e2eUsers.admin.password);

    await goToAdministrationSettings(page);
    await page.locator("#settings-create-username").fill(user.username);
    await page.locator("#settings-create-display-name").fill(user.displayName);
    await page.locator("#settings-create-email").fill(user.email);
    await page.locator("#settings-create-password").fill(user.password);
    await page.getByTestId("iam-create-user-submit").click();
    // The search box only filters on form submit (no auto-search on input).
    await page.getByTestId("iam-user-search").fill(user.username);
    await page.getByTestId("iam-user-search").press("Enter");
    await page.getByTestId(`iam-user-row-${user.username}`).click();

    const newPassword = "E2eScratch#Reset5678";
    await page.getByTestId("iam-user-reset-password-input").fill(newPassword);
    await page.getByTestId("iam-user-reset-password-submit").click();
    await confirmAction(page);
    // Wait for the confirm dialog to close (mutation has round-tripped)
    // before asserting against the backend.
    await expect(page.getByTestId("iam-confirm-dialog-confirm")).toHaveCount(0);

    expect((await apiLogin(request, user.username, newPassword)).status).toBe(200);
    expect((await apiLogin(request, user.username, user.password)).status).toBe(401);
  });
});
