# Notification API

## Purpose
Provides in-app notification-center endpoints and email outbox administration endpoints.

## Business Context
Notifications are emitted by approval, workflow stuck-day monitoring, and watchlist alert flows. Email outbox APIs let operators inspect and retry delivery records without exposing SMTP secrets.

## Authentication
All notification endpoints require Bearer JWT. Email admin routes are mounted only when the email handler and permission checker are configured.

## Authorization / Permission
User-facing notification-center routes have no function-permission middleware. Email outbox routes enforce `NOTIFICATION_VIEW`, `NOTIFICATION_RETRY`, `NOTIFICATION_TEST`, or `NOTIFICATION_HEALTH`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /notifications | List my in-app notifications | JWT required | Permission rule not found in code. | Implemented |
| GET | /notifications/email-outbox | List email outbox | JWT required | NOTIFICATION_VIEW | Implemented |
| GET | /notifications/email-outbox/{outbox_id} | Get email outbox detail | JWT required | NOTIFICATION_VIEW | Implemented |
| POST | /notifications/email-outbox/{outbox_id}/retry | Retry failed email | JWT required | NOTIFICATION_RETRY | Implemented |
| GET | /notifications/email/health | Email health | JWT required | NOTIFICATION_HEALTH | Implemented |
| POST | /notifications/email/test | Send test email | JWT required | NOTIFICATION_TEST | Implemented |
| POST | /notifications/read-all | Mark all notifications as read | JWT required | Permission rule not found in code. | Implemented |
| POST | /notifications/{id}/read | Mark one notification as read | JWT required | Permission rule not found in code. | Implemented |

## Endpoint Detail

### GET /notifications

#### Purpose
List my in-app notifications Returns the authenticated user's in-app notification center list joined with IAM user data.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| unread_only | boolean | No | Return only unread notifications |
| limit | integer | No | Page size (default 50, max 200) |
| offset | integer | No | Offset for paging |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | listResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### GET /notifications/email-outbox

#### Purpose
List email outbox Returns paged email delivery records for admin/operator use.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | string | No | Filter by status (PENDING, SENDING, SENT, FAILED, DEAD) |
| recipient_username | string | No | Filter by recipient username |
| recipient_email | string | No | Filter by recipient email |
| event_type | string | No | Filter by event type |
| event_category | string | No | Filter by event category |
| business_type | string | No | Filter by business type |
| business_reference | string | No | Filter by business reference |
| created_from | string | No | Filter created_at >= (RFC3339) |
| created_to | string | No | Filter created_at <= (RFC3339) |
| limit | integer | No | Page size (default 50) |
| offset | integer | No | Offset for paging |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | outboxListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### GET /notifications/email-outbox/{outbox_id}

#### Purpose
Get email outbox detail Returns one email outbox record with full body for troubleshooting.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| outbox_id | string | Yes | Outbox record UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | outboxDetailResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### POST /notifications/email-outbox/{outbox_id}/retry

#### Purpose
Retry failed email Moves a FAILED or DEAD outbox row back to PENDING for another send attempt.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| outbox_id | string | Yes | Outbox record UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | retryResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### GET /notifications/email/health

#### Purpose
Email health Returns email configuration and queue health without exposing secrets.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | healthResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### POST /notifications/email/test

#### Purpose
Send test email Creates a test email outbox row for SMTP verification. The background worker delivers it.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `testEmailRequest`.
```json
{
  "body": "string",
  "subject": "string",
  "to_email": "string",
  "to_username": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 202 | Accepted | testEmailResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### POST /notifications/read-all

#### Purpose
Mark all notifications as read Marks every unread in-app notification for the authenticated user as read.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | markAllReadResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

### POST /notifications/{id}/read

#### Purpose
Mark one notification as read Marks a single in-app notification as read for the authenticated user.

#### Business Rule
Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Notification ID (UUID) |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | markReadResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/notification/transport/router.go`.

## Source References
- `backend/internal/notification/module.go`
- `backend/internal/notification/transport/router.go`
- `backend/internal/notification/transport/handler/notification_handler.go`
- `backend/internal/notification/transport/handler/email_outbox_handler.go`
- `backend/internal/notification/permission/policies.go`
- `backend/docs/swagger.json`
