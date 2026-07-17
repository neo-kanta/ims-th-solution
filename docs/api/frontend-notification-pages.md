# Frontend Notification Pages Documentation

Status:  
Draft proposal / ready for frontend implementation

Scope:  
Notification center UI, email outbox admin UI, email health UI, test email UI, retry failed email UI.

Audience:  
Frontend developers, backend developers, product owner, demo operators.

## Header Name

Frontend Notification Pages

## Overview

The frontend consumes the IMS Notification API documented in `notification-email-api.md`.

The user notification center is for normal authenticated users. It displays in-app notifications created by backend business workflows and lets users mark notifications as read.

Email outbox, test email, retry, and health screens are admin/operator tools. They are for operational inspection, demo verification, and controlled retry actions.

The frontend does not know whether the backend uses Mailpit, Gmail SMTP, company SMTP, OCI Email Delivery, SendGrid SMTP, or any other SMTP provider. The frontend never handles SMTP credentials and must never expose SMTP password or secret fields.

The frontend must not send business emails directly or create a frontend-only direct email workflow. Business workflows trigger notifications internally on the backend. The frontend only:

1. Displays user notifications.
2. Lets users mark notifications as read.
3. Lets admin/operator users inspect email health and outbox.
4. Lets admin/operator users send test email if permission allows.
5. Lets admin/operator users retry failed email if permission allows.

## Page / Component Map

| Page / Component           | Route suggestion                                | User type          | Purpose                      |
| -------------------------- | ----------------------------------------------- | ------------------ | ---------------------------- |
| Notification Bell          | global layout/header                            | authenticated user | Quick unread notifications   |
| Notification Center Page   | `/notifications`                                | authenticated user | Full notification history    |
| Email Operations Dashboard | `/admin/notifications/email`                    | admin/operator     | Email health + queue summary |
| Email Outbox List          | `/admin/notifications/email/outbox`             | admin/operator     | Delivery log                 |
| Email Outbox Detail        | `/admin/notifications/email/outbox/{outbox_id}` | admin/operator     | Inspect one delivery record  |
| Test Email Panel           | `/admin/notifications/email/test` or embedded   | admin/operator     | Queue test email             |
| Retry Action               | outbox list/detail action                       | admin/operator     | Requeue failed/dead email    |

Canonical implementation units:

1. Notification Bell
2. Notification Center Page
3. Email Operations Dashboard
4. Email Outbox List
5. Email Outbox Detail
6. Test Email Panel
7. Retry Action

## API Mapping

All Notification API calls must go through the generated frontend API client from the current Swagger/OpenAPI contract. Do not hand-code `fetch`/`$fetch` calls with raw notification endpoint strings. If the generated client is missing one of these endpoints, stop frontend implementation for that endpoint and regenerate the client first.

### Notification Center APIs

#### `GET /api/v1/notifications`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | On notification bell load, notification center page load, refresh, filter change, pagination change, and after read actions if a fresh count/list is needed. |
| Required permission | Authenticated user only. User receives only their own notifications. |
| Request model | Query params: `unread_only?: boolean`, `limit?: number`, `offset?: number`. |
| Response fields used by UI | `items`, `total`, `unread`, and per item: `notification_id`, `recipient`, `event`, `title`, `body`, `context`, `action`, `is_read`, `read_at`, `created_at`. |
| Loading state | Bell shows count skeleton or neutral count. Center page shows list/table skeleton. |
| Empty state | Bell: no dropdown items or "No unread notifications." Center page: "No notifications found." |
| Error state | 401 follows app login handling. Network/API errors show retry action. Validation errors should not occur unless filters are malformed. |

#### `POST /api/v1/notifications/{notification_id}/read`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | When user clicks explicit mark-read action or after opening an unread notification if product chooses auto-read-on-open. |
| Required permission | Authenticated user only. Backend enforces ownership. |
| Request model | Path param: `notification_id`. Body: none. |
| Response fields used by UI | `notification_id`, `recipient`, `status`, `read_at`. |
| Loading state | Disable only the row/action being updated. Optimistic update is acceptable if rollback is implemented on failure. |
| Empty state | Not applicable. |
| Error state | 401 login handling. 404 means the notification is gone or not accessible. Network/server errors keep the item unread and show retry feedback. |

