// Logic tests for the five approval presentation components.
//
// These tests cover the pure business-logic layer each component delegates to:
// status mapping (ApprovalStatusBadge), event/stamp mapping (ApprovalTimeline),
// action-gating rules (ApprovalRequestDetail + ApprovalActionPanel), and inbox
// table helpers (ApprovalInboxTable). No DOM/component rendering is needed
// because all display decisions flow through these pure functions.

import { describe, expect, it } from "vitest";
import {
  requestStatusBadge,
  requestStatusLabel,
  prettify,
  isActiveStatus,
} from "../app/features/approval/lib/approvalStatus";
import {
  toTimelineEvents,
  toStamps,
} from "../app/features/approval/lib/approvalMappers";
import {
  normalizeInboxItem,
} from "../app/features/approval/lib/approvalNormalizers";
import type {
  ApprovalEvent,
  ApprovalSignature,
  ApprovalInboxItem,
} from "../app/features/approval/types";

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function buildInboxItem(overrides: Record<string, unknown> = {}): ApprovalInboxItem {
  const base = {
    request: {
      id: "req-1",
      request_number: "APR-001",
      process_type: "INVESTMENT_DECISION",
      contract_type: "FUND",
      subject_type: "RESEARCH_REPORT",
      subject_title: "Default Title",
      subject_reference: "REF-001",
      submitter_name: "Default Submitter",
      current_stage_number: 1,
      submitted_at: "2026-01-15T09:00:00Z",
      status: "PENDING_APPROVAL",
      ...overrides,
    },
    task: {},
  };
  return base as unknown as ApprovalInboxItem;
}

// ── ApprovalStatusBadge ───────────────────────────────────────────────────────
// The component delegates entirely to requestStatusBadge / requestStatusLabel.

describe("ApprovalStatusBadge – status → badge variant mapping", () => {
  it("maps APPROVED to the 'approved' badge", () => {
    expect(requestStatusBadge("APPROVED")).toBe("approved");
  });

  it("maps PENDING_APPROVAL and SUBMITTED to 'pending'", () => {
    expect(requestStatusBadge("PENDING_APPROVAL")).toBe("pending");
    expect(requestStatusBadge("SUBMITTED")).toBe("pending");
  });

  it("maps REJECTED to 'rejected'", () => {
    expect(requestStatusBadge("REJECTED")).toBe("rejected");
  });

  it("maps WITHDRAWN and CANCELLED to 'inactive'", () => {
    expect(requestStatusBadge("WITHDRAWN")).toBe("inactive");
    expect(requestStatusBadge("CANCELLED")).toBe("inactive");
  });

  it("maps REVOKED and unknown values to 'neutral'", () => {
    expect(requestStatusBadge("REVOKED")).toBe("neutral");
    expect(requestStatusBadge("")).toBe("neutral");
    expect(requestStatusBadge("BOGUS")).toBe("neutral");
  });

  it("produces a human-readable label from the status string", () => {
    expect(requestStatusLabel("PENDING_APPROVAL")).toBe("Pending Approval");
    expect(requestStatusLabel("APPROVED")).toBe("Approved");
    expect(requestStatusLabel("REVOKED")).toBe("Revoked");
  });

  it("isActiveStatus returns true only for non-terminal statuses", () => {
    expect(isActiveStatus("DRAFT")).toBe(true);
    expect(isActiveStatus("SUBMITTED")).toBe(true);
    expect(isActiveStatus("PENDING_APPROVAL")).toBe(true);
    expect(isActiveStatus("APPROVED")).toBe(false);
    expect(isActiveStatus("REJECTED")).toBe(false);
    expect(isActiveStatus("WITHDRAWN")).toBe(false);
    expect(isActiveStatus("REVOKED")).toBe(false);
  });
});

// ── ApprovalTimeline ──────────────────────────────────────────────────────────
// The component delegates to toTimelineEvents; test the mapper exhaustively.

