// Mappers from approval API shapes into the props expected by the shared IMS
// presentation primitives (IMSAuditTimeline, IMSApprovalStamp).

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
    actor: e.actor_name || e.actor_user_id || "System",
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
  if (e.delegated_from_user_id) meta.delegated_from = e.delegated_from_user_id;
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
    name: s.signer_display_name || s.signer_user_id || "—",
    role: s.stage_number != null ? `Stage ${s.stage_number}` : "Approver",
    timestamp: s.signed_at ?? null,
    delegated: Boolean(s.is_proxy_signature) || s.signature_label === "DELEGATED",
    status: "approved",
  }));
}
