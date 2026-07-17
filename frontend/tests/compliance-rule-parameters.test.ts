import { describe, expect, it } from "vitest";

import {
  asParameterRecord,
  formatParameterEntries,
} from "../app/features/compliance/lib/ruleParameters";

describe("asParameterRecord", () => {
  it("passes through a plain object", () => {
    expect(asParameterRecord({ max_pct: 5 })).toEqual({ max_pct: 5 });
  });

  it("treats an array as unusable — returns an empty record", () => {
    expect(asParameterRecord([1, 2, 3])).toEqual({});
  });

  it("treats null/undefined/primitives as an empty record", () => {
    expect(asParameterRecord(null)).toEqual({});
    expect(asParameterRecord(undefined)).toEqual({});
    expect(asParameterRecord("not-an-object")).toEqual({});
    expect(asParameterRecord(42)).toEqual({});
  });
});

describe("formatParameterEntries", () => {
  it("humanizes snake_case keys into readable labels", () => {
    const entries = formatParameterEntries({ max_percent_nav: 60 });
    expect(entries).toEqual([{ key: "max_percent_nav", label: "Max Percent Nav", value: "60" }]);
  });

  it("humanizes camelCase keys into readable labels", () => {
    const entries = formatParameterEntries({ maxPercentNav: 60 });
    expect(entries[0]?.label).toBe("Max Percent Nav");
  });

  it("sorts entries by key for stable rendering", () => {
    const entries = formatParameterEntries({ b_key: 1, a_key: 2 });
    expect(entries.map((e) => e.key)).toEqual(["a_key", "b_key"]);
  });

  it("returns an empty list for an empty or invalid parameter set", () => {
    expect(formatParameterEntries({})).toEqual([]);
    expect(formatParameterEntries(null)).toEqual([]);
  });
});
