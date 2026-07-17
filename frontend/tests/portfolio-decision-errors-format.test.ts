import { describe, expect, it } from "vitest";

import {
  describeDecisionError,
  extractMessage,
  requestIdOf,
  statusOf,
} from "../app/features/portfolio-decision/lib/decisionErrors";
import {
  decisionStatusLabel,
  decisionWorkflowStages,
  estimatedConsideration,
} from "../app/features/portfolio-decision/lib/decisionFormat";

describe("backend error display", () => {
  it("extracts the backend's `error` field from a legacy {error,code,details} envelope", () => {
    const err = { status: 422, data: { error: "compliance blocked: BLOCK verdict" } };
    expect(extractMessage(err, "fallback")).toBe("compliance blocked: BLOCK verdict");
    expect(statusOf(err)).toBe(422);
  });

  it("falls back to a generic message when the backend gives none", () => {
    expect(extractMessage({}, "Something went wrong.")).toBe("Something went wrong.");
    expect(extractMessage(null, "Something went wrong.")).toBe("Something went wrong.");
  });

  it("surfaces a request id from the response header without hiding it", () => {
    const err = {
      status: 500,
      data: { error: "internal error" },
      response: { headers: { get: (name: string) => (name === "x-request-id" ? "req-123" : null) } },
    };
    expect(requestIdOf(err)).toBe("req-123");
    expect(describeDecisionError(err, "fallback")).toBe("internal error (request req-123)");
  });

  it("omits the request-id suffix when none is available, rather than fabricating one", () => {
    const err = { status: 400, data: { error: "bad request" } };
    expect(requestIdOf(err)).toBeNull();
    expect(describeDecisionError(err, "fallback")).toBe("bad request");
  });
});

describe("decisionStatusLabel", () => {
  it("labels every lifecycle status the backend can return", () => {
    expect(decisionStatusLabel("DRAFT")).toBe("Draft");
    expect(decisionStatusLabel("PENDING_COMPLIANCE_RELEASE")).toBe("Pending Compliance Release");
    expect(decisionStatusLabel("PENDING_APPROVAL")).toBe("Pending Approval");
    expect(decisionStatusLabel("APPROVED")).toBe("Approved");
    expect(decisionStatusLabel("READY_FOR_EXECUTION")).toBe("Ready for Execution");
    expect(decisionStatusLabel("EXECUTED")).toBe("Executed");
    expect(decisionStatusLabel("REJECTED")).toBe("Rejected");
    expect(decisionStatusLabel("CANCELLED")).toBe("Cancelled");
  });
});

describe("decisionWorkflowStages", () => {
  it("marks Draft active and everything else upcoming for a fresh DRAFT decision", () => {
    const stages = decisionWorkflowStages("DRAFT");
    expect(stages.map((s) => s.state)).toEqual(["active", "upcoming", "upcoming", "upcoming", "upcoming"]);
  });

  it("marks Draft complete and Compliance & Approval active while pending approval", () => {
    const stages = decisionWorkflowStages("PENDING_APPROVAL");
    expect(stages[0]?.state).toBe("complete");
    expect(stages[1]?.state).toBe("active");
  });

  it("marks the compliance/approval stage blocked for a rejected decision", () => {
    const stages = decisionWorkflowStages("REJECTED");
    expect(stages[0]?.state).toBe("complete");
    expect(stages[1]?.state).toBe("blocked");
  });

  it("marks the compliance/approval stage blocked for a cancelled decision", () => {
    const stages = decisionWorkflowStages("CANCELLED");
    expect(stages[1]?.state).toBe("blocked");
  });
});

describe("estimatedConsideration", () => {
  it("multiplies quantity by limit price", () => {
    expect(estimatedConsideration("1000", "35.5")).toBe(35500);
  });

  it("uses an explicit amount when limit price is empty", () => {
    expect(estimatedConsideration("1", "", "11")).toBe(11);
  });

  it("prefers the explicit amount over quantity times limit price", () => {
    expect(estimatedConsideration("40", "35", "4000")).toBe(4000);
  });

  it("returns null when either input is missing or non-positive", () => {
    expect(estimatedConsideration("", "35.5")).toBeNull();
    expect(estimatedConsideration("1000", "")).toBeNull();
    expect(estimatedConsideration("0", "35.5")).toBeNull();
    expect(estimatedConsideration("1000", "-1")).toBeNull();
    expect(estimatedConsideration("1", "", "0")).toBeNull();
  });
});
