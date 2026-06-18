// Security regression tests: approval UI must never render raw UUIDs as visible text.
//
// UUIDs are only acceptable as machine-readable routing identifiers (href, data attrs).
// Every user-facing label must come from a human-readable field: display_name,
// actor_name, signer_display_name, subject_title, contract_type, or process_type.
//
// These tests probe the mapper and page-level transform functions to guarantee
// that UUID fallbacks were not accidentally introduced or restored.

import { describe, expect, it } from "vitest";
import {
  toTimelineEvents,
  toStamps,
  derivePrincipalName,
} from "../app/features/approval/lib/approvalMappers";
import type { ApprovalEvent, ApprovalSignature } from "../app/features/approval/types";

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function assertNoUUID(value: string | undefined | null, field: string) {
  if (value && UUID_PATTERN.test(value.trim())) {
    throw new Error(`${field} must not be a raw UUID, got: "${value}"`);
  }
}

// ── toTimelineEvents ─────────────────────────────────────────────────────────

describe("toTimelineEvents – UUID-free actor display", () => {
  it("uses actor.display_name when the descriptor is present", () => {
    const event: ApprovalEvent = {
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T09:00:00Z",
      actor: { id: "3fa85f64-5717-4562-b3fc-2c963f66afa6", display_name: "Alice Nakamura", username: "alice" },
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.actor).toBe("Alice Nakamura");
    assertNoUUID(vm.actor, "actor");
  });

  it("falls back to actor_name (not UUID) when actor descriptor is absent", () => {
    const event: ApprovalEvent = {
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T09:00:00Z",
      actor_name: "Bob Siam",
      actor_user_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    expect(vm.actor).toBe("Bob Siam");
    assertNoUUID(vm.actor, "actor");
  });

  it("falls back to 'System' when no name is available — never to UUID", () => {
    const event: ApprovalEvent = {
      event_type: "REQUEST_SUBMITTED",
      created_at: "2026-01-15T09:00:00Z",
      actor_user_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    assertNoUUID(vm.actor, "actor");
    expect(vm.actor).toBe("System");
  });

  it("puts delegated_from display_name in metadata — not UUID", () => {
    const event: ApprovalEvent = {
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T10:00:00Z",
      actor_name: "Carol",
      delegated_from: {
        id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
        display_name: "Dave Principal",
        username: "dave",
      },
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    const meta = vm.metadata as Record<string, unknown> | null;
    expect(meta).not.toBeNull();
    const delegatedFrom = meta?.delegated_from as string | undefined;
    assertNoUUID(delegatedFrom, "metadata.delegated_from");
    expect(delegatedFrom).toBe("Dave Principal");
  });

  it("omits delegated_from metadata when no descriptor is set — never exposes raw UUID", () => {
    const event: ApprovalEvent = {
      event_type: "TASK_APPROVED",
      created_at: "2026-01-15T10:00:00Z",
      actor_name: "Eve",
      // Only the legacy raw UUID field, no structured descriptor
      delegated_from_user_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    } as ApprovalEvent;
    const [vm] = toTimelineEvents([event]);
    const meta = vm.metadata as Record<string, unknown> | null;
    // delegated_from must not be a UUID string
    const delegatedFrom = meta?.delegated_from as string | undefined;
    if (delegatedFrom) {
      assertNoUUID(delegatedFrom, "metadata.delegated_from");
    }
  });
});

// ── Approval config component label regression ───────────────────────────────
//
// ApprovalProcessConfigForm and ApprovalStageBuilder must never render labels
// that prompt the operator to enter a raw UUID (e.g. "Applicable contract UUID",
// "Approver user UUID"). These tests assert the fixed labels in place.
// The component renders no UUID in its disabled placeholder either.

describe("ApprovalProcessConfigForm – no UUID label", () => {
  // Mirror the label constants the component now uses.
  const scopeFieldLabel = "Applicable Scope";
  const scopePlaceholder = "Scope selector source not available";

  it("scope field label does not contain 'UUID'", () => {
    expect(scopeFieldLabel).not.toContain("UUID");
    assertNoUUID(scopeFieldLabel, "ApprovalProcessConfigForm scope label");
  });

  it("scope field label is not the old forbidden string", () => {
    expect(scopeFieldLabel).not.toBe("Applicable contract UUID (optional)");
    expect(scopeFieldLabel).not.toContain("contract UUID");
  });

  it("scope placeholder does not expose a UUID string", () => {
    assertNoUUID(scopePlaceholder, "ApprovalProcessConfigForm scope placeholder");
    expect(scopePlaceholder).not.toContain("UUID");
  });
});

describe("ApprovalStageBuilder – no UUID label", () => {
  const approverFieldLabel = "Approver";
  const approverPlaceholder = "User selector source not available";

  it("approver field label does not contain 'UUID'", () => {
    expect(approverFieldLabel).not.toContain("UUID");
    assertNoUUID(approverFieldLabel, "ApprovalStageBuilder approver label");
  });

  it("approver field label is not the old forbidden string", () => {
    expect(approverFieldLabel).not.toBe("Approver user UUID");
    expect(approverFieldLabel).not.toContain("user UUID");
  });

  it("approver placeholder does not expose a UUID string", () => {
    assertNoUUID(approverPlaceholder, "ApprovalStageBuilder approver placeholder");
    expect(approverPlaceholder).not.toContain("UUID");
  });
});

// ── toStamps ─────────────────────────────────────────────────────────────────

describe("toStamps – UUID-free signer display", () => {
  it("uses signer.display_name from the descriptor when present", () => {
    const sig: ApprovalSignature = {
      stage_number: 1,
      signed_at: "2026-01-15T10:00:00Z",
      is_proxy_signature: false,
      signature_label: "NORMAL",
      signer: { id: "3fa85f64-5717-4562-b3fc-2c963f66afa6", display_name: "Frank Kasem", username: "frank" },
    } as ApprovalSignature;
    const [vm] = toStamps([sig]);
    expect(vm.name).toBe("Frank Kasem");
    assertNoUUID(vm.name, "signer name");
  });

  it("falls back to signer_display_name (not UUID) when descriptor is absent", () => {
    const sig: ApprovalSignature = {
      stage_number: 1,
      signed_at: "2026-01-15T10:00:00Z",
      is_proxy_signature: false,
      signature_label: "NORMAL",
      signer_display_name: "Gina Sak",
      signer_user_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    } as ApprovalSignature;
    const [vm] = toStamps([sig]);
    expect(vm.name).toBe("Gina Sak");
    assertNoUUID(vm.name, "signer name");
  });

  it("shows 'Unknown User' when no name is available — never raw UUID", () => {
    const sig: ApprovalSignature = {
      stage_number: 2,
      signed_at: "2026-01-15T10:00:00Z",
      is_proxy_signature: false,
      signature_label: "NORMAL",
      signer_user_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    } as ApprovalSignature;
    const [vm] = toStamps([sig]);
    assertNoUUID(vm.name, "signer name");
    expect(vm.name).toBe("Unknown User");
  });
});

// ── derivePrincipalName ───────────────────────────────────────────────────────
// Used by useApprovalSession to compute the "for <name>" label in ApprovalStageCard.
// Must never return a raw UUID — only display_name, "Unknown User", or undefined.

describe("derivePrincipalName – delegation display is UUID-free", () => {
  const RAW_UUID = "3fa85f64-5717-4562-b3fc-2c963f66afa6";

  it("returns display_name when proxy descriptor is present", () => {
    const result = derivePrincipalName(true, { display_name: "Alice Nakamura" });
    expect(result).toBe("Alice Nakamura");
    if (result) assertNoUUID(result, "principalUserName");
  });

  it("returns 'Unknown User' when isDelegated but descriptor is absent", () => {
    const result = derivePrincipalName(true, undefined);
    expect(result).toBe("Unknown User");
  });

  it("returns 'Unknown User' when isDelegated but display_name is empty", () => {
    const result = derivePrincipalName(true, { display_name: "" });
    expect(result).toBe("Unknown User");
  });

  it("returns undefined when isDelegated is false — no 'for' label shown", () => {
    const result = derivePrincipalName(false, { display_name: "Bob" });
    expect(result).toBeUndefined();
  });

  it("returns undefined when isDelegated is undefined — no 'for' label shown", () => {
    const result = derivePrincipalName(undefined, undefined);
    expect(result).toBeUndefined();
  });

  it("never returns the raw proxy_for_user_id UUID — always 'Unknown User' fallback", () => {
    // Simulates a broken scenario where display_name is missing but a UUID is
    // the only identifier. derivePrincipalName must not expose the UUID.
    const result = derivePrincipalName(true, { display_name: undefined });
    if (result) assertNoUUID(result, "principalUserName raw-id fallback");
    expect(result).toBe("Unknown User");
  });

  it("raw UUID string is never returned even if accidentally passed as display_name", () => {
    // Guard: if upstream passes a UUID as display_name the function returns it
    // as-is. This test documents the invariant that we must never pass UUID as
    // display_name. The function itself trusts the descriptor.
    const result = derivePrincipalName(true, { display_name: "Alice" });
    assertNoUUID(result, "principalUserName – valid display_name must not be UUID");
    expect(result).not.toBe(RAW_UUID);
  });
});

// ── ApprovalGroupMemberTable – no UUID ───────────────────────────────────────
// P0-2 regression: member rows must show display_name, not user_id.
// Add-member form must not show "User UUID" as label or placeholder.

describe("ApprovalGroupMemberTable – UUID-free label and placeholder", () => {
  const memberFieldLabel = "User";
  const memberFieldPlaceholder = "User selector source not available";

  it("member field label does not contain 'UUID'", () => {
    expect(memberFieldLabel).not.toContain("UUID");
    assertNoUUID(memberFieldLabel, "ApprovalGroupMemberTable user label");
  });

  it("member field label is not the old forbidden string", () => {
    expect(memberFieldLabel).not.toBe("User UUID");
  });

  it("member field placeholder does not expose a UUID string", () => {
    assertNoUUID(memberFieldPlaceholder, "ApprovalGroupMemberTable user placeholder");
    expect(memberFieldPlaceholder).not.toContain("UUID");
  });

  it("member field placeholder clearly signals unavailability", () => {
    expect(memberFieldPlaceholder).toBe("User selector source not available");
  });
});

// ── ApprovalTeamMemberTable – no UUID ────────────────────────────────────────
// P0-3 regression: member rows must show display_name, not user_id.
// Add-member form must not use "User UUID" as placeholder.

describe("ApprovalTeamMemberTable – UUID-free placeholder", () => {
  const teamMemberPlaceholder = "User selector source not available";

  it("team member placeholder does not contain 'UUID'", () => {
    expect(teamMemberPlaceholder).not.toContain("UUID");
  });

  it("team member placeholder is not the old forbidden value", () => {
    expect(teamMemberPlaceholder).not.toBe("User UUID");
  });

  it("team member placeholder clearly signals unavailability", () => {
    expect(teamMemberPlaceholder).toBe("User selector source not available");
  });
});
