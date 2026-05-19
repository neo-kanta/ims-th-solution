# IMS-TH-SOLUTION ERD Notes

Generated from `database/migrations/*.up.sql` on 2026-05-18.

## Inventory

- Total tables: 58
- IAM: 7
- Permissions: 5
- Workflow: 9
- Compliance: 9
- Investment: 25
- Market data: 3

## Draw.io Pages

- 00 - Overview
- 01 - IAM + Permissions
- 02 - Workflow + Scheduler
- 03 - Compliance
- 04 - Investment Reference + Master
- 05 - Investment Process + Ledger
- 06 - Market Data
- 07 - Critical Notes

## Critical Things To Notice

### Fund.id is the contract_id
investment__funds.id is the cross-module contract key used by workflow, compliance, scheduler, and permissions data rights. Most of those links are intentionally not declared as FKs to keep module boundaries loose.

### Soft delete is not uniform
funds, portfolios, instruments, and research reports use partial unique indexes with deleted_at IS NULL. iam_users and permissions_groups still have full unique constraints, so soft-deleted usernames/group names cannot be reused without a migration.

### Append-only enforcement differs by module
investment ledger/snapshot tables use rejecting triggers. iam_audit_events uses rejecting triggers. compliance_check_records and compliance_overrides rely on privilege revokes, which do not stop the table owner. workflow transition_log is append-only by design comments, but not enforced by a DB trigger today.

### Polymorphic scope columns need application validation
scope_type/scope_id appears in compliance bindings, workflow settings/rules, investment process assignments, and AUM snapshots. The database cannot enforce these conditional references directly.

### Research reports are deliberately loose in PoC scope
owner_user_id, author_user_id, applicable_contract_id, and instrument_code have no FKs yet. This keeps the feature scaffold flexible, but it allows orphaned user/contract/instrument references.

### Permissions data rights has a type mismatch
permissions_data_rights.contract_id is VARCHAR(50), while the investment contract key is investment__funds.id UUID. This is likely legacy or an integration boundary, and should be revisited before hard FK enforcement.

### Market data is symbol-driven
market_symbols and market_data_snapshots are not linked to investment__instruments. Mapping is by ticker/provider symbols, so reconciliation logic must handle mismatches and provider aliases.

### JSONB carries important domain state
instrument attributes, compliance parameters/evidence/snapshots, workflow metadata, and market raw_payload are JSONB. They are flexible, but they move some schema guarantees from DB constraints into code/tests.

### Optimistic locking marks mutable projections
version columns appear on users, funds, portfolios, day_states, portfolio_positions, and cash_balances. These are mutable current-state projections and should update with compare-and-increment semantics.

### Rule versioning preserves reproducibility
compliance_rule_instances points at the current version, while check records pin rule_instance_version and parameter_snapshot. Historical compliance decisions should be replayable against the exact parameters used.

## Inferred / Non-FK Relationships

- workflow__day_states.contract_id -> investment__funds.id: contract_id = fund.id, no FK
- workflow__scheduler_contracts.contract_id -> investment__funds.id: scheduler bridge contract, no FK
- workflow__scheduler_contracts.fund_id -> investment__funds.id: bridge fund_id, no FK
- workflow__transition_log.contract_id -> investment__funds.id: denormalized contract, no FK
- workflow__approval_records.contract_id -> investment__funds.id: denormalized contract, no FK
- workflow__control_decisions.contract_id -> investment__funds.id: decision scope, no FK
- workflow__scheduler_run_items.contract_id -> investment__funds.id: scheduler target, no FK
- compliance_check_records.contract_id -> investment__funds.id: compliance contract, no FK
- compliance_breaches.contract_id -> investment__funds.id: breach contract, no FK
- compliance_check_records.portfolio_id -> investment__portfolios.id: compliance portfolio, no FK
- compliance_breaches.portfolio_id -> investment__portfolios.id: breach portfolio, no FK
- permissions_data_rights.contract_id -> investment__funds.id: VARCHAR contract grant, fund.id is UUID
- investment__aum_snapshots.scope_id -> investment__funds.id: when scope_type = FUND
- investment__aum_snapshots.scope_id -> investment__portfolios.id: when scope_type = PORTFOLIO
- investment__research_reports.owner_user_id -> iam_users.id: plain UUID in PoC
- investment__research_reports.author_user_id -> iam_users.id: plain UUID in PoC
- investment__research_reports.applicable_contract_id -> investment__funds.id: plain UUID in PoC
- investment__research_reports.instrument_code -> investment__instruments.primary_ticker: plain string in PoC
- market_symbols.symbol -> investment__instruments.primary_ticker: symbol mapping, no FK

