import { describe, expect, it } from "vitest";

import {
  buildValuationHistorySeries,
  buildValuationSparkline,
  valuationHistoryChange,
} from "../app/features/portfolio-workspace/lib/valuationHistory";

describe("portfolio valuation history", () => {
  it("sorts snapshots, removes invalid values, and collapses duplicate dates", () => {
    const points = buildValuationHistorySeries(
      [
        { business_date: "2026-06-30", aum: "110" },
        { business_date: "2026-06-01", aum: "100" },
        { business_date: "2026-06-30", aum: "112" },
        { business_date: "not-a-date", aum: "999" },
        { business_date: "2026-06-15", aum: "0" },
      ],
      90,
    );

    expect(points).toEqual([
      { date: "2026-06-01", value: 100 },
      { date: "2026-06-30", value: 112 },
    ]);
  });

  it("anchors range filtering to the newest snapshot", () => {
    const points = buildValuationHistorySeries(
      [
        { business_date: "2025-01-01", aum: "80" },
        { business_date: "2026-05-01", aum: "100" },
        { business_date: "2026-06-30", aum: "110" },
      ],
      90,
    );

    expect(points.map((point) => point.date)).toEqual([
      "2026-05-01",
      "2026-06-30",
    ]);
  });

  it("builds bounded line and area paths", () => {
    const chart = buildValuationSparkline(
      [
        { date: "2026-06-01", value: 100 },
        { date: "2026-06-15", value: 90 },
        { date: "2026-06-30", value: 120 },
      ],
      300,
      110,
    );

    expect(chart.linePath).toMatch(/^M6\.00,/);
    expect(chart.linePath).toContain("L294.00,6.00");
    expect(chart.areaPath).toMatch(/Z$/);
  });

  it("reports the percentage change over the selected range", () => {
    expect(
      valuationHistoryChange([
        { date: "2026-06-01", value: 100 },
        { date: "2026-06-30", value: 125 },
      ]),
    ).toBe(25);
    expect(valuationHistoryChange([])).toBeNull();
  });
});
