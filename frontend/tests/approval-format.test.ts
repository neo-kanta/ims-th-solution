import { describe, expect, it } from "vitest";

import {
  isActiveStatus,
  prettify,
  requestStatusBadge,
  taskStatusBadge,
} from "../app/features/approval/lib/approvalStatus";
import {
  toStamps,
  toTimelineEvents,
} from "../app/features/approval/lib/approvalMappers";
import type {
  ApprovalEvent,
  ApprovalSignature,
} from "../app/features/approval/types";

describe("approval status helpers", () => {
  it("maps request statuses to badge keywords", () => {
    expect(requestStatusBadge("APPROVED")).toBe("approved");
    expect(requestStatusBadge("REJECTED")).toBe("rejected");
    expect(requestStatusBadge("PENDING_APPROVAL")).toBe("pending");
    expect(requestStatusBadge("WITHDRAWN")).toBe("inactive");
    expect(requestStatusBadge("DRAFT")).toBe("neutral");
  });

  it("maps task statuses to badge keywords", () => {
    expect(taskStatusBadge("PENDING")).toBe("pending");
    expect(taskStatusBadge("APPROVED")).toBe("approved");
    expect(taskStatusBadge("SKIPPED")).toBe("inactive");
  });

  it("prettifies enum tokens", () => {
    expect(prettify("INVESTMENT_ANALYSIS_REPORT")).toBe("Investment Analysis Report");
    expect(prettify("GROUP_ANY")).toBe("Group Any");
    expect(prettify("")).toBe("—");
  });

  it("classifies active (non-terminal) statuses", () => {
    expect(isActiveStatus("PENDING_APPROVAL")).toBe(true);
    expect(isActiveStatus("SUBMITTED")).toBe(true);
    expect(isActiveStatus("APPROVED")).toBe(false);
    expect(isActiveStatus("REJECTED")).toBe(false);
  });
});

describe("approval mappers", () => {
  it("maps approval events to timeline view-models", () => {
    const events: ApprovalEvent[] = [
      {
        id: "e1",
        event_type: "SUBMITTED",
        actor_name: "Ben",
        stage_number: 1,
        created_at: "2026-06-02T03:00:00Z",
        comment: "please review",
      } as ApprovalEvent,
    ];
    const vm = toTimelineEvents(events);
    expect(vm).toHaveLength(1);
    expect(vm[0].actor).toBe("Ben");
    expect(vm[0].action).toBe("Submitted");
    expect(vm[0].target).toBe("Stage 1");
    expect(vm[0].metadata).toMatchObject({ comment: "please review" });
  });

  it("maps signatures to stamp view-models and flags delegated proxies", () => {
    const sigs: ApprovalSignature[] = [
      {
        id: "s1",
        stage_number: 2,
        signer_user_id: "u1",
        signer_display_name: "Green",
        signed_at: "2026-06-02T04:00:00Z",
        is_proxy_signature: true,
        signature_label: "DELEGATED",
      } as ApprovalSignature,
      {
        id: "s2",
        stage_number: 1,
        signer_user_id: "u2",
        signer_display_name: "Admin",
        signed_at: "2026-06-02T03:30:00Z",
        is_proxy_signature: false,
        signature_label: "NORMAL",
      } as ApprovalSignature,
    ];
    const stamps = toStamps(sigs);
    expect(stamps).toHaveLength(2);
    expect(stamps[0].name).toBe("Green");
    expect(stamps[0].delegated).toBe(true);
    expect(stamps[1].delegated).toBe(false);
    expect(stamps[0].status).toBe("approved");
  });
});
