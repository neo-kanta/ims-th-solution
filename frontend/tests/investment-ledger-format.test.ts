import { describe, expect, it } from "vitest";

import {
  compareDecimal,
  formatMoney,
  formatMoneySigned,
  formatQuantity,
  isNonNegativeNumeric,
  isPositiveNumeric,
  severityTone,
  shortenId,
  statusBadgeVariant,
  verdictLabel,
  verdictTone,
} from "../app/features/investment-ledger/lib/ledgerFormat";

describe("formatMoney / formatMoneySigned", () => {
  it("returns the em dash when no value is supplied", () => {
    expect(formatMoney(null)).toBe("—");
    expect(formatMoney("")).toBe("—");
    expect(formatMoneySigned(undefined)).toBe("—");
  });

  it("formats with currency prefix and tabular numbers", () => {
    expect(formatMoney("1234.5", "THB")).toContain("THB");
    expect(formatMoney("1234.5", "THB")).toContain("1,234.50");
  });

  it("adds a + sign for positive amounts in formatMoneySigned", () => {
    expect(formatMoneySigned("100", "THB")).toMatch(/^\+/);
    // Negative amounts already carry a leading "-" inside the formatted body;
    // we don't add an extra prefix.
    expect(formatMoneySigned("-100", "THB")).not.toMatch(/^\+/);
    expect(formatMoneySigned("-100", "THB")).toContain("-100");
  });
});

describe("formatQuantity", () => {
  it("uses no decimals for whole numbers >= 1", () => {
    expect(formatQuantity("100000")).toBe("100,000");
  });

  it("uses up to 4 decimals for fractions", () => {
    const f = formatQuantity("0.1234");
    expect(f.includes("0.1234")).toBe(true);
  });
});

describe("compareDecimal", () => {
  it("returns the numeric delta", () => {
    expect(compareDecimal("10", "3")).toBe(7);
    expect(compareDecimal("3", "10")).toBe(-7);
    expect(compareDecimal("10", "10")).toBe(0);
  });

  it("pushes nullish values to the end", () => {
    expect(compareDecimal(null, "1")).toBeGreaterThan(0);
    expect(compareDecimal("1", null)).toBeLessThan(0);
  });
});

describe("isPositiveNumeric / isNonNegativeNumeric", () => {
  it("rejects negatives and zero for positive checks", () => {
    expect(isPositiveNumeric("1")).toBe(true);
    expect(isPositiveNumeric("0")).toBe(false);
    expect(isPositiveNumeric("-1")).toBe(false);
    expect(isPositiveNumeric("abc")).toBe(false);
  });

  it("accepts empty string as a valid optional input", () => {
    expect(isNonNegativeNumeric("")).toBe(true);
    expect(isNonNegativeNumeric("0")).toBe(true);
    expect(isNonNegativeNumeric("-0.01")).toBe(false);
  });
});

describe("verdict and severity helpers", () => {
  it("maps verdicts to human labels", () => {
    expect(verdictLabel("PASS")).toBe("Pass");
    expect(verdictLabel("BLOCK")).toBe("Block");
    expect(verdictLabel(null)).toBe("Unknown");
  });

  it("maps verdicts to tones for visual semantics", () => {
    expect(verdictTone("PASS")).toBe("success");
    expect(verdictTone("WARN")).toBe("warning");
    expect(verdictTone("BLOCK")).toBe("error");
    expect(verdictTone(undefined)).toBe("neutral");
  });

  it("maps severities consistently", () => {
    expect(severityTone("HIGH")).toBe("error");
    expect(severityTone("warn")).toBe("warning");
    expect(severityTone("low")).toBe("neutral");
  });
});

describe("statusBadgeVariant", () => {
  it("maps known transaction statuses", () => {
    expect(statusBadgeVariant("POSTED")).toBe("success");
    expect(statusBadgeVariant("PENDING")).toBe("warning");
    expect(statusBadgeVariant("REVERSED")).toBe("error");
    expect(statusBadgeVariant("SIMULATED")).toBe("info");
    expect(statusBadgeVariant("UNKNOWN")).toBe("neutral");
  });
});

describe("shortenId", () => {
  it("preserves short ids", () => {
    expect(shortenId("abc")).toBe("abc");
  });

  it("truncates long uuids", () => {
    expect(shortenId("11111111-1111-1111-1111-111111111111")).toMatch(/…/);
  });
});
