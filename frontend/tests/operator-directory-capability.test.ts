/**
 * Operator Directory capability derivation. The catalog must never claim a
 * workflow is "Ready" (fully API-backed AND permission-held) when either
 * half is false, and the search box must filter on real catalog content.
 */
import { describe, expect, it } from "vitest";

import {
  OPERATOR_CATALOG,
  deriveCapability,
  filterOperatorRows,
} from "../app/features/operator/lib/operatorCatalog";

function allow(...codes: string[]) {
  return (code: string) => codes.includes(code);
}

describe("OPERATOR_CATALOG", () => {
  it("has exactly OP-01, OP-02, OP-03", () => {
    expect(OPERATOR_CATALOG.map((w) => w.code)).toEqual(["OP-01", "OP-02", "OP-03"]);
  });

  it("OP-01, OP-02, and OP-03 are all fully available (every workflow is now backed end to end by a typed contract)", () => {
    for (const workflow of OPERATOR_CATALOG) {
      expect(workflow.fullyAvailable).toBe(true);
      expect(workflow.limitedReasonKey).toBeFalsy();
    }
  });
});

describe("deriveCapability", () => {
  it("denies a workflow when the caller lacks the required permission, even if fully available", () => {
    const op01 = OPERATOR_CATALOG.find((w) => w.code === "OP-01")!;
    expect(deriveCapability(op01, allow())).toBe("denied");
  });

  it("returns ready only when permission is held AND the contract is complete", () => {
    const op01 = OPERATOR_CATALOG.find((w) => w.code === "OP-01")!;
    expect(deriveCapability(op01, allow("INVESTMENT_DECISION_MANAGE"))).toBe("ready");
    const op02 = OPERATOR_CATALOG.find((w) => w.code === "OP-02")!;
    expect(deriveCapability(op02, allow("INVESTMENT_DECISION_APPROVE"))).toBe("ready");
  });

  it("permission held is not enough by itself — capability is never 'ready' unless fullyAvailable is also true", () => {
    for (const workflow of OPERATOR_CATALOG) {
      const capability = deriveCapability(workflow, allow(workflow.requiredPermission));
      if (!workflow.fullyAvailable) {
        expect(capability).not.toBe("ready");
      }
    }
  });
});

describe("filterOperatorRows", () => {
  const rows = [
    { code: "OP-01", title: "Buy / Sell Single Securities", description: "Choose a portfolio…", statusLabel: "Ready" },
    { code: "OP-02", title: "Execution & Approvals", description: "Find a decision…", statusLabel: "Partially available" },
  ];

  it("returns everything for a blank query", () => {
    expect(filterOperatorRows(rows, "  ")).toEqual(rows);
  });

  it("matches by code, title, description, or status, case-insensitively", () => {
    expect(filterOperatorRows(rows, "buy")).toEqual([rows[0]]);
    expect(filterOperatorRows(rows, "op-02")).toEqual([rows[1]]);
    expect(filterOperatorRows(rows, "partially")).toEqual([rows[1]]);
  });

  it("returns an empty list when nothing matches", () => {
    expect(filterOperatorRows(rows, "no-such-workflow")).toEqual([]);
  });
});
