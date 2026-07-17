# IMS API Documentation

## Purpose
This directory contains maintainable Markdown API documentation for the IMS backend. The documents are grounded in Go route registration, handlers, DTOs, services, permission middleware, and generated Swagger/OpenAPI metadata where annotations exist.

## Business Context
IMS is an Investment Management System. The API surface supports daily workflow control, stock investment management, permission governance, maker-checker approval, IRG / Compliance validation, portfolio ledger and valuation, market data, notifications, IAM, and audit.

## Documentation Index
| Module | Document | Status | Scope |
| --- | --- | --- | --- |
| Authentication | [auth-api.md](auth-api.md) | Implemented | Login, refresh, MFA, sessions, current user |
| IAM | [iam-api.md](iam-api.md) | Implemented | Admin account and session management |
| Permissions | [permission-api.md](permission-api.md) | Implemented | Account, Group, Function Permission, Data Permission, and change-request governance |
| Approval | [approval-api.md](approval-api.md) | Implemented | Generic maker-checker runtime, delegation, groups, teams, process config |
| Workflow | [workflow-api.md](workflow-api.md) | Implemented | Investment Day Start, Manager Approval, Transaction Closing, Accounting Closing |
| Investment | [investment-api.md](investment-api.md) | Implemented | Research, decision, execution, confirmation, fund/instrument APIs |
| Compliance / IRG | [compliance-api.md](compliance-api.md) | Implemented | Pre-trade/post-trade checks, breaches, overrides, rule instances |
| Portfolio | [portfolio-api.md](portfolio-api.md) | Implemented | Portfolio master, holdings, cash, ledger, valuation |
| Portfolio V2 | [portfolio-v2-api-ddd.md](portfolio-v2-api-ddd.md) | Target design | Portfolio-centric DDD redesign for API V2 |
| Market Data | [market-data-api.md](market-data-api.md) | Implemented | Quotes, history, imports, provider health, screen DTOs |
| Notification | [notification-api.md](notification-api.md) | Implemented | In-app notifications and email outbox admin |
| Audit | [audit-api.md](audit-api.md) | Implemented | System audit and permission-governance audit feeds |

## Not Implemented Yet
No requested API module is marked "Not implemented yet" in this repository. All required module files above have matching backend routes.

## Authentication Baseline
Most APIs are mounted under `/api/v1` behind Bearer JWT authentication. `/auth/login` and `/auth/refresh` are public but rate-limited. IAM admin and audit admin endpoints additionally inherit admin IP allowlist and admin/export rate limits.

## Authorization Baseline
Route-level authorization is documented from `middleware.RequirePermission(...)` calls where present. If a route is authenticated but no function-permission middleware or handler rule is visible, the endpoint states: "Permission rule not found in code."

## Automation Workflow
1. Update Go Swagger annotations when handler DTOs or routes change.
2. Regenerate backend Swagger with the repository's existing Swagger command (`make swagger` when available).
3. Run `python docs/api/_build_api_docs.py` from the repository root.
4. Compare route registration against `backend/docs/swagger.json`; router registration remains the source of truth when Swagger annotations are missing.
5. Preserve business terms from `docs/investment-module.md`, `docs/compliance-module.md`, `docs/handoff/workflow-backend.md`, and `docs/handoff/approval-module.md`.

## Source References
- `backend/cmd/server/main.go`
- `backend/docs/swagger.json`
- `backend/internal/*/module.go`
- `backend/internal/*/transport/router.go`
- `backend/internal/*/transport/handler/*.go`
- `backend/internal/*/transport/dto/**/*.go`
- `backend/internal/*/permission/policies.go`
- `docs/investment-module.md`
- `docs/compliance-module.md`
- `docs/handoff/workflow-backend.md`
- `docs/handoff/approval-module.md`
