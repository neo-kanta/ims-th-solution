# Notification Email API Documentation

Status: draft proposal, revised after API review  
Scope: notification center, email outbox, SMTP delivery demo, and future production provider swap  
Audience: backend developers, frontend developers, demo operators, product owners

## Header Name

Notification Email API

This document proposes the API and database design for the IMS notification email feature.

The API must be readable for business users and operators. UUIDs may remain in responses for technical traceability, but every response that references a user must also include username, display name, and email from `iam_users`.

Public API responses must not expose backend package names or implementation-only field names. Instead, responses use:

- `recipient`: the user who receives the notification.
- `event`: a backend-configured notification event type with a human label.
- `context`: the business object or business reference that caused the notification.
- `action`: the user-facing action label and URL.

## Design Principles

- Business logic must not depend on Gmail, Mailpit, SendGrid, OCI Email Delivery, or any provider-specific concept.
- SMTP configuration is provider-independent and loaded from environment variables.
- Email send requests should enqueue outbox rows first. The background worker sends later.
- API responses should contain usernames and business labels, not UUID-only data.
- Notification event types should be backend enum/config values, not free-form strings from controllers.
- Notification history should preserve recipient, event, and business context snapshots at the time the notification was created.
- Demo should work with Mailpit at zero cost.
- Real demo email should work by changing SMTP environment variables to Gmail SMTP.

## Business-Triggered Notification Flow

Email notifications are triggered by business workflows and processed through the notification outbox. Frontend applications call business APIs, while backend services create notification intents and enqueue email delivery records.

```text
business action
-> internal notification service
-> in-app notification
-> email outbox row
-> background worker
-> SMTP sender
-> mark SENT / FAILED / DEAD
```

Examples of business actions that may trigger email notification:

- Approval request submitted.
- Approval request approved.
- Approval request rejected.
- Workflow stuck-day job detects a delayed workflow day.
- Alert or risk threshold is breached.
- Permission is granted to a user.

Frontend responsibilities:

- Frontend calls only the business API, such as submit approval, approve request, reject request, or update permission.
- Frontend does not send email directly.
- Frontend does not know whether the email provider is Mailpit, Gmail SMTP, company SMTP, OCI Email Delivery, or SendGrid SMTP.
- Frontend can optionally show notification history, outbox status, or a demo test screen for admin/operator users.

Backend responsibilities:

- Business packages emit provider-independent notification intent.
- Notification package creates the in-app notification and email outbox row.
- Background worker sends the email asynchronously through provider-independent SMTP configuration.
- Public email APIs exist only for admin/demo/operator use: test email, health, outbox list/detail, and retry.

The test email endpoint is a demo and operations tool for verifying SMTP configuration and worker delivery. Business workflow notifications are created by backend services through internal notification intent.

## Backend Event Type Config

Notification event type should be a backend-owned enum, then rendered through a small event catalog.

Recommended implementation files:

- `backend/pkg/enum/notification_event.go`
- `backend/internal/notification/domain/valueobject/event_catalog.go`

Reason:

- `backend/pkg/enum` already stores shared workflow, investment, and approval enums.
- Approval, workflow, permission, alert, and future risk packages can depend on stable enum constants without depending on notification transport code.
- The notification package still owns the labels, categories, severity, default action labels, and email template mapping.

Example enum:

```go
package enum

type NotificationEventType string

const (
	NotificationEventApprovalTaskAssigned  NotificationEventType = "APPROVAL_TASK_ASSIGNED"
	NotificationEventApprovalCompleted     NotificationEventType = "APPROVAL_COMPLETED"
	NotificationEventApprovalRejected      NotificationEventType = "APPROVAL_REJECTED"
	NotificationEventWorkflowStuckDay      NotificationEventType = "WORKFLOW_STUCK_DAY"
	NotificationEventAlertThresholdBreached NotificationEventType = "ALERT_THRESHOLD_BREACHED"
	NotificationEventPermissionGranted     NotificationEventType = "PERMISSION_GRANTED"
	NotificationEventEmailTest             NotificationEventType = "EMAIL_TEST"
)
```

Example event catalog:

```go
type EventDefinition struct {
	Type               enum.NotificationEventType
	Label              string
	Category           string
	Severity           string
	DefaultActionLabel  string
	DefaultTemplateName string
}
```

Recommended first catalog:

| Event type | User-facing label | Category | Severity | Example user message |
| --- | --- | --- | --- | --- |
| `APPROVAL_TASK_ASSIGNED` | Approval task assigned | `APPROVAL` | `INFO` | Approval request APR-000123 needs your action. |
| `APPROVAL_COMPLETED` | Approval completed | `APPROVAL` | `INFO` | Approval request APR-000123 was approved. |
| `APPROVAL_REJECTED` | Approval rejected | `APPROVAL` | `WARNING` | Approval request APR-000123 was rejected. |
| `WORKFLOW_STUCK_DAY` | Workflow day needs attention | `WORKFLOW` | `WARNING` | Workflow day 2026-06-19 has not moved for 24 hours. |
| `ALERT_THRESHOLD_BREACHED` | Alert threshold breached | `ALERT` | `CRITICAL` | Portfolio risk threshold was breached. |
| `PERMISSION_GRANTED` | Permission granted | `PERMISSION` | `INFO` | You were granted permission by somchai.admin. |
| `EMAIL_TEST` | Email test | `SYSTEM` | `INFO` | IMS email test message. |

Remark:

The database can store `event_type` as `VARCHAR(80)` for portability, but the backend must validate it through this enum/catalog before creating notifications or email outbox rows.

## Internal Notification Service Contract

These are internal application service methods, not public HTTP APIs.

Business packages should call an internal notification service after the business transaction decides that a notification is required. The service should accept provider-independent intent and should not accept Gmail, Mailpit, SendGrid, OCI, or SMTP-specific fields.

Example internal interface:

```go
type NotificationService interface {
	NotifyApprovalTaskAssigned(ctx context.Context, input ApprovalTaskAssignedInput) error
	NotifyApprovalCompleted(ctx context.Context, input ApprovalCompletedInput) error
	NotifyApprovalRejected(ctx context.Context, input ApprovalRejectedInput) error
	NotifyWorkflowStuckDay(ctx context.Context, input WorkflowStuckDayInput) error
	NotifyAlertThresholdBreached(ctx context.Context, input AlertThresholdBreachedInput) error
	NotifyPermissionGranted(ctx context.Context, input PermissionGrantedInput) error
	SendTestEmailIntent(ctx context.Context, input TestEmailIntentInput) error
}
```

Common input fields:

| Field | Purpose |
| --- | --- |
| `recipient_user_id` | Primary recipient from IAM. Required for business notifications. |
| `recipient_username` | Optional convenience when the caller already has it; service should still resolve IAM data when needed. |
| `actor_user_id` | User who performed the business action, when available. |
| `event_type` | Backend enum such as `APPROVAL_TASK_ASSIGNED`. |
| `business_type` | Business object type such as `APPROVAL_REQUEST`, `WORKFLOW_DAY`, `RISK_ALERT`, `PERMISSION_GRANT`. |
| `business_id` | Internal business UUID for traceability. |
| `business_reference` | Human-readable business reference such as `APR-000123`. |
| `business_title` | Human-readable title or summary. |
| `action_url` | Frontend route where the recipient can act or inspect. |
| `idempotency_key` | Stable unique key to prevent duplicate notification/email creation. |

Service responsibilities:

1. Resolve recipient username, display name, and email from `iam_users`.
2. Validate `event_type` against the notification event catalog.
3. Create or skip duplicate in-app notification based on idempotency.
4. Create or skip duplicate email outbox row based on idempotency.
5. Never send SMTP synchronously from business packages.
6. Leave actual email delivery to the background worker.

Remark:

`SendTestEmailIntent` exists for the test email endpoint and demo operations only. It is not used by approval, workflow, investment, permission, alert, or risk business flows.