#### `POST /api/v1/notifications/read-all`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | From notification bell/dropdown or notification center "Mark all as read" action. |
| Required permission | Authenticated user only. Backend updates only the current user's notifications. |
| Request model | Body: none. |
| Response fields used by UI | `recipient`, `updated`. |
| Loading state | Disable "Mark all as read" while request is pending. |
| Empty state | If unread count is zero, hide or disable the action. |
| Error state | 401 login handling. Network/server errors leave current unread state visible and show retry feedback. |

### Email Operations APIs

#### `GET /api/v1/notifications/email/health`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | On Email Operations Dashboard load and manual refresh. |
| Required permission | `permission.notification.health` or `permission.notification.view`. |
| Request model | Body: none. |
| Response fields used by UI | `enabled`, `worker_enabled`, `smtp_host`, `smtp_port`, `smtp_tls_mode`, `from_address`, `from_name`, `send_real_email`, `allow_raw_email`, `test_endpoint_enabled`, `worker_interval`, `worker_batch_size`, `stale_sending_timeout`, `retry_policy`, `pending_count`, `failed_count`, `dead_count`. |
| Loading state | Dashboard health cards show skeleton values. |
| Empty state | Not applicable. If email is disabled, show disabled status and explain that outbox sending is not active. |
| Error state | 403 permission denied. 500/network errors show retryable server error. Never ask for SMTP credentials in the UI. |

#### `POST /api/v1/notifications/email/test`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | When admin/operator submits the Test Email Panel. |
| Required permission | `permission.notification.test` or `permission.notification.edit`. |
| Request model | Either `{ to_username, subject, body }` or `{ to_email, subject, body }` when raw email is allowed by backend. |
| Response fields used by UI | `outbox_id`, `recipient`, `event`, `status`. Expected immediate status is `QUEUED`. |
| Loading state | Disable submit button and prevent duplicate submits while pending. |
| Empty state | Not applicable. Form starts blank or with demo defaults if product wants. |
| Error state | 403 permission denied. 422 validation errors display beside fields. 500/network errors show retry action. |

### Email Outbox APIs

#### `GET /api/v1/notifications/email-outbox`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | On Email Outbox List load, filter change, pagination change, after retry, and from dashboard links. |
| Required permission | `permission.notification.view`. |
| Request model | Query params: `status`, `recipient_username`, `recipient_email`, `event_type`, `event_category`, `business_type`, `business_reference`, `created_from`, `created_to`, `limit`, `offset`. |
| Response fields used by UI | `items`, `total`, and per item: `outbox_id`, `notification_id`, `recipient`, `to_email`, `to_name`, `event`, `context`, `subject`, `status`, `attempts`, `max_attempts`, `next_attempt_at`, `last_error`, `provider_message_id`, `created_at`, `sent_at`. |
| Loading state | Table skeleton and disabled filter submit. |
| Empty state | "No email outbox records found." |
| Error state | 403 permission denied. API/network failure shows retry action and preserves current filters. |

#### `GET /api/v1/notifications/email-outbox/{outbox_id}`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | When admin/operator opens an outbox detail drawer/page. |
| Required permission | `permission.notification.view`. |
| Request model | Path param: `outbox_id`. Body: none. |
| Response fields used by UI | `outbox_id`, `notification_id`, `recipient`, `to_email`, `to_name`, `event`, `context`, `subject`, `body_text`, `body_html`, `status`, `attempts`, `max_attempts`, `next_attempt_at`, `last_error`, `provider_message_id`, `created_at`, `updated_at`, `sent_at`. |
| Loading state | Detail drawer/page skeleton. |
| Empty state | Not applicable. 404 uses not-found state. |
| Error state | 403 permission denied. 404 not found. 500/network errors show retry action. |

