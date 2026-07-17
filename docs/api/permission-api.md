# Permission API

## Purpose
Provides permission-governance APIs for accounts, groups, roles, function permissions, data permissions, effective access, change requests, labels, approval settings, and notification settings.

## Business Context
Permission APIs preserve the IMS Permission Management model: Account, Group, Account & Group, Function Permissions, and Data Permissions. Mutating access changes are routed through permission change requests and maker-checker approval steps.

## Authentication
All permission endpoints require Bearer JWT through `/api/v1`.

## Authorization / Permission
Route-level permissions use dot-notation codes such as `permission.users.view`, `permission.groups.view`, `permission.function_rights.view`, `permission.data_rights.view`, and `permission.change_request.approve`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /permissions/approval-settings | List permission approval settings | JWT required | permission.approval_settings.view | Implemented |
| GET | /permissions/approval-settings/{id} | Get permission approval setting | JWT required | permission.approval_settings.view | Implemented |
| GET | /permissions/change-requests | List permission change requests | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests | Create a permission change request | JWT required | permission.change_request.create | Implemented |
| GET | /permissions/change-requests/{id} | Get a permission change request | JWT required | permission.change_request.review | Implemented |
| PUT | /permissions/change-requests/{id} | Update a permission change request | JWT required | permission.change_request.create | Implemented |
| GET | /permissions/change-requests/{id}/approval-steps | List approval steps | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests/{id}/approval-steps/{stepId}/approve | Approve a permission approval step | JWT required | permission.change_request.approve | Implemented |
| POST | /permissions/change-requests/{id}/approval-steps/{stepId}/reject | Reject a permission approval step | JWT required | permission.change_request.reject | Implemented |
| POST | /permissions/change-requests/{id}/approval-steps/{stepId}/request-changes | Request changes on a permission approval step | JWT required | permission.change_request.approve | Implemented |
| POST | /permissions/change-requests/{id}/approve | Approve a permission change request | JWT required | permission.change_request.approve | Implemented |
| POST | /permissions/change-requests/{id}/cancel | Cancel a permission change request | JWT required | permission.change_request.cancel | Implemented |
| GET | /permissions/change-requests/{id}/checks | List change request checks | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests/{id}/close | Close a permission change request | JWT required | permission.change_request.close | Implemented |
| GET | /permissions/change-requests/{id}/comments | List change request comments | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests/{id}/comments | Add a change request comment | JWT required | permission.change_request.review | Implemented |
| PUT | /permissions/change-requests/{id}/comments/{commentId} | Update a change request comment | JWT required | permission.change_request.review | Implemented |
| DELETE | /permissions/change-requests/{id}/comments/{commentId} | Delete a change request comment | JWT required | permission.change_request.review | Implemented |
| GET | /permissions/change-requests/{id}/diff | Get change request diff | JWT required | permission.change_request.review | Implemented |
| GET | /permissions/change-requests/{id}/items | List change request items | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests/{id}/items | Add a change request item | JWT required | permission.change_request.create | Implemented |
| PUT | /permissions/change-requests/{id}/items/{itemId} | Update a change request item | JWT required | permission.change_request.create | Implemented |
| DELETE | /permissions/change-requests/{id}/items/{itemId} | Delete a change request item | JWT required | permission.change_request.create | Implemented |
| POST | /permissions/change-requests/{id}/labels | Add a label to a change request | JWT required | permission.change_request.create | Implemented |
| DELETE | /permissions/change-requests/{id}/labels/{labelId} | Remove a label from a change request | JWT required | permission.change_request.create | Implemented |
| POST | /permissions/change-requests/{id}/merge | Merge an approved permission change request | JWT required | permission.change_request.merge | Implemented |
| POST | /permissions/change-requests/{id}/reject | Reject a permission change request | JWT required | permission.change_request.reject | Implemented |
| POST | /permissions/change-requests/{id}/request-changes | Request changes on a permission change request | JWT required | permission.change_request.approve | Implemented |
| POST | /permissions/change-requests/{id}/rerun-checks | Rerun change request checks | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/change-requests/{id}/submit | Submit a permission change request | JWT required | permission.change_request.submit | Implemented |
| GET | /permissions/data-rights | List data rights | JWT required | permission.data_rights.view | Implemented |
| GET | /permissions/effective/users/{userId} | Get effective permissions for a user | JWT required | permission.users.view | Implemented |
| GET | /permissions/function-definitions | List function definitions | JWT required | permission.function_rights.view | Implemented |
| GET | /permissions/function-rights | List function rights | JWT required | permission.function_rights.view | Implemented |
| GET | /permissions/groups | List permission groups | JWT required | permission.groups.view | Implemented |
| GET | /permissions/groups/{id} | Get permission group detail | JWT required | permission.groups.view | Implemented |
| GET | /permissions/labels | List permission labels | JWT required | permission.change_request.review | Implemented |
| POST | /permissions/labels | Create or update a permission label | JWT required | permission.change_request.create | Implemented |
| PUT | /permissions/labels/{id} | Update a permission label | JWT required | permission.change_request.create | Implemented |
| GET | /permissions/notification-settings | List permission notification settings | JWT required | permission.notification.view | Implemented |
| PUT | /permissions/notification-settings | Update permission notification settings | JWT required | permission.notification.edit | Implemented |
| GET | /permissions/role-assignment-policies | List role assignment policies | JWT required | permission.roles.view | Implemented |
| GET | /permissions/roles | List roles | JWT required | permission.roles.view | Implemented |
| GET | /permissions/roles/{id} | Get a role | JWT required | permission.roles.view | Implemented |
| GET | /permissions/users | List permission users | JWT required | permission.users.view | Implemented |
| GET | /permissions/users/{id} | Get permission user detail | JWT required | permission.users.view | Implemented |
| POST | /permissions/users/{userId}/role-assignment-request | Create a role assignment request | JWT required | permission.change_request.create | Implemented |

