# Audit API

## Purpose
Provides immutable audit-event query and export endpoints for system and permission-governance audit trails.

## Business Context
Audit APIs support regulatory traceability across login/session events, IAM administration, permission change governance, workflow, approval, compliance, and investment lifecycle events.

## Authentication
All audit endpoints require Bearer JWT. `/admin/audit` endpoints also inherit IAM admin IP allowlist and admin/export rate limits.

## Authorization / Permission
`/admin/audit` and `/admin/audit/export` require `IAM_AUDIT_VIEW`; `/audit/logs` requires `permission.audit.view`; `/audit/logs/export` requires `permission.audit.export`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /admin/audit | List Audit Events | JWT required | IAM_AUDIT_VIEW | Implemented |
| GET | /admin/audit/export | Export Audit Events as CSV | JWT required | IAM_AUDIT_VIEW | Implemented |
| GET | /audit/logs | List permission-management audit logs | JWT required | permission.audit.view | Implemented |
| GET | /audit/logs/export | Export permission-management audit logs as CSV | JWT required | permission.audit.export | Implemented |

## Endpoint Detail

### GET /admin/audit

#### Purpose
List Audit Events Query paginated audit events with filters

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| actor_id | string | No | Actor UUID |
| event_type | string | No | Event type (e.g., LOGIN_SUCCESS) |
| target_type | string | No | Target type (e.g., user, session) |
| target_id | string | No | Target ID |
| since | string | No | Since datetime (ISO 8601) |
| until | string | No | Until datetime (ISO 8601) |
| offset | integer | No | Offset (default 0) |
| limit | integer | No | Limit (default 50, max 500) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | AuditListResponse |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/audit/module.go; backend/internal/audit/transport/handler/audit_handler.go`.

### GET /admin/audit/export

#### Purpose
Export Audit Events as CSV Export audit events matching filter criteria as CSV download

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| actor_id | string | No | Actor UUID |
| event_type | string | No | Event type |
| since | string | No | Since datetime |
| until | string | No | Until datetime |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | file |

#### Source
`backend/internal/audit/module.go; backend/internal/audit/transport/handler/audit_handler.go`.

### GET /audit/logs

#### Purpose
List permission-management audit logs

#### Business Rule
Reads or exports audit events for operational and regulatory traceability.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Paged permission audit log object |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /audit/logs/export

#### Purpose
Export permission-management audit logs as CSV

#### Business Rule
Reads or exports audit events for operational and regulatory traceability.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | CSV export | text/csv |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

## Source References
- `backend/internal/audit/module.go`
- `backend/internal/audit/transport/handler/audit_handler.go`
- `backend/internal/permissions/module.go`
- `backend/internal/permissions/transport/handler/permission_handler.go`
- `backend/internal/iam/module.go`
- `backend/docs/swagger.json`
