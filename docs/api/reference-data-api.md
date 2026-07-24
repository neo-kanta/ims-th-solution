# Reference Data API

**Status:** Current implementation as of 2026-07-20
**Base path:** `/api/v1`
**Endpoint count:** 10

## Purpose

Manages canonical IMS securities, provider-symbol mappings, and the unmapped-symbol review queue used by market-data ingestion.

## Authentication

Bearer JWT is required for every endpoint.

## Authorization

The current router has no function-permission middleware. The module's permission catalog is intentionally empty, so every authenticated user can currently call these routes.

## Runtime response conventions

Successful JSON handlers in this module use `{"data": ..., "message": ...}`; `message` is optional. A 204 response has no body. Error responses generally use `{"error": ..., "code": ..., "details": ...}`, although some newer typed-error paths use `{"error_code": ..., "message": ..., "request_id": ...}`. Clients should rely on the status and documented machine code when present, not exact human text.

## Endpoint summary

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/reference-data/securities` | Create canonical security | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/reference-data/securities/search` | Search canonical securities | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/reference-data/securities/{security_id}` | Get canonical security | Authenticated; no function-permission middleware | Implemented |
| PATCH | `/api/v1/reference-data/securities/{security_id}` | Patch canonical security | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/reference-data/securities/{security_id}/mappings` | List provider mappings for a security | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/reference-data/securities/{security_id}/mappings` | Add provider mapping to a security | Authenticated; no function-permission middleware | Implemented |
| DELETE | `/api/v1/reference-data/securities/{security_id}/mappings/{mapping_id}` | Soft-delete a provider mapping | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/reference-data/unmapped-candidates` | List unmapped provider symbol candidates | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/reference-data/unmapped-candidates/{candidate_id}/map` | Map a candidate to an existing security | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/reference-data/unmapped-candidates/{candidate_id}/reject` | Reject an unmapped candidate | Authenticated; no function-permission middleware | Implemented |

## Endpoint details

### POST `/api/v1/reference-data/securities`

Register a new canonical IMS security.

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

Schema: `CreateSecurityRequest`.

```json
{
  "asset_type": "string",
  "auto_build_ims_symbol": false,
  "country_code": "string",
  "currency": "string",
  "cusip": "string",
  "display_symbol": "string",
  "exchange_mic": "string",
  "figi": "string",
  "ims_symbol": "string",
  "isin": "string",
  "name": "string",
  "primary_identifier": "string",
  "status": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<SecurityDTO> |
| 400 | Bad Request | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### GET `/api/v1/reference-data/securities/search`

Search canonical securities by query, asset type, provider, and status.

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| query | query | string | No | Free-text query against ims_symbol/display_symbol/name/isin |
| asset_type | query | string | No | Filter by AssetType (EQUITY, BOND, FX, ...) |
| provider | query | string | No | Only securities with an ACTIVE mapping for this provider_code |
| status | query | string | No | Filter by SecurityStatus (ACTIVE, INACTIVE, SUSPENDED) |
| limit | query | integer | No | Result limit (default 50, max 200) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SearchResponse> |
| 400 | Bad Request | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### GET `/api/v1/reference-data/securities/{security_id}`

Get canonical security

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| security_id | path | string | Yes | Security id |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SecurityDTO> |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### PATCH `/api/v1/reference-data/securities/{security_id}`

Patch canonical security

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| security_id | path | string | Yes | Security id |

#### Request body

Schema: `UpdateSecurityRequest`.

```json
{
  "asset_type": "string",
  "country_code": "string",
  "currency": "string",
  "cusip": "string",
  "display_symbol": "string",
  "exchange_mic": "string",
  "figi": "string",
  "isin": "string",
  "name": "string",
  "primary_identifier": "string",
  "status": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SecurityDTO> |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### GET `/api/v1/reference-data/securities/{security_id}/mappings`

List provider mappings for a security

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| security_id | path | string | Yes | Security id |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<MappingsResponse> |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### POST `/api/v1/reference-data/securities/{security_id}/mappings`

Add provider mapping to a security

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| security_id | path | string | Yes | Security id |

#### Request body

Schema: `AddProviderMappingRequest`.

```json
{
  "confidence_score": "string",
  "is_primary": false,
  "priority": 0,
  "provider_asset_type": "string",
  "provider_code": "string",
  "provider_currency": "string",
  "provider_exchange": "string",
  "provider_symbol": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<ProviderMappingDTO> |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### DELETE `/api/v1/reference-data/securities/{security_id}/mappings/{mapping_id}`

Marks the mapping as INACTIVE rather than removing it physically.

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| security_id | path | string | Yes | Security id |
| mapping_id | path | string | Yes | Mapping id |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 204 | No Content |  |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### GET `/api/v1/reference-data/unmapped-candidates`

List unmapped provider symbol candidates

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| status | query | string | No | Filter by candidate_status |
| provider | query | string | No | Filter by provider_code |
| batch_id | query | string | No | Filter by import batch id |
| limit | query | integer | No | Result limit (default 100, max 500) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<CandidatesResponse> |

**Source:** `backend/internal/reference_data/transport/router.go`

### POST `/api/v1/reference-data/unmapped-candidates/{candidate_id}/map`

Map a candidate to an existing security

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| candidate_id | path | string | Yes | Candidate id |

#### Request body

Schema: `MapCandidateRequest`.

```json
{
  "security_id": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

### POST `/api/v1/reference-data/unmapped-candidates/{candidate_id}/reject`

Reject an unmapped candidate

**Permission:** Authenticated; no function-permission middleware

**Business rule:** Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| candidate_id | path | string | Yes | Candidate id |

#### Request body

Schema: `RejectCandidateRequest`.

```json
{
  "reason": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/reference_data/transport/router.go`

## Source references

- `backend/internal/reference_data/transport/router.go`
- `backend/internal/reference_data/transport/handler/securities_handler.go`
- `backend/internal/reference_data/application/service.go`
- `backend/internal/reference_data/permission/policies.go`
- `backend/docs/swagger.json`

## Notes

The current handler wraps successful JSON payloads in `{"data": ...}` even where older Swagger annotations name only the inner DTO.
