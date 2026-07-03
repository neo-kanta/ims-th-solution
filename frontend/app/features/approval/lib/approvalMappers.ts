// Mappers from approval API shapes into the props expected by the shared IMS
// presentation primitives (IMSAuditTimeline, IMSApprovalStamp).
//
// derivePrincipalName is also exported for unit-testable use in composables
// that map delegation/proxy signatures to the principalUserName display field.

/**
 * Resolves the display name shown after "for " in delegated/proxy signature rows.
 * Returns undefined when no delegation is in effect (isDelegated is falsy).
 * Returns the descriptor's display_name when available; otherwise "Unknown User".
 * Never returns a raw UUID string.
 */
export function derivePrincipalName(
  isDelegated: boolean | undefined,
  descriptor: { display_name?: string } | null | undefined,
): string | undefined {
  if (!isDelegated) return undefined;
  return descriptor?.display_name || "Unknown User";
}

import { prettify } from "./approvalStatus";
import type { ApprovalEvent, ApprovalSignature } from "../types";

export interface TimelineEventVM {
  actor: string;
  action: string;
  timestamp: string;
  target?: string;
  module?: string;
  metadata?: Record<string, unknown> | null;
}

/** Maps immutable approval events to IMSAuditTimeline event view-models. */
export function toTimelineEvents(events: ApprovalEvent[]): TimelineEventVM[] {
  return events.map((e) => ({
    actor: e.actor?.display_name || e.actor_name || "System",
    action: prettify(e.event_type ?? ""),
    timestamp: e.created_at ?? "",
    target: e.stage_number != null ? `Stage ${e.stage_number}` : undefined,
    module: "approval",
    metadata: buildMeta(e),
  }));
}

function buildMeta(e: ApprovalEvent): Record<string, unknown> | null {
  const meta: Record<string, unknown> = {};
  if (e.comment) meta.comment = e.comment;
  if (e.delegated_from?.display_name) meta.delegated_from = e.delegated_from.display_name;
  if (e.metadata && typeof e.metadata === "object") {
    Object.assign(meta, e.metadata as Record<string, unknown>);
  }
  return Object.keys(meta).length > 0 ? meta : null;
}

export interface StampVM {
  name: string;
  role: string;
  timestamp: string | null;
  delegated: boolean;
  status: "approved";
}

/** Maps signature records to IMSApprovalStamp props. */
export function toStamps(signatures: ApprovalSignature[]): StampVM[] {
  return signatures.map((s) => ({
    name: s.signer?.display_name || s.signer_display_name || "Unknown User",
    role: s.stage_number != null ? `Stage ${s.stage_number}` : "Approver",
    timestamp: s.signed_at ?? null,
    delegated: Boolean(s.is_proxy_signature) || s.signature_label === "DELEGATED",
    status: "approved",
  }));
}
