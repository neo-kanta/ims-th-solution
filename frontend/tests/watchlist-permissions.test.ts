import { describe, expect, it } from "vitest";
import { resolveRequiredPermission } from "../app/shared/routing/routeAccess";
import { buildDashboardNavigation } from "../app/features/shell/navigation";

describe("watchlist route permission", () => {
  it("requires WATCHLIST_VIEW via page meta", () => {
    expect(
      resolveRequiredPermission({ permission: "WATCHLIST_VIEW" }),
    ).toBe("WATCHLIST_VIEW");
  });
});

describe("watchlist navigation visibility", () => {
  const t = (key: string, fallback?: string) => fallback ?? key;

  it("shows watchlist item when user has WATCHLIST_VIEW", () => {
    const nav = buildDashboardNavigation(t, (code) => code === "WATCHLIST_VIEW");
    const investmentSection = nav.find((s) => s.items.some((i) => i.to === "/watchlist"));
    expect(investmentSection).toBeDefined();
    const item = investmentSection?.items.find((i) => i.to === "/watchlist");
    expect(item).toBeDefined();
    expect(item?.label).toBe("Watchlist");
  });

  it("hides watchlist item when user lacks WATCHLIST_VIEW", () => {
    const nav = buildDashboardNavigation(t, () => false);
    const hasWatchlist = nav.some((s) => s.items.some((i) => i.to === "/watchlist"));
    expect(hasWatchlist).toBe(false);
  });

  it("watchlist nav item requires WATCHLIST_VIEW permission", () => {
    const nav = buildDashboardNavigation(t, () => true);
    const item = nav.flatMap((s) => s.items).find((i) => i.to === "/watchlist");
    expect(item?.requiredPermissions).toContain("WATCHLIST_VIEW");
  });
});

describe("watchlist permission codes", () => {
  it("WATCHLIST_VIEW is required for viewing", () => {
    expect(resolveRequiredPermission({ permission: "WATCHLIST_VIEW" })).toBe("WATCHLIST_VIEW");
  });
});
