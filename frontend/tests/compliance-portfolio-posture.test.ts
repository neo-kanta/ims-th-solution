import { describe, expect, it } from "vitest";

import {
  derivePortfolioPosture,
  sortPortfolioBreaches,
} from "../app/features/compliance/lib/portfolioPosture";
import type {
  ApiPortfolioBreachView,
  ApiPortfolioRuleCatalogEntry,
} from "../app/features/portfolio-workspace/services/portfolioComplianceApi";

const TODAY = "2026-07-15";

function entry(overrides: Partial<ApiPortfolioRuleCatalogEntry> = {}): ApiPortfolioRuleCatalogEntry {
  return {
    rule_instance_id: "r1",
    rule_type_id: "cash.availability",
    name: "Cash availability",
    description: "",
    is_active: true,
    parameters: {},
    ...overrides,
  };
}

function breach(overrides: Partial<ApiPortfolioBreachView> = {}): ApiPortfolioBreachView {
  return {
    breach_id: "b1",
    business_date: TODAY,
    created_at: `${TODAY}T10:00:00Z`,
    message: "test",
    rule_type_id: "cash.availability",
    severity: "WARN",
    status: "OPEN",
    verdict: "WARN",
    ...overrides,
  };
}

describe("derivePortfolioPosture", () => {
  it("counts only OPEN breaches", () => {
    const posture = derivePortfolioPosture(
      [],
      [breach({ status: "OPEN" }), breach({ status: "OPEN" }), breach({ status: "OVERRIDDEN" })],
      TODAY,
    );
    expect(posture.openBreachCount).toBe(2);
  });

  it("counts currently-effective vs. scheduled bindings separately", () => {
    const effective = entry({
      rule_instance_id: "eff",
      binding: { binding_id: "b-eff", severity: "BLOCK", is_active: true, effective_from: "2026-01-01" },
    });
    const scheduled = entry({
      rule_instance_id: "sched",
      binding: { binding_id: "b-sched", severity: "WARN", is_active: true, effective_from: "2026-08-01" },
    });
    const deactivated = entry({
      rule_instance_id: "deact",
      binding: { binding_id: "b-deact", severity: "BLOCK", is_active: false, effective_from: "2026-01-01" },
    });
    const unbound = entry({ rule_instance_id: "unbound", binding: undefined });

    const posture = derivePortfolioPosture([effective, scheduled, deactivated, unbound], [], TODAY);
    expect(posture.effectiveCount).toBe(1);
    expect(posture.scheduledCount).toBe(1);
  });

  it("reports the strongest severity among effective bindings only", () => {
    const warnEffective = entry({
      rule_instance_id: "warn",
      binding: { binding_id: "b1", severity: "WARN", is_active: true, effective_from: "2026-01-01" },
    });
    const blockScheduled = entry({
      rule_instance_id: "block-sched",
      binding: { binding_id: "b2", severity: "BLOCK", is_active: true, effective_from: "2099-01-01" },
    });

    const posture = derivePortfolioPosture([warnEffective, blockScheduled], [], TODAY);
    // The stronger BLOCK binding is only SCHEDULED, not effective yet, so the
    // effective WARN binding must be what's reported — never a not-yet-active severity.
    expect(posture.strongestEffectiveSeverity).toBe("WARN");
  });

  it("reports null strongest severity when nothing is effective", () => {
    const posture = derivePortfolioPosture([], [], TODAY);
    expect(posture.strongestEffectiveSeverity).toBeNull();
  });
});

describe("sortPortfolioBreaches", () => {
  it("orders by severity then recency, same as the control-center queue", () => {
    const warn = breach({ breach_id: "warn", severity: "WARN", created_at: "2026-07-15T09:00:00Z" });
    const block = breach({ breach_id: "block", severity: "BLOCK", created_at: "2026-07-15T08:00:00Z" });
    const sorted = sortPortfolioBreaches([warn, block]);
    expect(sorted.map((b) => b.breach_id)).toEqual(["block", "warn"]);
  });
});
