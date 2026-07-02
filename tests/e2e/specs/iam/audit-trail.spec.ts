import { test, expect } from "@playwright/test";
import { apiLogin, withToken } from "../../support/apiClient";
import { e2eUsers } from "../../support/testUsers";

// Scenario H — Audit/security trail. Confirmed audited event types (see
// backend/internal/audit/domain/entity): LOGIN_SUCCESS, LOGIN_FAILURE,
// ACCOUNT_LOCKED, ACCOUNT_UNLOCKED, USER_CREATED, USER_ACTIVATED,
// USER_DEACTIVATED, PASSWORD_CHANGE, LOGOUT, and more — queryable via
// GET /admin/audit (IAM_AUDIT_VIEW). All assertions here filter by
// target_id/actor_id scoped to fixtures/records this run itself creates, so
// this file is safe to re-run without a database reset.
test.describe("IAM — Audit trail", () => {
  async function adminClient(request: import("@playwright/test").APIRequestContext) {
    const login = await apiLogin(request, e2eUsers.admin.username, e2eUsers.admin.password);
    return { client: withToken(request, login.accessToken!), adminId: await currentUserId(request, login.accessToken!) };
  }

  async function currentUserId(request: import("@playwright/test").APIRequestContext, token: string) {
    const resp = await withToken(request, token).get("/auth/me");
    const body = (await resp.json()) as { data: { user: { id: string } } };
    return body.data.user.id;
  }

  test("successful login is recorded", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.auditor.username, e2eUsers.auditor.password);
    expect(login.status).toBe(200);
    const userId = await currentUserId(request, login.accessToken!);

    const { client } = await adminClient(request);
    const resp = await client.get(`/admin/audit?event_type=LOGIN_SUCCESS&target_id=${userId}&limit=5`);
    expect(resp.status()).toBe(200);
    const body = (await resp.json()) as { data: { events: Array<{ event_type: string; target_id: string }> } };
    expect(body.data.events.some((e) => e.event_type === "LOGIN_SUCCESS" && e.target_id === userId)).toBe(true);
  });

  test("failed login is recorded", async ({ request }) => {
    const failed = await apiLogin(request, e2eUsers.admin.username, "definitely-wrong-password");
    expect(failed.status).toBe(401);

    const { client, adminId } = await adminClient(request);
    const resp = await client.get(`/admin/audit?event_type=LOGIN_FAILURE&target_id=${adminId}&limit=5`);
    expect(resp.status()).toBe(200);
    const body = (await resp.json()) as { data: { events: Array<{ event_type: string }> } };
    expect(body.data.events.length).toBeGreaterThan(0);
  });

  test("account creation and disable are recorded", async ({ request }) => {
    const { client } = await adminClient(request);
    const username = `e2e_scratch_audit_${Date.now().toString(36)}`;

    const createResp = await client.post("/admin/users", {
      username,
      display_name: "E2E Scratch Audit",
      email: `${username}@ims-e2e.local`,
      password: "E2eScratchAudit#1234",
    });
    expect(createResp.status()).toBe(201);
    const created = (await createResp.json()) as { id: string };
    const scratchId = created.id;

    const disableResp = await client.post(`/admin/users/${scratchId}/disable`);
    expect(disableResp.status()).toBe(204);

    const auditResp = await client.get(`/admin/audit?target_id=${scratchId}&limit=10`);
    const auditBody = (await auditResp.json()) as { data: { events: Array<{ event_type: string }> } };
    const eventTypes = auditBody.data.events.map((e) => e.event_type);
    expect(eventTypes).toContain("USER_CREATED");
    expect(eventTypes).toContain("USER_DEACTIVATED");
  });

  test("no-permission user cannot read the audit log", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.noPermission.username, e2eUsers.noPermission.password);
    const client = withToken(request, login.accessToken!);
    const resp = await client.get("/admin/audit");
    expect(resp.status()).toBe(403);
  });
});
