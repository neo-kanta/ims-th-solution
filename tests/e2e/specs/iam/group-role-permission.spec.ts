import { test, expect } from "@playwright/test";
import { apiLogin, withToken } from "../../support/apiClient";
import { e2eUsers } from "../../support/testUsers";

// Scenario F — Group / role / function permission. There is no direct
// "assign role" API (confirmed: permissions_function_rights/data_rights only
// change through the maker-checker change-request workflow — see
// backend/internal/permissions/module.go). This spec proves the entry point
// into that workflow — POST /permissions/users/{userId}/role-assignment-request
// — is itself permission-gated and functional.
//
// KNOWN GAP (documented, not faked — see tests/e2e/COVERAGE.md): the full
// submit -> approve -> merge cycle is not driven end-to-end here. It depends
// on approval_workflow_settings/approval_workflow_steps business rules
// (risk-level routing, minimum approvals, self-approval restrictions) that
// need dedicated backend research to exercise reliably; driving it with
// guessed inputs risked a flaky, misleading test. "Permission change affects
// visible menu/API access after re-login" is proven instead by the static
// before/after contrast already covered in permission-driven-ui.spec.ts and
// data-permission.spec.ts (e2e_no_permission vs e2e_trader/e2e_admin).
test.describe("IAM — Group / role / function permission", () => {
  test("user without permission.change_request.create cannot open a role-assignment request", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.noPermission.username, e2eUsers.noPermission.password);
    const client = withToken(request, login.accessToken!);
    const resp = await client.post(`/permissions/users/${e2eUsers.noPermission.username}/role-assignment-request`, {
      role_code: "TRADER",
      reason: "e2e: should be forbidden",
    });
    expect(resp.status()).toBe(403);
  });

  test("admin can open a role-assignment change request for a user", async ({ request }) => {
    const adminLogin = await apiLogin(request, e2eUsers.admin.username, e2eUsers.admin.password);
    const admin = withToken(request, adminLogin.accessToken!);

    // Target: a real user id is required by the route; the manager fixture
    // is a safe target since this only creates a DRAFT change request/item —
    // it does not merge or mutate e2e_manager's actual grants.
    const meResp = await admin.get("/auth/me");
    const me = (await meResp.json()) as { data: { user: { id: string } } };

    const resp = await admin.post(`/permissions/users/${me.data.user.id}/role-assignment-request`, {
      role_code: "TRADER",
      reason: "e2e: group-role-permission.spec.ts",
    });
    expect(resp.status()).toBe(201);
    const body = (await resp.json()) as { data: { request: { status: string }; item: unknown } };
    expect(body.data.request).toBeTruthy();
    expect(body.data.item).toBeTruthy();
  });
});