## Business Trigger Matrix

| Event type | Trigger source package | Trigger condition | Recipient | Email required? | In-app notification required? | Business context fields | Action URL |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `APPROVAL_TASK_ASSIGNED` | Approval | Approval submitted or moved to a stage that requires action. | Next assigned approver, delegated approver, or active eligible approver group. | yes | yes | `business_type=APPROVAL_REQUEST`, approval request ID, approval reference, approval title, stage/status. | `/approval/requests/{approval_request_id}` |
| `APPROVAL_COMPLETED` | Approval | Final approval step is approved and request becomes approved/completed. | Submitter and optional manager. | yes | yes | `business_type=APPROVAL_REQUEST`, approval request ID, approval reference, approval title, final status. | `/approval/requests/{approval_request_id}` |
| `APPROVAL_REJECTED` | Approval | Any approver rejects the request. | Submitter and optional manager. | yes | yes | `business_type=APPROVAL_REQUEST`, approval request ID, approval reference, approval title, rejection status/reason. | `/approval/requests/{approval_request_id}` |
| `WORKFLOW_STUCK_DAY` | Workflow job | Scheduled job detects a workflow day that has not progressed beyond the configured threshold. | Workflow operators, responsible manager, or configured support users. | yes | yes | `business_type=WORKFLOW_DAY`, workflow day ID/date, stuck state, elapsed time. | `/workflow/days/{workflow_day_id}` |
| `ALERT_THRESHOLD_BREACHED` | Alert/risk | Portfolio, risk, compliance, or alert rule crosses configured threshold. | Configured alert recipients, portfolio owner, risk owner, or compliance operator. | yes | yes | `business_type=RISK_ALERT`, alert ID, threshold name, measured value, threshold value. | `/alerts/{alert_id}` |
| `PERMISSION_GRANTED` | Permission/IAM | User is granted a permission, role, or access scope. | User who received the permission. | optional | yes | `business_type=PERMISSION_GRANT`, grant ID, permission key, granted-by username. | `/settings/permissions` |
| `EMAIL_TEST` | Notification admin/demo | Admin/operator submits test email endpoint. | Selected username or allowed raw email. | yes | no by default | `business_type=EMAIL_TEST`, optional request ID, subject. | none |

Rules:

- Business notifications should normally create both in-app notification and email outbox row.
- `EMAIL_TEST` normally creates only an outbox row, because it validates delivery rather than representing business work.
- If email is disabled by config, the service may still create in-app notifications and skip outbox creation.
- If an event does not require email, do not create an outbox row.
- The action URL must point to a business screen, not a provider or SMTP screen.

## Approval Notification Rules

Approval events should be generated by the approval package or its existing notification adapter after the approval transaction succeeds.

Rules:

1. When an approval request is submitted, notify the next approver.
2. When an approval request moves to a new approval stage, notify the approver or approver group for that stage.
3. When approval completes, notify the submitter and optionally the manager or owner configured by the approval workflow.
4. When approval is rejected, notify the submitter and optionally the manager or owner configured by the approval workflow.
5. If group approval exists, send to all active eligible approvers or follow the existing approval engine assignment result.
6. If delegation or agent approval applies, notify the delegated approver and preserve original approver context in the email body and outbox metadata.
7. Notification content must include business reference, title, current status, actor when available, and action URL.
8. Approval notifications must use stable idempotency keys so double-clicks, retries, or repeated event emission do not send duplicate emails.

Minimum approval context:

| Field | Example | Purpose |
| --- | --- | --- |
| `business_type` | `APPROVAL_REQUEST` | Allows filtering outbox by business object type. |
| `business_id` | `c996c25e-6b41-47b2-b1c5-7e4f2caa1111` | Technical traceability. |
| `business_reference` | `APR-000123` | Human-readable reference. |
| `business_title` | `Thai Equity Market Outlook` | User-facing title. |
| `status` | `PENDING_APPROVAL` | Current approval status. |
| `stage` | `MANAGER_REVIEW` | Current workflow stage when available. |
| `actor_username` | `somchai.manager` | Who submitted, approved, rejected, or delegated. |
| `original_approver_username` | `ben.approver` | Preserves original approver if delegation is used. |
| `delegated_approver_username` | `nina.delegate` | Actual delegated recipient if delegation is used. |
| `action_url` | `/approval/requests/{approval_request_id}` | Where recipient can act. |

## Shared Models

### User Summary Model

Used anywhere a response references a user.

```json
{
  "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
  "username": "ben.approver",
  "display_name": "Ben Approver",
  "email": "ben@example.com"
}
```

Field notes:

| Field | Description |
| --- | --- |
| `id` | Internal IAM UUID. Included for traceability, not as the main label. |
| `username` | Stable human-readable login identity. |
| `display_name` | Name shown to operators and business users. |
| `email` | Delivery email address. |

### Notification Event Model

Used by in-app notifications and email outbox APIs.

```json
{
  "type": "APPROVAL_TASK_ASSIGNED",
  "label": "Approval task assigned",
  "category": "APPROVAL",
  "severity": "INFO"
}
```

Field notes:

| Field | Description |
| --- | --- |
| `type` | Backend enum from `backend/pkg/enum/notification_event.go`. |
| `label` | User-facing label from the notification event catalog. |
| `category` | User-facing grouping such as `APPROVAL`, `WORKFLOW`, `ALERT`, `PERMISSION`, `SYSTEM`. |
| `severity` | User-facing severity such as `INFO`, `WARNING`, `CRITICAL`. |

### Business Context Model

Used to describe what business object caused the notification.

```json
{
  "business_type": "APPROVAL_REQUEST",
  "business_label": "Approval request",
  "business_id": "c996c25e-6b41-47b2-b1c5-7e4f2caa1111",
  "business_reference": "APR-000123",
  "business_title": "Thai Equity Market Outlook"
}
```

Field notes:

| Field | Description |
| --- | --- |
| `business_type` | Backend enum-like business object type, for example `APPROVAL_REQUEST`, `WORKFLOW_DAY`, `RISK_ALERT`, `PERMISSION_GRANT`. |
| `business_label` | User-facing name of the business object. |
| `business_id` | Internal UUID for traceability. It is not the display label. |
| `business_reference` | Human-readable business reference, for example `APR-000123`. |
| `business_title` | Human-readable title or summary. |

### Action Model

Used when the notification can navigate the user to a screen.

```json
{
  "label": "Open approval request",
  "url": "/approval/requests/c996c25e-6b41-47b2-b1c5-7e4f2caa1111"
}
```

## API Endpoint 1: List My In-App Notifications

```http
GET /api/v1/notifications?unread_only=true&limit=20&offset=0
```

### Purpose

Returns the authenticated user's in-app notification center list.

### Authentication And Permission

- Requires authenticated user.
- Uses current JWT subject only as the filtering key.
- No admin permission is required because users only read their own notifications.

### Logic And Process

1. Backend reads `current_user_id` from auth middleware claims.
2. Backend queries `notification__notifications` where `recipient_user_id` equals `current_user_id`.
3. Backend joins `iam_users` by `notification__notifications.recipient_user_id = iam_users.id`.
4. If `unread_only=true`, backend filters unread notifications.
5. Backend orders results by `created_at desc`.
6. Backend maps the stored `event_type` through the backend event catalog to get user-facing `label`, `category`, and `severity`.
7. Backend returns notification rows with `recipient`, `event`, `context`, and `action`.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Current user filter | JWT claims from auth middleware. |
| Recipient username/display/email | `iam_users`, joined by `notification__notifications.recipient_user_id`. |
| Notification title/body/read state | `notification__notifications`. |
| Event type | `notification__notifications.event_type` or current `category` during migration. |
| Event label/category/severity | Backend notification event catalog. |
| Business context | `notification__notifications.business_type`, `business_id`, `business_reference`, `business_title`. |
| Action label/URL | `notification__notifications.action_label`, `action_url`; current implementation may map existing `link`. |
| Total and unread counts | Aggregate count from `notification__notifications`. |