## Endpoint Detail

### GET /permissions/approval-settings

#### Purpose
List permission approval settings

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[ApprovalSetting] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/approval-settings/{id}

#### Purpose
Get permission approval setting

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ApprovalSetting |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests

#### Purpose
List permission change requests

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| status | string | No | Request status |
| risk | string | No | Risk level |
| label | string | No | Label filter |
| module | string | No | Module filter |
| search | string | No | Search text |
| requester | uuid | No | Requester user ID |
| reviewer | uuid | No | Reviewer user ID |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Paged change request object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests

#### Purpose
Create a permission change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
```json
{
  "title": "string",
  "description": "string",
  "request_type": "FUNCTION_PERMISSION",
  "risk_level": "MEDIUM",
  "target_entity_type": "USER",
  "target_entity_id": "uuid"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}

#### Purpose
Get a permission change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### PUT /permissions/change-requests/{id}

#### Purpose
Update a permission change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "title": "string",
  "description": "string",
  "request_type": "FUNCTION_PERMISSION",
  "risk_level": "MEDIUM",
  "target_entity_type": "USER",
  "target_entity_id": "uuid"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}/approval-steps

#### Purpose
List approval steps

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[ApprovalStep] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/approval-steps/{stepId}/approve

#### Purpose
Approve a permission approval step

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| stepId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/approval-steps/{stepId}/reject

#### Purpose
Reject a permission approval step

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| stepId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/approval-steps/{stepId}/request-changes

#### Purpose
Request changes on a permission approval step

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| stepId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/approve

#### Purpose
Approve a permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/cancel

#### Purpose
Cancel a permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}/checks

#### Purpose
List change request checks

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[Check] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/close

#### Purpose
Close a permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}/comments

#### Purpose
List change request comments

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[Comment] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/comments

#### Purpose
Add a change request comment

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | Comment |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### PUT /permissions/change-requests/{id}/comments/{commentId}

#### Purpose
Update a change request comment

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| commentId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### DELETE /permissions/change-requests/{id}/comments/{commentId}

#### Purpose
Delete a change request comment

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| commentId | string | Yes | Path parameter |

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
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}/diff

#### Purpose
Get change request diff

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Diff object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/change-requests/{id}/items

#### Purpose
List change request items

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[ChangeItem] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/items

#### Purpose
Add a change request item

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "item_type": "FUNCTION_RIGHT",
  "target_table": "permissions_function_rights",
  "target_id": "uuid",
  "action_type": "ADD",
  "before_json": {},
  "after_json": {}
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ChangeItem |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### PUT /permissions/change-requests/{id}/items/{itemId}

#### Purpose
Update a change request item

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| itemId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "item_type": "FUNCTION_RIGHT",
  "target_table": "permissions_function_rights",
  "target_id": "uuid",
  "action_type": "UPDATE",
  "before_json": {},
  "after_json": {}
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### DELETE /permissions/change-requests/{id}/items/{itemId}

#### Purpose
Delete a change request item

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| itemId | string | Yes | Path parameter |

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
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/labels

#### Purpose
Add a label to a change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "label_code": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### DELETE /permissions/change-requests/{id}/labels/{labelId}

#### Purpose
Remove a label from a change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |
| labelId | string | Yes | Path parameter |

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
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/merge

#### Purpose
Merge an approved permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/reject

#### Purpose
Reject a permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/request-changes

#### Purpose
Request changes on a permission change request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "comment": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/rerun-checks

#### Purpose
Rerun change request checks

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[Check] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/change-requests/{id}/submit

#### Purpose
Submit a permission change request

#### Business Rule
Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ChangeRequest |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/data-rights

#### Purpose
List data rights

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[DataRight] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/effective/users/{userId}

#### Purpose
Get effective permissions for a user

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| userId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | EffectivePermission |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/function-definitions

#### Purpose
List function definitions

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[FunctionDefinition] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/function-rights

#### Purpose
List function rights

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[FunctionRight] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/groups

#### Purpose
List permission groups

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| search | string | No | Search text |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Paged GroupSummary object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/groups/{id}

#### Purpose
Get permission group detail

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | GroupSummary |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/labels

#### Purpose
List permission labels

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[Label] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/labels

#### Purpose
Create or update a permission label

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
```json
{
  "label_code": "string",
  "label_name": "string",
  "label_type": "RISK",
  "color": "#ff0000",
  "description": "string",
  "is_active": true
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | Label |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### PUT /permissions/labels/{id}

#### Purpose
Update a permission label

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "label_code": "string",
  "label_name": "string",
  "label_type": "RISK",
  "color": "#ff0000",
  "description": "string",
  "is_active": true
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Label |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/notification-settings

#### Purpose
List permission notification settings

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[NotificationSetting] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### PUT /permissions/notification-settings

#### Purpose
Update permission notification settings

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
```json
{
  "settings": [
    {
      "event_code": "string",
      "channel": "email",
      "enabled": true,
      "target_scope": "REQUESTER",
      "template_subject": "string",
      "template_body": "string"
    }
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/role-assignment-policies

#### Purpose
List role assignment policies

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[RoleAssignmentPolicy] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/roles

#### Purpose
List roles

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

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
| 200 | OK | array[Role] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/roles/{id}

#### Purpose
Get a role

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Role |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/users

#### Purpose
List permission users

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| search | string | No | Search text |
| page | integer | No | Page number |
| limit | integer | No | Page size |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Paged UserSummary object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### GET /permissions/users/{id}

#### Purpose
Get permission user detail

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | UserSummary |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

### POST /permissions/users/{userId}/role-assignment-request

#### Purpose
Create a role assignment request

#### Business Rule
Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| userId | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "role_id": "uuid",
  "role_code": "string",
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | Role assignment request object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go`.

## Source References
- `backend/internal/permissions/module.go`
- `backend/internal/permissions/transport/handler/permission_handler.go`
- `backend/internal/permissions/application/service/permission_service.go`
- `backend/internal/permissions/domain/models.go`
- `backend/internal/permissions/permission/policies.go`
## Notes
The permissions module is more granular than the legacy `PERMISSIONS_VIEW` and `PERMISSIONS_MANAGE` provider codes. The fine-grained dot-notation permissions are seeded separately.
