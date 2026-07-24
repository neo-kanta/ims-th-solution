# IMS Current API Reference

**Verified against workspace code:** 2026-07-20
**Runtime source of truth:** `backend/cmd/server/main.go` plus each module's route registration
**Payload metadata:** `backend/docs/swagger.json` and `backend/docs/v2/v2_swagger.json`

## Scope and authority

This document is the canonical catalog of routes mounted by the current server. Router registration wins when Markdown and Swagger disagree. Swagger supplies request/response metadata only where an operation is annotated.

The current server registers 234 possible V1 operations, 25 V2 operations, plus `/health` and `/metrics`. The four Chat operations are conditional: they are present only when the configured LLM provider initializes. Swagger UI and raw-spec routes are discovery endpoints and are listed separately in `system-api.md`.

## Base URLs and interactive documentation

| Surface | Local URL | Notes |
| --- | --- | --- |
| V1 API | `http://localhost:8080/api/v1` | Primary modular-monolith API |
| Portfolio V2 API | `http://localhost:8080/api/v2` | Additive portfolio-code routes |
| V1 Swagger UI | `http://localhost:8080/swagger/index.html` | Swagger 2.0 V1 spec |
| V2 Swagger UI | `http://localhost:8080/swagger/v2/index.html` | Separate Swagger 2.0 spec because V2 has a different base path |
| Health | `http://localhost:8080/health` | Outside both API base paths |
| Metrics | `http://localhost:8080/metrics` | Prometheus; restrict at the network boundary |

## Authentication and request conventions

`POST /api/v1/auth/login` and `POST /api/v1/auth/refresh` are public and rate-limited. Other V1 routes and every V2 route require `Authorization: Bearer <access-token>`. Authentication validates issuer, audience, expiry, active user status, and active server-side session when the token carries a session ID.

JSON success handlers normally return `{"data": ..., "message": ...}` with optional `message`; 204 responses have no body. Legacy errors use `{"error": ..., "code": ..., "details": ...}` while typed domain-error paths use `{"error_code": ..., "message": ..., "details": ..., "request_id": ...}`. Some authentication middleware errors are JSON text written through `http.Error`, so clients must primarily use the HTTP status and parse machine codes when present.

Money, prices, quantities, and percentages are generally decimal strings. Timestamps are RFC3339/UTC and business dates use `YYYY-MM-DD`. Pagination is endpoint-specific: older APIs use either `page`/`limit` or `offset`/`limit`, so use the parameters documented for that operation.

## Module coverage

| Module | Surface | Operations | Runtime status | Route source |
| --- | --- | --- | --- | --- |
| Approval | V1 | 35 | Implemented | backend/internal/approval/transport/router.go |
| Audit | V1 | 4 | Implemented | backend/internal/audit/module.go; backend/internal/permissions/module.go |
| Authentication | V1 | 13 | Implemented | backend/internal/iam/module.go |
| Chat | V1 | 4 | Conditional | backend/internal/chat/transport/router.go |
| Compliance | V1 | 7 | Implemented | backend/internal/compliance/module.go |
| IAM administration | V1 | 9 | Implemented | backend/internal/iam/module.go |
| Integration | V1 | 4 | Implemented | backend/internal/integration/transport/router.go |
| Investment V1 | V1 | 66 | Implemented | backend/internal/investment/module.go |
| Market Data | V1 | 10 | Implemented | backend/internal/market_data/transport/http/router.go |
| Notification | V1 | 8 | Implemented | backend/internal/notification/transport/router.go |
| Permissions | V1 | 47 | Implemented | backend/internal/permissions/module.go |
| Reference Data | V1 | 10 | Implemented | backend/internal/reference_data/transport/router.go |
| Watchlist | V1 | 7 | Implemented | backend/internal/watchlist/transport/http/router.go |
| Workflow | V1 | 10 | Implemented | backend/internal/workflow/transport/router.go |
| Portfolio V2 | V2 | 25 | Implemented | backend/internal/investment/module.go:RegisterRoutesV2 |
| System/operations | Root | 2 | Implemented | backend/cmd/server/main.go |

## OpenAPI coverage gaps

The V1 Swagger artifact describes 174 mounted V1 operations after excluding its misplaced `/health` entry. The following 60 registered V1 operations are not present in that Swagger artifact and therefore will not appear in generated frontend types unless annotations are added:

| Method | Runtime path | Module |
| --- | --- | --- |
| GET | `/api/v1/audit/logs` | Audit |
| GET | `/api/v1/audit/logs/export` | Audit |
| GET | `/api/v1/investment/decisions/{id}` | Investment V1 |
| PUT | `/api/v1/investment/decisions/{id}` | Investment V1 |
| GET | `/api/v1/investment/executions` | Investment V1 |
| GET | `/api/v1/investment/executions/{id}` | Investment V1 |
| POST | `/api/v1/investment/executions/{id}/cancel` | Investment V1 |
| POST | `/api/v1/investment/executions/{id}/fill` | Investment V1 |
| GET | `/api/v1/investment/trade-confirmations` | Investment V1 |
| POST | `/api/v1/investment/trade-confirmations` | Investment V1 |
| POST | `/api/v1/investment/trade-confirmations/batch` | Investment V1 |
| GET | `/api/v1/investment/trade-confirmations/{id}` | Investment V1 |
| POST | `/api/v1/investment/trade-confirmations/{id}/resolve` | Investment V1 |
| GET | `/api/v1/permissions/approval-settings` | Permissions |
| GET | `/api/v1/permissions/approval-settings/{id}` | Permissions |
| GET | `/api/v1/permissions/change-requests` | Permissions |
| POST | `/api/v1/permissions/change-requests` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}` | Permissions |
| PUT | `/api/v1/permissions/change-requests/{id}` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}/approval-steps` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/approve` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/reject` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/request-changes` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/approve` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/cancel` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}/checks` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/close` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}/comments` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/comments` | Permissions |
| PUT | `/api/v1/permissions/change-requests/{id}/comments/{commentId}` | Permissions |
| DELETE | `/api/v1/permissions/change-requests/{id}/comments/{commentId}` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}/diff` | Permissions |
| GET | `/api/v1/permissions/change-requests/{id}/items` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/items` | Permissions |
| PUT | `/api/v1/permissions/change-requests/{id}/items/{itemId}` | Permissions |
| DELETE | `/api/v1/permissions/change-requests/{id}/items/{itemId}` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/labels` | Permissions |
| DELETE | `/api/v1/permissions/change-requests/{id}/labels/{labelId}` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/merge` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/reject` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/request-changes` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/rerun-checks` | Permissions |
| POST | `/api/v1/permissions/change-requests/{id}/submit` | Permissions |
| GET | `/api/v1/permissions/data-rights` | Permissions |
| GET | `/api/v1/permissions/effective/users/{userId}` | Permissions |
| GET | `/api/v1/permissions/function-definitions` | Permissions |
| GET | `/api/v1/permissions/function-rights` | Permissions |
| GET | `/api/v1/permissions/groups` | Permissions |
| GET | `/api/v1/permissions/groups/{id}` | Permissions |
| GET | `/api/v1/permissions/labels` | Permissions |
| POST | `/api/v1/permissions/labels` | Permissions |
| PUT | `/api/v1/permissions/labels/{id}` | Permissions |
| GET | `/api/v1/permissions/notification-settings` | Permissions |
| PUT | `/api/v1/permissions/notification-settings` | Permissions |
| GET | `/api/v1/permissions/role-assignment-policies` | Permissions |
| GET | `/api/v1/permissions/roles` | Permissions |
| GET | `/api/v1/permissions/roles/{id}` | Permissions |
| GET | `/api/v1/permissions/users` | Permissions |
| GET | `/api/v1/permissions/users/{id}` | Permissions |
| POST | `/api/v1/permissions/users/{userId}/role-assignment-request` | Permissions |

The V2 Swagger artifact covers all 25 currently registered V2 routes. Its success schemas often name the inner DTO even though the handlers wrap JSON responses in `SuccessResponse`; `portfolio-v2-api.md` records the runtime envelope.

## Endpoint catalog