### Request Model

Query parameters:

| Name | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `unread_only` | boolean | no | `false` | Return only unread notifications. |
| `limit` | integer | no | `50` | Page size. Recommended max: `200`. |
| `offset` | integer | no | `0` | Offset for paging. |

Example:

```http
GET /api/v1/notifications?unread_only=true&limit=20&offset=0
```

### Response Model

```json
{
  "items": [
    {
      "notification_id": "9f7b2a7e-2c2e-4a21-b9c3-4a0a6c1c1234",
      "recipient": {
        "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
        "username": "ben.approver",
        "display_name": "Ben Approver",
        "email": "ben@example.com"
      },
      "event": {
        "type": "APPROVAL_TASK_ASSIGNED",
        "label": "Approval task assigned",
        "category": "APPROVAL",
        "severity": "INFO"
      },
      "title": "Approval request APR-000123 needs your action",
      "body": "Research report: Thai Equity Market Outlook",
      "context": {
        "business_type": "APPROVAL_REQUEST",
        "business_label": "Approval request",
        "business_id": "c996c25e-6b41-47b2-b1c5-7e4f2caa1111",
        "business_reference": "APR-000123",
        "business_title": "Thai Equity Market Outlook"
      },
      "action": {
        "label": "Open approval request",
        "url": "/approval/requests/c996c25e-6b41-47b2-b1c5-7e4f2caa1111"
      },
      "is_read": false,
      "read_at": null,
      "created_at": "2026-06-19T08:12:00Z"
    }
  ],
  "total": 1,
  "unread": 1
}
```

### Remark

The response intentionally uses `notification_id`, not a vague top-level `id`. The current user is shown in `recipient`, with username loaded from `iam_users`.

If the current implementation still returns `id`, keep it only as a temporary backward-compatible alias and prefer `notification_id` in new frontend code.

### Dependencies

- `backend/internal/notification`
- `backend/pkg/enum/notification_event.go`
- Notification event catalog
- `notification__notifications`
- `iam_users`
- IAM auth middleware
- JWT token claims

## API Endpoint 2: Mark One In-App Notification As Read

```http
POST /api/v1/notifications/{notification_id}/read
```

### Purpose

Marks one notification as read for the authenticated user.

### Authentication And Permission

- Requires authenticated user.
- User can only mark their own notification as read.

### Logic And Process

1. Backend reads `current_user_id` from JWT claims.
2. Backend parses `{notification_id}` from the URL.
3. Backend updates `notification__notifications` where `id` equals `{notification_id}` and `recipient_user_id` equals `current_user_id`.
4. Backend sets `is_read=true` and `read_at=now()`.
5. Backend joins or reads `iam_users` for the response recipient summary.
6. If the row is already read, response should still be successful.
7. If the row belongs to another user, response should avoid leaking ownership. Treat as not found or soft success based on current handler behavior.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Notification ID | URL path `{notification_id}`. |
| Current user filter | JWT claims from auth middleware. |
| Recipient username/display/email | `iam_users`. |
| Read timestamp | Server time in UTC. |
| Updated row | `notification__notifications`. |

### Request Model

Path parameters:

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `notification_id` | UUID | yes | Notification record ID. |

Body: none.

### Response Model

```json
{
  "notification_id": "9f7b2a7e-2c2e-4a21-b9c3-4a0a6c1c1234",
  "recipient": {
    "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
    "username": "ben.approver",
    "display_name": "Ben Approver",
    "email": "ben@example.com"
  },
  "status": "READ",
  "read_at": "2026-06-19T08:20:00Z"
}
```

### Remark

This API should not accept `recipient_user_id` in the request body. The recipient must always come from the authenticated session and be displayed with username in the response.

### Dependencies

- `notification__notifications`
- `iam_users`
- IAM auth middleware

## API Endpoint 3: Mark All My In-App Notifications As Read

```http
POST /api/v1/notifications/read-all
```

### Purpose

Marks every unread notification for the authenticated user as read.

### Authentication And Permission

- Requires authenticated user.
- User can only update their own notifications.

### Logic And Process

1. Backend reads `current_user_id` from JWT claims.
2. Backend updates all rows in `notification__notifications` for the current user where `is_read=false`.
3. Backend sets `is_read=true` and `read_at=now()`.
4. Backend reads `iam_users` for the response recipient summary.
5. Backend returns the number of updated rows.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Current user filter | JWT claims from auth middleware. |
| Recipient username/display/email | `iam_users`. |
| Notification rows | `notification__notifications`. |
| Updated count | Database command tag rows affected. |

### Request Model

Body: none.

### Response Model

```json
{
  "recipient": {
    "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
    "username": "ben.approver",
    "display_name": "Ben Approver",
    "email": "ben@example.com"
  },
  "updated": 5
}
```

### Remark

This is a user convenience endpoint. It should not create audit noise per row.

### Dependencies

- `notification__notifications`
- `iam_users`
- IAM auth middleware

## API Endpoint 4: List Email Outbox

```http
GET /api/v1/notifications/email-outbox?status=FAILED&recipient_username=ben.approver&event_type=APPROVAL_TASK_ASSIGNED&limit=50&offset=0
```

### Purpose

Returns email delivery records for operators, admins, and demos.

This is not a mailbox. It is an operational delivery log showing what IMS attempted to send, who it was sent to, the business reason, and the delivery status.

### Authentication And Permission

- Requires authenticated user.
- Recommended permission: `NOTIFICATION_CONFIG`.
- If integrating with the older permission workflow UI, map read access to `permission.notification.view`.

### Logic And Process

1. Backend validates operator permission.
2. Backend builds filters from query parameters.
3. Backend queries `notification__email_outbox`.
4. Backend should prefer outbox snapshot fields for historical readability.
5. Backend may join `iam_users` only to show the current user profile beside the snapshot if needed.
6. Backend maps `event_type` through the notification event catalog.
7. Backend returns a paged list with recipient, event, context, delivery status, attempts, next retry time, last error, and timestamps.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Recipient username/display/email | Snapshot columns in `notification__email_outbox`; optional current `iam_users` join. |
| Event type | `notification__email_outbox.event_type`. |
| Event label/category/severity | Backend notification event catalog, with snapshot fallback columns if stored. |
| Business context | `notification__email_outbox.business_type`, `business_reference`, `business_title`. |
| Delivery status | `notification__email_outbox.status`. |
| Attempt counts | `attempts`, `max_attempts`. |
| Retry schedule | `next_attempt_at`. |
| Error details | `last_error`. |
| SMTP provider message ID | `provider_message_id`. |

### Request Model

Query parameters:

| Name | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `status` | string | no | none | Filter by `PENDING`, `SENDING`, `SENT`, `FAILED`, `DEAD`, `CANCELLED`. |
| `recipient_username` | string | no | none | Filter by readable username. |
| `recipient_email` | string | no | none | Filter by target email address. |
| `event_type` | string | no | none | Filter by backend event enum, for example `APPROVAL_TASK_ASSIGNED`. |
| `event_category` | string | no | none | Filter by user-facing group, for example `APPROVAL` or `ALERT`. |
| `business_type` | string | no | none | Filter by business object type, for example `APPROVAL_REQUEST`. |
| `business_reference` | string | no | none | Filter by business reference, for example `APR-000123`. |
| `created_from` | string/date-time | no | none | Filter created timestamp greater than or equal. |
| `created_to` | string/date-time | no | none | Filter created timestamp less than or equal. |
| `limit` | integer | no | `50` | Page size. Recommended max: `200`. |
| `offset` | integer | no | `0` | Offset for paging. |

Example:

