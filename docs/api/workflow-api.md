# Workflow API

## Purpose
Provides daily workflow state, transition execution, scheduler controls, transition rules, and workflow approver settings.

## Business Context
Workflow APIs implement the IMS daily control sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing. Investment ledger and decision submission checks depend on this state.

## Authentication
All workflow endpoints require Bearer JWT through the authenticated `/api/v1` route group.

## Authorization / Permission
Legacy UUID day-state reads require `WORKFLOW_VIEW`; scheduler requires `WORKFLOW_RUN_SCHEDULER`; legacy transition execution resolves per-action workflow permissions in the handler. Business-readable daily endpoints use settings-based approver/admin checks instead of `RequirePermission` middleware.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /workflow/daily | Get Daily Workflow State | JWT required | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| POST | /workflow/daily/execute | Execute Daily Workflow Transition | JWT required | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| GET | /workflow/daily/transitions | Get Daily Transition History | JWT required | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| GET | /workflow/day-states/{contractId} | Get Workflow Day State | JWT required | WORKFLOW_VIEW | Implemented |
| GET | /workflow/day-states/{contractId}/history | Get Workflow Transition History | JWT required | WORKFLOW_VIEW | Implemented |
| POST | /workflow/day-states/{contractId}/transitions | Execute Workflow Transition | JWT required | Per-action workflow permission | Implemented |
| POST | /workflow/scheduler/run-once | Run Workflow Scheduler Once | JWT required | WORKFLOW_RUN_SCHEDULER | Implemented |
| GET | /workflow/settings | Get Workflow Approval Settings | JWT required | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| PUT | /workflow/settings | Update Workflow Approval Settings | JWT required | Admin role required by handler; no RequirePermission middleware | Implemented |
| GET | /workflow/transition-rules | Get Workflow Transition Rules | JWT required | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |

## Endpoint Detail

### GET /workflow/daily

#### Purpose
Get Daily Workflow State Returns aggregated global workflow state for a business date.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| businessDate | string | Yes | Business date (YYYY-MM-DD) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DailyWorkflowResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### POST /workflow/daily/execute

#### Purpose
Execute Daily Workflow Transition Execute a workflow transition using business-readable identifiers and canonical operation types.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

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
Schema: `DailyExecuteRequest`.
```json
{
  "accountingDate": "string",
  "attestationReason": "string",
  "businessDate": "string",
  "notes": "string",
  "operationType": "string",
  "reason": "string",
  "remark": "string",
  "zeroTransactionAttestation": false
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DailyWorkflowResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### GET /workflow/daily/transitions

#### Purpose
Get Daily Transition History Returns paginated global workflow transition history.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| businessDate | string | Yes | Business date (YYYY-MM-DD) |
| page | integer | No | Page number (1-based, default 1) |
| pageSize | integer | No | Page size (default 20, max 100) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DailyTransitionsResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### GET /workflow/day-states/{contractId}

#### Purpose
Get Workflow Day State Get the current workflow state for a contract and business date.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| contractId | string | Yes | Contract UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| businessDate | string | Yes | Business date (YYYY-MM-DD) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | WorkflowStateResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### GET /workflow/day-states/{contractId}/history

#### Purpose
Get Workflow Transition History Get transition history for a contract and business date.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| contractId | string | Yes | Contract UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| businessDate | string | Yes | Business date (YYYY-MM-DD) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | HistoryResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### POST /workflow/day-states/{contractId}/transitions

#### Purpose
Execute Workflow Transition Execute a workflow transition such as OPEN_DAY, APPROVE, CLOSE_TRANSACTIONS, or ROLLBACK_ACCOUNTING_CLOSE.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| contractId | string | Yes | Contract UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ExecuteTransitionRequest`.
```json
{
  "accountingDate": "string",
  "action": "string",
  "attestationReason": "string",
  "businessDate": "string",
  "idempotencyKey": "string",
  "notes": "string",
  "reason": "string",
  "zeroTransactionAttestation": false
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TransitionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### POST /workflow/scheduler/run-once

#### Purpose
Run Workflow Scheduler Once Manually execute one workflow day scheduler tick.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

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
| 200 | OK | array[object] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### GET /workflow/settings

#### Purpose
Get Workflow Approval Settings Returns the configured approvers for each workflow operation type.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

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
| 200 | OK | WorkflowSettingsResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### PUT /workflow/settings

#### Purpose
Update Workflow Approval Settings Replace the configured approvers for a workflow operation type. Admin only. Note: approver account codes are stored without IAM validation (no cross-module IAM lookup interface exists).

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

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
Schema: `DailySettingsUpdateRequest`.
```json
{
  "approvers": [
    "<ApproverInput>"
  ],
  "operationType": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | WorkflowSettingsResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

### GET /workflow/transition-rules

#### Purpose
Get Workflow Transition Rules Returns the full state machine topology as a list of valid from->operation->to transitions.

#### Business Rule
Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing.

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
| 200 | OK | TransitionRulesResponse |
| 401 | Unauthorized | ErrorResponse |

#### Source
`backend/internal/workflow/transport/router.go`.

## Source References
- `backend/internal/workflow/transport/router.go`
- `backend/internal/workflow/transport/handler/workflow_handler.go`
- `backend/internal/workflow/transport/handler/daily_handler.go`
- `backend/internal/workflow/permission/policies.go`
- `backend/docs/swagger.json`
