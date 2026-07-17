import { describe, expect, it } from "vitest";

import { countOrNull, deviceLocalIsoDate } from "../app/features/compliance/lib/asOfDate";

describe("countOrNull", () => {
  it("returns the real value when not loading and no error", () => {
    expect(countOrNull(5, false, false)).toBe(5);
  });

  it("returns null while loading — never a misleading zero", () => {
    expect(countOrNull(0, true, false)).toBeNull();
  });

  it("returns null on error, even if a stale count is available", () => {
    expect(countOrNull(3, false, true)).toBeNull();
  });
});

describe("deviceLocalIsoDate", () => {
  it("returns a YYYY-MM-DD date string", () => {
    expect(deviceLocalIsoDate()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it("uses the device-local calendar day rather than UTC", () => {
    const nearUtcMidnight = new Date(2026, 6, 16, 0, 15, 0);
    expect(deviceLocalIsoDate(nearUtcMidnight)).toBe("2026-07-16");
  });
});