```http
GET /api/v1/notifications/email-outbox?status=FAILED&recipient_username=ben.approver&event_type=APPROVAL_TASK_ASSIGNED&limit=50&offset=0
```

### Response Model

```json
{
  "items": [
    {
      "outbox_id": "2a83e30a-2127-4527-8fe5-1a48c4781111",
      "notification_id": "9f7b2a7e-2c2e-4a21-b9c3-4a0a6c1c1234",
      "recipient": {
        "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
        "username": "ben.approver",
        "display_name": "Ben Approver",
        "email": "ben@example.com"
      },
      "to_email": "ben@example.com",
      "to_name": "Ben Approver",
      "event": {
        "type": "APPROVAL_TASK_ASSIGNED",
        "label": "Approval task assigned",
        "category": "APPROVAL",
        "severity": "INFO"
      },
      "context": {
        "business_type": "APPROVAL_REQUEST",
        "business_label": "Approval request",
        "business_id": "c996c25e-6b41-47b2-b1c5-7e4f2caa1111",
        "business_reference": "APR-000123",
        "business_title": "Thai Equity Market Outlook"
      },
      "subject": "Approval request APR-000123 needs your action",
      "status": "FAILED",
      "attempts": 2,
      "max_attempts": 5,
      "next_attempt_at": "2026-06-19T08:17:00Z",
      "last_error": "smtp: authentication failed",
      "provider_message_id": null,
      "created_at": "2026-06-19T08:12:00Z",
      "sent_at": null
    }
  ],
  "total": 1
}
```

### Remark

The list response should not include the full email body by default. Body content can be long and may contain sensitive business context. Use the detail endpoint for body inspection.

### Dependencies

- `notification__email_outbox`
- `iam_users` for optional current user enrichment
- Notification event catalog
- Notification permission catalog
- IAM auth and permission middleware

## API Endpoint 5: Get Email Outbox Detail

```http
GET /api/v1/notifications/email-outbox/{outbox_id}
```

### Purpose

Returns one email outbox record with full email body and delivery details.

### Authentication And Permission

- Requires authenticated user.
- Recommended permission: `NOTIFICATION_CONFIG`.
- If integrating with the older permission workflow UI, map read access to `permission.notification.view`.

### Logic And Process

1. Backend validates operator permission.
2. Backend parses `{outbox_id}` from the URL.
3. Backend reads the outbox row from `notification__email_outbox`.
4. Backend maps `event_type` through the notification event catalog.
5. Backend returns recipient snapshot, event, business context, full email body, provider message ID, delivery status, and timestamps.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Recipient object | Snapshot columns in `notification__email_outbox`. |
| Event object | `event_type` plus backend event catalog. |
| Business context | `business_type`, `business_id`, `business_reference`, `business_title`. |
| Email subject/body | `subject`, `body_text`, `body_html`. |
| Delivery status | `status`, `attempts`, `max_attempts`, `next_attempt_at`, `last_error`. |

### Request Model

Path parameters:

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `outbox_id` | UUID | yes | Email outbox record ID. |

Body: none.

### Response Model

```json
{
  "outbox_id": "2a83e30a-2127-4527-8fe5-1a48c4781111",
  "notification_id": "9f7b2a7e-2c2e-4a21-b9c3-4a0a6c1c1234",
  "recipient": {
    "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
    "username": "ben.approver",
    "display_name": "Ben Approver",
    "email": "ben@example.com"
  },
  "to_email": "ben@example.com",
  "to_name": "Ben Approver",
  "event": {
    "type": "APPROVAL_TASK_ASSIGNED",
    "label": "Approval task assigned",
    "category": "APPROVAL",
    "severity": "INFO"
  },
  "context": {
    "business_type": "APPROVAL_REQUEST",
    "business_label": "Approval request",
    "business_id": "c996c25e-6b41-47b2-b1c5-7e4f2caa1111",
    "business_reference": "APR-000123",
    "business_title": "Thai Equity Market Outlook"
  },
  "subject": "Approval request APR-000123 needs your action",
  "body_text": "Hello Ben Approver,\n\nApproval request APR-000123 needs your action.\n\nSubject: Thai Equity Market Outlook",
  "body_html": "<p>Hello Ben Approver,</p><p>Approval request <strong>APR-000123</strong> needs your action.</p>",
  "status": "SENT",
  "attempts": 1,
  "max_attempts": 5,
  "next_attempt_at": null,
  "last_error": null,
  "provider_message_id": "smtp-message-id",
  "created_at": "2026-06-19T08:12:00Z",
  "updated_at": "2026-06-19T08:12:30Z",
  "sent_at": "2026-06-19T08:12:30Z"
}
```

### Remark

This endpoint is for troubleshooting. It should be permission-gated and should not be used by normal business users.

### Dependencies

- `notification__email_outbox`
- Notification event catalog
- IAM auth and permission middleware

## API Endpoint 6: Retry Failed Email

```http
POST /api/v1/notifications/email-outbox/{outbox_id}/retry
```

### Purpose

Moves a failed or dead email back to the pending queue for another send attempt.

### Authentication And Permission

- Requires authenticated user.
- Recommended permission: `NOTIFICATION_CONFIG`.
- If integrating with older permission workflow UI, map write access to `permission.notification.edit`.

### Logic And Process

1. Backend validates operator permission.
2. Backend parses `{outbox_id}` from the URL.
3. Backend locks the outbox row.
4. Backend allows retry only when status is `FAILED` or `DEAD`.
5. Backend sets status to `PENDING`.
6. Backend clears `locked_at`, `locked_by`, and `last_error`.
7. Backend sets `next_attempt_at=now()`.
8. Backend does not reset `attempts` by default, so the history remains honest.
9. Worker picks up the row on its next tick.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Outbox ID | URL path `{outbox_id}`. |
| Recipient username | `notification__email_outbox.recipient_username`. |
| Event type | `notification__email_outbox.event_type`. |
| Business reference | `notification__email_outbox.business_reference`. |
| Current status/attempts | `notification__email_outbox`. |
| Operator identity | JWT claims, used for audit if implemented. |
| Retry timestamp | Server time in UTC. |

### Request Model

Path parameters:

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `outbox_id` | UUID | yes | Email outbox record ID. |

Optional body:

```json
{
  "reason": "SMTP credentials rotated; retry after fix."
}
```

### Response Model

```json
{
  "outbox_id": "2a83e30a-2127-4527-8fe5-1a48c4781111",
  "recipient_username": "ben.approver",
  "event_type": "APPROVAL_TASK_ASSIGNED",
  "business_reference": "APR-000123",
  "status": "PENDING",
  "attempts": 2,
  "max_attempts": 5,
  "next_attempt_at": "2026-06-19T08:20:00Z"
}
```

### Remark

Retry should not send the email synchronously. It should only requeue the outbox row. This preserves one delivery path for demos and production.

### Dependencies

- `notification__email_outbox`
- Email outbox worker
- IAM auth and permission middleware
- Optional audit recorder

## API Endpoint 7: Send Test Email

```http
POST /api/v1/notifications/email/test
```

### Purpose

Creates a test email outbox row for Mailpit, Gmail SMTP, or future SMTP provider verification.

This endpoint is for demo and operations. It proves the same outbox and worker path used by real approval, workflow, permission, alert, and risk notifications.

### Authentication And Permission

- Requires authenticated user.
- Recommended permission: `NOTIFICATION_CONFIG`.
- If integrating with older permission workflow UI, map write access to `permission.notification.edit`.

### Logic And Process

1. Backend validates operator permission.
2. Backend accepts either `to_username` or `to_email`.
3. If `to_username` is provided, backend resolves user from `iam_users`.
4. If `to_email` is provided without username, backend sends to the raw email and leaves `recipient_user_id` null.
5. Backend creates a `notification__email_outbox` row with event type `EMAIL_TEST`.
6. Backend returns status `QUEUED`.
7. Email worker sends the row asynchronously.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| `to_username` | Request body. |
| User display/email | `iam_users` when username is supplied. |
| `to_email` | Request body or resolved from `iam_users.email`. |
| Event type | Backend constant `EMAIL_TEST`. |
| Subject/body | Request body. |
| Outbox row | `notification__email_outbox`. |
| SMTP send result | Later worker execution, not the immediate API response. |

