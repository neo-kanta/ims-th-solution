// UI helpers for mapping approval enums to display labels and badge variants.

import type { RequestStatus } from "../types";

/** Maps an approval request status to an AppStatusBadge status keyword. */
export function requestStatusBadge(status: string): string {
  switch (status) {
    case "APPROVED":
      return "approved";
    case "REJECTED":
      return "rejected";
    case "PENDING_APPROVAL":
    case "SUBMITTED":
      return "pending";
    case "DRAFT":
      return "neutral";
    case "CANCELLED":
    case "WITHDRAWN":
      return "inactive";
    default:
      return "neutral";
  }
}

/** Maps a task status to an AppStatusBadge status keyword. */
export function taskStatusBadge(status: string): string {
  switch (status) {
    case "APPROVED":
      return "approved";
    case "REJECTED":
      return "rejected";
    case "PENDING":
      return "pending";
    case "SKIPPED":
    case "CANCELLED":
      return "inactive";
    default:
      return "neutral";
  }
}

/** Human-readable label for a request status. */
export function requestStatusLabel(status: string): string {
  return prettify(status);
}

/** Human-readable label for a process / subject / approver-mode enum. */
export function prettify(value: string): string {
  if (!value) return "—";
  return value
    .toLowerCase()
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

/** True when a request is still in flight (actionable / not terminal). */
export function isActiveStatus(status: RequestStatus | string): boolean {
  return (
    status === "DRAFT" ||
    status === "SUBMITTED" ||
    status === "PENDING_APPROVAL"
  );
}
