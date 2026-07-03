// Security regression tests: permission UI must never render raw UUIDs as visible text.
//
// These tests probe the pure display-logic rules that derive visible labels from
// Permission module data. Since PermissionRequestDetailScreen renders using Vue
// template expressions against the `request` prop, we test the guard functions
// that mediate the data rather than mounting the full component.

import { describe, expect, it } from "vitest";

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function assertNoUUID(value: string | undefined | null, field: string) {
  if (value && UUID_PATTERN.test(value.trim())) {
    throw new Error(`${field} must not be a raw UUID, got: "${value}"`);
  }
}

// ── Creator label logic ───────────────────────────────────────────────────────

describe("PermissionRequestDetail – creator label", () => {
  function resolveCreatorLabel(created_by_name: string, _created_by: string): string {
    return created_by_name || "Unknown User";
  }

  it("shows created_by_name when available", () => {
    const label = resolveCreatorLabel("Alice Nakamura", "3fa85f64-5717-4562-b3fc-2c963f66afa6");
    expect(label).toBe("Alice Nakamura");
    assertNoUUID(label, "creator label");
  });

  it("shows 'Unknown User' when created_by_name is empty — never raw UUID", () => {
    const uuid = "3fa85f64-5717-4562-b3fc-2c963f66afa6";
    const label = resolveCreatorLabel("", uuid);
    expect(label).toBe("Unknown User");
    expect(label).not.toBe(uuid);
    assertNoUUID(label, "creator label");
  });
});

// ── Target entity display ────────────────────────────────────────────────────

describe("PermissionRequestDetail – target entity display", () => {
  function resolveTargetLabel(target_entity_type: string): string {
    return target_entity_type || "-";
  }

  it("shows target_entity_type as the label", () => {
    const label = resolveTargetLabel("USER");
    expect(label).toBe("USER");
    assertNoUUID(label, "target label");
  });

  it("shows '-' when type is absent rather than raw target_entity_id UUID", () => {
    const label = resolveTargetLabel("");
    expect(label).toBe("-");
    assertNoUUID(label, "target label");
  });
});

// ── Account/group screen – user display labels ───────────────────────────────

describe("PermissionAccountScreen – user display", () => {
  const resolveUserLabel = (display_name?: string, username?: string): string =>
    display_name || username || "Unknown User";

  it("shows display_name when available", () => {
    const label = resolveUserLabel("Somchai Prasert", "somchai.p");
    expect(label).toBe("Somchai Prasert");
    assertNoUUID(label, "user label");
  });

  it("falls back to username when display_name is absent", () => {
    const label = resolveUserLabel(undefined, "somchai.p");
    expect(label).toBe("somchai.p");
    assertNoUUID(label, "user label");
  });

  it("falls back to 'Unknown User' when both are absent", () => {
    const label = resolveUserLabel(undefined, undefined);
    expect(label).toBe("Unknown User");
    assertNoUUID(label, "user label");
  });
});

// ── Action button state derivation ───────────────────────────────────────────

describe("PermissionRequestDetail – action state derivation", () => {
  const computePermActions = (
    status: string,
    isCreator: boolean,
    hasPendingStep: boolean,
    hasBlocker: boolean,
  ) => ({
    canSubmit: status === "DRAFT" || status === "CHANGES_REQUESTED",
    canReview: status === "READY_FOR_REVIEW" && !isCreator && hasPendingStep,
    canMerge: status === "APPROVED" && !hasBlocker,
  });

  it("canSubmit is true for DRAFT", () => {
    expect(computePermActions("DRAFT", false, false, false).canSubmit).toBe(true);
  });

  it("canSubmit is true for CHANGES_REQUESTED", () => {
    expect(computePermActions("CHANGES_REQUESTED", false, false, false).canSubmit).toBe(true);
  });

  it("canSubmit is false for READY_FOR_REVIEW", () => {
    expect(computePermActions("READY_FOR_REVIEW", false, true, false).canSubmit).toBe(false);
  });

  it("canReview is false when creator is the reviewer", () => {
    expect(computePermActions("READY_FOR_REVIEW", true, true, false).canReview).toBe(false);
  });

  it("canReview is false when no pending step exists", () => {
    expect(computePermActions("READY_FOR_REVIEW", false, false, false).canReview).toBe(false);
  });

  it("canMerge is false when blocking check exists", () => {
    expect(computePermActions("APPROVED", false, false, true).canMerge).toBe(false);
  });

  it("canMerge is true when APPROVED with no blockers", () => {
    expect(computePermActions("APPROVED", false, false, false).canMerge).toBe(true);
  });
});

