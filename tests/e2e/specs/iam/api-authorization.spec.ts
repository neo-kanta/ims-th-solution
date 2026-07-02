import { test, expect } from "@playwright/test";
import { apiLogin, withToken, API_BASE_URL } from "../../support/apiClient";
import { e2eUsers } from "../../support/testUsers";

// Scenario C — Backend authorization enforcement, asserted directly against
// the API (no browser). This is the proof that frontend route/menu guards
// are not the only thing standing between an unauthorized user and the data —
// the exact requirement the brief calls out: "Frontend behavior and backend
// behavior must both be verified."
test.describe("IAM — Backend authorization enforcement", () => {
  test("unauthenticated request to an admin endpoint returns 401", async ({ request }) => {
    const resp = await request.get(`${API_BASE_URL}/admin/users`);
    expect(resp.status()).toBe(401);
  });

  test("authenticated user without IAM permission gets 403 from admin API", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.noPermission.username, e2eUsers.noPermission.password);
    expect(login.status).toBe(200);

    const client = withToken(request, login.accessToken!);
    const resp = await client.get("/admin/users");
    expect(resp.status()).toBe(403);
  });

  test("admin request succeeds against admin API", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.admin.username, e2eUsers.admin.password);
    expect(login.status).toBe(200);

    const client = withToken(request, login.accessToken!);
    expect((await client.get("/admin/users")).status()).toBe(200);
    expect((await client.get("/admin/audit")).status()).toBe(200);
  });

  test("disabled/locked fixtures cannot obtain a token to begin with", async ({ request }) => {
    expect((await apiLogin(request, e2eUsers.disabled.username, e2eUsers.disabled.password)).status).toBe(401);
    expect((await apiLogin(request, e2eUsers.locked.username, e2eUsers.locked.password)).status).toBe(401);
  });

  // The per-request re-check (a still-valid token for a user disabled
  // *mid-session* must get 403, not just fail at next login) is covered in
  // account-management.spec.ts, where a scratch user is created and disabled
  // via the real admin flow while holding an already-issued token.
});
