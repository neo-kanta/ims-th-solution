import { test, expect } from "@playwright/test";
import { loginViaUI, expectLoginError } from "../../support/auth";
import { e2eUsers } from "../../support/testUsers";

// Scenario A — Authentication (browser-driven; no localStorage/API bypass —
// this file is the "one real UI login test" other specs' loginViaApi shortcut
// depends on).
test.describe("IAM — Authentication", () => {
  test("successful login with e2e_admin redirects off the login page", async ({ page }) => {
    await loginViaUI(page, e2eUsers.admin.username, e2eUsers.admin.password);
    await expect(page).toHaveURL(/^(?!.*\/auth\/login).*$/);
    await expect(page.locator("#login-username")).toHaveCount(0);
  });

  test("failed login with wrong password shows an error and stays on the login page", async ({ page }) => {
    await loginViaUI(page, e2eUsers.admin.username, "definitely-wrong-password");
    await expectLoginError(page);
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test("disabled user cannot log in", async ({ page }) => {
    await loginViaUI(page, e2eUsers.disabled.username, e2eUsers.disabled.password);
    await expectLoginError(page);
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test("locked user cannot log in", async ({ page }) => {
    await loginViaUI(page, e2eUsers.locked.username, e2eUsers.locked.password);
    await expectLoginError(page);
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test("logout clears the session and blocks protected pages", async ({ page }) => {
    await loginViaUI(page, e2eUsers.admin.username, e2eUsers.admin.password);
    await expect(page).toHaveURL(/^(?!.*\/auth\/login).*$/);

    await page.getByTestId("iam-user-menu-toggle").click();
    await page.getByTestId("iam-logout-button").click();
    await expect(page).toHaveURL(/\/auth\/login/);

    await page.goto("/permissions/accounts");
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test("session survives a page reload", async ({ page }) => {
    await loginViaUI(page, e2eUsers.admin.username, e2eUsers.admin.password);
    await expect(page).toHaveURL(/^(?!.*\/auth\/login).*$/);

    await page.reload();
    await expect(page).not.toHaveURL(/\/auth\/login/);
  });
});
