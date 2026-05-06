# Module: iam

The IAM module is the most implemented backend module in this repository.

It currently owns:

- login and token issuance
- refresh and logout flows
- MFA enrollment and verification support
- authenticated profile/session queries
- admin user lifecycle operations
- admin session management

## Current Status

This module is active and wired into the running backend.

Compared with the other `backend/internal/*` modules, `iam` is not just a scaffold. It already contains application logic, persistence implementations, transport handlers, tests, and module wiring.

## Directory Structure

```text
iam/
|-- application/
|   |-- command/          # Login, logout, admin actions, password flows
|   |-- dto/
|   |-- query/            # Me, users, sessions
|   `-- service/          # Authorization, session, MFA challenge helpers
|-- domain/
|   |-- entity/
|   |-- valueobject/
|   `-- repository.go
|-- infrastructure/
|   |-- adapter/
|   `-- persistence/
|-- jobs/
|-- permission/
|-- transport/
|   |-- dto/
|   `-- handler/
|-- module.go
`-- README.md
```

## Key Files

- `application/command/login.go`
- `application/command/admin_user.go`
- `application/command/refresh_token.go`
- `application/query/get_me.go`
- `application/query/list_users.go`
- `application/query/list_sessions.go`
- `transport/handler/auth_handler.go`
- `transport/handler/admin_handler.go`
- `transport/handler/session_handler.go`
- `transport/handler/mfa_handler.go`
- `module.go`

## Security Notes

- Password and credential rules live in the domain/value-object layer
- Access tokens are issued here and validated by platform middleware
- Session and MFA persistence live under `infrastructure/persistence/`
- Admin and auth-sensitive actions are the current source of most backend security behavior

## Boundary Notes

- audit trail querying, export, event persistence, and event definitions now belong to the dedicated `audit` module
- IAM emits audit records through an audit recorder port and remains responsible only for identity and access behavior

## System Actors

Two stable system users are seeded by migration `20260506000003_iam__seed_system_actors`:

| ID                                       | Username             | Purpose                                           |
| ---------------------------------------- | -------------------- | ------------------------------------------------- |
| `00000000-0000-0000-0000-000000000001`   | `system.market_data` | Phase 5 market-data ingestion writer              |
| `00000000-0000-0000-0000-000000000002`   | `system.scheduler`   | ValuationRunner and other scheduled-job invokers  |

Both rows carry `password_hash = '!system'` (a sentinel that cannot match any bcrypt-verified credential), `is_active = true`, and `force_password_change = false`. They MUST NOT be assigned to permission groups; their authority comes from the calling subsystem rather than from a function-permission grant. The bcrypt comparator rejects the sentinel hash as malformed, so login attempts against these usernames fail naturally without any special-case branch.