### Approval

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/approval-config/groups` | List approval groups | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/groups` | Create an approval group | APPROVAL_GROUP_MANAGE | Implemented |
| PUT | `/api/v1/approval-config/groups/{id}` | Update an approval group | APPROVAL_GROUP_MANAGE | Implemented |
| GET | `/api/v1/approval-config/groups/{id}/members` | List approval group members | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/groups/{id}/members` | Add an approval group member | APPROVAL_GROUP_MANAGE | Implemented |
| POST | `/api/v1/approval-config/groups/{id}/members/reorder` | Reorder approval group members | APPROVAL_GROUP_MANAGE | Implemented |
| PUT | `/api/v1/approval-config/groups/{id}/members/{memberId}` | Update an approval group member | APPROVAL_GROUP_MANAGE | Implemented |
| POST | `/api/v1/approval-config/groups/{id}/members/{memberId}/approve` | Approve an approval group member | APPROVAL_GROUP_MANAGE | Implemented |
| POST | `/api/v1/approval-config/groups/{id}/members/{memberId}/revoke` | Revoke an approval group member | APPROVAL_GROUP_MANAGE | Implemented |
| GET | `/api/v1/approval-config/processes` | List approval process configs | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/processes` | Create an approval process config | APPROVAL_PROCESS_MANAGE | Implemented |
| GET | `/api/v1/approval-config/processes/{id}` | Get an approval process config | APPROVAL_CONFIG_VIEW | Implemented |
| PUT | `/api/v1/approval-config/processes/{id}` | Update an approval process config | APPROVAL_PROCESS_MANAGE | Implemented |
| POST | `/api/v1/approval-config/processes/{id}/activate` | Activate an approval process config | APPROVAL_PROCESS_MANAGE | Implemented |
| POST | `/api/v1/approval-config/processes/{id}/deactivate` | Deactivate an approval process config | APPROVAL_PROCESS_MANAGE | Implemented |
| GET | `/api/v1/approval-config/teams` | List approval teams | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/teams` | Create an approval team | APPROVAL_TEAM_MANAGE | Implemented |
| PUT | `/api/v1/approval-config/teams/{id}` | Update an approval team | APPROVAL_TEAM_MANAGE | Implemented |
| GET | `/api/v1/approval-config/teams/{id}/contracts` | List a team's contract assignments | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/teams/{id}/contracts` | Assign a contract/fund to an approval team | APPROVAL_TEAM_MANAGE | Implemented |
| GET | `/api/v1/approval-config/teams/{id}/members` | List approval team members | APPROVAL_CONFIG_VIEW | Implemented |
| POST | `/api/v1/approval-config/teams/{id}/members` | Add an approval team member | APPROVAL_TEAM_MANAGE | Implemented |
| PUT | `/api/v1/approval-config/teams/{id}/members/{memberId}` | Update an approval team member | APPROVAL_TEAM_MANAGE | Implemented |
| DELETE | `/api/v1/approval-config/teams/{id}/members/{memberId}` | Remove an approval team member | APPROVAL_TEAM_MANAGE | Implemented |
| GET | `/api/v1/approvals/inbox` | List my approval inbox | APPROVAL_VIEW_INBOX | Implemented |
| GET | `/api/v1/approvals/requests` | List approval requests | APPROVAL_VIEW_REQUEST | Implemented |
| GET | `/api/v1/approvals/requests/{requestId}` | Get approval request detail | APPROVAL_VIEW_REQUEST | Implemented |
| POST | `/api/v1/approvals/requests/{requestId}/cancel` | Cancel an approval request | APPROVAL_CANCEL | Implemented |
| POST | `/api/v1/approvals/requests/{requestId}/revoke` | Revoke an approved request | APPROVAL_REVOKE | Implemented |
| GET | `/api/v1/approvals/requests/{requestId}/timeline` | Get approval timeline | APPROVAL_AUDIT_VIEW | Implemented |
| POST | `/api/v1/approvals/requests/{requestId}/withdraw` | Withdraw an approval request | APPROVAL_WITHDRAW | Implemented |
| GET | `/api/v1/approvals/subjects/{subjectType}/{subjectId}/status` | Get subject approval status | APPROVAL_VIEW_REQUEST | Implemented |
| POST | `/api/v1/approvals/submit` | Submit a subject for approval | APPROVAL_SUBMIT | Implemented |
| POST | `/api/v1/approvals/tasks/{taskId}/approve` | Approve an approval task | APPROVAL_APPROVE | Implemented |
| POST | `/api/v1/approvals/tasks/{taskId}/reject` | Reject an approval task | APPROVAL_REJECT | Implemented |

