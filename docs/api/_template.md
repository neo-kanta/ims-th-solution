# {Module Name} API

## Purpose
Explain the API module in business terms. State the backend source package and the base path.

## Business Context
Explain which IMS business process this API supports. Preserve workflow, investment-process, permission-management, approval/delegation, and IRG/compliance terms when applicable.

## Authentication
State whether Bearer JWT authentication is required. Identify any public endpoints, active-session checks, IP allowlist, or rate limits found in code.

## Authorization / Permission
List required permission code, role, or access rule found in route middleware or handler logic. If no rule is found, write: "Permission rule not found in code."

## Endpoint Summary

| Method | Path | Purpose | Auth | Permission | Status |
|---|---|---|---|---|---|
| GET | `/example` | Example purpose | JWT required | `EXAMPLE_VIEW` | Implemented |

## Endpoint Detail

### {METHOD} {PATH}

#### Purpose
Explain what this endpoint does.

#### Business Rule
Explain workflow, approval, permission, compliance, or validation rules involved. If no business rule is visible in code, write "Business rule not found in code."

#### Request

##### Path Parameters
| Name | Type | Required | Description |
|---|---|---|---|
| id | UUID string | Yes | Business object ID |

##### Query Parameters
| Name | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Page number |

##### Request Body
Schema: `{RequestDtoName}`.
```json
{}
```

#### Response
| HTTP Status | Description | Schema |
|---|---|---|
| 200 | OK | `{ResponseDtoName}` |
| 400 | Bad request / validation error | `ErrorResponse` |
| 401 | Missing or invalid JWT | `ErrorResponse` |
| 403 | Missing permission or access rule | `ErrorResponse` |

#### Source
`backend/internal/{module}/transport/router.go`; handler, DTO, service, permission, and tests used to validate this endpoint.

## Source References
- `backend/internal/{module}/transport/router.go`
- `backend/internal/{module}/transport/handler/...`
- `backend/internal/{module}/transport/dto/...`
- `backend/internal/{module}/application/...`
- `backend/internal/{module}/permission/policies.go`
- `backend/docs/swagger.json` when annotations exist
