# IMS Thailand Solution

Enterprise Investment Management System for Thai Market

> **Status**: Proof of Concept (PoC) | **Phase**: Active Development

IMS Thailand is a modern, modular web application for investment management operations. Built with Go for the backend and Nuxt.js 3 for the frontend, it provides a comprehensive platform for managing investment workflows, approvals, permissions, and compliance.

## 🏗 Architecture

- **Backend**: Go 1.23+, modular monolith with `go-chi/chi` routing, `jackc/pgx` for PostgreSQL, JWT authentication, and structured logging
- **Frontend**: Nuxt.js 3, Vue 3 Composition API, Pinia state management, Tailwind CSS, and i18n (EN/TH/ZH)
- **Database**: PostgreSQL 16 with migrations and seed data
- **Infrastructure**: Docker & Docker Compose for local development and deployment
- **Features**: Light/Dark theme support, multi-language UI, permission-based access control, audit logging

## ✨ Key Features

- **Workflow Management**: Day-start operations, contract closing, and workflow state management
- **Investment Operations**: 4-step investment decision and execution flow
- **Approval Workflows**: Role-based approval routing and delegation
- **Permission System**: Function-level and data-level (contract) permissions
- **Multi-Language UI**: English, Thai, and Traditional Chinese support
- **Theme Support**: Light and Dark theme with system preference detection
- **Audit Logging**: Comprehensive audit trail for compliance and traceability
- **Data Integration**: ETL capabilities for market data and reference data

## 🚀 Quick Start

The project uses a `Makefile` to simplify local development. Make sure you have **Docker**, **Docker Compose**, **Go 1.23+**, **Node.js 18+**, and **npm** installed.

### 1. Start Everything

To start the database, backend API, and frontend server:

```bash
make dev
```

**Available at:**

- 🌐 Frontend: http://localhost:3000
- 🔌 Backend API: http://localhost:8080
- 📚 API Docs: http://localhost:8080/swagger/index.html#/

### 2. Database Management

Reset the database completely (stops containers, wipes data, runs migrations, and seeds):

```bash
make db-reset
```

Run pending migrations:

```bash
make migrate-up
```

Load initial seed data (includes dev accounts `guest` and `admin`):

```bash
make seed
```

### 3. Development Accounts

The development environment includes pre-configured user accounts:

| Username | Password   | Role          | Permissions                  |
| -------- | ---------- | ------------- | ---------------------------- |
| `admin`  | `admin123` | Administrator | All functions, all contracts |
| `guest`  | `guest123` | Viewer        | Limited read-only access     |

> ⚠️ **Important**: These credentials are for **local development only**. Use unique, strong passwords in production.

### 4. Other Useful Commands

**Backend Only:**

```bash
make dev-backend      # Run backend server without Docker
make build            # Build backend binary
```

**Frontend Only:**

```bash
make dev-frontend     # Run frontend dev server (requires backend)
make build-frontend   # Build frontend for production
```

**Database Operations:**

```bash
make migrate-new module=workflow name=add_feature    # Create new migration pair
make migrate-up                                      # Run pending migrations
make migrate-down                                    # Rollback last migration
make db-reset                                        # Full database reset
```

**Building & Deployment:**

```bash
make docker-build     # Build Docker images
make build            # Build backend binary + frontend static
```

## 🧪 Testing and Quality

Run unit and integration tests (Backend):

```bash
make test
```

Run E2E tests using Playwright:

```bash
make test-e2e
```

Run linters (Go and Frontend Typecheck):

```bash
make lint
```

## 🔒 Security

For a detailed code-first review of the current application security posture, please refer to the [Security Gap Analysis](docs/security_gap_analysis.md).

## 📂 Project Structure

```text
ims-th-solution/
├── backend/
│   ├── cmd/                     # Application entry points
│   │   ├── server/              # API server
│   │   ├── migrate/             # Database migrations
│   │   └── seed/                # Seed data
│   ├── internal/                # Domain modules (modular monolith)
│   │   ├── iam/                 # Identity & Access Management
│   │   ├── permissions/         # Permission system
│   │   ├── workflow/            # Workflow operations
│   │   ├── investment/          # Investment management
│   │   ├── approval/            # Approval workflows
│   │   ├── audit/               # Audit logging
│   │   ├── notification/        # Notifications
│   │   └── compliance/          # Compliance/IRG
│   ├── platform/                # Cross-cutting concerns
│   │   ├── middleware/          # HTTP middleware
│   │   ├── config/              # Configuration
│   │   ├── database/            # Database helpers
│   │   ├── logging/             # Structured logging
│   │   └── errors/              # Error handling
│   ├── pkg/                     # Shared packages
│   │   ├── types/               # Shared value types
│   │   ├── enum/                # Enumerations
│   │   └── contract/            # Inter-module interfaces
│   └── api/                     # OpenAPI/Swagger specs
│
├── frontend/
│   ├── app/
│       ├── pages/               # Nuxt pages
│       ├── layouts/             # Layout components
│       ├── middleware/          # Nuxt middleware
│       ├── plugins/             # Nuxt plugins
│       └── assets/              # CSS, images
│
├── database/
│   └── migrations/              # SQL migrations
│
├── infra/                       # Infrastructure
│   └── docker/                  # Docker Compose
│
├── docs/                        # Documentation
│   └── adr/                     # Architecture decisions
│
├── Makefile                     # Development commands
├── CLAUDE.md                    # Developer instructions
└── AIREAD.md                    # Project context
```

## 🎯 Core Modules

| Module           | Purpose                          | Status         |
| ---------------- | -------------------------------- | -------------- |
| **IAM**          | Authentication, Sessions, JWT    | ✅ In Progress |
| **Permissions**  | Access control (function & data) | ✅ In Progress |
| **Workflow**     | Day-start, operations, closing   | 🔄 Planning    |
| **Investment**   | 4-step investment flow           | 🔄 Planning    |
| **Approval**     | Approval routing                 | 🔄 Planning    |
| **Audit**        | Compliance logging               | 🔄 Planning    |
| **Notification** | User notifications               | 🔄 Planning    |

## 📜 License

MIT License

Copyright (c) 2026 neo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