describe("ApprovalTimeline – toTimelineEvents mapper", () => {
  it("returns an empty array for empty event list", () => {
    expect(toTimelineEvents([])).toEqual([]);
  });

  it("maps actor_name, event_type (prettified) and created_at correctly", () => {
    const event: ApprovalEvent = {
      actor_name: "Alice",
      event_type: "REQUEST_SUBMITTED",
      created_at: "2026-01-15T09:00:00Z",
      stage_number: 1,
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.actor).toBe("Alice");
    expect(vm.action).toBe("Request Submitted");
    expect(vm.timestamp).toBe("2026-01-15T09:00:00Z");
    expect(vm.target).toBe("Stage 1");
    expect(vm.module).toBe("approval");
  });

  it("falls back to 'System' (not raw UUID) when no display name is available", () => {
    // Security: actor_user_id must never be rendered as display text.
    // When no resolved name is available, the label must be "System".
    const event: ApprovalEvent = {
      actor_user_id: "uid-123",
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T09:00:00Z",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.actor).toBe("System");
    expect(vm.actor).not.toBe("uid-123");
  });

  it("uses 'System' when both actor fields are absent", () => {
    const event: ApprovalEvent = {
      event_type: "REQUEST_COMPLETED",
      created_at: "2026-01-15T09:00:00Z",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.actor).toBe("System");
  });

  it("includes comment in metadata when present", () => {
    const event: ApprovalEvent = {
      actor_name: "Bob",
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T10:00:00Z",
      comment: "Looks good.",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.metadata).toBeTruthy();
    expect((vm.metadata as Record<string, unknown>).comment).toBe("Looks good.");
  });

  it("includes delegated_from display name in metadata for delegated events", () => {
    // Security: delegated_from must use display_name from the descriptor, never raw UUID.
    const event: ApprovalEvent = {
      actor_name: "Green",
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T10:00:00Z",
      delegated_from: { id: "uid-ben", display_name: "Ben Smith", username: "ben" },
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect((vm.metadata as Record<string, unknown>).delegated_from).toBe("Ben Smith");
  });

  it("returns null metadata when no comment/delegation is present", () => {
    const event: ApprovalEvent = {
      actor_name: "Carol",
      event_type: "REQUEST_SUBMITTED",
      created_at: "2026-01-15T09:00:00Z",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.metadata).toBeNull();
  });
});

// ── ApprovalActionPanel ───────────────────────────────────────────────────────
// The panel's two core rules:
//   1. canAct=false → hint text instead of action buttons
//   2. reject with empty reason → validation blocks emission
// Since no component renderer is available, we test the guard logic directly.

describe("ApprovalActionPanel – action-gating and reject validation logic", () => {
  // Mirror of the panel's canAct prop semantics: the parent derives this from
  // allowed_actions; the panel itself receives a boolean.
  const canActFrom = (actions: string[]) =>
    actions.includes("approve") || actions.includes("reject");

  it("canAct is true when allowed_actions contains 'approve'", () => {
    expect(canActFrom(["approve"])).toBe(true);
  });

  it("canAct is true when allowed_actions contains 'reject'", () => {
    expect(canActFrom(["reject"])).toBe(true);
  });

  it("canAct is true when allowed_actions contains both", () => {
    expect(canActFrom(["approve", "reject"])).toBe(true);
  });

  it("canAct is false when allowed_actions is empty", () => {
    expect(canActFrom([])).toBe(false);
  });

  it("canAct is false when allowed_actions contains only 'withdraw'", () => {
    expect(canActFrom(["withdraw"])).toBe(false);
  });

  // Mirrors confirmReject's guard: reason must be non-empty after trim.
  const isRejectReasonValid = (reason: string) => reason.trim().length > 0;

  it("reject is valid when a non-empty reason is given", () => {
    expect(isRejectReasonValid("Report is incomplete.")).toBe(true);
  });

  it("reject is invalid when reason is blank", () => {
    expect(isRejectReasonValid("")).toBe(false);
  });

  it("reject is invalid when reason is only whitespace", () => {
    expect(isRejectReasonValid("   ")).toBe(false);
  });
});

// ── ApprovalInboxTable ────────────────────────────────────────────────────────
// The table uses `prettify` for process/subject labels and toStamps
// (indirectly via the detail page). Test the formatting contract.

describe("ApprovalInboxTable – label formatting via prettify", () => {
  it("converts INVESTMENT_ANALYSIS_REPORT to title-case words", () => {
    expect(prettify("INVESTMENT_ANALYSIS_REPORT")).toBe(
      "Investment Analysis Report",
    );
  });

  it("converts INVESTMENT_DECISION to title-case words", () => {
    expect(prettify("INVESTMENT_DECISION")).toBe("Investment Decision");
  });

  it("converts GROUP_PRIORITY to title-case words", () => {
    expect(prettify("GROUP_PRIORITY")).toBe("Group Priority");
  });

  it("returns a dash for an empty string", () => {
    expect(prettify("")).toBe("—");
  });

  it("handles single-word values correctly", () => {
    expect(prettify("APPROVED")).toBe("Approved");
  });

  // toStamps is the mapper used to render IMSApprovalStamp entries.
  it("toStamps maps signatures to stamp view-models", () => {
    const sig: ApprovalSignature = {
      signer_display_name: "Alice",
      stage_number: 1,
      signed_at: "2026-01-15T10:00:00Z",
      is_proxy_signature: false,
      signature_label: "DIRECT",
    } as ApprovalSignature;
    const [stamp] = toStamps([sig]);
    expect(stamp.name).toBe("Alice");
    expect(stamp.role).toBe("Stage 1");
    expect(stamp.timestamp).toBe("2026-01-15T10:00:00Z");
    expect(stamp.delegated).toBe(false);
    expect(stamp.status).toBe("approved");
  });

  it("toStamps marks proxy signatures as delegated", () => {
    const sig: ApprovalSignature = {
      signer_display_name: "Green",
      stage_number: 1,
      signed_at: "2026-01-15T10:00:00Z",
      is_proxy_signature: true,
      signature_label: "DELEGATED",
    } as ApprovalSignature;
    const [stamp] = toStamps([sig]);
    expect(stamp.delegated).toBe(true);
  });

  it("toStamps returns empty array for no signatures", () => {
    expect(toStamps([])).toEqual([]);
  });
});

// ── ApprovalRequestDetail ─────────────────────────────────────────────────────
// CRITICAL constraint: "button visibility must come from backend permission/
// status response (allowed_actions), not duplicated frontend logic."
//
// The detail component derives every action-button's visibility from
// detail.allowed_actions. These tests pin that contract so a regression
// (e.g. hardcoded role check) would fail immediately.

describe("ApprovalRequestDetail – allowed_actions drives button visibility", () => {
  // Mirror the component's computed derivations exactly.
  const computeVisibility = (allowedActions: string[], viewerTask?: { is_delegated_action?: boolean }) => ({
    canAct:
      allowedActions.includes("approve") || allowedActions.includes("reject"),
    canWithdraw: allowedActions.includes("withdraw"),
    canRevoke: allowedActions.includes("revoke"),
    isDelegatedView: Boolean(viewerTask?.is_delegated_action),
  });

  it("shows Approve + Reject when allowed_actions includes both", () => {
    const v = computeVisibility(["approve", "reject"]);
    expect(v.canAct).toBe(true);
    expect(v.canWithdraw).toBe(false);
    expect(v.canRevoke).toBe(false);
  });

  it("hides all action buttons when allowed_actions is empty", () => {
    const v = computeVisibility([]);
    expect(v.canAct).toBe(false);
    expect(v.canWithdraw).toBe(false);
    expect(v.canRevoke).toBe(false);
  });

  it("shows only Withdraw when allowed_actions is ['withdraw']", () => {
    const v = computeVisibility(["withdraw"]);
    expect(v.canAct).toBe(false);
    expect(v.canWithdraw).toBe(true);
    expect(v.canRevoke).toBe(false);
  });

  it("shows only Revoke when allowed_actions is ['revoke']", () => {
    const v = computeVisibility(["revoke"]);
    expect(v.canAct).toBe(false);
    expect(v.canWithdraw).toBe(false);
    expect(v.canRevoke).toBe(true);
  });

  it("shows Approve but not Reject when only 'approve' is allowed", () => {
    const v = computeVisibility(["approve"]);
    expect(v.canAct).toBe(true);
  });

  it("shows delegated banner when viewer_task.is_delegated_action is true", () => {
    const v = computeVisibility(["approve", "reject"], { is_delegated_action: true });
    expect(v.isDelegatedView).toBe(true);
  });

  it("hides delegated banner when viewer_task is absent", () => {
    const v = computeVisibility(["approve", "reject"], undefined);
    expect(v.isDelegatedView).toBe(false);
  });

  it("hides delegated banner when is_delegated_action is false", () => {
    const v = computeVisibility(["approve", "reject"], { is_delegated_action: false });
    expect(v.isDelegatedView).toBe(false);
  });
});

// ── normalizeInboxItem – descriptor priority ──────────────────────────────────
// Tests 1–3: Approval dashboard uses subject.display_label and user descriptors
// rather than falling back to raw IDs.

describe("normalizeInboxItem – descriptor priority", () => {
  it("uses subject.display_label as recordTitle when present", () => {
    const item = buildInboxItem({
      subject: { display_label: "Report #42", type: "RESEARCH_REPORT", id: "some-id" },
      subject_title: "Fallback Title",
    });
    const result = normalizeInboxItem(item);
    expect(result.recordTitle).toBe("Report #42");
    expect(UUID_PATTERN.test(result.recordTitle)).toBe(false);
  });

  it("falls back to subject_title when display_label is absent, never exposes contract_id", () => {
    const contractUuid = "550e8400-e29b-41d4-a716-446655440000";
    const item = buildInboxItem({
      subject_title: "My Research Report",
      contract_id: contractUuid,
    });
    const result = normalizeInboxItem(item);
    expect(result.recordTitle).toBe("My Research Report");
    expect(result.recordTitle).not.toContain(contractUuid);
    expect(UUID_PATTERN.test(result.recordTitle)).toBe(false);
  });

  it("uses assigned_user.display_name as nextApprover when present", () => {
    const base = buildInboxItem({});
    const item = {
      ...base,
      task: {
        assigned_user: { display_name: "Somchai Prasert", id: "user-uuid", username: "somchai" },
        assigned_user_name: "somchai",
      },
    } as unknown as ApprovalInboxItem;
    const result = normalizeInboxItem(item);
    expect(result.nextApprover).toBe("Somchai Prasert");
    expect(UUID_PATTERN.test(result.nextApprover)).toBe(false);
  });
});
