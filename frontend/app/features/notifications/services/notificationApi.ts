import { useOpenApiClient, unwrapOpenApiResponse } from "~/api/openapi";
import type { components } from "~/api/ims-api";
import type {
  NotificationList,
  MarkReadResult,
  MarkAllReadResult,
  EmailHealth,
  OutboxList,
  OutboxDetail,
  RetryResult,
  TestEmailBody,
  TestEmailResult,
  OutboxFilter,
} from "../types";

export const notificationApi = {
  // ── User notification center ──────────────────────────────────────────────
  async list(params: { unread_only?: boolean; limit?: number; offset?: number } = {}): Promise<NotificationList> {
    const client = useOpenApiClient();
    const result = await client.GET("/notifications", {
      params: { query: params },
    });
    return unwrapOpenApiResponse(result) as components["schemas"]["listResponse"];
  },

  async markRead(id: string): Promise<MarkReadResult> {
    const client = useOpenApiClient();
    const result = await client.POST("/notifications/{id}/read", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse(result) as components["schemas"]["markReadResponse"];
  },

  async markAllRead(): Promise<MarkAllReadResult> {
    const client = useOpenApiClient();
    const result = await client.POST("/notifications/read-all");
    return unwrapOpenApiResponse(result) as components["schemas"]["markAllReadResponse"];
  },

  // ── Email admin ───────────────────────────────────────────────────────────
  async emailHealth(): Promise<EmailHealth> {
    const client = useOpenApiClient();
    const result = await client.GET("/notifications/email/health");
    return unwrapOpenApiResponse(result) as components["schemas"]["healthResponse"];
  },

  async sendTestEmail(body: TestEmailBody): Promise<TestEmailResult> {
    const client = useOpenApiClient();
    const result = await client.POST("/notifications/email/test", { body });
    return unwrapOpenApiResponse(result) as components["schemas"]["testEmailResponse"];
  },

  async listOutbox(filter: OutboxFilter = {}): Promise<OutboxList> {
    const client = useOpenApiClient();
    const result = await client.GET("/notifications/email-outbox", {
      params: { query: filter },
    });
    return unwrapOpenApiResponse(result) as components["schemas"]["outboxListResponse"];
  },

  async getOutboxDetail(outbox_id: string): Promise<OutboxDetail> {
    const client = useOpenApiClient();
    const result = await client.GET("/notifications/email-outbox/{outbox_id}", {
      params: { path: { outbox_id } },
    });
    return unwrapOpenApiResponse(result) as components["schemas"]["outboxDetailResponse"];
  },

  async retryOutbox(outbox_id: string): Promise<RetryResult> {
    const client = useOpenApiClient();
    const result = await client.POST("/notifications/email-outbox/{outbox_id}/retry", {
      params: { path: { outbox_id } },
    });
    return unwrapOpenApiResponse(result) as components["schemas"]["retryResponse"];
  },
};

export function notificationErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object") {
    const data = (err as { data?: { error?: unknown } }).data;
    if (data && typeof data.error === "string" && data.error.trim()) return data.error;
    const status = (err as { status?: number }).status;
    if (status === 403) return "You do not have permission to perform this action.";
    const message = (err as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) return message;
  }
  return fallback;
}