### Audit

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/admin/audit` | List Audit Events | IAM_AUDIT_VIEW | Implemented |
| GET | `/api/v1/admin/audit/export` | Export Audit Events as CSV | IAM_AUDIT_VIEW | Implemented |
| GET | `/api/v1/audit/logs` | List permission-management audit logs | permission.audit.view | Implemented |
| GET | `/api/v1/audit/logs/export` | Export permission-management audit logs as CSV | permission.audit.export | Implemented |

### Authentication

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/auth/change-password` | Change Password | Authenticated | Implemented |
| POST | `/api/v1/auth/login` | User Login | Public | Implemented |
| POST | `/api/v1/auth/logout` | User Logout | Authenticated | Implemented |
| POST | `/api/v1/auth/logout-all` | Logout All Sessions | Authenticated | Implemented |
| GET | `/api/v1/auth/me` | Get Current User | Authenticated | Implemented |
| GET | `/api/v1/auth/mfa/dev/totp-code` | Get Current TOTP Code (Dev Only) | Authenticated | Implemented |
| POST | `/api/v1/auth/mfa/disable` | Disable MFA | Authenticated | Implemented |
| POST | `/api/v1/auth/mfa/enroll` | Start MFA Enrollment | Authenticated | Implemented |
| GET | `/api/v1/auth/mfa/status` | Get MFA Status | Authenticated | Implemented |
| POST | `/api/v1/auth/mfa/verify` | Verify and Enable MFA | Authenticated | Implemented |
| POST | `/api/v1/auth/refresh` | Refresh Token | Public | Implemented |
| GET | `/api/v1/auth/sessions` | List My Sessions | Authenticated | Implemented |
| POST | `/api/v1/auth/sessions/{id}/revoke` | Revoke a Session | Authenticated | Implemented |

### Chat

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/chat` | Send a chat message and stream the response | Authenticated; each MCP tool is permission-gated; writes also require CHAT_WRITE_ENABLED | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions` | List chat sessions | Owner-scoped | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions/{session_id}` | Get a chat session | Owner-scoped; IAM_AUDIT_VIEW may read with audit trail | Conditionally mounted when the chat provider initializes |
| GET | `/api/v1/chat/sessions/{session_id}/messages` | List chat session messages | Owner-scoped; IAM_AUDIT_VIEW may read with audit trail | Conditionally mounted when the chat provider initializes |

### Compliance

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/compliance/breaches` | List Compliance Breaches | IRG_VIEW_RULES | Implemented |
| POST | `/api/v1/compliance/breaches/{breachID}/override` | Override Compliance Breach | IRG_OVERRIDE_BREACH | Implemented |
| POST | `/api/v1/compliance/checks/post-trade` | Run Post-Trade Compliance Check | WORKFLOW_EXECUTE | Implemented |
| POST | `/api/v1/compliance/checks/pre-trade` | Run Pre-Trade Compliance Check | WORKFLOW_EXECUTE | Implemented |
| GET | `/api/v1/compliance/checks/{groupID}` | Get Compliance Check Group | IRG_VIEW_RULES | Implemented |
| GET | `/api/v1/compliance/rules` | List Compliance Rule Instances | IRG_VIEW_RULES | Implemented |
| POST | `/api/v1/compliance/rules` | Create Compliance Rule Instance | IRG_EDIT_RULE_INSTANCE | Implemented |

### IAM administration

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/admin/sessions/{id}/revoke` | Revoke Any Session (Admin) | IAM_USER_UPDATE | Implemented |
| GET | `/api/v1/admin/users` | List Users | IAM_USER_VIEW | Implemented |
| POST | `/api/v1/admin/users` | Create User | IAM_USER_CREATE | Implemented |
| POST | `/api/v1/admin/users/{id}/disable` | Disable User | IAM_USER_DEACTIVATE | Implemented |
| POST | `/api/v1/admin/users/{id}/enable` | Enable User | IAM_USER_DEACTIVATE | Implemented |
| POST | `/api/v1/admin/users/{id}/lock` | Lock User | IAM_USER_UPDATE | Implemented |
| POST | `/api/v1/admin/users/{id}/reset-password` | Admin Password Reset | IAM_USER_UPDATE | Implemented |
| GET | `/api/v1/admin/users/{id}/sessions` | List User Sessions (Admin) | IAM_USER_UPDATE | Implemented |
| POST | `/api/v1/admin/users/{id}/unlock` | Unlock User | IAM_USER_UPDATE | Implemented |

### Integration

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/integration/dashboard/me` | Personal dashboard snapshot | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |
| GET | `/api/v1/integration/dashboard/valuation-summary` | Dashboard AUM / P&L summary | Authenticated; company aggregate is permission-independent, mine is data-scoped | Implemented |
| GET | `/api/v1/integration/tasks/my` | Personal task list | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |
| GET | `/api/v1/integration/tasks/my/summary` | Personal task summary | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |

