# Authentication API

## Purpose
Provides login, token refresh, current-user identity, password change, MFA enrollment/verification, and session management for IMS users.

## Business Context
Auth APIs establish the identity used by workflow, approval, permission, compliance, and investment modules. JWT claims carry user subject, username, roles/groups, and session identity used by downstream permission checks.

## Authentication
Login and refresh are public but rate-limited. All other endpoints require a valid Bearer JWT and active server-side session.

## Authorization / Permission
No function-permission middleware is mounted for user self-service auth endpoints. Sensitive operations are protected by JWT, active-session validation, and rate limiting; permission rule not found in code.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| POST | /auth/change-password | Change Password | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/login | User Login | Not required | Permission rule not found in code. | Implemented |
| POST | /auth/logout | User Logout | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/logout-all | Logout All Sessions | JWT required | Permission rule not found in code. | Implemented |
| GET | /auth/me | Get Current User | JWT required | Permission rule not found in code. | Implemented |
| GET | /auth/mfa/dev/totp-code | Get Current TOTP Code (Dev Only) | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/mfa/disable | Disable MFA | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/mfa/enroll | Start MFA Enrollment | JWT required | Permission rule not found in code. | Implemented |
| GET | /auth/mfa/status | Get MFA Status | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/mfa/verify | Verify and Enable MFA | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/refresh | Refresh Token | Not required | Permission rule not found in code. | Implemented |
| GET | /auth/sessions | List My Sessions | JWT required | Permission rule not found in code. | Implemented |
| POST | /auth/sessions/{id}/revoke | Revoke a Session | JWT required | Permission rule not found in code. | Implemented |

## Endpoint Detail

### POST /auth/change-password

#### Purpose
Change Password Change active user's password and revoke active sessions

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `ChangePasswordRequest`.
```json
{
  "new_password": "string",
  "old_password": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/login

#### Purpose
User Login Authenticate user and return access and refresh tokens. Returns mfa_required if MFA is active.

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `LoginRequest`.
```json
{
  "mfa_token": "string",
  "password": "string",
  "recovery_code": "string",
  "totp_code": "string",
  "username": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | LoginResponse |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/logout

#### Purpose
User Logout Revoke a single refresh token session

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `LogoutRequest`.
```json
{
  "refresh_token": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/logout-all

#### Purpose
Logout All Sessions Revoke all refresh token sessions globally for the user

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 204 | No Content |  |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### GET /auth/me

#### Purpose
Get Current User Retrieve the profile and permissions of the currently authenticated user

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 200 | OK | MeResponse |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### GET /auth/mfa/dev/totp-code

#### Purpose
Get Current TOTP Code (Dev Only) Development/test-only helper to retrieve the current TOTP code for the authenticated user

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 200 | OK | MFADevTOTPCodeResponse |
| 401 | Unauthorized | object |
| 404 | Not Found | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/mfa/disable

#### Purpose
Disable MFA Disable MFA for the current user (requires valid TOTP code)

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `MFADisableRequest`.
```json
{
  "totp_code": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/mfa/enroll

#### Purpose
Start MFA Enrollment Generate TOTP secret and recovery codes for MFA setup

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 200 | OK | MFAEnrollResponse |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### GET /auth/mfa/status

#### Purpose
Get MFA Status Get the current MFA enrollment status for the authenticated user

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 200 | OK | MFAStatusResponse |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/mfa/verify

#### Purpose
Verify and Enable MFA Verify TOTP code to complete MFA enrollment and activate

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `MFAVerifyRequest`.
```json
{
  "totp_code": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/refresh

#### Purpose
Refresh Token Issue a new access token using a refresh token (enforces idle and absolute timeouts)

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
Schema: `RefreshRequest`.
```json
{
  "refresh_token": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RefreshResponse |
| 400 | Bad Request | object |
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### GET /auth/sessions

#### Purpose
List My Sessions List all active sessions for the current user

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 200 | OK | array[SessionResponse] |
| 401 | Unauthorized | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

### POST /auth/sessions/{id}/revoke

#### Purpose
Revoke a Session Revoke a specific active session

#### Business Rule
Supports identity authentication, session issuance, MFA, and session lifecycle controls.

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
| 401 | Unauthorized | object |
| 403 | Forbidden | object |

#### Source
`backend/internal/iam/module.go; backend/internal/iam/transport/handler`.

## Source References
- `backend/internal/iam/module.go`
- `backend/internal/iam/transport/handler/auth_handler.go`
- `backend/internal/iam/transport/handler/mfa_handler.go`
- `backend/internal/iam/transport/handler/session_handler.go`
- `backend/docs/swagger.json`
