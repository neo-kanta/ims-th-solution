import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import type { RouteLocationNormalizedLoaded } from "vue-router";

import { portfolioWorkspaceDashboardTabs } from "../app/features/portfolio-workspace/navigation/dashboardTabs";
import { enPortfolioMessages } from "../app/shared/i18n/messages/en/portfolio";
import { thPortfolioMessages } from "../app/shared/i18n/messages/th/portfolio";
import { zhPortfolioMessages } from "../app/shared/i18n/messages/zh/portfolio";

function portfolioRoute(path: string, portfolioCode = "TH/EQ 01") {
  return {
    path,
    params: { portfolioCode },
  } as unknown as RouteLocationNormalizedLoaded;
}

function setupTabs(path: string, portfolioCode?: string) {
  return portfolioWorkspaceDashboardTabs.setup({
    route: portfolioRoute(path, portfolioCode),
    t: (key) => key,
  });
}

describe("portfolio decision workspace navigation", () => {
  it("registers the portfolio workspace tab provider exactly once", () => {
    const registry = readFileSync(
      new URL("../app/features/shell/tabs/dashboardTabRegistry.ts", import.meta.url),
      "utf8",
    );

    expect(registry.match(/^\s+portfolioWorkspaceDashboardTabs,$/gm)).toHaveLength(1);
  });

  it("includes a Decisions tab with an encoded portfolio-code URL", () => {
    const tabs = setupTabs("/portfolios/TH%2FEQ%2001/decisions");
    const decisionTab = tabs.items.value.find((item) => item.key === "decisions");

    expect(decisionTab).toMatchObject({
      key: "decisions",
      label: "portfolio.workspaceTabs.decisions",
      to: "/portfolios/TH%2FEQ%2001/decisions",
    });
  });

  it.each([
    "/portfolios/TH%2FEQ%2001/decisions",
    "/portfolios/TH%2FEQ%2001/decisions/new",
    "/portfolios/TH%2FEQ%2001/decisions/decision%2F123",
  ])("keeps Decisions active for %s", (path) => {
    expect(setupTabs(path).activeKey.value).toBe("decisions");
  });

  it("provides the Decisions label in every supported locale", () => {
    expect(enPortfolioMessages.portfolio.workspaceTabs.decisions).toBeTruthy();
    expect(thPortfolioMessages.portfolio.workspaceTabs.decisions).toBeTruthy();
    expect(zhPortfolioMessages.portfolio.workspaceTabs.decisions).toBeTruthy();
  });
});