### Request Model

Use username:

```json
{
  "to_username": "ben.approver",
  "subject": "IMS email test",
  "body": "This is a test email from IMS."
}
```

Use raw email:

```json
{
  "to_email": "ben@example.com",
  "subject": "IMS email test",
  "body": "This is a test email from IMS."
}
```

Validation rules:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `to_username` | string | conditional | Required when `to_email` is not supplied. |
| `to_email` | string/email | conditional | Required when `to_username` is not supplied. |
| `subject` | string | yes | Email subject. Recommended max: 255 characters. |
| `body` | string | yes | Plain text body. |

### Response Model

```json
{
  "outbox_id": "2a83e30a-2127-4527-8fe5-1a48c4781111",
  "recipient": {
    "id": "0d8f4d10-2f90-4db8-85c5-6c1a1d99abcd",
    "username": "ben.approver",
    "display_name": "Ben Approver",
    "email": "ben@example.com"
  },
  "event": {
    "type": "EMAIL_TEST",
    "label": "Email test",
    "category": "SYSTEM",
    "severity": "INFO"
  },
  "status": "QUEUED"
}
```

Response when using raw email only:

```json
{
  "outbox_id": "2a83e30a-2127-4527-8fe5-1a48c4781111",
  "recipient": {
    "id": null,
    "username": "",
    "display_name": "",
    "email": "ben@example.com"
  },
  "event": {
    "type": "EMAIL_TEST",
    "label": "Email test",
    "category": "SYSTEM",
    "severity": "INFO"
  },
  "status": "QUEUED"
}
```

### Remark

This endpoint must not bypass the outbox worker. Direct synchronous SMTP send would make the demo path different from production behavior.

### Dependencies

- `iam_users`
- `backend/pkg/enum/notification_event.go`
- `notification__email_outbox`
- Email template/outbox service
- Email outbox worker
- SMTP sender
- IAM auth and permission middleware

## API Endpoint 8: Email Health

```http
GET /api/v1/notifications/email/health
```

### Purpose

Returns email configuration and queue health without exposing secrets.

### Authentication And Permission

- Requires authenticated user.
- Recommended permission: `NOTIFICATION_CONFIG`.
- If integrating with older permission workflow UI, map read access to `permission.notification.view`.

### Logic And Process

1. Backend validates operator permission.
2. Backend reads non-secret email config from `AppConfig`.
3. Backend aggregates pending, failed, and dead queue counts from `notification__email_outbox`.
4. Backend returns worker and SMTP readiness information.
5. Backend must never return SMTP password.

### Where The Information Comes From

| Information | Data origin |
| --- | --- |
| Enabled flags | Environment variables loaded into `AppConfig`. |
| SMTP host/port/TLS mode | Environment variables loaded into `AppConfig`. |
| From address/name | Environment variables loaded into `AppConfig`. |
| Queue counts | Aggregate query on `notification__email_outbox`. |
| Worker interval/batch size | Environment variables loaded into `AppConfig`. |

### Request Model

Body: none.

### Response Model

```json
{
  "enabled": true,
  "worker_enabled": true,
  "smtp_host": "mailpit",
  "smtp_port": 1025,
  "smtp_tls_mode": "none",
  "from_address": "no-reply@ims-demo.local",
  "from_name": "IMS Thailand Demo",
  "send_real_email": false,
  "test_endpoint_enabled": true,
  "worker_interval": "10s",
  "worker_batch_size": 25,
  "stale_sending_timeout": "10m",
  "retry_policy": ["1m", "5m", "15m", "1h"],
  "pending_count": 3,
  "failed_count": 1,
  "dead_count": 0
}
```

### Remark

Health response should intentionally redact:

- `SMTP_PASSWORD`
- `SMTP_USERNAME` if the username is sensitive in the target organization
- provider credentials

### Dependencies

- `backend/platform/config`
- `notification__email_outbox`
- IAM auth and permission middleware

## Permission Seed Requirements

Required permission keys:

| Permission key | Purpose |
| --- | --- |
| `permission.notification.view` | Read notification/email operational data. |
| `permission.notification.edit` | General notification administration permission. |
| `permission.notification.retry` | Retry failed/dead email outbox rows. |
| `permission.notification.test` | Send demo/test email through outbox. |
| `permission.notification.health` | Read email health/config status without secrets. |

Endpoint permission mapping:

| Endpoint | Required permission |
| --- | --- |
| `GET /api/v1/notifications/email/health` | `permission.notification.health` or `permission.notification.view` |
| `GET /api/v1/notifications/email-outbox` | `permission.notification.view` |
| `GET /api/v1/notifications/email-outbox/{outbox_id}` | `permission.notification.view` |
| `POST /api/v1/notifications/email-outbox/{outbox_id}/retry` | `permission.notification.retry` or `permission.notification.edit` |
| `POST /api/v1/notifications/email/test` | `permission.notification.test` or `permission.notification.edit` |

Remark:

User-facing notification center endpoints should remain available to any authenticated user for their own notifications. Admin/demo email endpoints must be permission-gated.

## Swagger / OpenAPI Requirements

Every HTTP endpoint in this document must have Swagger/OpenAPI annotations.

Requirements:

1. Handler functions must include request, response, security, permission, and error annotations.
2. Routes must be registered in the notification router.
3. OpenAPI output must be regenerated after backend changes.
4. Frontend API client/types must be regenerated after OpenAPI changes, if the project uses generated clients.
5. Backend must be restarted after route or config changes.
6. Feature flags must be documented in the endpoint remarks when they can hide an endpoint.

If an API is missing from Swagger, check:

- Handler not implemented.
- Route not registered.
- Swagger annotations missing or malformed.
- OpenAPI build was not regenerated.
- Generated frontend client was not refreshed.
- Backend was not restarted.
- Feature flag disabled the endpoint.

Swagger/OpenAPI should document only the supported notification HTTP endpoints: notification center, email health, test email, email outbox list/detail, and retry.

## Database Changes

### Existing In-App Notification Table Adjustment

The current `notification__notifications` table can continue to exist. For the next clean step, add user-facing event and context columns so API responses do not need to expose implementation names.

```sql
ALTER TABLE notification__notifications
    ADD COLUMN IF NOT EXISTS event_type VARCHAR(80),
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(180),
    ADD COLUMN IF NOT EXISTS business_type VARCHAR(80),
    ADD COLUMN IF NOT EXISTS business_label VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS business_id UUID,
    ADD COLUMN IF NOT EXISTS business_reference VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS business_title VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS action_label VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS action_url TEXT NOT NULL DEFAULT '';

UPDATE notification__notifications
   SET event_type = category
 WHERE event_type IS NULL OR event_type = '';

ALTER TABLE notification__notifications
    ALTER COLUMN event_type SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_notification_notifications_idempotency
ON notification__notifications (idempotency_key)
WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
```

Remark:

The backend should treat older origin fields as internal migration details only. They must not appear in the public response model.

### Table: `notification__email_outbox`

Purpose: durable email delivery queue and operational delivery log.