### Investment V1

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/investment/decisions` | List Investment Decisions | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v1/investment/decisions` | Create Investment Decision | INVESTMENT_DECISION_MANAGE | Implemented |
| GET | `/api/v1/investment/decisions/approval-items` | List Investment Decision Approval Items | INVESTMENT_DECISION_APPROVE | Implemented |
| POST | `/api/v1/investment/decisions/batch-approve` | Batch Approve Investment Decisions | INVESTMENT_DECISION_APPROVE | Implemented |
| POST | `/api/v1/investment/decisions/batch-reject` | Batch Reject Investment Decisions | INVESTMENT_DECISION_APPROVE | Implemented |
| GET | `/api/v1/investment/decisions/{id}` | Get Investment Decision | INVESTMENT_DECISION_VIEW | Implemented |
| PUT | `/api/v1/investment/decisions/{id}` | Update Investment Decision | INVESTMENT_DECISION_MANAGE | Implemented |
| POST | `/api/v1/investment/decisions/{id}/cancel` | Cancel Investment Decision | INVESTMENT_DECISION_CANCEL | Implemented |
| GET | `/api/v1/investment/decisions/{id}/details` | Get Investment Decision With Lines | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v1/investment/decisions/{id}/submit` | Submit Investment Decision | INVESTMENT_DECISION_SUBMIT | Implemented |
| GET | `/api/v1/investment/executions` | List Trade Executions | INVESTMENT_EXECUTION_VIEW | Implemented |
| POST | `/api/v1/investment/executions` | Create Investment Execution | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | `/api/v1/investment/executions/{id}` | Get Trade Execution | INVESTMENT_EXECUTION_VIEW | Implemented |
| POST | `/api/v1/investment/executions/{id}/cancel` | Cancel Trade Execution | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | `/api/v1/investment/executions/{id}/fill` | Fill Trade Execution | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | `/api/v1/investment/funds` | List Funds | INVESTMENT_FUND_VIEW | Implemented |
| POST | `/api/v1/investment/funds` | Create Fund | INVESTMENT_FUND_MANAGE | Implemented |
| GET | `/api/v1/investment/funds/{id}` | Get Fund | INVESTMENT_FUND_VIEW | Implemented |
| PUT | `/api/v1/investment/funds/{id}` | Update Fund | INVESTMENT_FUND_MANAGE | Implemented |
| DELETE | `/api/v1/investment/funds/{id}` | Delete Fund | INVESTMENT_FUND_MANAGE | Implemented |
| GET | `/api/v1/investment/funds/{id}/allocation` | Get Fund Allocation | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | `/api/v1/investment/funds/{id}/aum/compute` | Compute Fund AUM | INVESTMENT_FUND_AUM_COMPUTE | Implemented |
| GET | `/api/v1/investment/funds/{id}/holdings/valuation` | Get Mark-to-Market Holdings Valuation | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | `/api/v1/investment/funds/{id}/market-data/refresh` | Refresh Fund Market Data | INVESTMENT_VALUATION_RUN | Implemented |
| GET | `/api/v1/investment/funds/{id}/market-data/status` | Get Fund Market Data Status | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v1/investment/funds/{id}/nav-history` | Get Fund NAV History | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v1/investment/funds/{id}/nav/latest` | Get Latest Fund NAV | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v1/investment/instruments` | List Instruments | INVESTMENT_INSTRUMENT_VIEW | Implemented |
| POST | `/api/v1/investment/instruments` | Create Instrument | INVESTMENT_INSTRUMENT_MANAGE | Implemented |
| GET | `/api/v1/investment/instruments/{id}` | Get Instrument | INVESTMENT_INSTRUMENT_VIEW | Implemented |
| PUT | `/api/v1/investment/instruments/{id}` | Update Instrument | INVESTMENT_INSTRUMENT_MANAGE | Implemented |
| POST | `/api/v1/investment/instruments/{id}/prices` | Post Price Snapshot | INVESTMENT_PRICE_POST | Implemented |
| GET | `/api/v1/investment/portfolios` | List Portfolios | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | `/api/v1/investment/portfolios` | Create Portfolio | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| GET | `/api/v1/investment/portfolios/{id}` | Get Portfolio | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| PUT | `/api/v1/investment/portfolios/{id}` | Update Portfolio | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| DELETE | `/api/v1/investment/portfolios/{id}` | Delete Portfolio | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| GET | `/api/v1/investment/portfolios/{id}/cash` | List Portfolio Cash Balances | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v1/investment/portfolios/{id}/holdings` | List Portfolio Holdings | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v1/investment/portfolios/{id}/transactions` | List Portfolio Transactions | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | `/api/v1/investment/portfolios/{id}/transactions` | Post Portfolio Transaction | INVESTMENT_LEDGER_POST | Implemented |
| POST | `/api/v1/investment/portfolios/{id}/transactions/simulate` | Simulate Portfolio Transaction | INVESTMENT_LEDGER_SIMULATE | Implemented |
| POST | `/api/v1/investment/portfolios/{id}/transactions/{txnId}/reverse` | Reverse Portfolio Transaction | INVESTMENT_LEDGER_REVERSE | Implemented |
| GET | `/api/v1/investment/portfolios/{id}/valuations` | List Portfolio Valuations | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v1/investment/portfolios/{id}/valuations/latest` | Get Latest Portfolio Valuation | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | `/api/v1/investment/portfolios/{id}/valuations/run` | Run Portfolio Valuation | INVESTMENT_VALUATION_RUN | Implemented |
| GET | `/api/v1/investment/reference/asset-classes` | List Asset Classes | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/asset-subtypes` | List Asset Subtypes | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/countries` | List Countries | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/fund-categories` | List Fund Categories | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/investment-styles` | List Investment Styles | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/regions` | List Regions | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/reference/sectors` | List Sectors | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | `/api/v1/investment/research-reports` | List Investment Research Reports | INVESTMENT_RESEARCH_VIEW | Implemented |
| POST | `/api/v1/investment/research-reports` | Create Investment Research Report | INVESTMENT_RESEARCH_CREATE | Implemented |
| GET | `/api/v1/investment/research-reports/{id}` | Get Investment Research Report | INVESTMENT_RESEARCH_VIEW | Implemented |
| PUT | `/api/v1/investment/research-reports/{id}` | Update Investment Research Report | INVESTMENT_RESEARCH_UPDATE | Implemented |
| DELETE | `/api/v1/investment/research-reports/{id}` | Delete Investment Research Report | INVESTMENT_RESEARCH_DELETE | Implemented |
| POST | `/api/v1/investment/research-reports/{id}/cancel-submit` | Cancel Submission Of Investment Research Report | INVESTMENT_RESEARCH_CANCEL_SUBMIT | Implemented |
| POST | `/api/v1/investment/research-reports/{id}/invalidate` | Invalidate Investment Research Report | INVESTMENT_RESEARCH_INVALIDATE | Implemented |
| POST | `/api/v1/investment/research-reports/{id}/submit` | Submit Investment Research Report | INVESTMENT_RESEARCH_SUBMIT | Implemented |
| GET | `/api/v1/investment/trade-confirmations` | List Trade Confirmations | INVESTMENT_CONFIRMATION_VIEW | Implemented |
| POST | `/api/v1/investment/trade-confirmations` | Record Trade Confirmation | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| POST | `/api/v1/investment/trade-confirmations/batch` | Import Trade Confirmation Batch | INVESTMENT_CONFIRMATION_IMPORT | Implemented |
| GET | `/api/v1/investment/trade-confirmations/{id}` | Get Trade Confirmation | INVESTMENT_CONFIRMATION_VIEW | Implemented |
| POST | `/api/v1/investment/trade-confirmations/{id}/resolve` | Resolve Trade Confirmation | INVESTMENT_CONFIRMATION_MANAGE | Implemented |

### Market Data

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/market-data/history` | Get Market Price History | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/market-data/import` | Import Market Data | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/market-data/import-batches` | Create Market Data Import Batch | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/import-batches/{batch_id}` | Get Market Data Import Batch | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/import-batches/{batch_id}/errors` | List Market Data Import Batch Errors | Authenticated; no function-permission middleware | Implemented |
| POST | `/api/v1/market-data/import-batches/{batch_id}/run` | Run Market Data Import Batch | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/provider-health` | Get Market Data Provider Health | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/quote` | Get Market Quote | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/screen/search` | Market Data screen — search | Authenticated; no function-permission middleware | Implemented |
| GET | `/api/v1/market-data/screen/watchlist` | Market Data screen — watchlist | Authenticated; no function-permission middleware | Implemented |

### Notification

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/notifications` | List my in-app notifications | Authenticated; current-user scope | Implemented |
| GET | `/api/v1/notifications/email-outbox` | List email outbox | NOTIFICATION_VIEW | Implemented |
| GET | `/api/v1/notifications/email-outbox/{outbox_id}` | Get email outbox detail | NOTIFICATION_VIEW | Implemented |
| POST | `/api/v1/notifications/email-outbox/{outbox_id}/retry` | Retry failed email | NOTIFICATION_RETRY | Implemented |
| GET | `/api/v1/notifications/email/health` | Email health | NOTIFICATION_HEALTH | Implemented |
| POST | `/api/v1/notifications/email/test` | Send test email | NOTIFICATION_TEST | Implemented |
| POST | `/api/v1/notifications/read-all` | Mark all notifications as read | Authenticated; current-user scope | Implemented |
| POST | `/api/v1/notifications/{id}/read` | Mark one notification as read | Authenticated; current-user scope | Implemented |

