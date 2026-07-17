# Approval API

## Purpose
Provides the generic maker-checker approval runtime and approval-group, team, and process configuration APIs.

## Business Context
Approval APIs support review, approve, reject, submit, withdraw, cancel, revoke, delegation, and maker-checker controls for IMS subjects such as Investment Research Reports, Investment Decisions, and Compliance Release requests.

## Authentication
All approval endpoints are mounted under `/api/v1` behind Bearer JWT authentication.

## Authorization / Permission
Every approval route is gated by `middleware.RequirePermission` using the approval permission catalog. Subject-level access is also checked through registered subject access ports for investment subjects.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /approval-config/groups | List approval groups | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/groups | Create an approval group | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| PUT | /approval-config/groups/{id} | Update an approval group | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| GET | /approval-config/groups/{id}/members | List approval group members | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/groups/{id}/members | Add an approval group member | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| POST | /approval-config/groups/{id}/members/reorder | Reorder approval group members | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| PUT | /approval-config/groups/{id}/members/{memberId} | Update an approval group member | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| POST | /approval-config/groups/{id}/members/{memberId}/approve | Approve an approval group member | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| POST | /approval-config/groups/{id}/members/{memberId}/revoke | Revoke an approval group member | JWT required | APPROVAL_GROUP_MANAGE | Implemented |
| GET | /approval-config/processes | List approval process configs | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/processes | Create an approval process config | JWT required | APPROVAL_PROCESS_MANAGE | Implemented |
| GET | /approval-config/processes/{id} | Get an approval process config | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| PUT | /approval-config/processes/{id} | Update an approval process config | JWT required | APPROVAL_PROCESS_MANAGE | Implemented |
| POST | /approval-config/processes/{id}/activate | Activate an approval process config | JWT required | APPROVAL_PROCESS_MANAGE | Implemented |
| POST | /approval-config/processes/{id}/deactivate | Deactivate an approval process config | JWT required | APPROVAL_PROCESS_MANAGE | Implemented |
| GET | /approval-config/teams | List approval teams | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/teams | Create an approval team | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| PUT | /approval-config/teams/{id} | Update an approval team | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| GET | /approval-config/teams/{id}/contracts | List a team's contract assignments | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/teams/{id}/contracts | Assign a contract/fund to an approval team | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| GET | /approval-config/teams/{id}/members | List approval team members | JWT required | APPROVAL_CONFIG_VIEW | Implemented |
| POST | /approval-config/teams/{id}/members | Add an approval team member | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| PUT | /approval-config/teams/{id}/members/{memberId} | Update an approval team member | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| DELETE | /approval-config/teams/{id}/members/{memberId} | Remove an approval team member | JWT required | APPROVAL_TEAM_MANAGE | Implemented |
| GET | /approvals/inbox | List my approval inbox | JWT required | APPROVAL_VIEW_INBOX | Implemented |
| GET | /approvals/requests | List approval requests | JWT required | APPROVAL_VIEW_REQUEST | Implemented |
| GET | /approvals/requests/{requestId} | Get approval request detail | JWT required | APPROVAL_VIEW_REQUEST | Implemented |
| POST | /approvals/requests/{requestId}/cancel | Cancel an approval request | JWT required | APPROVAL_CANCEL | Implemented |
| POST | /approvals/requests/{requestId}/revoke | Revoke an approved request | JWT required | APPROVAL_REVOKE | Implemented |
| GET | /approvals/requests/{requestId}/timeline | Get approval timeline | JWT required | APPROVAL_AUDIT_VIEW | Implemented |
| POST | /approvals/requests/{requestId}/withdraw | Withdraw an approval request | JWT required | APPROVAL_WITHDRAW | Implemented |
| GET | /approvals/subjects/{subjectType}/{subjectId}/status | Get subject approval status | JWT required | APPROVAL_VIEW_REQUEST | Implemented |
| POST | /approvals/submit | Submit a subject for approval | JWT required | APPROVAL_SUBMIT | Implemented |
| POST | /approvals/tasks/{taskId}/approve | Approve an approval task | JWT required | APPROVAL_APPROVE | Implemented |
| POST | /approvals/tasks/{taskId}/reject | Reject an approval task | JWT required | APPROVAL_REJECT | Implemented |