#### `POST /api/v1/notifications/email-outbox/{outbox_id}/retry`

| Item | Frontend rule |
| --- | --- |
| When frontend calls it | When admin/operator confirms retry for an outbox row with status `FAILED` or `DEAD`. |
| Required permission | `permission.notification.retry` or `permission.notification.edit`. |
| Request model | Path param: `outbox_id`. Optional body: `{ reason?: string }`. |
| Response fields used by UI | `outbox_id`, `recipient_username`, `event_type`, `business_reference`, `status`, `attempts`, `max_attempts`, `next_attempt_at`. |
| Loading state | Disable retry button for that row/detail while pending. |
| Empty state | Not applicable. Hide retry action unless status is `FAILED` or `DEAD`. |
| Error state | 403 permission denied. 404 not found. 422/409 if status is not retryable. 500/network errors show retry action. |

## Permission Rules

Normal authenticated users can access their own notifications without admin/operator notification permissions.

Admin/operator pages require notification permissions. Use route guards for protected routes and permission-based menu visibility so unauthorized users do not see admin email tools.

Suggested permission keys:

| Permission key | Purpose |
| --- | --- |
| `permission.notification.view` | Read notification/email operational data. |
| `permission.notification.edit` | General notification administration permission. |
| `permission.notification.retry` | Retry failed/dead email outbox rows. |
| `permission.notification.test` | Send demo/test email through outbox. |
| `permission.notification.health` | Read email health/config status without secrets. |

Permission mapping:

| Frontend area/action | Required access |
| --- | --- |
| Notification Bell | Authenticated user |
| Notification Center Page | Authenticated user |
| Mark one notification as read | Authenticated user, own notification |
| Mark all notifications as read | Authenticated user, own notifications |
| Email Operations Dashboard | `permission.notification.health` or `permission.notification.view` |
| Email Outbox List | `permission.notification.view` |
| Email Outbox Detail | `permission.notification.view` |
| Retry Action | `permission.notification.retry` or `permission.notification.edit` |
| Test Email Panel | `permission.notification.test` or `permission.notification.edit` |

## Refresh / Polling Strategy

Use simple refresh and polling. Do not add websocket/SSE realtime behavior unless the project already has a shared realtime notification layer.

Notification bell:

- Fetch on authenticated layout mount.
- Refresh when the browser tab becomes visible again.
- Refresh after mark-read and mark-all-read actions.
- Optional polling: every 60 seconds while the tab is visible and the user is authenticated.

Notification center page:

- Fetch on page load, filter change, and pagination change.
- Refresh after mark-read and mark-all-read actions.
- Manual refresh button is recommended.
- Avoid aggressive polling on the full history page.

Email operations dashboard:

- Fetch health on page load.
- Provide manual refresh.
- Optional polling: every 30 seconds while the page is visible.
- Stop polling when the user leaves the route or loses permission.

Email outbox list/detail:

- Fetch on page/detail load, filter change, pagination change, and after retry/test-email success.
- Optional short polling is allowed during demo verification: every 5 to 10 seconds for rows in `PENDING` or `SENDING`.
- Stop short polling once visible rows are terminal: `SENT`, `DEAD`, or `CANCELLED`.
- Do not poll hidden tabs or background routes.

## Notification Bell

UI behavior:

- Show unread count from `GET /api/v1/notifications`.
- Fetch latest 5 to 10 notifications, usually with `limit=5` or `limit=10`.
- Display `event.label`, `title`, `context.business_reference`, and `created_at`.
- Use `event.severity` for badge tone, not for business logic.
- Clicking a notification navigates to `action.url` if available.
- Mark notification as read after click or through an explicit row action, depending on product decision.
- Include "View all notifications".
- Include "Mark all as read" when unread count is greater than zero.

Data fields to render:

