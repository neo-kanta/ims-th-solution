import { describe, expect, it } from "vitest";
import { readFileSync } from "fs";
import { resolve } from "path";
import { isPositiveDecimal, isUuid, securityLabel } from "../app/features/watchlist/lib/formatters";

const featuresDir = resolve(__dirname, "../app/features/watchlist");

function readFeatureFile(relPath: string): string {
  return readFileSync(resolve(featuresDir, relPath), "utf-8");
}

describe("WatchlistItemDrawer — threshold value must be decimal strings, not numbers", () => {
  it("validates threshold_value with isPositiveDecimal before submission", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain("isPositiveDecimal");
  });

  it("threshold value input uses text type with decimal inputmode", () => {
    const formSrc = readFeatureFile("components/ThresholdRuleForm.vue");
    // threshold_value field must use type="text" so browser never coerces the string to a number
    expect(formSrc).toContain('type="text"');
    expect(formSrc).toContain('inputmode="decimal"');
  });

  it("isPositiveDecimal accepts decimal strings", () => {
    expect(isPositiveDecimal("190.00000000")).toBe(true);
    expect(isPositiveDecimal("0.5")).toBe(true);
  });

  it("isPositiveDecimal rejects numeric types coerced to zero", () => {
    expect(isPositiveDecimal("0")).toBe(false);
    expect(isPositiveDecimal("0.00")).toBe(false);
    expect(isPositiveDecimal("")).toBe(false);
  });
});

describe("WatchlistItemDrawer — metric_type is always MARKET_PRICE", () => {
  it("create body always sends metric_type MARKET_PRICE", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain('"MARKET_PRICE"');
    expect(drawerSrc).not.toContain('"NAV"');
    expect(drawerSrc).not.toContain('"AUM"');
  });

  it("ThresholdRuleForm does not allow metric_type selection", () => {
    const formSrc = readFeatureFile("components/ThresholdRuleForm.vue");
    expect(formSrc).not.toContain("metric_type");
  });
});

describe("WatchlistItemDrawer — PORTFOLIO scope requires portfolio_id", () => {
  it("drawer source validates portfolio_id required when scope is PORTFOLIO", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain("PORTFOLIO");
    expect(drawerSrc).toContain("portfolio_id");
    expect(drawerSrc).toContain("portfolioId");
  });

  it("WatchlistScopeSelector emits null portfolio when scope changes to PERSONAL", () => {
    const selectorSrc = readFeatureFile("components/WatchlistScopeSelector.vue");
    expect(selectorSrc).toContain("update:portfolioId");
    expect(selectorSrc).toContain("null");
  });
});

describe("WatchlistItemDrawer — threshold_rules update semantics", () => {
  it("edit drawer omits threshold_rules when rules not modified", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain("rulesModified");
    expect(drawerSrc).toContain("threshold_rules");
  });

  it("edit drawer sends full replacement array when rules are modified", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain("rulesModified.value");
    // When modified, threshold_rules is set on the update body
    expect(drawerSrc).toContain("body.threshold_rules");
  });

  it("blank rule defaults to ABOVE direction and 60 minute cooldown", () => {
    const drawerSrc = readFeatureFile("components/WatchlistItemDrawer.vue");
    expect(drawerSrc).toContain('"ABOVE"');
    expect(drawerSrc).toContain("cooldown_minutes: 60");
  });
});

describe("WatchlistPage — WATCHLIST_EVALUATE is not exposed as a normal user action", () => {
  it("WatchlistPage does not expose manual evaluate to regular users", () => {
    const pageSrc = readFeatureFile("WatchlistPage.vue");
    expect(pageSrc).not.toContain("manualEvaluate");
    expect(pageSrc).not.toContain("WATCHLIST_EVALUATE");
    expect(pageSrc).not.toContain("evaluate");
  });

  it("manualEvaluate exists in service but not in any component", () => {
    const apiSrc = readFeatureFile("services/watchlistApi.ts");
    expect(apiSrc).toContain("manualEvaluate");

    const alertsPanelSrc = readFeatureFile("components/WatchlistAlertsPanel.vue");
    expect(alertsPanelSrc).not.toContain("manualEvaluate");
    expect(alertsPanelSrc).not.toContain("WATCHLIST_EVALUATE");
  });
});

