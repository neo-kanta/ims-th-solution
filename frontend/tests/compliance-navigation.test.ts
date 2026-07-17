import { ref } from "vue";
import { describe, expect, it, vi } from "vitest";
import type { RouteLocationNormalizedLoaded } from "vue-router";

vi.mock("../app/features/compliance/composables/useComplianceRuleDirectory", () => ({
  useComplianceRuleDirectory: () => ({
    loading: ref(false),
    error: ref(null),
    total: ref(5),
    ensureLoaded: vi.fn(),
  }),
}));

vi.mock("../app/features/compliance/composables/useComplianceBreaches", () => ({
  useComplianceBreachesList: () => ({
    loading: ref(false),
    error: ref(null),
    total: ref(4),
    fetchList: vi.fn(),
  }),
}));

import { complianceDashboardTabs } from "../app/features/compliance/navigation/dashboardTabs";
import { buildDashboardNavigation } from "../app/features/shell/navigation";

function route(path: string) {
  return { path } as RouteLocationNormalizedLoaded;
}

describe("compliance navigation", () => {
  it("omits simulator, post-trade, and exception routes from the side navigation", () => {
    const sections = buildDashboardNavigation(
      (key, fallback) => typeof fallback === "string" ? fallback : key,
      () => true,
    );
    const routes = sections.flatMap((section) => section.items.map((item) => item.to));

    expect(routes).not.toContain("/compliance/pre-trade");
    expect(routes).not.toContain("/compliance/post-trade");
    expect(routes).not.toContain("/compliance/exceptions");
  });

  it("omits Exceptions and Settings from the compliance header tabs", () => {
    const tabs = complianceDashboardTabs.setup({
      route: route("/compliance"),
      t: (key) => key,
    });

    expect(tabs.items.value.map((item) => item.key)).toEqual([
      "overview",
      "library",
      "approvals",
      "breaches",
      "audit",
    ]);
  });
});