- `notification_id`
- `event.label`
- `event.severity`
- `title`
- `body`
- `context.business_reference`
- `action.url`
- `is_read`
- `created_at`

## Notification Center Page

Filters:

- unread only
- event category
- severity
- date range if useful

Table/list columns:

- status/read
- event label
- title
- business reference
- business title
- created date
- action

Actions:

- open business object through `action.url`
- mark one notification as read
- mark all notifications as read

Empty state:

- "No notifications found."

Error state:

- auth error: follow app login handling
- network error: show retry action
- API validation error: show filter/form error and keep current page stable

## Email Operations Dashboard

The dashboard calls `GET /api/v1/notifications/email/health`.

Show:

- email enabled
- worker enabled
- SMTP host
- SMTP port
- TLS mode
- from address
- from name
- send real email flag
- raw email allowed flag
- test endpoint enabled
- pending count
- failed count
- dead count
- retry policy

Rules:

- Must not show SMTP password or secret fields.
- Avoid showing SMTP username unless backend explicitly returns a non-sensitive value.
- Warn if worker is disabled.
- Warn if failed count or dead count is greater than zero.
- Link to Email Outbox List page with useful filters, such as `status=FAILED` or `status=DEAD`.

## Email Outbox List

Filters:

- status
- recipient username
- recipient email
- event type
- event category
- business type
- business reference
- created date range

Columns:

- status
- recipient username/display name/email
- event label
- business reference
- subject
- attempts
- next attempt
- created at
- sent at
- last error preview

Row actions:

- view detail
- retry if status is `FAILED` or `DEAD`

Status badge mapping:

| Status | UI meaning | Suggested tone |
| --- | --- | --- |
| `PENDING` | Queued and waiting for worker | neutral/info |
| `SENDING` | Claimed by worker and sending | info/loading |
| `SENT` | Delivered to SMTP provider successfully | success |
| `FAILED` | Failed but may be retried | warning |
| `DEAD` | Max attempts exhausted | danger |
| `CANCELLED` | Cancelled and not active | muted |

Empty state:

- "No email outbox records found."

Error state:

- permission denied
- API failure

## Email Outbox Detail

Display full record:

- outbox id
- notification id
- recipient
- event
- context
- subject
- body text
- body HTML preview if safe
- status
- attempts
- max attempts
- next attempt
- last error
- provider message id
- created/updated/sent timestamps

Rules:

- Retry button is visible only for `FAILED` or `DEAD`.
- Do not allow editing body from this page.
- Do not allow editing SMTP config from this page.
- Render `body_html` only through the app's safe/sanitized HTML pattern. If no safe preview exists, show text body only.
- UUIDs can appear here because this is a technical detail page.

## Test Email Panel

The panel calls `POST /api/v1/notifications/email/test`.

Support two modes:

1. Send by `to_username`.
2. Send by raw `to_email`, only if backend health/config allows it.

Fields:

- to username
- to email
- subject
- body

Show queued result:

- outbox id
- recipient
- event
- status = `QUEUED`

After success, provide a link to the outbox detail page or filtered outbox list.

Rules:

- Prefer `to_username` mode for normal demo flows.
- Show raw `to_email` mode only when the health response exposes `allow_raw_email=true`.
- Disable the test form when `test_endpoint_enabled=false`.
- Warn that this is a test/demo operation only.
- The frontend should not send SMTP credentials and should not imply that the email was delivered immediately. Delivery happens later through the backend worker.

## Retry Action

The retry action calls `POST /api/v1/notifications/email-outbox/{outbox_id}/retry`.

Placement:

- Email Outbox List row action.
- Email Outbox Detail primary/secondary action.

Visibility:

- Show only for admin/operator users with `permission.notification.retry` or `permission.notification.edit`.
- Show only when the outbox status is `FAILED` or `DEAD`.
- Hide for `PENDING`, `SENDING`, `SENT`, and `CANCELLED`.

Behavior:

