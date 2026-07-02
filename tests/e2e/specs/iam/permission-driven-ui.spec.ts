import { test, expect } from "@playwright/test";
import { loginViaApi } from "../../support/auth";
import { apiLogin, withToken } from "../../support/apiClient";
import { e2eUsers } from "../../support/testUsers";

// Scenario D — Permission-driven UI. Sidebar nav items are rendered by
// frontend/app/features/shell/navigation.ts's buildDashboardNavigation,
// which filters entirely based on authStore.hasPermission(...) — items are
// NuxtLinks (real <a href>), so we assert on the rendered href rather than
// adding a new data-testid per nav item.
//
// Critically: every "menu hidden" assertion here is paired with a direct
// backend API assertion proving the hidden menu isn't the only thing
// standing between the user and the data (the brief's explicit requirement).
test.describe("IAM — Permission-driven UI", () => {
  const administrationLink = 'a.nav-link[href="/administration/settings"]';
  const accountsLink = 'a.nav-link[href="/permissions/accounts"]';

  test("admin sees IAM/administration navigation items", async ({ page, context }) => {
    await loginViaApi(context, e2eUsers.admin.username, e2eUsers.admin.password);
    await page.goto("/");
    await expect(page.locator(administrationLink)).toBeVisible();
    await expect(page.locator(accountsLink)).toBeVisible();
  });

  test("no-permission user does not see IAM/administration navigation items", async ({ page, context, request }) => {
    await loginViaApi(context, e2eUsers.noPermission.username, e2eUsers.noPermission.password);
    await page.goto("/");
    await expect(page.locator(administrationLink)).toHaveCount(0);
    await expect(page.locator(accountsLink)).toHaveCount(0);

    // Hidden menu is not sufficient security: confirm the underlying API
    // still rejects a direct call regardless of what the menu shows.
    const login = await apiLogin(request, e2eUsers.noPermission.username, e2eUsers.noPermission.password);
    const client = withToken(request, login.accessToken!);
    expect((await client.get("/admin/users")).status()).toBe(403);
  });
});