```sql
CREATE TABLE notification__email_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(180) NOT NULL,

    notification_id UUID REFERENCES notification__notifications(id) ON DELETE SET NULL,
    recipient_user_id UUID REFERENCES iam_users(id) ON DELETE SET NULL,

    recipient_username VARCHAR(100) NOT NULL DEFAULT '',
    recipient_display_name VARCHAR(255) NOT NULL DEFAULT '',
    recipient_email VARCHAR(255) NOT NULL DEFAULT '',

    to_email VARCHAR(255) NOT NULL,
    to_name VARCHAR(255) NOT NULL DEFAULT '',

    event_type VARCHAR(80) NOT NULL,
    event_label VARCHAR(160) NOT NULL DEFAULT '',
    event_category VARCHAR(40) NOT NULL DEFAULT '',
    event_severity VARCHAR(20) NOT NULL DEFAULT 'INFO',

    business_type VARCHAR(80),
    business_label VARCHAR(160) NOT NULL DEFAULT '',
    business_id UUID,
    business_reference VARCHAR(120) NOT NULL DEFAULT '',
    business_title VARCHAR(255) NOT NULL DEFAULT '',

    action_label VARCHAR(120) NOT NULL DEFAULT '',
    action_url TEXT NOT NULL DEFAULT '',

    subject VARCHAR(255) NOT NULL,
    body_text TEXT NOT NULL,
    body_html TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    locked_at TIMESTAMPTZ,
    locked_by VARCHAR(120),

    sent_at TIMESTAMPTZ,
    last_error TEXT,
    provider_message_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_notification_email_status CHECK (
        status IN ('PENDING', 'SENDING', 'SENT', 'FAILED', 'DEAD', 'CANCELLED')
    ),
    CONSTRAINT chk_notification_email_attempts CHECK (attempts >= 0),
    CONSTRAINT chk_notification_email_max_attempts CHECK (max_attempts > 0),
    CONSTRAINT uq_notification_email_outbox_idempotency UNIQUE (idempotency_key)
);
```

### Indexes

```sql
CREATE INDEX idx_notification_email_outbox_due
ON notification__email_outbox (status, next_attempt_at, created_at)
WHERE status IN ('PENDING', 'FAILED');

CREATE INDEX idx_notification_email_outbox_recipient
ON notification__email_outbox (recipient_username, created_at DESC);

CREATE INDEX idx_notification_email_outbox_recipient_email
ON notification__email_outbox (to_email, created_at DESC);

CREATE INDEX idx_notification_email_outbox_event
ON notification__email_outbox (event_type, created_at DESC);

CREATE INDEX idx_notification_email_outbox_event_category
ON notification__email_outbox (event_category, created_at DESC);

CREATE INDEX idx_notification_email_outbox_business
ON notification__email_outbox (business_type, business_id);

CREATE INDEX idx_notification_email_outbox_business_reference
ON notification__email_outbox (business_reference, created_at DESC);

CREATE INDEX idx_notification_email_outbox_status_created
ON notification__email_outbox (status, created_at DESC);
```

### Worker Claim Query

Worker should claim rows using `FOR UPDATE SKIP LOCKED` so multiple backend instances do not send the same email.

```sql
WITH due AS (
    SELECT id
    FROM notification__email_outbox
    WHERE status IN ('PENDING', 'FAILED')
      AND next_attempt_at <= NOW()
      AND attempts < max_attempts
    ORDER BY next_attempt_at ASC, created_at ASC
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
UPDATE notification__email_outbox e
SET status = 'SENDING',
    locked_at = NOW(),
    locked_by = $2,
    updated_at = NOW()
FROM due
WHERE e.id = due.id
RETURNING e.*;
```

### Mark Sent

```sql
UPDATE notification__email_outbox
SET status = 'SENT',
    sent_at = NOW(),
    last_error = NULL,
    provider_message_id = $2,
    locked_at = NULL,
    locked_by = NULL,
    updated_at = NOW()
WHERE id = $1;
```

### Mark Failed

```sql
UPDATE notification__email_outbox
SET status = CASE
        WHEN attempts + 1 >= max_attempts THEN 'DEAD'
        ELSE 'FAILED'
    END,
    attempts = attempts + 1,
    next_attempt_at = CASE
        WHEN attempts + 1 >= max_attempts THEN next_attempt_at
        ELSE NOW() + ($2::interval)
    END,
    last_error = $3,
    locked_at = NULL,
    locked_by = NULL,
    updated_at = NOW()
WHERE id = $1;
```

## Idempotency and Duplicate Prevention

Notification creation must be idempotent. The same business event should not create duplicate in-app notifications or duplicate emails when users double-click, workers restart, API calls retry, or events are emitted more than once.

Database requirements:

- Add `idempotency_key VARCHAR(180) UNIQUE` to `notification__email_outbox`.
- Add optional `idempotency_key VARCHAR(180)` to `notification__notifications`.
- Add a partial unique index for in-app notifications when `idempotency_key` is present.
- Generate the key in the application service before insert.

Example idempotency keys:

| Event type | Example key |
| --- | --- |
| `APPROVAL_TASK_ASSIGNED` | `APPROVAL_TASK_ASSIGNED:{approval_request_id}:{stage}:{recipient_user_id}` |
| `APPROVAL_COMPLETED` | `APPROVAL_COMPLETED:{approval_request_id}:{submitter_user_id}` |
| `APPROVAL_REJECTED` | `APPROVAL_REJECTED:{approval_request_id}:{submitter_user_id}` |
| `WORKFLOW_STUCK_DAY` | `WORKFLOW_STUCK_DAY:{workflow_day_id}:{state}:{recipient_user_id}` |
| `ALERT_THRESHOLD_BREACHED` | `ALERT_THRESHOLD_BREACHED:{alert_id}:{threshold_name}:{recipient_user_id}` |
| `PERMISSION_GRANTED` | `PERMISSION_GRANTED:{grant_id}:{recipient_user_id}` |
| `EMAIL_TEST` | `EMAIL_TEST:{request_id}:{recipient_email}` |

Insert behavior:

```sql
INSERT INTO notification__email_outbox (
    idempotency_key,
    recipient_user_id,
    recipient_username,
    recipient_display_name,
    recipient_email,
    to_email,
    event_type,
    subject,
    body_text
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING id;
```

If `RETURNING id` returns no row, the service should treat the notification/email as already queued and should not return an error to the business package.

Why this matters:

- Prevents duplicate emails when a user double-clicks submit/approve.
- Prevents duplicate emails when API retry middleware repeats a request.
- Prevents duplicate emails when the backend restarts during a transaction retry.
- Prevents duplicate emails when scheduled jobs re-detect the same workflow day or alert.
- Keeps worker retry behavior separate from notification creation behavior.

## Stale SENDING Recovery

Problem:

The worker may claim a row and mark it `SENDING`, then crash before it marks the row `SENT` or `FAILED`. Without recovery, the row can remain stuck forever.

Default stale lock timeout:

```env
NOTIFICATION_EMAIL_STALE_SENDING_TIMEOUT=10m
```

Recovery rule:

- If `status='SENDING'` and `locked_at` is older than the stale timeout, move the row back to `FAILED` or `PENDING`.
- Recommended first version: move stale rows to `FAILED`, increment attempts, set `next_attempt_at` using retry policy.
- If attempts are exhausted, mark as `DEAD`.

SQL example:

```sql
UPDATE notification__email_outbox
SET status = CASE
        WHEN attempts + 1 >= max_attempts THEN 'DEAD'
        ELSE 'FAILED'
    END,
    attempts = attempts + 1,
    next_attempt_at = CASE
        WHEN attempts + 1 >= max_attempts THEN next_attempt_at
        ELSE NOW() + $2::interval
    END,
    last_error = COALESCE(last_error, 'email worker crashed or timed out while sending'),
    locked_at = NULL,
    locked_by = NULL,
    updated_at = NOW()
WHERE status = 'SENDING'
  AND locked_at < NOW() - $1::interval
RETURNING id, status, attempts, next_attempt_at;
```

Recommended worker startup behavior:

1. Run stale `SENDING` recovery once during worker startup.
2. Run stale recovery periodically before claiming the next batch.
3. Log the number of recovered rows.
4. Do not recover rows locked more recently than the configured timeout.

## Outbox Status State Machine

Valid statuses:

