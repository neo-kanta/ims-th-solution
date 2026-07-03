import type { APIRequestContext } from "@playwright/test";

export const API_BASE_URL = process.env.E2E_API_BASE_URL ?? "http://localhost:8080/api/v1";

export interface ApiLoginResult {
  status: number;
  accessToken?: string;
  refreshToken?: string;
  accessTokenExpiresAt?: string;
  error?: string;
}

/** Calls the real backend login endpoint directly (no browser involved). */
export async function apiLogin(
  request: APIRequestContext,
  username: string,
  password: string,
): Promise<ApiLoginResult> {
  const resp = await request.post(`${API_BASE_URL}/auth/login`, {
    data: { username, password, totp_code: "", recovery_code: "" },
  });
  const body = await resp.json().catch(() => ({}) as Record<string, unknown>);

  if (resp.ok()) {
    const data = (body as { data?: Record<string, string> }).data ?? {};
    return {
      status: resp.status(),
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      accessTokenExpiresAt: data.access_token_expires_at,
    };
  }
  return { status: resp.status(), error: (body as { error?: string }).error };
}

/** Thin bearer-token-aware wrapper for direct backend calls (401/403 assertions independent of the browser). */
export function withToken(request: APIRequestContext, token: string) {
  const headers = { Authorization: `Bearer ${token}` };
  return {
    get: (path: string) => request.get(`${API_BASE_URL}${path}`, { headers }),
    post: (path: string, data?: unknown) => request.post(`${API_BASE_URL}${path}`, { headers, data }),
  };
}