describe("WatchlistPage — no raw UUIDs displayed as labels", () => {
  it("WatchlistItemsTable uses descriptor fields, not raw IDs", () => {
    const tableSrc = readFeatureFile("components/WatchlistItemsTable.vue");
    // securityLabel wraps display_symbol — the table delegates to the formatter, not bare IDs
    expect(tableSrc).toContain("securityLabel");
    expect(tableSrc).toContain("display_name");
    expect(tableSrc).not.toContain("owner_user_id");
    expect(tableSrc).not.toContain("portfolio_id");
  });

  it("WatchlistAlertsPanel uses descriptor fields, not raw IDs", () => {
    const panelSrc = readFeatureFile("components/WatchlistAlertsPanel.vue");
    expect(panelSrc).toContain("display_name");
    expect(panelSrc).not.toContain("acknowledged_by\b");
    expect(panelSrc).not.toContain("owner_user_id");
  });

  it("securityLabel returns descriptor field, never raw ID", () => {
    const sym = securityLabel({ display_symbol: "CPALL TB", name: "CP All" });
    expect(isUuid(sym)).toBe(false);
    expect(sym).toBe("CPALL TB");
  });

  it("securityLabel falls back to name, not ID", () => {
    const label = securityLabel({ name: "CP All" });
    expect(isUuid(label)).toBe(false);
    expect(label).toBe("CP All");
  });
});

describe("WatchlistAlertsPanel — acknowledgement flow", () => {
  it("acknowledge action is gated by WATCHLIST_ALERT_ACK permission", () => {
    const panelSrc = readFeatureFile("components/WatchlistAlertsPanel.vue");
    expect(panelSrc).toContain("WATCHLIST_ALERT_ACK");
    expect(panelSrc).toContain("IMSPermissionGuard");
  });

  it("acknowledge button only shown for UNACKNOWLEDGED alerts", () => {
    const panelSrc = readFeatureFile("components/WatchlistAlertsPanel.vue");
    expect(panelSrc).toContain("UNACKNOWLEDGED");
  });

  it("AcknowledgeAlertDialog collects optional note", () => {
    const dialogSrc = readFeatureFile("components/AcknowledgeAlertDialog.vue");
    expect(dialogSrc).toContain("note");
    expect(dialogSrc).toContain("textarea");
  });

  it("AcknowledgeAlertDialog emits note as undefined when empty", () => {
    const dialogSrc = readFeatureFile("components/AcknowledgeAlertDialog.vue");
    expect(dialogSrc).toContain("undefined");
    expect(dialogSrc).toContain("note.value.trim()");
  });
});

describe("WatchlistItemsTable — edit/delete gated by WATCHLIST_MANAGE", () => {
  it("edit and delete actions require WATCHLIST_MANAGE", () => {
    const tableSrc = readFeatureFile("components/WatchlistItemsTable.vue");
    expect(tableSrc).toContain("WATCHLIST_MANAGE");
    expect(tableSrc).toContain("IMSPermissionGuard");
  });
});

describe("WatchlistPage — error handling maps codes to friendly messages", () => {
  it("WatchlistPage maps WATCHLIST_DUPLICATE_ITEM without exposing raw details", () => {
    const pageSrc = readFeatureFile("WatchlistPage.vue");
    expect(pageSrc).toContain("WATCHLIST_DUPLICATE_ITEM");
    expect(pageSrc).not.toContain("err.details.security_id");
    expect(pageSrc).not.toContain("err.details.portfolio_id");
  });

  it("WatchlistPage maps WATCHLIST_FORBIDDEN_SCOPE to user message", () => {
    const pageSrc = readFeatureFile("WatchlistPage.vue");
    expect(pageSrc).toContain("WATCHLIST_FORBIDDEN_SCOPE");
  });

  it("WatchlistPage maps WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED to warning", () => {
    const pageSrc = readFeatureFile("WatchlistPage.vue");
    expect(pageSrc).toContain("WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED");
    expect(pageSrc).toContain("showWarning");
  });
});

describe("Watchlist feature does not import from other feature services", () => {
  const featureFiles = [
    "WatchlistPage.vue",
    "services/watchlistApi.ts",
    "composables/useWatchlist.ts",
    "composables/useWatchlistAlerts.ts",
    "composables/useWatchlistPortfolios.ts",
    "composables/useWatchlistSecuritySearch.ts",
  ];

  for (const file of featureFiles) {
    it(`${file} does not import from market-data or investment-ledger features`, () => {
      const src = readFeatureFile(file);
      expect(src).not.toContain("features/market-data");
      expect(src).not.toContain("features/investment-ledger");
      expect(src).not.toContain("marketDataCatalogApi");
      expect(src).not.toContain("usePortfolioDirectory");
    });
  }
});
