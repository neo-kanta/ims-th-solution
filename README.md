# IMS Thailand Solution

Investment Management Platform for Thai Market

> **Status**: Proof of Concept (PoC)

IMS Thailand is a modern web application for investment management, featuring a Go-based microservices architecture and a Nuxt.js frontend. 

## 🏗 Architecture

- **Backend**: Go (1.25+), `go-chi/chi` routing, `jackc/pgx` for PostgreSQL access, and JWT authentication.
- **Frontend**: Nuxt.js, Vue 3, Pinia (State Management), and Tailwind CSS.
- **Database**: PostgreSQL 16.
- **Infrastructure**: Docker & Docker Compose for local development and deployment.

## 🚀 Getting Started

The project uses a `Makefile` to simplify local development and operations. Make sure you have **Docker**, **Go**, and **Node.js/npm** installed.

### 1. Start the Environment

To start the database, backend API, and frontend server:

```bash
make dev
```
- Frontend will be available at: http://localhost:3000
- Backend API will be available at: http://localhost:8080
- Swagger UI will be available at: http://localhost:8080/api/v1/swagger/index.html

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

The development environment is seeded with the following accounts:
- **Admin User**: `admin` / `admin123`
- **Guest User**: `guest` / `guest123`

*(Note: These are only for local development and must differ in production).*

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
├── backend/            # Go backend application
│   ├── cmd/            # Entry points (server, migrate, tools)
│   ├── internal/       # Domain modules (iam, approval, workflow, etc.)
│   └── platform/       # Cross-cutting concerns (config, db, middleware, logging)
├── frontend/           # Nuxt.js frontend application
│   ├── app/            # Pages, layouts, plugins, middleware
│   └── shared/         # Shared stores, components, composables, API clients
├── database/           # SQL Migrations and Seed data
├── infra/              # Docker and Environment configurations
├── docs/               # Project documentation and architecture records
└── scripts/            # Shell/utility scripts
```

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