## Endpoint Detail

### GET /approval-config/groups

#### Purpose
List approval groups

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| active_only | boolean | No | Only active groups |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[GroupResponse] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/groups

#### Purpose
Create an approval group

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

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
Schema: `GroupRequest`.
```json
{
  "group_code": "string",
  "group_name": "string",
  "is_active": false,
  "remarks": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | GroupResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### PUT /approval-config/groups/{id}

#### Purpose
Update an approval group

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `GroupRequest`.
```json
{
  "group_code": "string",
  "group_name": "string",
  "is_active": false,
  "remarks": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | GroupResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/groups/{id}/members

#### Purpose
List approval group members

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[GroupMemberResponse] |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/groups/{id}/members

#### Purpose
Add an approval group member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `GroupMemberRequest`.
```json
{
  "is_active": false,
  "member_type": "string",
  "priority_order": 0,
  "status": "string",
  "user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | GroupMemberResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/groups/{id}/members/reorder

#### Purpose
Reorder approval group members

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ReorderRequest`.
```json
{
  "member_ids": [
    "string"
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[GroupMemberResponse] |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### PUT /approval-config/groups/{id}/members/{memberId}

#### Purpose
Update an approval group member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |
| memberId | string | Yes | Member UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `GroupMemberRequest`.
```json
{
  "is_active": false,
  "member_type": "string",
  "priority_order": 0,
  "status": "string",
  "user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | GroupMemberResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/groups/{id}/members/{memberId}/approve

#### Purpose
Approve an approval group member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |
| memberId | string | Yes | Member UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | GroupMemberResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/groups/{id}/members/{memberId}/revoke

#### Purpose
Revoke an approval group member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Group UUID |
| memberId | string | Yes | Member UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | GroupMemberResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/processes

#### Purpose
List approval process configs

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| process_type | string | No | Process type filter |
| active_only | boolean | No | Only active configs |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[ProcessConfigResponse] |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/processes

#### Purpose
Create an approval process config

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

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
Schema: `ProcessConfigRequest`.
```json
{
  "contract_id": "string",
  "contract_type": "string",
  "effective_date": "string",
  "group_approval_enabled": false,
  "is_active": false,
  "process_code": "string",
  "process_name": "string",
  "process_type": "string",
  "require_team_approval": false,
  "stages": [
    "<StageRequest>"
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ProcessConfigResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/processes/{id}

#### Purpose
Get an approval process config

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Process config UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ProcessConfigResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### PUT /approval-config/processes/{id}

#### Purpose
Update an approval process config

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Process config UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ProcessConfigRequest`.
```json
{
  "contract_id": "string",
  "contract_type": "string",
  "effective_date": "string",
  "group_approval_enabled": false,
  "is_active": false,
  "process_code": "string",
  "process_name": "string",
  "process_type": "string",
  "require_team_approval": false,
  "stages": [
    "<StageRequest>"
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ProcessConfigResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/processes/{id}/activate

#### Purpose
Activate an approval process config

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Process config UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ProcessConfigResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/processes/{id}/deactivate

#### Purpose
Deactivate an approval process config

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Process config UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ProcessConfigResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/teams

#### Purpose
List approval teams

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

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
| 200 | OK | array[TeamResponse] |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/teams

#### Purpose
Create an approval team

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

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
Schema: `TeamRequest`.
```json
{
  "has_co_manager": false,
  "is_active": false,
  "max_allowed_stamps": 0,
  "min_required_stamps": 0,
  "remarks": "string",
  "team_code": "string",
  "team_name": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TeamResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### PUT /approval-config/teams/{id}

#### Purpose
Update an approval team

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `TeamRequest`.
```json
{
  "has_co_manager": false,
  "is_active": false,
  "max_allowed_stamps": 0,
  "min_required_stamps": 0,
  "remarks": "string",
  "team_code": "string",
  "team_name": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TeamResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/teams/{id}/contracts

#### Purpose
List a team's contract assignments

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[TeamContractResponse] |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/teams/{id}/contracts

#### Purpose
Assign a contract/fund to an approval team

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `TeamContractRequest`.
```json
{
  "contract_id": "string",
  "effective_date": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TeamContractResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approval-config/teams/{id}/members

#### Purpose
List approval team members

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[TeamMemberResponse] |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approval-config/teams/{id}/members

#### Purpose
Add an approval team member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `TeamMemberRequest`.
```json
{
  "is_active": false,
  "member_type": "string",
  "priority_order": 0,
  "user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TeamMemberResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### PUT /approval-config/teams/{id}/members/{memberId}

#### Purpose
Update an approval team member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |
| memberId | string | Yes | Member UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `TeamMemberRequest`.
```json
{
  "is_active": false,
  "member_type": "string",
  "priority_order": 0,
  "user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TeamMemberResponse |
| 400 | Bad Request | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### DELETE /approval-config/teams/{id}/members/{memberId}

#### Purpose
Remove an approval team member

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Team UUID |
| memberId | string | Yes | Member UUID |

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
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approvals/inbox

#### Purpose
List my approval inbox Returns the authenticated user's pending approval tasks joined with their requests.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | string | No | Task status filter (default PENDING) |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | InboxListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approvals/requests

#### Purpose
List approval requests Lists approval requests with optional filters.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | string | No | Request status |
| process_type | string | No | Process type |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approvals/requests/{requestId}

#### Purpose
Get approval request detail Returns a request with its tasks, immutable timeline and signature records.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| requestId | string | Yes | Approval request UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestDetailResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/requests/{requestId}/cancel

#### Purpose
Cancel an approval request Cancels an in-flight approval request (privileged).

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| requestId | string | Yes | Approval request UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/requests/{requestId}/revoke

#### Purpose
Revoke an approved request Revokes a previously approved request, returning the subject to an un-approved state. Requires a reason.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| requestId | string | Yes | Approval request UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ActionRequest`.
```json
{
  "comment": "string",
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approvals/requests/{requestId}/timeline

#### Purpose
Get approval timeline Returns the immutable, ordered approval timeline for a request.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| requestId | string | Yes | Approval request UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[EventResponse] |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/requests/{requestId}/withdraw

#### Purpose
Withdraw an approval request The submitter withdraws their own active approval request.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| requestId | string | Yes | Approval request UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### GET /approvals/subjects/{subjectType}/{subjectId}/status

#### Purpose
Get subject approval status Returns the latest approval request for a business object (subject), if any.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| subjectType | string | Yes | Subject type |
| subjectId | string | Yes | Subject UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | SubjectStatusResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/submit

#### Purpose
Submit a subject for approval Creates an approval request and the first stage tasks for a business object.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

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
Schema: `SubmitRequest`.
```json
{
  "contract_id": "string",
  "contract_type": "string",
  "portfolio_id": "string",
  "process_type": "string",
  "subject_id": "string",
  "subject_reference": "string",
  "subject_title": "string",
  "subject_type": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | RequestResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/tasks/{taskId}/approve

#### Purpose
Approve an approval task Records an approval on an assigned task and advances the request.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| taskId | string | Yes | Approval task UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ActionRequest`.
```json
{
  "comment": "string",
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

### POST /approvals/tasks/{taskId}/reject

#### Purpose
Reject an approval task Records a rejection (reason required); one rejection stops the request.

#### Business Rule
Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| taskId | string | Yes | Approval task UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ActionRequest`.
```json
{
  "comment": "string",
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RequestResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/approval/transport/router.go`.

## Source References
- `backend/internal/approval/transport/router.go`
- `backend/internal/approval/transport/handler/runtime_handler.go`
- `backend/internal/approval/transport/handler/config_handler.go`
- `backend/internal/approval/application/service/runtime_service.go`
- `backend/internal/approval/permission/policies.go`
- `backend/docs/swagger.json`
## Notes
Maker-checker is enforced in the approval runtime: the submitter cannot approve their own request. Delegation is checked when computing the viewer task and when authorizing approve/reject actions.
