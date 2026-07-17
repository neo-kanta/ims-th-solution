# Compliance API

## Purpose
Provides IRG rule execution, compliance check-group lookup, breach listing, breach override, and rule instance administration.

## Business Context
Compliance APIs support pre-trade validation before Investment Decision approval and post-trade validation during workflow Transaction Closing. BLOCK/WARN results become audit-ready check records and breaches.

## Authentication
All compliance endpoints require Bearer JWT through `/api/v1`.

## Authorization / Permission
Route groups enforce `WORKFLOW_EXECUTE`, `IRG_VIEW_RULES`, `IRG_OVERRIDE_BREACH`, and `IRG_EDIT_RULE_INSTANCE`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /compliance/breaches | List Compliance Breaches | JWT required | IRG_VIEW_RULES | Implemented |
| POST | /compliance/breaches/{breachID}/override | Override Compliance Breach | JWT required | IRG_OVERRIDE_BREACH | Implemented |
| POST | /compliance/checks/post-trade | Run Post-Trade Compliance Check | JWT required | WORKFLOW_EXECUTE | Implemented |
| POST | /compliance/checks/pre-trade | Run Pre-Trade Compliance Check | JWT required | WORKFLOW_EXECUTE | Implemented |
| GET | /compliance/checks/{groupID} | Get Compliance Check Group | JWT required | IRG_VIEW_RULES | Implemented |
| GET | /compliance/rules | List Compliance Rule Instances | JWT required | IRG_VIEW_RULES | Implemented |
| POST | /compliance/rules | Create Compliance Rule Instance | JWT required | IRG_EDIT_RULE_INSTANCE | Implemented |

## Endpoint Detail

### GET /compliance/breaches

#### Purpose
List Compliance Breaches List compliance breaches with optional portfolio, contract, rule, status, and date filters.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| portfolio_id | string | No | Portfolio UUID |
| contract_id | string | No | Contract UUID |
| status | string | No | Breach status |
| rule_type_id | string | No | Rule type ID |
| date_from | string | No | Start business date (YYYY-MM-DD) |
| date_to | string | No | End business date (YYYY-MM-DD) |
| offset | integer | No | Offset (default 0) |
| limit | integer | No | Limit (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ListBreachesResult |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### POST /compliance/breaches/{breachID}/override

#### Purpose
Override Compliance Breach Override an open, overridable compliance breach with an audit reason.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| breachID | string | Yes | Breach UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `OverrideRequest`.
```json
{
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | Override |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### POST /compliance/checks/post-trade

#### Purpose
Run Post-Trade Compliance Check Evaluate compliance rules after trade capture or during periodic replay.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

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
Schema: `PostTradeRequest`.
```json
{
  "business_date": "string",
  "contract_id": "string",
  "portfolio_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | PostTradeCheckResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### POST /compliance/checks/pre-trade

#### Purpose
Run Pre-Trade Compliance Check Evaluate compliance rules for a proposed order before execution.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

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
Schema: `PreTradeRequest`.
```json
{
  "business_date": "string",
  "contract_id": "string",
  "currency": "string",
  "exchange": "string",
  "fees": "string",
  "order_id": "string",
  "portfolio_id": "string",
  "price": "string",
  "quantity": "string",
  "side": "string",
  "ticker": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | PreTradeCheckResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### GET /compliance/checks/{groupID}

#### Purpose
Get Compliance Check Group Retrieve check records and breaches associated with one compliance check group.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| groupID | string | Yes | Check group UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | CheckGroupResult |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### GET /compliance/rules

#### Purpose
List Compliance Rule Instances List configured compliance rule instances with optional type and active filters.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| rule_type_id | string | No | Rule type ID |
| is_active | boolean | No | Filter by active flag |
| offset | integer | No | Offset (default 0) |
| limit | integer | No | Limit (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ListRuleInstancesResult |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

### POST /compliance/rules

#### Purpose
Create Compliance Rule Instance Create a compliance rule instance and its initial parameter version.

#### Business Rule
Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason.

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
Schema: `CreateRuleInstanceRequest`.
```json
{
  "change_reason": "string",
  "description": "string",
  "effective_from": "string",
  "effective_to": "string",
  "is_active": false,
  "name": "string",
  "parameters": {},
  "rule_type_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | CreateRuleInstanceResult |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go`.

## Source References
- `backend/internal/compliance/module.go`
- `backend/internal/compliance/transport/router.go`
- `backend/internal/compliance/transport/handler/compliance_handler.go`
- `backend/internal/compliance/application/command`
- `backend/docs/swagger.json`
## Notes
Pre-trade and post-trade HTTP endpoints are permission-gated with `WORKFLOW_EXECUTE`. In-process compliance checks invoked by Investment or Workflow use contract adapters and do not pass through HTTP permission middleware.
