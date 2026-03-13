# Module: approval

Approval Workflow — approval groups, teams, configurable approval chains, digital signatures.

## Key Domain Rules

TODO: Document key domain rules and business constraints for this module.

## Layer Structure

- `domain/` — Entities, value objects, events, policies, repository interface
- `application/` — Command handlers (writes), query handlers (reads), DTOs
- `infrastructure/` — SQL persistence, external adapters
- `transport/` — HTTP handlers, request/response DTOs, validators, routes
- `jobs/` — Background/scheduled tasks
- `permission/` — Module permission code declarations