| Status | Meaning |
| --- | --- |
| `PENDING` | Ready to be claimed when `next_attempt_at <= now()`. |
| `SENDING` | Claimed by a worker and currently being sent. |
| `SENT` | SMTP send succeeded. Terminal state. |
| `FAILED` | SMTP send failed but may retry. |
| `DEAD` | Max attempts exhausted. Requires manual decision or admin retry. |
| `CANCELLED` | Optional future admin state. Terminal unless explicitly requeued. |

Valid transitions:

| From | To | Trigger |
| --- | --- | --- |
| `PENDING` | `SENDING` | Worker claim query. |
| `SENDING` | `SENT` | SMTP sender succeeds. |
| `SENDING` | `FAILED` | SMTP sender fails and attempts remain. |
| `SENDING` | `DEAD` | SMTP sender fails and max attempts are exhausted. |
| `FAILED` | `PENDING` | Retry endpoint or retry scheduler requeues row. |
| `FAILED` | `DEAD` | Max attempts exhausted. |
| `PENDING` | `CANCELLED` | Optional future admin cancel. |
| `FAILED` | `CANCELLED` | Optional future admin cancel. |
| Stale `SENDING` | `FAILED` | Stale lock recovery. |

Invalid transitions:

- `SENT` should not move back to `PENDING`.
- `CANCELLED` should not be sent.
- `DEAD` should not send automatically unless an admin retry explicitly requeues it.

## Environment Variables

### Local Mailpit Defaults

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_WORKER_ENABLED=true
NOTIFICATION_EMAIL_WORKER_INTERVAL=10s
NOTIFICATION_EMAIL_WORKER_BATCH_SIZE=25
NOTIFICATION_EMAIL_MAX_ATTEMPTS=5
NOTIFICATION_EMAIL_STALE_SENDING_TIMEOUT=10m
NOTIFICATION_EMAIL_RETRY_POLICY=1m,5m,15m,1h
NOTIFICATION_EMAIL_ALLOW_RAW_EMAIL=false
NOTIFICATION_EMAIL_ALLOWED_DOMAINS=gmail.com,systemweb.com.th
NOTIFICATION_EMAIL_TEST_ENDPOINT_ENABLED=true
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=false

SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_TLS_MODE=none
SMTP_FROM_ADDRESS=no-reply@ims-demo.local
SMTP_FROM_NAME=IMS Thailand Demo
SMTP_TIMEOUT=10s

APP_PUBLIC_BASE_URL=http://localhost:3000
```

### Gmail SMTP Demo

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_TLS_MODE=starttls
SMTP_USERNAME=<gmail-address>
SMTP_PASSWORD=<gmail-app-password>
SMTP_FROM_ADDRESS=<gmail-address>
SMTP_FROM_NAME=IMS Thailand Demo
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=true
NOTIFICATION_EMAIL_ALLOW_RAW_EMAIL=false
NOTIFICATION_EMAIL_ALLOWED_DOMAINS=gmail.com,systemweb.com.th
NOTIFICATION_EMAIL_TEST_ENDPOINT_ENABLED=true
```

### Remark

Gmail password must be an app password, not the normal account password. Store it in a gitignored env file or a secret manager, not in committed config.

## Worker Configuration

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `NOTIFICATION_EMAIL_WORKER_ENABLED` | `true` | Starts or disables the email outbox worker. |
| `NOTIFICATION_EMAIL_WORKER_INTERVAL` | `10s` | How often the worker claims due rows. |
| `NOTIFICATION_EMAIL_WORKER_BATCH_SIZE` | `25` | Maximum rows claimed per tick. |
| `NOTIFICATION_EMAIL_MAX_ATTEMPTS` | `5` | Max send attempts before `DEAD`. |
| `NOTIFICATION_EMAIL_STALE_SENDING_TIMEOUT` | `10m` | Age after which `SENDING` rows are considered stale. |
| `NOTIFICATION_EMAIL_RETRY_POLICY` | `1m,5m,15m,1h` | Retry delay by failed attempt number. |

Worker behavior:

1. If `NOTIFICATION_EMAIL_WORKER_ENABLED=false`, API can still enqueue rows but no email will be sent by this process.
2. Worker must claim rows using `FOR UPDATE SKIP LOCKED`.
3. Worker must process at most `NOTIFICATION_EMAIL_WORKER_BATCH_SIZE` rows per tick.
4. Worker must run stale `SENDING` recovery before or during the claim cycle.
5. Worker must update status to `SENT`, `FAILED`, or `DEAD`; it must not leave rows in `SENDING` without `locked_at`.

## Demo Safety and Raw Email Policy

Safety variables:

| Environment variable | Local Mailpit value | Gmail demo value | Purpose |
| --- | --- | --- | --- |
| `NOTIFICATION_EMAIL_ALLOW_RAW_EMAIL` | `false` | `false` by default | Allows raw `to_email` only for test endpoint when explicitly enabled. |
| `NOTIFICATION_EMAIL_ALLOWED_DOMAINS` | `gmail.com,systemweb.com.th` | `gmail.com,systemweb.com.th` | Restricts raw/test recipients to approved domains. |
| `NOTIFICATION_EMAIL_TEST_ENDPOINT_ENABLED` | `true` | `true` | Enables `POST /api/v1/notifications/email/test`. |
| `NOTIFICATION_EMAIL_SEND_REAL_EMAIL` | `false` | `true` | Distinguishes Mailpit capture from real outbound email. |

Policy:

1. Business notifications must normally resolve recipients from `iam_users`.
2. Raw `to_email` is allowed only for the test endpoint and only when the feature flag permits it.
3. In production, raw email should usually be disabled or restricted to approved domains.
4. Email health API must never expose `SMTP_PASSWORD`.
5. Email health API should avoid exposing `SMTP_USERNAME` if the organization treats it as sensitive.
6. Provider credentials must come from environment variables or a secret manager, never from request body.
7. Mailpit local testing must cost 0 dollars and must not send real external email.

## Backend Files To Create Later

These are implementation notes only. This document update does not create backend code.

| File | Purpose |
| --- | --- |
| `backend/pkg/enum/notification_event.go` | Shared notification event type constants. |
| `backend/internal/notification/domain/valueobject/event_catalog.go` | Maps event types to labels, categories, severity, action labels, and templates. |
| `backend/internal/notification/domain/entity/email_outbox.go` | Email outbox entity. |
| `backend/internal/notification/domain/email_sender.go` | Provider-independent email sender interface. |
| `backend/internal/notification/infrastructure/email/smtp_sender.go` | SMTP implementation for Mailpit, Gmail SMTP, company SMTP, OCI SMTP, SendGrid SMTP. |
| `backend/internal/notification/infrastructure/persistence/email_outbox_repository.go` | Outbox repository and worker claim queries. |
| `backend/internal/notification/application/service/email_notification_service.go` | Creates outbox rows from notification intent. |
| `backend/internal/notification/application/service/email_template_service.go` | Renders provider-independent text/HTML email bodies. |
| `backend/internal/notification/jobs/email_outbox_worker.go` | Background send/retry worker. |
| `backend/internal/notification/transport/handler/email_outbox_handler.go` | Admin/demo email outbox endpoints. |
| Permission seed migration | Adds notification view/edit/retry/test/health permissions. |

## Backend Files To Modify Later

| File | Purpose |
| --- | --- |
| `backend/platform/config/config.go` | Add provider-independent SMTP and email worker config. |
| Notification package wiring file | Wire sender, repository, service, routes, and worker. |
| `backend/internal/notification/transport/router.go` | Register email health, test, outbox, and retry endpoints. |
| `backend/internal/notification/infrastructure/adapter/approval_notifier.go` | Emit `APPROVAL_TASK_ASSIGNED` with event/context/action fields. |
| `backend/internal/notification/infrastructure/adapter/workflow_stuck_day_notifier.go` | Emit `WORKFLOW_STUCK_DAY` with event/context/action fields. |
| `backend/internal/notification/infrastructure/persistence/notification_repository.go` | Join `iam_users` for list response and map event/context fields. |
| Swagger/OpenAPI generation files | Add annotations and regenerate API contract. |
| Docker compose files | Add Mailpit SMTP and web UI service. |