- Require confirmation before retrying.
- Confirmation copy should mention recipient username/email, event label, business reference, and current status.
- Optional reason can be collected only if the product wants audit context.
- Retry does not send synchronously. It only requeues the outbox row to `PENDING`.
- After success, refresh the affected row/detail and queue summary counts.

States:

- Loading: disable the retry button for the selected row/detail.
- Success: show requeued status and link back to outbox list if used from detail.
- Error: show permission denied, not found, not retryable status, or retryable server/network error.

## UX Rules

- Keep UI business-readable.
- Prefer username, display name, and business reference over UUID.
- UUID can appear in technical detail pages only.
- Do not expose backend package names.
- Do not expose SMTP password, provider credentials, or secret fields.
- Use confirmation before retrying failed email.
- Keep pages simple for demo.
- Do not expose or create provider selection controls.
- Do not create a direct business email form.

## Frontend Data Models

These TypeScript-like interfaces document expected API shapes. They are examples for planning and client typing, not implementation source.

```ts
type UUID = string;
type ISODateTime = string;

type NotificationSeverity = "INFO" | "WARNING" | "CRITICAL" | string;
type EmailOutboxStatus = "PENDING" | "SENDING" | "SENT" | "FAILED" | "DEAD" | "CANCELLED";

interface UserSummary {
  id: UUID | null;
  username: string;
  display_name: string;
  email: string;
}

interface NotificationEvent {
  type: string;
  label: string;
  category: string;
  severity: NotificationSeverity;
}

interface BusinessContext {
  business_type: string;
  business_label: string;
  business_id: UUID | null;
  business_reference: string;
  business_title: string;
}

interface Action {
  label: string;
  url: string;
}

interface NotificationListItem {
  notification_id: UUID;
  recipient: UserSummary;
  event: NotificationEvent;
  title: string;
  body: string;
  context: BusinessContext;
  action: Action | null;
  is_read: boolean;
  read_at: ISODateTime | null;
  created_at: ISODateTime;
}

interface EmailHealth {
  enabled: boolean;
  worker_enabled: boolean;
  smtp_host: string;
  smtp_port: number;
  smtp_tls_mode: string;
  from_address: string;
  from_name: string;
  send_real_email: boolean;
  allow_raw_email: boolean;
  test_endpoint_enabled: boolean;
  worker_interval: string;
  worker_batch_size: number;
  stale_sending_timeout: string;
  retry_policy: string[];
  pending_count: number;
  failed_count: number;
  dead_count: number;
}

interface EmailOutboxListItem {
  outbox_id: UUID;
  notification_id: UUID | null;
  recipient: UserSummary;
  to_email: string;
  to_name: string;
  event: NotificationEvent;
  context: BusinessContext;
  subject: string;
  status: EmailOutboxStatus;
  attempts: number;
  max_attempts: number;
  next_attempt_at: ISODateTime | null;
  last_error: string | null;
  provider_message_id: string | null;
  created_at: ISODateTime;
  sent_at: ISODateTime | null;
}

interface EmailOutboxDetail extends EmailOutboxListItem {
  body_text: string;
  body_html: string | null;
  updated_at: ISODateTime;
}

interface TestEmailRequest {
  to_username?: string;
  to_email?: string;
  subject: string;
  body: string;
}

interface TestEmailResponse {
  outbox_id: UUID;
  recipient: UserSummary;
  event: NotificationEvent;
  status: "QUEUED";
}
```

## State Management Recommendation

Use the generated frontend API client as the only source for Notification API calls.

Composable/service functions should be thin wrappers around generated client methods. They may normalize loading/error state for Vue components, but they should not duplicate endpoint paths, request models, or response models by hand.

Use a store or composable only if the project already uses that pattern. Do not add a new global state framework only for notifications.

Suggested composables/services:

- `useNotifications()`
- `useEmailHealth()`
- `useEmailOutbox()`
- `useTestEmail()`

Keep admin email state separate from user notification state. Notification bell unread count should not depend on admin outbox screens.

Generated client rules:

