# Chat API

**Status:** Current implementation as of 2026-07-20
**Base path:** `/api/v1`
**Endpoint count:** 4

## Purpose

Provides the optional AI financial assistant stream and persisted conversation-history reads.

## Authentication

Bearer JWT is required. The module is mounted only when the configured LLM provider initializes successfully.

## Authorization

Session reads are owner-scoped. `IAM_AUDIT_VIEW` permits audited cross-owner reads. Every MCP tool call passes a permission gate; mutating tools also require `CHAT_WRITE_ENABLED=true`.

## Runtime response conventions

Successful JSON handlers in this module use `{"data": ..., "message": ...}`; `message` is optional. A 204 response has no body. Error responses generally use `{"error": ..., "code": ..., "details": ...}`, although some newer typed-error paths use `{"error_code": ..., "message": ..., "request_id": ...}`. Clients should rely on the status and documented machine code when present, not exact human text.

## Endpoint summary

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/chat` | Send a chat message and stream the response | Authenticated; each MCP tool is permission-gated; writes also require CHAT_WRITE_ENABLED | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions` | List chat sessions | Owner-scoped | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions/{session_id}` | Get a chat session | Owner-scoped; IAM_AUDIT_VIEW may read with audit trail | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions/{session_id}/messages` | List chat session messages | Owner-scoped; IAM_AUDIT_VIEW may read with audit trail | Conditionally mounted when the chat provider initializes |

## Endpoint details

### POST `/api/v1/chat`

Streams text/event-stream. Each frame's `data:` payload is a ChatStreamEvent (see response model). Frames have event names: session_started, text, done, error. Use fetch() + ReadableStream on the client — EventSource cannot carry an Authorization header.

**Permission:** Authenticated; each MCP tool is permission-gated; writes also require CHAT_WRITE_ENABLED

**Business rule:** Chat is mounted only when its configured LLM provider initializes. Session reads are owner-scoped; auditor access requires `IAM_AUDIT_VIEW` and is audited. Tool execution is permission-gated, and mutating tools additionally require `CHAT_WRITE_ENABLED=true`.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

Schema: `SendMessageRequest`.

```json
{
  "content": "Summarize last quarter's NAV trend for Fund A.",
  "model": "claude-haiku-4-5-20251001",
  "provider": "anthropic",
  "session_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | One example frame payload; the response is a stream of these | text/event-stream frames of ChatStreamEvent |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/chat/transport/router.go`

### GET `/api/v1/chat/sessions`

Lists the caller's chat sessions. Auditors (IAM_AUDIT_VIEW) may pass user_id to list another user's sessions; such access is strictly audited.

**Permission:** Owner-scoped

**Business rule:** Chat is mounted only when its configured LLM provider initializes. Session reads are owner-scoped; auditor access requires `IAM_AUDIT_VIEW` and is audited. Tool execution is permission-gated, and mutating tools additionally require `CHAT_WRITE_ENABLED=true`.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| page | query | integer | No | Page (default 1) |
| limit | query | integer | No | Page size (default 20, max 100) |
| user_id | query | string | No | Auditor-only: list this user's sessions (requires IAM_AUDIT_VIEW) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SessionListResponse> |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

**Source:** `backend/internal/chat/transport/router.go`

### GET `/api/v1/chat/sessions/{session_id}`

Returns one session's metadata. Owner or auditor (IAM_AUDIT_VIEW). Auditor access is strictly audited.

**Permission:** Owner-scoped; IAM_AUDIT_VIEW may read with audit trail

**Business rule:** Chat is mounted only when its configured LLM provider initializes. Session reads are owner-scoped; auditor access requires `IAM_AUDIT_VIEW` and is audited. Tool execution is permission-gated, and mutating tools additionally require `CHAT_WRITE_ENABLED=true`.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| session_id | path | string | Yes | Session UUID |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SessionSummaryResponse> |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/chat/transport/router.go`

### GET `/api/v1/chat/sessions/{session_id}/messages`

Returns a session's messages with safe provenance (no raw provider payload). Owner or auditor (IAM_AUDIT_VIEW). Auditor access is strictly audited.

**Permission:** Owner-scoped; IAM_AUDIT_VIEW may read with audit trail

**Business rule:** Chat is mounted only when its configured LLM provider initializes. Session reads are owner-scoped; auditor access requires `IAM_AUDIT_VIEW` and is audited. Tool execution is permission-gated, and mutating tools additionally require `CHAT_WRITE_ENABLED=true`.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| session_id | path | string | Yes | Session UUID |
| page | query | integer | No | Page (default 1) |
| limit | query | integer | No | Page size (default 20, max 100) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<SessionMessagesResponse> |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/chat/transport/router.go`

## Source references

- `backend/internal/chat/module.go`
- `backend/internal/chat/transport/router.go`
- `backend/internal/chat/transport/handler/chat_handler.go`
- `backend/internal/chat/transport/handler/session_handler.go`
- `backend/internal/chat/application/service`
- `backend/docs/swagger.json`

## Notes

`POST /api/v1/chat` is Server-Sent Events. Use `fetch` plus a readable stream because browser `EventSource` cannot send the Authorization header.
