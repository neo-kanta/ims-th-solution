import { describe, expect, it } from "vitest";

import { normalizeTranslateArgs, translateMessage } from "../app/shared/i18n/core";
import { messages } from "../app/shared/i18n/messages";
import {
  buildAumMetric,
  buildPnlMetric,
  buildScopeOptions,
  defaultAumScope,
} from "../app/features/dashboard/lib/dashboard";
import { normalizeValuationSummary } from "../app/features/dashboard/lib/valuationSummaryMapping";
import type { ValuationSummaryDTO } from "../app/features/dashboard/types";

// Real translate function backed by the actual EN message catalog, so these
// tests also catch a typo'd or missing i18n key for the new dashboard
// scope/metric copy.
function t(
  key: string,
  paramsOrFallback?: Record<string, unknown> | string,
  fallback?: string,
): string {
  const normalized = normalizeTranslateArgs(paramsOrFallback, fallback);
  return translateMessage({
    key,
    locale: "en",
    messages,
    fallback: normalized.fallback,
    params: normalized.params,
  });
}

function makeSummary(overrides: Partial<ValuationSummaryDTO> = {}): ValuationSummaryDTO {
  return {
    scope: "company",
    username: null,
    businessDate: "2026-07-14",
    currency: "THB",
    aumToday: "1234567.89",
    todayPnl: "12345.67",
    todayPnlPercent: "1.01",
    asOf: "2026-07-14T14:30:00Z",
    dataAvailable: true,
    ...overrides,
  };
}

describe("normalizeValuationSummary (API integration mapping)", () => {
  it("maps the snake_case wire shape to the camelCase DTO", () => {
    const dto = normalizeValuationSummary({
      scope: "mine",
      username: "neo",
      business_date: "2026-07-14",
      currency: "THB",
      aum_today: "1234567.89",
      today_pnl: "12345.67",
      today_pnl_percent: "1.01",
      as_of: "2026-07-14T14:30:00Z",
      data_available: true,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any);

    expect(dto).toEqual({
      scope: "mine",
      username: "neo",
      businessDate: "2026-07-14",
      currency: "THB",
      aumToday: "1234567.89",
      todayPnl: "12345.67",
      todayPnlPercent: "1.01",
      asOf: "2026-07-14T14:30:00Z",
      dataAvailable: true,
    });
  });

  it("defaults to an explicit unavailable state when the payload is missing, never a fabricated zero", () => {
    const dto = normalizeValuationSummary(undefined);
    expect(dto.dataAvailable).toBe(false);
    expect(dto.scope).toBe("company");
    expect(dto.username).toBeNull();
    expect(dto.aumToday).toBe("");
  });

  it("treats any scope value other than \"mine\" as company", () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    expect(normalizeValuationSummary({ scope: "bogus" } as any).scope).toBe("company");
  });
});

describe("defaultAumScope (permission-driven default)", () => {
  it("defaults to company only when the caller has wildcard data scope", () => {
    expect(defaultAumScope(true)).toBe("company");
  });

  it("defaults to mine for a restricted (non-wildcard) data scope", () => {
    expect(defaultAumScope(false)).toBe("mine");
  });
});

describe("buildScopeOptions (permission-driven visibility)", () => {
  it("offers both options to a caller with wildcard company-wide access", () => {
    const options = buildScopeOptions(true, t);
    expect(options.map((o) => o.key)).toEqual(["company", "mine"]);
  });

  it("hides the company option for a restricted caller — no unauthorized company-wide option is exposed", () => {
    const options = buildScopeOptions(false, t);
    expect(options.map((o) => o.key)).toEqual(["mine"]);
  });
});

describe("buildAumMetric", () => {
  it("renders an explicit not-available state (not a zero) when no valuation snapshot exists", () => {
    const metric = buildAumMetric(null, false, t);
    expect(metric.value).toBe("—");
    expect(metric.changeTone).toBe("neutral");
    expect(metric.helperText).toBe("Not yet available");
  });

  it("renders an explicit not-available state when dataAvailable is false", () => {
    const metric = buildAumMetric(makeSummary({ dataAvailable: false }), false, t);
    expect(metric.value).toBe("—");
  });

  it("renders a distinct load-error state on fetch failure", () => {
    const metric = buildAumMetric(null, true, t);
    expect(metric.value).toBe("—");
    expect(metric.helperText).toBe("Could not load");
  });

  it("formats the aggregate AUM with currency and an as-of timestamp when data is available", () => {
    const metric = buildAumMetric(makeSummary(), false, t);
    expect(metric.value).not.toBe("—");
    expect(metric.value).toContain("฿");
    // asOf "2026-07-14T14:30:00Z" rendered in Asia/Bangkok (UTC+7) is 21:30,
    // not the raw UTC hour — the dashboard must display local Bangkok time.
    expect(metric.helperText).toContain("21:30");
  });
});

describe("buildPnlMetric", () => {
  it("renders success tone and a signed positive value for positive P&L", () => {
    const metric = buildPnlMetric(
      makeSummary({ todayPnl: "12345.67", todayPnlPercent: "1.01" }),
      false,
      t,
    );
    expect(metric.tone).toBe("success");
    expect(metric.changeTone).toBe("success");
    expect(metric.value.startsWith("+")).toBe(true);
    expect(metric.changeLabel).toContain("+");
  });

  it("renders danger tone and a signed negative value for negative P&L", () => {
    const metric = buildPnlMetric(
      makeSummary({ todayPnl: "-500", todayPnlPercent: "-0.5" }),
      false,
      t,
    );
    expect(metric.tone).toBe("danger");
    expect(metric.changeTone).toBe("danger");
    expect(metric.value.startsWith("+")).toBe(false);
  });

  it("does not look like a zero when data is unavailable", () => {
    const metric = buildPnlMetric(null, false, t);
    expect(metric.value).toBe("—");
    expect(metric.value).not.toBe("0");
    expect(metric.changeTone).toBe("neutral");
  });

  it("renders a distinct load-error state on fetch failure", () => {
    const metric = buildPnlMetric(null, true, t);
    expect(metric.helperText).toBe("Could not load");
  });
});
