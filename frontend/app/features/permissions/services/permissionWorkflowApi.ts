import { useApi } from "~/composables/useApi";

import type {
  EffectivePermissions,
  PermissionChangeItem,
  PermissionChangeRequest,
  PermissionGroupSummary,
  PermissionLabel,
  PermissionRole,
  PermissionUserSummary,
} from "../types";

interface ApiEnvelope<T> {
  data: T;
}

interface ListPayload<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

async function apiData<T>(url: string, options: Record<string, unknown> = {}) {
  const { apiFetch } = useApi();
  const response = await apiFetch<ApiEnvelope<T>>(url, options);
  return response.data;
}

export const permissionWorkflowApi = {
  listRequests(query: Record<string, string | number | undefined> = {}) {
    return apiData<ListPayload<PermissionChangeRequest>>("/permissions/change-requests", {
      query,
    });
  },

  getRequest(id: string) {
    return apiData<PermissionChangeRequest>(`/permissions/change-requests/${id}`);
  },

  createRequest(payload: Partial<PermissionChangeRequest>) {
    return apiData<PermissionChangeRequest>("/permissions/change-requests", {
      method: "POST",
      body: payload,
    });
  },

  updateRequest(id: string, payload: Partial<PermissionChangeRequest>) {
    return apiData<PermissionChangeRequest>(`/permissions/change-requests/${id}`, {
      method: "PUT",
      body: payload,
    });
  },

  submit(id: string) {
    return apiData<PermissionChangeRequest>(`/permissions/change-requests/${id}/submit`, {
      method: "POST",
    });
  },

  approve(id: string, comment = "", stepId?: string) {
    const path = stepId
      ? `/permissions/change-requests/${id}/approval-steps/${stepId}/approve`
      : `/permissions/change-requests/${id}/approve`;
    return apiData<PermissionChangeRequest>(path, {
      method: "POST",
      body: { comment },
    });
  },

  requestChanges(id: string, comment = "", stepId?: string) {
    const path = stepId
      ? `/permissions/change-requests/${id}/approval-steps/${stepId}/request-changes`
      : `/permissions/change-requests/${id}/request-changes`;
    return apiData<PermissionChangeRequest>(path, {
      method: "POST",
      body: { comment },
    });
  },

  reject(id: string, comment = "", stepId?: string) {
    const path = stepId
      ? `/permissions/change-requests/${id}/approval-steps/${stepId}/reject`
      : `/permissions/change-requests/${id}/reject`;
    return apiData<PermissionChangeRequest>(path, {
      method: "POST",
      body: { comment },
    });
  },

  merge(id: string) {
    return apiData<PermissionChangeRequest>(`/permissions/change-requests/${id}/merge`, {
      method: "POST",
    });
  },

  rerunChecks(id: string) {
    return apiData(`/permissions/change-requests/${id}/rerun-checks`, {
      method: "POST",
    });
  },

  addItem(id: string, item: Partial<PermissionChangeItem>) {
    return apiData<PermissionChangeItem>(`/permissions/change-requests/${id}/items`, {
      method: "POST",
      body: item,
    });
  },

  addComment(id: string, comment: string) {
    return apiData(`/permissions/change-requests/${id}/comments`, {
      method: "POST",
      body: { comment },
    });
  },

  listLabels() {
    return apiData<PermissionLabel[]>("/permissions/labels");
  },

  listRoles() {
    return apiData<PermissionRole[]>("/permissions/roles");
  },

  listUsers(query: Record<string, string | number | undefined> = {}) {
    return apiData<ListPayload<PermissionUserSummary>>("/permissions/users", {
      query,
    });
  },

  listGroups(query: Record<string, string | number | undefined> = {}) {
    return apiData<ListPayload<PermissionGroupSummary>>("/permissions/groups", {
      query,
    });
  },

  getEffectivePermissions(userId: string) {
    return apiData<EffectivePermissions>(`/permissions/effective/users/${userId}`);
  },

  listFunctionDefinitions() {
    return apiData<any[]>("/permissions/function-definitions");
  },

  requestRoleAssignment(userId: string, payload: { role_code: string; role_id?: string; reason: string }) {
    return apiData<{ request: PermissionChangeRequest; item: PermissionChangeItem }>(
      `/permissions/users/${userId}/role-assignment-request`,
      {
        method: "POST",
        body: payload,
      }
    );
  },
};
