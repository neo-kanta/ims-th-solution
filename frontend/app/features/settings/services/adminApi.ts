import { useApi } from "~/composables/useApi";
import type { ApiResponse } from "~/types/api.types";

import type {
  AdminSession,
  AdminUserFilters,
  AdminUserListPayload,
  AdminUserStatusAction,
  CreateAdminUserInput,
} from "../types/admin.types";

const ADMIN_BASE = "/admin";

function buildUserQuery(filters: AdminUserFilters = {}) {
  const query: Record<string, string | number | boolean> = {};

  if (filters.search?.trim()) {
    query.search = filters.search.trim();
  }

  if (typeof filters.is_active === "boolean") {
    query.is_active = filters.is_active;
  }

  if (typeof filters.is_locked === "boolean") {
    query.is_locked = filters.is_locked;
  }

  if (typeof filters.offset === "number") {
    query.offset = filters.offset;
  }

  if (typeof filters.limit === "number") {
    query.limit = filters.limit;
  }

  return query;
}

export const adminApi = {
  listUsers(filters: AdminUserFilters = {}) {
    const { apiFetch } = useApi();

    return apiFetch<ApiResponse<AdminUserListPayload>>(`${ADMIN_BASE}/users`, {
      method: "GET",
      query: buildUserQuery(filters),
    });
  },

  createUser(payload: CreateAdminUserInput) {
    const { apiFetch } = useApi();

    return apiFetch<{ id: string }>(`${ADMIN_BASE}/users`, {
      method: "POST",
      body: payload,
    });
  },

  setUserStatus(userId: string, action: AdminUserStatusAction) {
    const { apiFetch } = useApi();

    return apiFetch<void>(`${ADMIN_BASE}/users/${userId}/${action}`, {
      method: "POST",
    });
  },

  resetPassword(userId: string, newPassword: string) {
    const { apiFetch } = useApi();

    return apiFetch<void>(`${ADMIN_BASE}/users/${userId}/reset-password`, {
      method: "POST",
      body: { new_password: newPassword },
    });
  },

  listUserSessions(userId: string) {
    const { apiFetch } = useApi();

    return apiFetch<ApiResponse<AdminSession[]>>(`${ADMIN_BASE}/users/${userId}/sessions`, {
      method: "GET",
    });
  },

  revokeSession(sessionId: string) {
    const { apiFetch } = useApi();

    return apiFetch<void>(`${ADMIN_BASE}/sessions/${sessionId}/revoke`, {
      method: "POST",
    });
  },
};