### Permissions

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/permissions/approval-settings` | List permission approval settings | permission.approval_settings.view | Implemented |
| GET | `/api/v1/permissions/approval-settings/{id}` | Get permission approval setting | permission.approval_settings.view | Implemented |
| GET | `/api/v1/permissions/change-requests` | List permission change requests | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests` | Create a permission change request | permission.change_request.create | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}` | Get a permission change request | permission.change_request.review | Implemented |
| PUT | `/api/v1/permissions/change-requests/{id}` | Update a permission change request | permission.change_request.create | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}/approval-steps` | List approval steps | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/approve` | Approve a permission approval step | permission.change_request.approve | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/reject` | Reject a permission approval step | permission.change_request.reject | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/approval-steps/{stepId}/request-changes` | Request changes on a permission approval step | permission.change_request.approve | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/approve` | Approve a permission change request | permission.change_request.approve | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/cancel` | Cancel a permission change request | permission.change_request.cancel | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}/checks` | List change request checks | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/close` | Close a permission change request | permission.change_request.close | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}/comments` | List change request comments | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/comments` | Add a change request comment | permission.change_request.review | Implemented |
| PUT | `/api/v1/permissions/change-requests/{id}/comments/{commentId}` | Update a change request comment | permission.change_request.review | Implemented |
| DELETE | `/api/v1/permissions/change-requests/{id}/comments/{commentId}` | Delete a change request comment | permission.change_request.review | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}/diff` | Get change request diff | permission.change_request.review | Implemented |
| GET | `/api/v1/permissions/change-requests/{id}/items` | List change request items | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/items` | Add a change request item | permission.change_request.create | Implemented |
| PUT | `/api/v1/permissions/change-requests/{id}/items/{itemId}` | Update a change request item | permission.change_request.create | Implemented |
| DELETE | `/api/v1/permissions/change-requests/{id}/items/{itemId}` | Delete a change request item | permission.change_request.create | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/labels` | Add a label to a change request | permission.change_request.create | Implemented |
| DELETE | `/api/v1/permissions/change-requests/{id}/labels/{labelId}` | Remove a label from a change request | permission.change_request.create | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/merge` | Merge an approved permission change request | permission.change_request.merge | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/reject` | Reject a permission change request | permission.change_request.reject | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/request-changes` | Request changes on a permission change request | permission.change_request.approve | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/rerun-checks` | Rerun change request checks | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/change-requests/{id}/submit` | Submit a permission change request | permission.change_request.submit | Implemented |
| GET | `/api/v1/permissions/data-rights` | List data rights | permission.data_rights.view | Implemented |
| GET | `/api/v1/permissions/effective/users/{userId}` | Get effective permissions for a user | permission.users.view | Implemented |
| GET | `/api/v1/permissions/function-definitions` | List function definitions | permission.function_rights.view | Implemented |
| GET | `/api/v1/permissions/function-rights` | List function rights | permission.function_rights.view | Implemented |
| GET | `/api/v1/permissions/groups` | List permission groups | permission.groups.view | Implemented |
| GET | `/api/v1/permissions/groups/{id}` | Get permission group detail | permission.groups.view | Implemented |
| GET | `/api/v1/permissions/labels` | List permission labels | permission.change_request.review | Implemented |
| POST | `/api/v1/permissions/labels` | Create or update a permission label | permission.change_request.create | Implemented |
| PUT | `/api/v1/permissions/labels/{id}` | Update a permission label | permission.change_request.create | Implemented |
| GET | `/api/v1/permissions/notification-settings` | List permission notification settings | permission.notification.view | Implemented |
| PUT | `/api/v1/permissions/notification-settings` | Update permission notification settings | permission.notification.edit | Implemented |
| GET | `/api/v1/permissions/role-assignment-policies` | List role assignment policies | permission.roles.view | Implemented |
| GET | `/api/v1/permissions/roles` | List roles | permission.roles.view | Implemented |
| GET | `/api/v1/permissions/roles/{id}` | Get a role | permission.roles.view | Implemented |
| GET | `/api/v1/permissions/users` | List permission users | permission.users.view | Implemented |
| GET | `/api/v1/permissions/users/{id}` | Get permission user detail | permission.users.view | Implemented |
| POST | `/api/v1/permissions/users/{userId}/role-assignment-request` | Create a role assignment request | permission.change_request.create | Implemented |

### Reference Data

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

### Watchlist

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/watchlists` | List watchlist items | WATCHLIST_VIEW + owner/portfolio data scope | Implemented |
| GET | `/api/v1/watchlists/alerts` | List alert events | WATCHLIST_VIEW + owner/portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/alerts/{id}/acknowledge` | Acknowledge an alert event | WATCHLIST_ALERT_ACK + owner/portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/evaluate` | Manually trigger rule evaluation | WATCHLIST_EVALUATE + optional portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/items` | Create a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |
| PATCH | `/api/v1/watchlists/items/{id}` | Update a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |
| DELETE | `/api/v1/watchlists/items/{id}` | Delete a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |

### Workflow

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/workflow/daily` | Get Daily Workflow State | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| POST | `/api/v1/workflow/daily/execute` | Execute Daily Workflow Transition | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| GET | `/api/v1/workflow/daily/transitions` | Get Daily Transition History | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| GET | `/api/v1/workflow/day-states/{contractId}` | Get Workflow Day State | WORKFLOW_VIEW | Implemented |
| GET | `/api/v1/workflow/day-states/{contractId}/history` | Get Workflow Transition History | WORKFLOW_VIEW | Implemented |
| POST | `/api/v1/workflow/day-states/{contractId}/transitions` | Execute Workflow Transition | Per-action workflow permission | Implemented |
| POST | `/api/v1/workflow/scheduler/run-once` | Run Workflow Scheduler Once | WORKFLOW_RUN_SCHEDULER | Implemented |
| GET | `/api/v1/workflow/settings` | Get Workflow Approval Settings | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |
| PUT | `/api/v1/workflow/settings` | Update Workflow Approval Settings | Admin role required by handler; no RequirePermission middleware | Implemented |
| GET | `/api/v1/workflow/transition-rules` | Get Workflow Transition Rules | Settings-based approver/admin rule; no RequirePermission middleware | Implemented |

