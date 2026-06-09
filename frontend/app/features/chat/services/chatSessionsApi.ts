// Chat session-history API service. All calls go through useApi() (auth header
// + base URL) and use the generated OpenAPI types — nothing here is hand-typed
// (per CLAUDE.md). Responses are wrapped in the standard {data} envelope.

import { useApi } from "~/composables/useApi";

import type { components } from "~/api/ims-api";

export type SessionSummary = components["schemas"]["SessionSummaryResponse"];
export type SessionListResponse = components["schemas"]["SessionListResponse"];
export type SessionMessagesResponse = components["schemas"]["SessionMessagesResponse"];
export type ChatMessageResponse = components["schemas"]["ChatMessageResponse"];

interface Envelope<T> {
  data: T;
  message?: string;
}

function unwrap<T>(e: Envelope<T>): T {
  return e.data;
}

function query(params: Record<string, string | number | undefined>): string {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === "") {
      continue;
    }
    sp.set(k, String(v));
  }
  const s = sp.toString();
  return s ? `?${s}` : "";
}

export const chatSessionsApi = {
  async list(params: { page?: number; limit?: number } = {}): Promise<SessionListResponse> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<SessionListResponse>>(`/chat/sessions${query(params)}`));
  },

  async get(sessionId: string): Promise<SessionSummary> {
    const { apiFetch } = useApi();
    return unwrap(await apiFetch<Envelope<SessionSummary>>(`/chat/sessions/${sessionId}`));
  },

  async messages(
    sessionId: string,
    params: { page?: number; limit?: number } = {},
  ): Promise<SessionMessagesResponse> {
    const { apiFetch } = useApi();
    return unwrap(
      await apiFetch<Envelope<SessionMessagesResponse>>(
        `/chat/sessions/${sessionId}/messages${query(params)}`,
      ),
    );
  },
};