## Frontend Impact

Frontend does not need email sending logic. It only needs readable API fields.

Recommended changes only if the demo needs them:

| Area | Purpose |
| --- | --- |
| Existing notifications page | Render `recipient.username`, `event.label`, `context.business_reference`, and `action.label`. |
| Optional admin email log page | Show outbox status, recipient username, event label, business reference, attempts, and last error. |
| Optional email test page | Call `POST /api/v1/notifications/email/test` using `to_username` or `to_email`. |

## Background Worker Process

### Process Flow

```text
business action
-> business package emits notification event type and business context
-> notification package creates in-app notification
-> notification package creates email outbox row
-> worker claims due outbox rows
-> SMTP sender sends email
-> worker marks row SENT or FAILED
-> failed rows retry until max_attempts
-> exhausted rows become DEAD
```

### Retry Policy

Recommended first version:

| Attempt | Next Retry |
| --- | --- |
| 1 | 1 minute |
| 2 | 5 minutes |
| 3 | 15 minutes |
| 4 | 1 hour |
| 5 | Mark `DEAD` |

This can be hardcoded in the worker for the first implementation.

## Dependencies

### Backend Dependencies

| Dependency | Purpose |
| --- | --- |
| `backend/pkg/enum` | Shared notification event type constants. |
| `backend/internal/notification` | Owns notification APIs, outbox, templates, worker, sender interface, and event catalog. |
| `backend/internal/approval` | Emits approval notification intent. |
| `backend/internal/workflow` | Emits workflow stuck-day notification intent. |
| `backend/internal/iam` | Provides authenticated user context and recipient user data. |
| `backend/platform/config` | Loads SMTP and worker environment variables. |
| `backend/platform/database` | Provides PostgreSQL transaction helpers and connection pool. |
| `net/smtp` or small SMTP wrapper | Sends email through provider-independent SMTP. |

### Database Dependencies

| Dependency | Purpose |
| --- | --- |
| `notification__notifications` | Existing in-app notification baseline. |
| `notification__email_outbox` | New durable email outbox and delivery log. |
| `iam_users` | Recipient lookup by username, display name, and email. |

### Infrastructure Dependencies

| Dependency | Purpose |
| --- | --- |
| Mailpit | Zero-cost local SMTP capture and UI. |
| Gmail SMTP | Real demo email delivery. |
| Company SMTP / OCI / SendGrid SMTP | Future production provider options without business logic changes. |

## Minimum Demo Flow

### Local Mailpit Demo

1. Start database, backend, frontend, and Mailpit.
2. Configure SMTP to Mailpit:

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_WORKER_ENABLED=true
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=false
SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_TLS_MODE=none
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_ADDRESS=no-reply@ims-demo.local
SMTP_FROM_NAME=IMS Thailand Demo
```

3. Login as an admin/operator with `permission.notification.test` and `permission.notification.view`.
4. Call `POST /api/v1/notifications/email/test`.
5. Verify an outbox row is created with `event.type=EMAIL_TEST`.
6. Verify worker claims the row and moves status from `PENDING` to `SENDING`.
7. Open Mailpit web UI.
8. Verify the test email appears in Mailpit.
9. Verify outbox status becomes `SENT`.

### Gmail SMTP Real Demo

1. Stop backend or prepare config reload according to project behavior.
2. Configure SMTP to Gmail SMTP:

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_WORKER_ENABLED=true
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=true
NOTIFICATION_EMAIL_ALLOWED_DOMAINS=gmail.com,systemweb.com.th
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_TLS_MODE=starttls
SMTP_USERNAME=<gmail-address>
SMTP_PASSWORD=<gmail-app-password>
SMTP_FROM_ADDRESS=<gmail-address>
SMTP_FROM_NAME=IMS Thailand Demo
```

3. Restart backend.
4. Call `GET /api/v1/notifications/email/health`.
5. Confirm health shows SMTP host/port/TLS mode but does not expose password.
6. Call `POST /api/v1/notifications/email/test` to an allowed test recipient.
7. Verify outbox row is created.
8. Verify worker sends the email.
9. Verify outbox status becomes `SENT`.
10. Verify recipient receives the real email.

### Approval Auto Email Demo

1. Login as a user who can submit an approval request.
2. Submit approval through the existing approval business API or UI.
3. Do not call any email API from the frontend.
4. Backend approval flow calls the internal notification service.
5. Verify in-app notification row is created for the next approver.
6. Verify email outbox row is created with `event_type=APPROVAL_TASK_ASSIGNED`.
7. Verify recipient snapshot includes username, display name, and email from `iam_users`.
8. Verify worker sends email.
9. Verify outbox status becomes `SENT`.
10. Login as approver and verify notification center shows username-friendly, business-friendly data.

## Implementation-Ready Checklist

Use this checklist for the next coding step:

- [ ] Add DB migration for `notification__email_outbox`.
- [ ] Add DB migration for in-app notification event/context fields.
- [ ] Add optional `notification__notifications.idempotency_key` and partial unique index.
- [ ] Add `notification__email_outbox.idempotency_key VARCHAR(180) UNIQUE`.
- [ ] Add notification event enum in `backend/pkg/enum`.
- [ ] Add notification event catalog in notification domain/valueobject.
- [ ] Add internal notification service contract.
- [ ] Implement approval notification intents.
- [ ] Implement workflow stuck-day notification intent.
- [ ] Prepare alert/risk notification intent extension point.
- [ ] Implement outbox repository.
- [ ] Implement provider-independent `EmailSender` interface.
- [ ] Implement SMTP sender using provider-independent SMTP config.
- [ ] Implement email template service.
- [ ] Implement background worker claim/send/retry loop.
- [ ] Implement stale `SENDING` recovery.
- [ ] Implement duplicate prevention through idempotency keys.
- [ ] Implement email health handler.
- [ ] Implement test email handler.
- [ ] Implement email outbox list/detail handlers.
- [ ] Implement retry handler.
- [ ] Register routes in notification router.
- [ ] Add Swagger/OpenAPI annotations for all HTTP endpoints.
- [ ] Regenerate OpenAPI output and frontend client/types if used.
- [ ] Add config loading for SMTP and worker variables.
- [ ] Add Docker Compose Mailpit service.
- [ ] Add local Mailpit env example.
- [ ] Add Gmail SMTP env example without secrets committed.
- [ ] Add permission seed records.
- [ ] Add permission checks for admin/demo endpoints.
- [ ] Update notification center response to include `recipient.username` from `iam_users`.
- [ ] Add unit tests for event catalog and idempotency key generation.
- [ ] Add repository tests for outbox insert/claim/mark sent/mark failed/retry.
- [ ] Add worker tests for retry and stale `SENDING` recovery.
- [ ] Add handler tests for health/test/outbox/retry permission behavior.
- [ ] Run local Mailpit demo end to end.
- [ ] Run Gmail SMTP demo only with approved demo credentials.

## Final Remark

The API should be designed for business users and operators, not only developers.

UUIDs should remain available for traceability, but users should see usernames, event labels, business references, titles, statuses, and retry information.

The clean minimum implementation is:

1. Add notification event enum and event catalog.
2. Update the in-app notification response to join `iam_users` and return `recipient.username`.
3. Add `notification__email_outbox`.
4. Add provider-independent `EmailSender`.
5. Add `SmtpEmailSender`.
6. Add email outbox repository.
7. Add background worker.
8. Add Mailpit config.
9. Add Gmail SMTP config path.
10. Add admin/demo endpoints only if the UI or demo script needs them.
