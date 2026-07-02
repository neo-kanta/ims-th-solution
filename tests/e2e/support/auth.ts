import type { BrowserContext, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { API_BASE_URL } from "./apiClient";

/** Drives the real login form. Use this for the auth flow itself (auth.spec.ts). */
export async function loginViaUI(page: Page, username: string, password: string) {
  await page.goto("/auth/login");
  // Nuxt hydrates client-side after the initial SSR paint; filling before
  // hydration attaches the v-model listeners silently no-ops (the DOM shows
  // the typed value, but the underlying ref — and therefore the submit
  // button's disabled state — never updates). Wait for the network to go
  // idle, which corresponds to the client bundle finishing load/hydration,
  // before interacting with the form.
  await page.waitForLoadState("networkidle");

  await page.locator("#login-username").fill(username);
  await page.locator("#login-password").fill(password);
  await page.getByTestId("iam-login-submit").click();
}

export async function expectLoginError(page: Page) {
  const alert = page.getByTestId("iam-login-error");
  await expect(alert).toBeVisible();
  return alert;
}

/**
 * Logs in via the real API and seeds the browser context's auth cookies
 * directly, bypassing the UI form. This is a setup-optimization only: the
 * app stores its session in the `auth_token` / `auth_token_expires_at`
 * cookies (frontend/app/stores/useAuthStore.ts), not localStorage, so that's
 * what this helper writes. Only use this after auth.spec.ts already proves
 * the real UI login flow works end to end.
 */
export async function loginViaApi(context: BrowserContext, username: string, password: string) {
  const resp = await context.request.post(`${API_BASE_URL}/auth/login`, {
    data: { username, password, totp_code: "", recovery_code: "" },
  });
  if (!resp.ok()) {
    throw new Error(`API login failed for ${username}: HTTP ${resp.status()}`);
  }
  const body = (await resp.json()) as {
    data: { access_token: string; access_token_expires_at: string };
  };

  const baseURL = new URL(process.env.E2E_BASE_URL ?? "http://localhost:3000");
  await context.addCookies([
    {
      name: "auth_token",
      value: body.data.access_token,
      domain: baseURL.hostname,
      path: "/",
    },
    {
      name: "auth_token_expires_at",
      value: body.data.access_token_expires_at,
      domain: baseURL.hostname,
      path: "/",
    },
  ]);

  return body.data;
}
