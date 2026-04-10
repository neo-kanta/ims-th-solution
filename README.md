# IMS Thailand Solution

IMS Thailand is a modular investment management platform with a Go backend, a Nuxt frontend, PostgreSQL migrations, and local infrastructure for end-to-end development.

## Current State

- Backend: modular monolith under `backend/internal`
- Frontend: Nuxt 4 app under `frontend/app`
- Database: PostgreSQL schema managed through SQL migrations
- Local infrastructure: Docker Compose from `infra/`
- Most implemented domain area today: IAM, auth, admin user management, sessions, MFA, and audit-related admin APIs
- Other backend domain modules exist as reserved scaffolds and boundaries, not as fully implemented features yet

## Tech Stack

- Backend: Go, Chi, pgx, Swagger
- Frontend: Nuxt 4, Vue 3, Pinia, TypeScript
- Database: PostgreSQL
- Tooling: Docker Compose, Make, Vitest, Playwright

## Repository Layout

```text
ims-th-solution/
|-- backend/                  # Go API, modular monolith, Swagger generation
|-- database/                 # SQL migrations and seed files
|-- docs/                     # Runbooks, ADRs, security docs, API docs
|-- frontend/                 # Nuxt 4 web application
|-- infra/                    # Dockerfiles, compose files, env files
|-- tests/                    # End-to-end tests
|-- tools/                    # Supporting tooling such as the MCP server
|-- Makefile                  # Common local development commands
|-- CLAUDE.md                 # AI/dev assistant guidance
`-- AIREAD.md                 # Project context notes
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.23+
- Node.js 20+
- npm 10+

### Start the stack

From the repository root:

```bash
make dev
```

This starts PostgreSQL, the backend API, and the frontend dev server.

### Local URLs

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Common Commands

### Development

```bash
make dev
make dev-backend
make dev-frontend
```

### Database

```bash
make migrate-up
make migrate-down
make migrate-new module=iam name=add_mfa_fields
make seed
make db-reset
```

### Quality

```bash
make test
make test-unit
make test-integration
make test-e2e
make lint
```

### Build

```bash
make build
make docker-build
make swagger
```

## Module Status

| Module | Status | Notes |
| --- | --- | --- |
| `iam` | Active | Authentication, sessions, MFA, admin user flows, audit/admin APIs |
| `approval` | Scaffolded | Folder structure and router/repository placeholders only |
| `audit` | Scaffolded | Reserved boundary; current audit admin listing lives in IAM |
| `compliance` | Scaffolded | Reserved boundary only |
| `integration` | Scaffolded | Reserved boundary only |
| `investment` | Scaffolded | Reserved boundary only |
| `leave_delegation` | Scaffolded | Reserved boundary only |
| `market_data` | Scaffolded | Reserved boundary only |
| `notification` | Scaffolded | Reserved boundary only |
| `permissions` | Scaffolded | Reserved boundary; current permission checks are handled through IAM and platform middleware |
| `reference_data` | Scaffolded | Reserved boundary only |
| `workflow` | Scaffolded | Reserved boundary only |

## Frontend Notes

The frontend has been reorganized around real ownership boundaries:

- `app/features/` for feature-owned code
- `app/shared/ui/` for shared components
- `app/shared/i18n/` for translations and locale helpers
- `app/pages/` as thin route shells where possible

See [frontend/README.md](C:/Users/kanta/source/repos/ims-th-solution/frontend/README.md) for the frontend-specific layout and conventions.

## Database Notes

- Migration files live in [database/migrations](C:/Users/kanta/source/repos/ims-th-solution/database/migrations)
- Development seed SQL lives in `database/seeds/`
- The current seed file is a placeholder template, so do not assume built-in demo accounts unless you have added them yourself

See [database/migrations/README.md](C:/Users/kanta/source/repos/ims-th-solution/database/migrations/README.md) for migration workflow details.

## Additional Documentation

- [Frontend README](C:/Users/kanta/source/repos/ims-th-solution/frontend/README.md)
- [Migrations README](C:/Users/kanta/source/repos/ims-th-solution/database/migrations/README.md)
- [MCP Server README](C:/Users/kanta/source/repos/ims-th-solution/tools/mcp/ims-mcp-server/README.md)
- `docs/SECURITY_CHECKLIST.md`
- `docs/SECURITY_IMPLEMENTATION.md`
- `docs/API_DOCUMENTATION.md`

## License

MIT
