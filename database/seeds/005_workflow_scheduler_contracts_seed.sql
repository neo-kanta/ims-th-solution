-- =============================================================================
-- Workflow scheduler contract bridge seed
-- =============================================================================
-- Temporary bridge data until a contract/fund master module owns active
-- contracts. Scheduler OPEN_DAY reads this table, not workflow history.

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
