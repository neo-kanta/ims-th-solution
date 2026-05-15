import {
  assertOpenApiResponse,
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths } from "~/api/ims-api";

import type {
  ChangePersonalPasswordInput,
  PersonalAccountPayload,
  PersonalAccountSession,
  PersonalAccountUser,
  PersonalMfaStatus,
} from "../account.types";

type MeResponse =
  paths["/auth/me"]["get"]["responses"][200]["content"]["application/json"];
type UserResponse = NonNullable<MeResponse["user"]>;
type PermissionsResponse = NonNullable<MeResponse["permissions"]>;
type ChangePasswordRequest =
  paths["/auth/change-password"]["post"]["requestBody"]["content"]["application/json"];
type MfaStatusResponse =
  paths["/auth/mfa/status"]["get"]["responses"][200]["content"]["application/json"];
type SessionListResponse =
  paths["/auth/sessions"]["get"]["responses"][200]["content"]["application/json"];
type SessionResponse = SessionListResponse[number];

function normalizeUser(user: UserResponse | undefined): PersonalAccountUser {
  return {
    id: user?.id ?? "",
    username: user?.username ?? "",
    display_name: user?.display_name ?? "",
    email: user?.email ?? "",
    groups: user?.groups ?? [],
  };
}

function normalizePermissions(permissions: PermissionsResponse | undefined) {
  return {
    functions: permissions?.functions ?? [],
    contracts: permissions?.contracts ?? [],
  };
}

function normalizeMe(response: MeResponse): PersonalAccountPayload {
  return {
    user: normalizeUser(response.user),
    permissions: normalizePermissions(response.permissions),
  };
}

function normalizeMfaStatus(response: MfaStatusResponse): PersonalMfaStatus {
  return {
    enrolled: Boolean(response.enrolled),
    enabled: Boolean(response.enabled),
    recovery_codes_left: response.recovery_codes_left ?? 0,
  };
}

function normalizeSession(session: SessionResponse): PersonalAccountSession {
  return {
    id: session.id ?? "",
    ip_address: session.ip_address ?? "",
    user_agent: session.user_agent ?? "",
    last_activity_at: session.last_activity_at ?? "",
    created_at: session.created_at ?? "",
    expires_at: session.expires_at ?? "",
  };
}

export const accountApi = {
  async me() {
    const client = useOpenApiClient();
    const response = await client.GET("/auth/me");

    return normalizeMe(unwrapOpenApiResponse<MeResponse>(response));
  },

  async changePassword(payload: ChangePersonalPasswordInput) {
    const client = useOpenApiClient();
    const body: ChangePasswordRequest = payload;
    const response = await client.POST("/auth/change-password", {
      body,
    });

    assertOpenApiResponse(response);
  },

  async mfaStatus() {
    const client = useOpenApiClient();
    const response = await client.GET("/auth/mfa/status");

    return normalizeMfaStatus(unwrapOpenApiResponse<MfaStatusResponse>(response));
  },

  async listSessions() {
    const client = useOpenApiClient();
    const response = await client.GET("/auth/sessions");

    return unwrapOpenApiResponse<SessionListResponse>(response).map(normalizeSession);
  },

  async revokeSession(sessionId: string) {
    const client = useOpenApiClient();
    const response = await client.POST("/auth/sessions/{id}/revoke", {
      params: {
        path: {
          id: sessionId,
        },
      },
    });

    assertOpenApiResponse(response);
  },
};
