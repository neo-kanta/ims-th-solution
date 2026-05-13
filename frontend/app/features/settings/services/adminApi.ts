import {
  assertOpenApiResponse,
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths } from "~/api/ims-api";

import type {
  AdminUser,
  AdminSession,
  AdminUserFilters,
  AdminUserListPayload,
  AdminUserStatusAction,
  CreateAdminUserInput,
} from "../admin.types";

type AdminUsersQuery = NonNullable<
  paths["/admin/users"]["get"]["parameters"]["query"]
>;
type AdminUserListResponse =
  paths["/admin/users"]["get"]["responses"][200]["content"]["application/json"];
type CreateUserRequest =
  paths["/admin/users"]["post"]["requestBody"]["content"]["application/json"];
type CreateUserResponse =
  paths["/admin/users"]["post"]["responses"][201]["content"]["application/json"];
type AdminResetPasswordRequest =
  paths["/admin/users/{id}/reset-password"]["post"]["requestBody"]["content"]["application/json"];
type AdminUserResponse = AdminUserListResponse["users"][number];
type AdminSessionResponse =
  paths["/admin/users/{id}/sessions"]["get"]["responses"][200]["content"]["application/json"];
type SessionResponse = AdminSessionResponse[number];
type AdminUserStatusPath =
  | "/admin/users/{id}/disable"
  | "/admin/users/{id}/enable"
  | "/admin/users/{id}/lock"
  | "/admin/users/{id}/unlock";

const USER_STATUS_PATHS: Record<AdminUserStatusAction, AdminUserStatusPath> = {
  disable: "/admin/users/{id}/disable",
  enable: "/admin/users/{id}/enable",
  lock: "/admin/users/{id}/lock",
  unlock: "/admin/users/{id}/unlock",
};

function buildUserQuery(filters: AdminUserFilters = {}) {
  const query: AdminUsersQuery = {};

  if (filters.search?.trim()) {
    query.search = filters.search.trim();
  }

  if (typeof filters.is_active === "boolean") {
    query.is_active = String(filters.is_active);
  }

  if (typeof filters.is_locked === "boolean") {
    query.is_locked = String(filters.is_locked);
  }

  if (typeof filters.offset === "number") {
    query.offset = filters.offset;
  }

  if (typeof filters.limit === "number") {
    query.limit = filters.limit;
  }

  return query;
}

function normalizeAdminUser(user: AdminUserResponse): AdminUser {
  return {
    id: user.id ?? "",
    username: user.username ?? "",
    display_name: user.display_name ?? "",
    email: user.email ?? "",
    is_active: Boolean(user.is_active),
    is_locked: Boolean(user.is_locked),
    locked_until: user.locked_until ?? null,
    failed_login_attempts: user.failed_login_attempts ?? 0,
    force_password_change: Boolean(user.force_password_change),
    last_login_at: user.last_login_at ?? null,
    password_changed_at: user.password_changed_at ?? null,
    groups: user.groups ?? [],
    created_at: user.created_at ?? "",
    updated_at: user.updated_at ?? "",
  };
}

function normalizeUserList(response: AdminUserListResponse): AdminUserListPayload {
  return {
    users: (response.users ?? []).map(normalizeAdminUser),
    total: response.total ?? 0,
    offset: response.offset ?? 0,
    limit: response.limit ?? 0,
  };
}

function normalizeSession(session: SessionResponse): AdminSession {
  return {
    id: session.id ?? "",
    ip_address: session.ip_address ?? "",
    user_agent: session.user_agent ?? "",
    last_activity_at: session.last_activity_at ?? "",
    created_at: session.created_at ?? "",
    expires_at: session.expires_at ?? "",
  };
}

export const adminApi = {
  async listUsers(filters: AdminUserFilters = {}) {
    const client = useOpenApiClient();
    const response = await client.GET("/admin/users", {
      params: {
        query: buildUserQuery(filters),
      },
    });

    return normalizeUserList(unwrapOpenApiResponse<AdminUserListResponse>(response));
  },

  async createUser(payload: CreateAdminUserInput) {
    const client = useOpenApiClient();
    const body: CreateUserRequest = payload;
    const response = await client.POST("/admin/users", {
      body,
    });

    return unwrapOpenApiResponse<CreateUserResponse>(response);
  },

  async setUserStatus(userId: string, action: AdminUserStatusAction) {
    const client = useOpenApiClient();
    const response = await client.POST(USER_STATUS_PATHS[action], {
      params: {
        path: {
          id: userId,
        },
      },
    });

    assertOpenApiResponse(response);
  },

  async resetPassword(userId: string, newPassword: string) {
    const client = useOpenApiClient();
    const body: AdminResetPasswordRequest = { new_password: newPassword };
    const response = await client.POST("/admin/users/{id}/reset-password", {
      params: {
        path: {
          id: userId,
        },
      },
      body,
    });

    assertOpenApiResponse(response);
  },

  async listUserSessions(userId: string) {
    const client = useOpenApiClient();
    const response = await client.GET("/admin/users/{id}/sessions", {
      params: {
        path: {
          id: userId,
        },
      },
    });

    return unwrapOpenApiResponse<AdminSessionResponse>(response).map(normalizeSession);
  },

  async revokeSession(sessionId: string) {
    const client = useOpenApiClient();
    const response = await client.POST("/admin/sessions/{id}/revoke", {
      params: {
        path: {
          id: sessionId,
        },
      },
    });

    assertOpenApiResponse(response);
  },
};
