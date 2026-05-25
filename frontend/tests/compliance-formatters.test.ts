import { describe, expect, it } from "vitest";

import {
  deriveRuleStatus,
  formatEffectiveWindow,
  formatIsoDate,
  isPositiveDecimal,
  isUuid,
  ruleStatusLabel,
  ruleStatusTone,
  severityLabel,
  severityTone,
  verdictLabel,
  verdictTone,
} from "../app/features/compliance/lib/formatters";
import type { ComplianceRule } from "../app/features/compliance/types";

function makeRule(overrides: Partial<ComplianceRule> = {}): ComplianceRule {
  return {
    id: "11111111-1111-1111-1111-111111111111",
    ruleTypeID: "concentration.single_issuer",
    name: "Test rule",
    description: "",
    currentVersion: 1,
    isActive: true,
    effectiveWindow: undefined,
    createdBy: "22222222-2222-2222-2222-222222222222",
    createdAt: "2026-05-01T00:00:00Z",
    updatedAt: "2026-05-15T00:00:00Z",
    ...overrides,
  };
}

describe("verdict helpers", () => {
  it("maps each verdict to a status tone", () => {
    expect(verdictTone("PASS")).toBe("success");
    expect(verdictTone("WARN")).toBe("warning");
    expect(verdictTone("BLOCK")).toBe("error");
  });

  it("returns the verdict label verbatim (used in screen-reader text)", () => {
    expect(verdictLabel("PASS")).toBe("PASS");
    expect(verdictLabel("WARN")).toBe("WARN");
    expect(verdictLabel("BLOCK")).toBe("BLOCK");
  });
});

describe("severity helpers", () => {
  it("maps BLOCK severity to the error tone (highest urgency)", () => {
    expect(severityTone("BLOCK")).toBe("error");
  });

  it("treats MONITOR as informational neutral, not warning", () => {
    expect(severityTone("MONITOR")).toBe("neutral");
  });

  it("labels REQUIRE_APPROVAL with the spec copy", () => {
    expect(severityLabel("REQUIRE_APPROVAL")).toBe("Requires approval");
  });
});

describe("deriveRuleStatus", () => {
  // Frozen "today" for deterministic effective-window comparisons.
  const NOW = new Date("2026-05-22T00:00:00Z");

  it("returns DISABLED when isActive is false, regardless of window", () => {
    const rule = makeRule({
      isActive: false,
      effectiveWindow: { from: "2025-01-01", to: "2099-01-01" },
    });
    expect(deriveRuleStatus(rule, NOW)).toBe("DISABLED");
  });

  it("returns SCHEDULED when effective window starts in the future", () => {
    const rule = makeRule({
      effectiveWindow: { from: "2099-01-01", to: null },
    });
    expect(deriveRuleStatus(rule, NOW)).toBe("SCHEDULED");
  });

  it("returns EXPIRED when effective window ended in the past", () => {
    const rule = makeRule({
      effectiveWindow: { from: "2020-01-01", to: "2025-12-31" },
    });
    expect(deriveRuleStatus(rule, NOW)).toBe("EXPIRED");
  });

  it("returns ACTIVE when the window straddles today", () => {
    const rule = makeRule({
      effectiveWindow: { from: "2026-01-01", to: "2099-01-01" },
    });
    expect(deriveRuleStatus(rule, NOW)).toBe("ACTIVE");
  });

  it("returns ACTIVE when the window is missing entirely (open-ended)", () => {
    const rule = makeRule({ effectiveWindow: undefined });
    expect(deriveRuleStatus(rule, NOW)).toBe("ACTIVE");
  });

  it("derived statuses each map to a distinct tone", () => {
    expect(ruleStatusTone("ACTIVE")).toBe("success");
    expect(ruleStatusTone("SCHEDULED")).toBe("info");
    expect(ruleStatusTone("EXPIRED")).toBe("warning");
    expect(ruleStatusTone("DISABLED")).toBe("neutral");

    expect(ruleStatusLabel("ACTIVE")).toBe("Active");
    expect(ruleStatusLabel("DISABLED")).toBe("Disabled");
  });
});

describe("formatEffectiveWindow", () => {
  it("renders both-sided window with an arrow", () => {
    expect(formatEffectiveWindow({ from: "2026-01-01", to: "2026-12-31" })).toBe(
      "2026-01-01 → 2026-12-31",
    );
  });

  it("renders left-sided window with 'From' prefix", () => {
    expect(formatEffectiveWindow({ from: "2026-01-01", to: null })).toBe(
      "From 2026-01-01",
    );
  });

  it("renders right-sided window with 'Until' prefix", () => {
    expect(formatEffectiveWindow({ from: null, to: "2026-12-31" })).toBe(
      "Until 2026-12-31",
    );
  });

  it("returns '—' when both ends are missing or undefined", () => {
    expect(formatEffectiveWindow(undefined)).toBe("—");
    expect(formatEffectiveWindow({})).toBe("—");
  });
});

describe("formatIsoDate", () => {
  it("trims an ISO datetime down to the date component", () => {
    expect(formatIsoDate("2026-05-22T11:30:00Z")).toBe("2026-05-22");
  });

  it("returns '—' for null/empty input", () => {
    expect(formatIsoDate(null)).toBe("—");
    expect(formatIsoDate(undefined)).toBe("—");
  });
});

describe("input validators", () => {
  it("accepts canonical UUIDs and rejects malformed strings", () => {
    expect(isUuid("11111111-1111-1111-1111-111111111111")).toBe(true);
    expect(isUuid("not-a-uuid")).toBe(false);
    expect(isUuid("")).toBe(false);
    // Trailing spaces should be tolerated — backend will trim.
    expect(isUuid("  11111111-1111-1111-1111-111111111111  ")).toBe(true);
  });

  it("only accepts strictly positive decimal numbers", () => {
    expect(isPositiveDecimal("1")).toBe(true);
    expect(isPositiveDecimal("100.5")).toBe(true);
    expect(isPositiveDecimal("0")).toBe(false);
    expect(isPositiveDecimal("-1")).toBe(false);
    expect(isPositiveDecimal("abc")).toBe(false);
    expect(isPositiveDecimal("")).toBe(false);
  });
});
