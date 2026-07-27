-- =============================================================================
-- Workflow scheduler contract bridge seed (development/test only)
-- =============================================================================
-- Formerly the always-run database/seeds/005_workflow_scheduler_contracts_seed.sql.
-- Moved here because every row it inserts is synthetic demo data: six
-- "IMS-DEMO-*" contracts (five active with no end date, one inactive/stale)
-- backing the temporary workflow__scheduler_contracts bridge table that
-- backend/internal/workflow/infrastructure/persistence/scheduler_repository.go
-- (PostgresWorkflowSchedulerContractSource.ListActiveContracts) reads to
-- decide which contracts the workflow scheduler opens/closes each business
-- day. Letting synthetic contracts reach the production scheduler is exactly
-- the class of defect this seed reorganization exists to close, so this file
-- now only runs when APP_ENV is development or test (see
-- backend/cmd/seed/sql_seeds.go).
--
-- Verified before moving this file: workflow__scheduler_contracts has no
-- foreign key pointing at it, and no other reference seed (checked:
-- 002_workflow_seed.sql, which seeds workflow__day_states/
-- __transition_log/__approval_records) joins or depends on these rows —
-- workflow__day_states.contract_id is a plain UUID column with a fixed
-- legacy placeholder value, not a foreign key to this table. Moving this
-- file does not break any reference seed.
--
-- Practical implication, intentional: with this file demo-only, a production
-- (or any non-opted-in) deployment now seeds ZERO rows into
-- workflow__scheduler_contracts. This is correct by design, not an oversight
-- — see database/seeds/README.md, "Production has no seeded scheduler
-- contracts by design": the table is an explicitly temporary bridge ("until a
-- contract/fund master module owns active contracts", per this file's own
-- original header) and production scheduling must be driven by real
-- contract/fund data once that module exists, never by synthetic demo rows.
-- =============================================================================

BEGIN;

INSERT INTO workflow__scheduler_contracts (
    id,
    contract_id,
    contract_code,
    contract_name,
    fund_id,
    fund_code,
    is_active,
    effective_from,
    effective_to,
    source,
    metadata,
    created_by,
    updated_by
)
VALUES
    (
        '78000000-0000-0000-0000-000000000001',
        'c0000000-0000-0000-0000-000000000001',
        'IMS-DEMO-001',
        'IMS demo contract - day open',
        'f0000000-0000-0000-0000-000000000001',
        'IMS-FUND-A',
        true,
        DATE '2026-01-01',
        NULL,
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"workflow"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '78000000-0000-0000-0000-000000000002',
        'c0000000-0000-0000-0000-000000000002',
        'IMS-DEMO-002',
        'IMS demo contract - manager approved',
        'f0000000-0000-0000-0000-000000000001',
        'IMS-FUND-A',
        true,
        DATE '2026-01-01',
        NULL,
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"workflow"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '78000000-0000-0000-0000-000000000003',
        'c0000000-0000-0000-0000-000000000003',
        'IMS-DEMO-003',
        'IMS demo contract - transaction closed',
        'f0000000-0000-0000-0000-000000000002',
        'IMS-FUND-B',
        true,
        DATE '2026-01-01',
        NULL,
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"workflow"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '78000000-0000-0000-0000-000000000004',
        'c0000000-0000-0000-0000-000000000004',
        'IMS-DEMO-004',
        'IMS demo contract - accounting closed',
        'f0000000-0000-0000-0000-000000000002',
        'IMS-FUND-B',
        true,
        DATE '2026-01-01',
        NULL,
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"workflow"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '78000000-0000-0000-0000-000000000005',
        'c0000000-0000-0000-0000-000000000005',
        'IMS-DEMO-005',
        'IMS demo brand-new active contract without workflow history',
        'f0000000-0000-0000-0000-000000000003',
        'IMS-FUND-C',
        true,
        DATE '2026-01-01',
        NULL,
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"brand_new_no_workflow_history"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '78000000-0000-0000-0000-000000000099',
        'c0000000-0000-0000-0000-000000000099',
        'IMS-DEMO-099',
        'IMS demo inactive stale contract',
        'f0000000-0000-0000-0000-000000000099',
        'IMS-FUND-Z',
        false,
        DATE '2025-01-01',
        DATE '2025-12-31',
        'TEMP_WORKFLOW_BRIDGE',
        '{"seed":"inactive_stale"}'::JSONB,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_wf_scheduler_contracts_contract DO UPDATE
SET
    contract_code = EXCLUDED.contract_code,
    contract_name = EXCLUDED.contract_name,
    fund_id = EXCLUDED.fund_id,
    fund_code = EXCLUDED.fund_code,
    is_active = EXCLUDED.is_active,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    source = EXCLUDED.source,
    metadata = EXCLUDED.metadata,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

COMMIT;
