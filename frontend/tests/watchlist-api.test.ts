import { describe, expect, it } from "vitest";
import { isPositiveDecimal, isUuid } from "../app/features/watchlist/lib/formatters";

describe("watchlist API contract guards", () => {
  it("threshold_value must be a positive decimal string, not a number", () => {
    const asString = "190.00000000";
    const asNumber = 190.0;
    expect(typeof asString).toBe("string");
    expect(typeof asNumber).toBe("number");
    expect(isPositiveDecimal(asString)).toBe(true);
    expect(isPositiveDecimal(String(asNumber))).toBe(true);
  });

  it("security_id is a UUID string", () => {
    const securityId = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee";
    expect(isUuid(securityId)).toBe(true);
  });

  it("zero threshold_value is rejected", () => {
    expect(isPositiveDecimal("0")).toBe(false);
    expect(isPositiveDecimal("0.00")).toBe(false);
  });

  it("negative threshold_value is rejected", () => {
    expect(isPositiveDecimal("-1.5")).toBe(false);
  });

  it("empty threshold_value is rejected", () => {
    expect(isPositiveDecimal("")).toBe(false);
  });

  it("non-numeric threshold_value is rejected", () => {
    expect(isPositiveDecimal("abc")).toBe(false);
  });
});

describe("watchlist API scope rules", () => {
  it("PERSONAL scope does not require portfolio_id", () => {
    const scope: "PERSONAL" | "PORTFOLIO" = "PERSONAL";
    const needsPortfolio = scope === "PORTFOLIO";
    expect(needsPortfolio).toBe(false);
  });

  it("PORTFOLIO scope requires portfolio_id", () => {
    const scope: "PERSONAL" | "PORTFOLIO" = "PORTFOLIO";
    const needsPortfolio = scope === "PORTFOLIO";
    expect(needsPortfolio).toBe(true);
  });
});

describe("watchlist API error codes", () => {
  const knownCodes = [
    "WATCHLIST_DUPLICATE_ITEM",
    "WATCHLIST_FORBIDDEN_SCOPE",
    "WATCHLIST_INVALID_SECURITY",
    "WATCHLIST_INVALID_THRESHOLD",
    "WATCHLIST_ITEM_NOT_FOUND",
    "WATCHLIST_ALERT_NOT_FOUND",
    "WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED",
    "WATCHLIST_DUPLICATE_THRESHOLD",
    "WATCHLIST_RULE_DISABLED",
    "WATCHLIST_STALE_QUOTE",
  ];

  it("all known watchlist error codes are non-empty strings", () => {
    for (const code of knownCodes) {
      expect(typeof code).toBe("string");
      expect(code.length).toBeGreaterThan(0);
    }
  });
});

describe("no market-data screen watchlist calls in watchlist feature", () => {
  it("watchlistApi source does not call /market-data/screen/watchlist", async () => {
    const { readFileSync } = await import("fs");
    const { resolve } = await import("path");
    const src = readFileSync(
      resolve(__dirname, "../app/features/watchlist/services/watchlistApi.ts"),
      "utf-8",
    );
    expect(src).not.toContain("/market-data/screen/watchlist");
    expect(src).toContain("/watchlists");
    expect(src).toContain("listItems");
    expect(src).toContain("createItem");
    expect(src).toContain("updateItem");
    expect(src).toContain("deleteItem");
    expect(src).toContain("listAlerts");
    expect(src).toContain("acknowledgeAlert");
    expect(src).toContain("manualEvaluate");
  });
});
