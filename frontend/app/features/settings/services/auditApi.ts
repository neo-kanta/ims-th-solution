import {
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths } from "~/api/ims-api";

import type { AuditEvent, AuditFilters, AuditListPayload } from "../audit.types";

type AuditQuery = NonNullable<
  paths["/admin/audit"]["get"]["parameters"]["query"]
>;
type AuditExportQuery = NonNullable<
  paths["/admin/audit/export"]["get"]["parameters"]["query"]
>;
type AuditListResponse =
  paths["/admin/audit"]["get"]["responses"][200]["content"]["application/json"];
type AuditEventResponse = AuditListResponse["events"][number];

export function buildAuditQuery(filters: AuditFilters = {}) {
  const query: AuditQuery = {};

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

function buildAuditExportQuery(filters: AuditFilters = {}) {
  const query: AuditExportQuery = {};

  if (filters.actor_id?.trim()) {
    query.actor_id = filters.actor_id.trim();
  }

  if (filters.event_type?.trim()) {
    query.event_type = filters.event_type.trim();
  }

  if (filters.since) {
    query.since = filters.since;
  }

  if (filters.until) {
    query.until = filters.until;
  }

  return query;
}

function normalizeAuditEvent(event: AuditEventResponse): AuditEvent {
  return {
    id: event.id ?? "",
    actor_id: event.actor_id ?? null,
    event_type: event.event_type ?? "",
    target_type: event.target_type ?? "",
    target_id: event.target_id ?? "",
    ip_address: event.ip_address ?? "",
    user_agent: event.user_agent ?? "",
    metadata: event.metadata,
    created_at: event.created_at ?? "",
  };
}

function normalizeAuditList(response: AuditListResponse): AuditListPayload {
  return {
    events: (response.events ?? []).map(normalizeAuditEvent),
    total: response.total ?? 0,
    offset: response.offset ?? 0,
    limit: response.limit ?? 0,
  };
}

export const auditApi = {
  async listEvents(filters: AuditFilters = {}) {
    const client = useOpenApiClient();
    const response = await client.GET("/admin/audit", {
      params: {
        query: buildAuditQuery(filters),
      },
    });

    return normalizeAuditList(unwrapOpenApiResponse<AuditListResponse>(response));
  },

  async exportEvents(filters: AuditFilters = {}) {
    const client = useOpenApiClient();
    const response = await client.GET("/admin/audit/export", {
      params: {
        query: buildAuditExportQuery(filters),
      },
      parseAs: "blob",
    });

    return unwrapOpenApiResponse<Blob>(response);
  },
};
