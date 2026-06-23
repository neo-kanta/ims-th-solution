// Notification feature types.
//
// API shapes are NOT hand-written: every type below is a direct alias of a
// schema generated into `~/api/ims-api` from the backend Swagger.
// UI-only helpers live at the bottom.

import type { components } from "~/api/ims-api";

type Schemas = components["schemas"];

// ── Response types ────────────────────────────────────────────────────────────
export type NotificationItem = Schemas["notificationResponse"];
export type NotificationList = Schemas["listResponse"];
export type NotificationAction = Schemas["actionResponse"];
export type NotificationContext = Schemas["contextResponse"];
export type NotificationEvent = Schemas["eventResponse"];
export type NotificationRecipient = Schemas["userSummaryResponse"];
export type MarkReadResult = Schemas["markReadResponse"];
export type MarkAllReadResult = Schemas["markAllReadResponse"];

export type EmailHealth = Schemas["healthResponse"];

export type OutboxItem = Schemas["outboxItemResponse"];
export type OutboxList = Schemas["outboxListResponse"];
export type OutboxDetail = Schemas["outboxDetailResponse"];
export type RetryResult = Schemas["retryResponse"];

export type TestEmailBody = Schemas["testEmailRequest"];
export type TestEmailResult = Schemas["testEmailResponse"];

// ── UI-only helpers ───────────────────────────────────────────────────────────
export type OutboxStatus = "PENDING" | "SENDING" | "SENT" | "FAILED" | "DEAD";

export interface OutboxFilter {
  status?: string;
  recipient_username?: string;
  recipient_email?: string;
  event_type?: string;
  event_category?: string;
  business_type?: string;
  business_reference?: string;
  created_from?: string;
  created_to?: string;
  limit?: number;
  offset?: number;
}
