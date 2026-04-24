# Module: approval

This module is a reserved backend boundary for approval workflow functionality.

## Intended Scope

- approval groups and teams
- configurable approval chains
- approval actions and decision history
- signature or attestation support

## Current Status

This module is scaffolded only.

Today it contains the standard folder layout plus placeholder files such as:

- `module.go`
- `domain/repository.go`
- `transport/router.go`

There is no implemented application logic, persistence layer, or mounted HTTP surface yet.

## Directory Contract

- `domain/` for aggregates, value objects, policies, and repository interfaces
- `application/` for commands, queries, and DTOs
- `infrastructure/` for adapters and persistence
- `transport/` for HTTP routing, request/response DTOs, and validation
- `jobs/` for scheduled/background work
- `permission/` for permission-code declarations

## Implementation Guidance

When this module becomes active:

1. keep the business rules here instead of leaking them into `workflow` or `investment`
2. add real route registration in `transport/router.go`
3. wire the module through `module.go`
4. update the root README once the module moves beyond scaffold status
