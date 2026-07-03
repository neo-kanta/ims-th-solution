import { describe, expect, it } from "vitest";
import { readFileSync } from "fs";
import { resolve } from "path";

function readAppFile(relPath: string): string {
  return readFileSync(resolve(__dirname, "../app", relPath), "utf-8");
}

const apiSrc = readAppFile(
  "features/investment-workspace/services/intradayValuationApi.ts",
);
const pageSrc = readAppFile(
  "pages/investment/funds/[fundId]/holdings/index.vue",
);
const tableSrc = readAppFile(
  "features/investment-workspace/components/LivePositionsTable.vue",
);
const myFundsApiSrc = readAppFile("features/my-funds/services/myFundsApi.ts");

describe("intradayValuationApi — business_date threading", () => {
  it("getFundValuation accepts an optional businessDate parameter", () => {
    expect(apiSrc).toMatch(/getFundValuation\(fundId: string, businessDate\?: string\)/);
  });

  it("passes business_date as a query param, not a hand-built URL", () => {
    expect(apiSrc).toContain("business_date: businessDate");
  });

  it("does not fabricate a valuation on error — 404 maps to null, other errors rethrow", () => {
    expect(apiSrc).toContain("if (err instanceof OpenApiRequestError && err.status === 404) return null;");
    expect(apiSrc).toContain("throw err;");
  });
});

describe("holdings page — date picker refetches the valuation (regression: previously only refreshed fund detail)", () => {
  it("onChangeAsOf refetches the intraday valuation, not just fund detail", () => {
    const match = pageSrc.match(/async function onChangeAsOf\([\s\S]*?\n}/);
    expect(match, "onChangeAsOf function not found").toBeTruthy();
    const body = match![0];
    expect(body).toContain("refreshIntradayValuation");
  });

  it("onChangeAsOf also refetches allocation so it never disagrees with holdings for the new date", () => {
    const match = pageSrc.match(/async function onChangeAsOf\([\s\S]*?\n}/);
    const body = match![0];
    expect(body).toContain("refreshAllocation");
  });

  it("refreshIntradayValuation passes the selected as-of date to the API", () => {
    const match = pageSrc.match(/async function refreshIntradayValuation\([\s\S]*?\n}/);
    expect(match, "refreshIntradayValuation function not found").toBeTruthy();
    const body = match![0];
    expect(body).toContain("getFundValuation(");
    expect(body).toContain("holdings.asOf.value");
  });

  it("refresh-all handler reloads valuation, allocation, and feed status together", () => {
    const match = pageSrc.match(/async function onRefreshAll\([\s\S]*?\n}/);
    expect(match, "onRefreshAll function not found").toBeTruthy();
    const body = match![0];
    expect(body).toContain("refreshIntradayValuation");
    expect(body).toContain("refreshAllocation");
    expect(body).toContain("refreshFeedStatus");
  });

  it("watch on fundId refetches the intraday valuation so switching funds does not show stale data", () => {
    const watchBlock = pageSrc.match(/watch\(\s*\(\) => fundIdParam\.value,[\s\S]*?\n\);/);
    expect(watchBlock, "fundId watcher not found").toBeTruthy();
    expect(watchBlock![0]).toContain("refreshIntradayValuation");
  });

  it("CSV export and KPIs read exclusively from the backend valuation response — no hardcoded fallback numbers", () => {
    // Every reference to official/estimated AUM in the export must come from
    // `intraday.value?.` — never a literal numeric fallback.
    expect(pageSrc).not.toMatch(/intraday\.value\s*\?\?\s*["']?\d/);
    expect(pageSrc).toContain('intraday.value?.official_aum ?? ""');
    expect(pageSrc).toContain('intraday.value?.estimated_aum ?? ""');
  });
});

describe("allocation — same business_date as holdings, so the two pages agree", () => {
  it("getFundAllocation accepts an optional businessDate and forwards it as business_date", () => {
    expect(myFundsApiSrc).toMatch(/getFundAllocation\(fundId: string, businessDate\?: string\)/);
    expect(myFundsApiSrc).toContain("business_date: businessDate");
  });

  it("refreshAllocation passes the selected as-of date to the API", () => {
    const match = pageSrc.match(/async function refreshAllocation\([\s\S]*?\n}/);
    expect(match, "refreshAllocation function not found").toBeTruthy();
    const body = match![0];
    expect(body).toContain("getFundAllocation(");
    expect(body).toContain("holdings.asOf.value");
  });
});

describe("LivePositionsTable — mark-to-market source/effective-date/stale rendering", () => {
  it("renders the price source label alongside the feed badge", () => {
    expect(tableSrc).toContain("sourceLabel(row.source)");
  });

  it("renders the price effective date per position", () => {
    expect(tableSrc).toContain("row.price_effective_date");
  });

  it("keeps the existing stale/live indicator driven by row.is_stale", () => {
    expect(tableSrc).toContain("row.is_stale ? 'lp-feed--stale' : 'lp-feed--live'");
  });

  it("maps every known backend source tag to a human label", () => {
    const knownSources = [
      "live_quote",
      "market_data_snapshot",
      "official_price_snapshot",
      "cost_carry",
    ];
    for (const source of knownSources) {
      expect(tableSrc).toContain(`${source}:`);
    }
  });
});
