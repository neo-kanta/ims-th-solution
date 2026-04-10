import { useApi } from "~/composables/useApi";
import type { ApiResponse } from "~/types/api.types";

import type { AuditFilters, AuditListPayload } from "../types/audit.types";

const ADMIN_BASE = "/admin";

export function buildAuditQuery(filters: AuditFilters = {}) {
  const query: Record<string, string | number> = {};

  if (filters.actor_id?.trim()) {
    query.actor_id = filters.actor_id.trim();
  }

  if (filters.event_type?.trim()) {
    query.event_type = filters.event_type.trim();
  }

  if (filters.target_type?.trim()) {
    query.target_type = filters.target_type.trim();
  }

  if (filters.target_id?.trim()) {
    query.target_id = filters.target_id.trim();
  }

  if (filters.since) {
    query.since = filters.since;
  }

  if (filters.until) {
    query.until = filters.until;
  }

  if (typeof filters.offset === "number") {
    query.offset = filters.offset;
  }

  if (typeof filters.limit === "number") {
    query.limit = filters.limit;
  }

  return query;
}

export const auditApi = {
  listEvents(filters: AuditFilters = {}) {
    const { apiFetch } = useApi();

    return apiFetch<ApiResponse<AuditListPayload>>(`${ADMIN_BASE}/audit`, {
      method: "GET",
      query: buildAuditQuery(filters),
    });
  },
};