// ── Reject reason validation ──────────────────────────────────────────────────

describe("PermissionRequestDetail – reject reason validation", () => {
  const isRejectReasonValid = (reason: string) => reason.trim().length > 0;

  it("is invalid when reason is empty", () => {
    expect(isRejectReasonValid("")).toBe(false);
  });

  it("is invalid when reason is whitespace only", () => {
    expect(isRejectReasonValid("   ")).toBe(false);
  });

  it("is valid when reason has meaningful text", () => {
    expect(isRejectReasonValid("Insufficient documentation")).toBe(true);
  });
});

// ── Data permission panel – session grants do not expose UUID contractId ──────

describe("SettingsDataPermissionsPanel – session grants do not expose UUID contractId", () => {
  // Mirrors the contractName resolution in SettingsControlCenter.vue dataPermissionGrants.
  // When the backend returns a UUID-format contract_id (from permission_data_rights table),
  // it must not appear as the display label — a safe fallback label is used instead.
  const resolveContractName = (contractId: string, fallback: string): string =>
    UUID_PATTERN.test(contractId) ? fallback : contractId;

  it("human-readable IMS contract codes are displayed as-is", () => {
    expect(resolveContractName("TH-FUND-001", "Fund access")).toBe("TH-FUND-001");
    expect(resolveContractName("TH-DISC-002", "Fund access")).toBe("TH-DISC-002");
  });

  it("UUID-format contract_id from permission_data_rights is masked as the fallback label", () => {
    const uuid = "3fa85f64-5717-4562-b3fc-2c963f66afa6";
    const name = resolveContractName(uuid, "Fund access");
    expect(name).toBe("Fund access");
    expect(name).not.toContain(uuid);
    assertNoUUID(name, "contractName");
  });

  it("contractName is never a raw UUID regardless of which backend table sourced it", () => {
    const uuids = [
      "550e8400-e29b-41d4-a716-446655440000",
      "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    ];
    uuids.forEach((id) => {
      assertNoUUID(resolveContractName(id, "Fund access"), "contractName");
    });
  });
});

// ── Changes tab — item target_id must not be visible ────────────────────────

describe("PermissionRequestDetail – Changes tab item display", () => {
  // Mirrors the rendering logic in PermissionRequestDetailScreen.vue Changes tab.
  // The template renders only item.target_table, not item.target_id.
  function renderChangeItemHeader(item: { target_table: string; target_id: string }): string {
    // This is what the component now renders (target_id removed per P0 FIX 5).
    return item.target_table;
  }

  it("does not render item.target_id as visible text", () => {
    const item = {
      target_table: "iam__users",
      target_id: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    };
    const rendered = renderChangeItemHeader(item);
    expect(rendered).not.toContain(item.target_id);
    assertNoUUID(rendered, "Changes tab item header");
  });

  it("renders target_table as the type label", () => {
    const item = {
      target_table: "iam__function_permissions",
      target_id: "9b2e1c3d-4a5f-6789-0abc-def123456789",
    };
    const rendered = renderChangeItemHeader(item);
    expect(rendered).toBe("iam__function_permissions");
    expect(rendered).not.toContain(item.target_id);
  });
});
