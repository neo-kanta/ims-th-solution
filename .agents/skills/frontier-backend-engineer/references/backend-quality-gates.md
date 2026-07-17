# Backend Quality Gates

Load this reference for implementation planning, migration review, or final
backend verification.

## Domain and Financial Logic

- Identify authoritative portfolio, fund, contract, order, actor, and
  business-date sources.
- Preserve decimal precision, currency, valuation mode, and effective date.
- Verify approval -> compliance -> execution -> posting lifecycle order.
- Prove rejection creates no partial financial state.
- Confirm PASS, WARN, BLOCK, override, and unavailable-control behavior.
- Do not invent regulation, policy, rating, allocation, NAV, or AUM thresholds.

## Application and Transactions

- Keep command/query responsibilities explicit.
- Define transaction boundaries around all state that must succeed atomically.
- Check duplicate requests, retries, idempotency keys, and optimistic or
  pessimistic concurrency behavior.
- Preserve audit evidence outside ambiguous partial-failure windows.
- Treat a mandatory production dependency as fail-closed when policy requires.

## Persistence and Migration

- Test existing rows against new constraints before migration.
- Review foreign keys, cascade behavior, indexes, NULL semantics, soft deletes,
  and partial uniqueness.
- Pair every up migration with an honest down migration.
- Rehearse on a disposable PostgreSQL database with representative data.
- Verify both schema and behavior after migration; refresh table snapshots.

## API and Security

- Validate JSON and UUID/decimal/date fields at the correct boundary.
- Map invalid input to 400, business rejection to the agreed 4xx, and unexpected
  failures to a scrubbed 500.
- Enforce function permission and data scope in the backend.
- Never expose domain entities directly or leak internal errors/secrets.
- Verify public API compatibility and regenerate Swagger from source.

## Minimum Evidence by Risk

| Risk | Minimum evidence |
| --- | --- |
| Local backend behavior | Focused test, module test, build, diff review |
| Shared contract or lifecycle | Regression tests, broad backend tests, generated contract review |
| Schema, permission, or financial write | Disposable-database/HTTP proof, rollback review, independent reviewer, no-write failure test |

## Completion Report

```text
Backend outcome:
Domain/data invariants:
API contract and error statuses:
Migrations and rollback semantics:
Commands run and results:
Independent review:
Not proved or deferred:
Git state:
Frontend/integration handoff:
```
