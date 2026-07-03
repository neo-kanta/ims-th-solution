import { test, expect } from "@playwright/test";
import { apiLogin, withToken } from "../../support/apiClient";
import { e2eUsers, e2eFunds } from "../../support/testUsers";

// Scenario G — Data permission. e2e_trader is seeded (backend/cmd/seed-e2e)
// with permissions_data_rights granting Fund A only. Reuses the existing
// demo funds (database/seeds/zz_demo/01_demo_funds.sql) rather than
// duplicating the investment reference-data schema — see
// backend/cmd/seed-e2e/main.go's fundAID/fundBID comment.
//
// Enforcement is asserted directly against the backend API
// (GET /investment/funds/{id} -> accessibleFundIDs / hasFundAccess in
// backend/internal/investment/transport/handler/investment_handler.go), not
// just frontend filtering — the brief's explicit requirement.
test.describe("IAM — Data permission", () => {
  test("user with access to Fund A can read it", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.trader.username, e2eUsers.trader.password);
    const client = withToken(request, login.accessToken!);
    const resp = await client.get(`/investment/funds/${e2eFunds.fundA.id}`);
    expect(resp.status()).toBe(200);
  });

  test("same user cannot read Fund B", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.trader.username, e2eUsers.trader.password);
    const client = withToken(request, login.accessToken!);
    const resp = await client.get(`/investment/funds/${e2eFunds.fundB.id}`);
    expect(resp.status()).toBe(403);
  });

  test("the fund list is filtered to only accessible funds", async ({ request }) => {
    const login = await apiLogin(request, e2eUsers.trader.username, e2eUsers.trader.password);
    const client = withToken(request, login.accessToken!);
    const resp = await client.get("/investment/funds?limit=200");
    expect(resp.status()).toBe(200);
    const body = (await resp.json()) as { data: { items: Array<{ id: string }> } };
    const ids = body.data.items.map((f) => f.id);
    expect(ids).toContain(e2eFunds.fundA.id);
    expect(ids).not.toContain(e2eFunds.fundB.id);
  });
});
