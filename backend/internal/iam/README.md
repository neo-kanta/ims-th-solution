# Module: IAM (Identity and Access Management)

The IAM module provides a standalone identity foundation for the IMS platform. It handles user authentication, session management, fine-grained authorization, and administrative user lifecycle.

## 🏛 Architecture (DDD)

The module follows a layered Domain-Driven Design (DDD) approach:

- **`domain/`**: The core of the module.
  - `entity/`: `User` (aggregate root), `Session`, `AuditEvent`.
  - `valueobject/`: `Credentials` (hashing/validation).
  - `repository.go`: Interfaces for data persistence.
- **`application/`**: Orchestrates domain logic to fulfill use cases.
  - `command/`: Write operations (Login, Refresh, Logout, Admin actions).
  - `query/`: Read-only operations (Get Profile, Permissions).
  - `token_service.go`: JWT generation and management.
- **`infrastructure/`**: Implementation of domain interfaces.
  - `persistence/`: PostgreSQL implementations for User, Session, and Audit repositories.
  - `adapter/`: Bridges to other modules or systems (e.g., `PermissionsFetcher`).
- **`transport/`**: External entry points (HTTP).
  - `handler/`: `AuthHandler` and `AdminHandler` mapping HTTP to application commands.
  - `dto/`: Request and Response data structures with Swagger annotations.
- **`module.go`**: The composition root that wires all dependencies.

## 🔐 Key Security Constraints

- **Passwords**: Hashed with `bcrypt` (cost 12), enforced minimum length (8), and common password blocklist.
- **Sessions**: JWT access tokens (15m) + Server-side Refresh Tokens with **Rotation** and **Token Family Breach Detection**.
- **Authorization**:
  - **Stateless + Stateful**: JWT signature validation is supplemented by a synchronous database "Active Status" recheck on every protected request.
  - **RBAC**: Function permissions (e.g., `IAM_ADMIN`) and contract-based Data Scopes.
- **Audit**: Every security-sensitive mutation (login, lock, password change) is recorded in an immutable audit log with IP and UserAgent tracing.

## 🚀 Key Components

- **`LoginCommand`**: Handles credentials, lockouts, session creation, and initial permission hydration.
- **`AdminUserCommand`**: Centralized administrative lifecycle (Create, Lock/Unlock, Disable/Enable, Reset Password).
- **`Auth middleware`**: Injected via the `iam.Module` implementing `UserStatusChecker` and `PermissionChecker` adapters.

## 📡 API Groups

- **Public**: `/auth/login`, `/auth/refresh`
- **Self-Service**: `/auth/me`, `/auth/logout`, `/auth/logout-all`, `/auth/change-password`
- **Admin**: `/admin/users/**` (Requires `IAM_ADMIN`)
