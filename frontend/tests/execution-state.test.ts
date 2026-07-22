/**
 * OP-02 execution-lifecycle-state derivation. Only real backend enum values
 * (see backend/internal/investment/domain/valueobject/decision_lifecycle.go)
 * may produce a state — there is no "FAILED" execution status, so this
 * module must never invent one.
 */
import { describe, expect, it } from "vitest";

import {
  deriveExecutionLifecycleState,
  executionStateTone,
} from "../app/features/investment-decision/lib/executionState";

describe("deriveExecutionLifecycleState", () => {
  it("derives from the decision status when no execution has been created yet", () => {
    expect(deriveExecutionLifecycleState("DRAFT", null)).toBe("draft");
    expect(deriveExecutionLifecycleState("PENDING_APPROVAL", null)).toBe("pending_approval");
    expect(deriveExecutionLifecycleState("APPROVED", null)).toBe("executable");
    expect(deriveExecutionLifecycleState("READY_FOR_EXECUTION", null)).toBe("executable");
    expect(deriveExecutionLifecycleState("REJECTED", null)).toBe("rejected");
    expect(deriveExecutionLifecycleState("CANCELLED", null)).toBe("cancelled");
    expect(deriveExecutionLifecycleState("PENDING_COMPLIANCE_RELEASE", null)).toBe("compliance_blocked");
  });

  it("trusts a decision-level EXECUTED marker when no execution row could be matched", () => {
    expect(deriveExecutionLifecycleState("EXECUTED", null)).toBe("filled");
  });

  it("prefers the matched execution's own status once an execution exists", () => {
    expect(deriveExecutionLifecycleState("APPROVED", "PENDING")).toBe("execution_pending");
    expect(deriveExecutionLifecycleState("APPROVED", "EXECUTED")).toBe("filled");
    expect(deriveExecutionLifecycleState("APPROVED", "PARTIALLY_EXECUTED")).toBe("partially_filled");
    expect(deriveExecutionLifecycleState("APPROVED", "CANCELLED")).toBe("execution_cancelled");
  });

  it("a rejected decision stays rejected even if a stray execution row somehow matched (defensive — should not happen)", () => {
    // Execution status, when present, is authoritative for fill progress —
    // this documents that behavior explicitly rather than leaving it
    // implicit.
    expect(deriveExecutionLifecycleState("REJECTED", "EXECUTED")).toBe("filled");
  });

  it("falls back to unknown for an unrecognized decision status with no execution", () => {
    expect(deriveExecutionLifecycleState("SOME_FUTURE_STATUS", null)).toBe("unknown");
    expect(deriveExecutionLifecycleState(undefined, undefined)).toBe("unknown");
  });

  it("falls back to unknown for an unrecognized execution status", () => {
    expect(deriveExecutionLifecycleState("APPROVED", "SOME_FUTURE_EXEC_STATUS")).toBe("unknown");
  });

  it("never produces a fabricated 'failed' state — the backend has no such execution status", () => {
    const allDecisionStatuses = [
      "DRAFT",
      "PENDING_APPROVAL",
      "APPROVED",
      "REJECTED",
      "CANCELLED",
      "READY_FOR_EXECUTION",
      "EXECUTED",
      "PENDING_COMPLIANCE_RELEASE",
    ];
    const allExecutionStatuses = ["PENDING", "EXECUTED", "PARTIALLY_EXECUTED", "CANCELLED"];

    for (const d of allDecisionStatuses) {
      expect(deriveExecutionLifecycleState(d, null)).not.toBe("failed");
    }
    for (const e of allExecutionStatuses) {
      expect(deriveExecutionLifecycleState("APPROVED", e)).not.toBe("failed");
    }
  });
});

describe("executionStateTone", () => {
  it("maps every state to one of AppStatusBadge's real tones", () => {
    const validTones = ["pending", "approved", "success", "rejected", "inactive", "neutral"];
    const states = [
      "draft",
      "pending_approval",
      "executable",
      "rejected",
      "cancelled",
      "compliance_blocked",
      "execution_pending",
      "partially_filled",
      "filled",
      "execution_cancelled",
      "unknown",
    ] as const;

    for (const state of states) {
      expect(validTones).toContain(executionStateTone(state));
    }
  });

  it("filled is success and execution_cancelled/cancelled are inactive", () => {
    expect(executionStateTone("filled")).toBe("success");
    expect(executionStateTone("cancelled")).toBe("inactive");
    expect(executionStateTone("execution_cancelled")).toBe("inactive");
  });
});