- Regenerate the frontend API client after Swagger/OpenAPI changes.
- Use generated request/response types when available.
- Add local TypeScript interfaces only as UI-facing aliases or temporary documentation until generated types exist.
- Do not create parallel manual API services with raw URL strings for notification endpoints.

## Error Handling

| Status/error | Frontend behavior |
| --- | --- |
| 401 | Use existing redirect/login handling. |
| 403 | Hide unavailable admin routes/menu items, or show permission denied when a guarded route is opened directly. |
| 404 | Show not found for missing notification/outbox detail. |
| 422 | Show validation errors near filters/forms. |
| 500 | Show retryable server error. |
| Network error | Show retry action and preserve current form/filter state. |

## Demo Flow

1. Login as admin/operator.
2. Open Email Operations Dashboard.
3. Confirm health config.
4. Send test email.
5. Open Email Outbox List.
6. Confirm row moves from `PENDING`/`SENDING` to `SENT`.
7. Submit approval from existing approval UI.
8. Confirm Notification Bell unread count increases.
9. Open Notification Center.
10. Open approval notification action URL.
11. Mark notifications as read.

## Out of Scope

- frontend SMTP credential editing
- frontend provider selection
- creating direct send-email API
- stop-loss/risk page implementation
- production alert escalation UI
- email template editor
- websocket/SSE realtime notification unless already existing

## Backend Dependency

Frontend implementation depends on:

- backend routes implemented
- Swagger/OpenAPI regenerated
- frontend API client regenerated and committed if the repository tracks generated clients
- permissions seeded
- demo users have email in `iam_users`
- Mailpit/Gmail SMTP configured on backend

## Files To Create Later

These are suggested frontend files only. Adapt paths to the existing Nuxt/Vue repository structure.

Nuxt route file naming:

- Use `index.vue` for route directory indexes.
- Use `[outbox_id].vue` for the dynamic outbox detail route if the project keeps snake_case route params.
- If the existing project standard is camelCase params, `[outboxId].vue` is acceptable, but map it explicitly to backend `outbox_id`.
- If the detail opens as a drawer instead of a route page, keep the list route as the owner and create a drawer component under the feature/components area.

| Future file area | Purpose |
| --- | --- |
| `pages/notifications/index.vue` | Notification center page. |
| `pages/admin/notifications/email/index.vue` | Email admin dashboard page. |
| `pages/admin/notifications/email/outbox/index.vue` | Email outbox list page. |
| `pages/admin/notifications/email/outbox/[outbox_id].vue` | Email outbox detail page when using route-based detail. |
| `pages/admin/notifications/email/test.vue` | Test email route if not embedded in the dashboard. |
| `components/notifications/TestEmailPanel.vue` | Test email component. |
| `components/notifications/NotificationBell.vue` | Notification bell/dropdown component. |
| `components/notifications/EmailOutboxDetailDrawer.vue` | Email outbox detail drawer when detail is embedded in the list page. |
| `composables/useNotifications.ts` | Thin wrapper around generated client methods for user notification API calls. |
| `composables/useEmailHealth.ts` | Thin wrapper around generated client methods for email health API calls. |
| `composables/useEmailOutbox.ts` | Thin wrapper around generated client methods for outbox list/detail/retry API calls. |
| `types/notification.ts` | UI aliases only if generated types are missing or too backend-shaped for component props. |

## Implementation-Ready Checklist

- [ ] routes
- [ ] permission guards
- [ ] generated API client methods
- [ ] no manual raw endpoint strings for notification API calls
- [ ] types
- [ ] notification bell
- [ ] notification center
- [ ] email health panel
- [ ] email outbox list
- [ ] email outbox detail
- [ ] retry action
- [ ] test email form
- [ ] loading/error/empty states
- [ ] demo verification

## Final Recommendation

Implement the frontend notification center first.

Then implement the admin email operations page.

Then implement outbox detail, retry, and test email.

Do not implement stop-loss/risk frontend yet.

Do not create frontend direct email sending logic.
