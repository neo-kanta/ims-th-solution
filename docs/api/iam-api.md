# IAM API

## Purpose
Provides administrative account and session operations for IAM users.

## Business Context
IAM supports the Account part of IMS permission management by maintaining user accounts, lock/disable state, password reset state, and active sessions used by JWT validation.

## Authentication
All IAM admin endpoints require Bearer JWT, active session validation, admin IP allowlist, and admin rate limiting.

## Authorization / Permission
Route-level permission checks are enforced with `IAM_USER_VIEW`, `IAM_USER_CREATE`, `IAM_USER_UPDATE`, or `IAM_USER_DEACTIVATE`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| POST | /admin/sessions/{id}/revoke | Revoke Any Session (Admin) | JWT required | IAM_USER_UPDATE | Implemented |
| GET | /admin/users | List Users | JWT required | IAM_USER_VIEW | Implemented |
| POST | /admin/users | Create User | JWT required | IAM_USER_CREATE | Implemented |
| POST | /admin/users/{id}/disable | Disable User | JWT required | IAM_USER_DEACTIVATE | Implemented |
| POST | /admin/users/{id}/enable | Enable User | JWT required | IAM_USER_DEACTIVATE | Implemented |
| POST | /admin/users/{id}/lock | Lock User | JWT required | IAM_USER_UPDATE | Implemented |
| POST | /admin/users/{id}/reset-password | Admin Password Reset | JWT required | IAM_USER_UPDATE | Implemented |
| GET | /admin/users/{id}/sessions | List User Sessions (Admin) | JWT required | IAM_USER_UPDATE | Implemented |
| POST | /admin/users/{id}/unlock | Unlock User | JWT required | IAM_USER_UPDATE | Implemented |

## Endpoint Detail

### POST /admin/sessions/{id}/revoke

#### Purpose
Revoke Any Session (Admin) Admin revoke any session by ID

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Session UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### GET /admin/users

#### Purpose
List Users Paginated user list for admin management with filtering

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
| search | string | No | Search by username, display name, or email |
| is_active | string | No | Filter by active status (true/false) |
| is_locked | string | No | Filter by locked status (true/false) |
| offset | integer | No | Offset (default 0) |
| limit | integer | No | Limit (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | AdminUserListResponse |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users

#### Purpose
Create User Administrative creation of a new IAM user

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
| None |  |  |  |

##### Request Body
Schema: `CreateUserRequest`.
```json
{
  "display_name": "string",
  "email": "string",
  "password": "string",
  "username": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | object |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users/{id}/disable

#### Purpose
Disable User Disable a specific user account preventing future logins

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users/{id}/enable

#### Purpose
Enable User Re-enable a disabled user account

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users/{id}/lock

#### Purpose
Lock User Immediately lock a user out of the system

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users/{id}/reset-password

#### Purpose
Admin Password Reset Forcefully reset a user's password and flag it for mandatory change on next login

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `AdminResetPasswordRequest`.
```json
{
  "new_password": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### GET /admin/users/{id}/sessions

#### Purpose
List User Sessions (Admin) List all active sessions for a specific user (admin only)

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[SessionResponse] |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

### POST /admin/users/{id}/unlock

#### Purpose
Unlock User Unlock an account that was either manually or automatically locked

#### Business Rule
Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | User UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go`.

## Source References
- `backend/internal/iam/module.go`
- `backend/internal/iam/permission/policies.go`
- `backend/internal/iam/transport/handler/admin_handler.go`
- `backend/internal/iam/transport/handler/session_handler.go`
- `backend/docs/swagger.json`