### Portfolio V2

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v2/portfolios/{portfolioCode}` | Get Portfolio By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/cash` | Get Portfolio Cash Balances By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/compliance/breaches` | List Compliance Breaches For Portfolio (V2) | IRG_VIEW_RULES | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/checks/post-trade` | Run Portfolio Post-Trade Compliance Check (V2) | WORKFLOW_EXECUTE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/checks/pre-trade` | Run Portfolio Pre-Trade Compliance Check (V2) | WORKFLOW_EXECUTE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/compliance/rules` | List Compliance Rules For Portfolio (V2) | IRG_VIEW_RULES | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings` | Bind Compliance Rule To Portfolio (V2) | IRG_EDIT_BINDING | Implemented |
| DELETE | `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}` | Deactivate Portfolio Compliance Rule Binding (V2) | IRG_EDIT_BINDING | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve` | Resolve Portfolio Trade Confirmation By Code | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/decisions` | List Portfolio Decisions By Code | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions` | Create Portfolio Decision By Code | INVESTMENT_DECISION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}` | Get Portfolio Decision By Code | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/cancel` | Cancel Portfolio Decision By Code | INVESTMENT_DECISION_CANCEL | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions` | Create Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/submit` | Submit Portfolio Decision By Code | INVESTMENT_DECISION_SUBMIT | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/cancel` | Cancel Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations` | Record Portfolio Trade Confirmation By Code | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/fill` | Fill Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/holdings` | Get Portfolio Holdings By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/transactions` | List Portfolio Transactions By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions` | Post Portfolio Transaction By Code | INVESTMENT_LEDGER_POST | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions/simulate` | Simulate Portfolio Transaction By Code | INVESTMENT_LEDGER_SIMULATE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse` | Reverse Portfolio Transaction By Code | INVESTMENT_LEDGER_REVERSE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/valuations` | List Portfolio Valuations By Code | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/valuations/latest` | Get Latest Portfolio Valuation By Code | INVESTMENT_VALUATION_VIEW | Implemented |

## Source references

- `backend/cmd/server/main.go`
- `backend/internal/*/module.go`
- `backend/internal/*/transport/router.go`
- `backend/internal/*/transport/http/router.go`
- `backend/docs/swagger.json`
- `backend/docs/v2/v2_swagger.json`
- `docs/api/_build_api_docs.py` for registered V1 routes missing from Swagger

## Regeneration

Run `python -B docs/api/_build_current_api_docs.py` from the repository root after regenerating Swagger. The script writes only the current-code reference files listed in its `main` function and does not rewrite the legacy module documents.
