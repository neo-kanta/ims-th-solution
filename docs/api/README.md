# IMS API Documentation

## Start here

[current-api-reference.md](current-api-reference.md) is the canonical catalog of the routes mounted by the current server. It reconciles Chi route registration with both generated Swagger specifications and identifies registered routes that Swagger does not currently expose.

Current runtime inventory as of 2026-07-20:

- 234 possible operations under `/api/v1`; the four Chat operations are conditional on successful LLM-provider initialization.
- 25 implemented Portfolio V2 operations under `/api/v2`.
- Public operational endpoints at `/health` and `/metrics`.
- Separate V1 and V2 Swagger UIs at `/swagger/index.html` and `/swagger/v2/index.html`.

Router registration is authoritative when Markdown and Swagger disagree.

## Current implementation documents

| Module | Document | Runtime status | Scope |
| --- | --- | --- | --- |
| Complete catalog | [current-api-reference.md](current-api-reference.md) | Current | Every mounted V1/V2 route, permissions, conditional routes, and OpenAPI gaps |
| System/discovery | [system-api.md](system-api.md) | Current | Health, metrics, Swagger UI, and raw specs |
| Authentication | [auth-api.md](auth-api.md) | Implemented | Login, refresh, MFA, sessions, and current user |
| IAM | [iam-api.md](iam-api.md) | Implemented | Admin account and session management |
| Audit | [audit-api.md](audit-api.md) | Implemented | System and permission-governance audit feeds |
| Permissions | [permission-api.md](permission-api.md) | Implemented | Users, groups, roles, rights, and change-request governance |
| Approval | [approval-api.md](approval-api.md) | Implemented | Maker-checker runtime, groups, teams, and process configuration |
| Workflow | [workflow-api.md](workflow-api.md) | Implemented | Daily workflow state, transitions, scheduler, and settings |
| Compliance / IRG | [compliance-api.md](compliance-api.md) | Implemented | Checks, breaches, overrides, and rule instances |
| Investment V1 | [investment-api.md](investment-api.md) | Implemented | Funds, instruments, research, decisions, execution, confirmation, and valuation |
| Portfolio V1 | [portfolio-api.md](portfolio-api.md) | Implemented | Portfolio master, holdings, cash, ledger, and valuation |
| Portfolio V2 | [portfolio-v2-api.md](portfolio-v2-api.md) | Implemented subset | 25 portfolio-code operations under `/api/v2` |
| Market Data | [market-data-api.md](market-data-api.md) | Implemented | Quotes, history, imports, provider health, and screen DTOs |
| Reference Data | [reference-data-api.md](reference-data-api.md) | Current | Canonical securities, provider mappings, and unmapped candidates |
| Integration | [integration-api.md](integration-api.md) | Current | Dashboard, task feeds, and company/mine valuation summary |
| Notification | [notification-api.md](notification-api.md) | Implemented | In-app notifications and email operations |
| Watchlist | [watchlist-current-api.md](watchlist-current-api.md) | Current | Watchlist items, alerts, acknowledgement, and evaluation |
| Chat | [chat-api.md](chat-api.md) | Conditional | SSE assistant endpoint and conversation history |

## Design and supporting documents

These files contain design rationale, frontend guidance, or historical proposals. They are useful context but are not the authoritative runtime contract:

- [portfolio-v2-api-ddd.md](portfolio-v2-api-ddd.md) — broader Portfolio V2 DDD design, including future routes.
- [watchlist-api.md](watchlist-api.md) — original Watchlist design/API contract and future-policy discussion.
- [notification-email-api.md](notification-email-api.md) — notification/email design and operating model.
- [frontend-notification-pages.md](frontend-notification-pages.md) — frontend consumption guidance.
- [approval-permission-user-facing-dto.md](approval-permission-user-facing-dto.md) — display-safe approval/permission DTO guidance.

## Authentication baseline

`POST /api/v1/auth/login` and `POST /api/v1/auth/refresh` are public and rate-limited. Other V1 endpoints and all V2 endpoints require a Bearer JWT. Authentication checks token issuer, audience, expiry, active account status, and the backing server-side session when a session ID is present.

IAM admin and audit endpoints additionally inherit the admin IP allowlist and admin/export rate limits. Route-level function permissions and application-level owner/data-scope rules are recorded per endpoint.

## Response baseline

Successful JSON handlers normally use:

```json
{
  "data": {},
  "message": "optional"
}
```

The codebase currently has two error-envelope generations. Legacy handlers use `error`, optional `code`, and optional `details`; typed domain-error handlers use `error_code`, `message`, optional `details`, and `request_id`. Authentication middleware also has a few JSON-text errors written through `http.Error`. Consumers should primarily handle HTTP status and stable machine codes when present.

## OpenAPI and generated clients

The generated source artifacts are:

- `backend/docs/swagger.json` for `/api/v1`.
- `backend/docs/v2/v2_swagger.json` for `/api/v2` Portfolio routes.
- `frontend/app/api/ims-api.d.ts` for generated frontend V1 types.

The current catalog identifies 60 mounted V1 operations missing from the V1 Swagger artifact: 49 permission/audit operations and 11 investment decision/execution/confirmation operations. They are real runtime routes, but they are absent from generated frontend types until their authoritative Go handlers receive Swagger annotations and the client is regenerated.

The V1 Swagger artifact also lists `/health` under a `/api/v1` base path even though the live route is `/health`. Use [system-api.md](system-api.md) for the correct operational paths.

## Regeneration workflow

1. Update Go Swagger annotations at the authoritative handlers.
2. Run `make swagger` from the repository root.
3. Run `python -B docs/api/_build_current_api_docs.py` to refresh the canonical catalog and current module documents.
4. Run `python -B docs/api/_build_api_docs.py` only when intentionally refreshing the legacy generated module documents.
5. Compare registered routes against both Swagger artifacts and investigate any unexplained difference.
6. Run `make api-client` when the accepted V1 contract must be propagated to the frontend.

The current-code generator writes only:

- `current-api-reference.md`
- `system-api.md`
- `integration-api.md`
- `reference-data-api.md`
- `chat-api.md`
- `watchlist-current-api.md`
- `portfolio-v2-api.md`

## Source hierarchy

1. `backend/cmd/server/main.go` for base paths, middleware, conditional modules, and public endpoints.
2. `backend/internal/*/module.go` and transport routers for mounted methods and paths.
3. Handlers, application services, and permission catalogs for validation, authorization, and business rules.
4. Generated Swagger for annotated request/response metadata.
5. Markdown documents for consumer guidance and rationale.
