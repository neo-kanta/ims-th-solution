-- =============================================================================
-- Workflow seed: IMS investment day workflow
-- =============================================================================
-- Source model:
--   Investment Day-start
--     -> investment workbench subflow
--        (Analysis Report, Investment Decisions, Save,
--         Investment Execution, Investment Review,
--         Confirmation and Settlement, Review)
--     -> Supervisor Approval
--     -> Transaction Day-end
--     -> Accounting Day-end
--
-- Backend state mapping:
--   Investment Day-start  -> DAY_OPEN
--   Supervisor Approval   -> MANAGER_APPROVED
--   Transaction Day-end   -> TRANSACTION_CLOSED
--   Accounting Day-end    -> ACCOUNTING_CLOSED
--
-- Demo contract IDs are stable so Bruno/local API calls can target known data.
-- =============================================================================

BEGIN;

-- Grant the concrete workflow permissions used by the HTTP handlers.
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('WORKFLOW_VIEW'),
    ('WORKFLOW_OPEN_DAY'),
    ('WORKFLOW_APPROVE'),
    ('WORKFLOW_CANCEL_DAY_START'),
    ('WORKFLOW_CANCEL_APPROVAL'),
    ('WORKFLOW_CLOSE_TRANSACTIONS'),
    ('WORKFLOW_CANCEL_TRANSACTION_CLOSE'),
    ('WORKFLOW_CLOSE_ACCOUNTING'),
    ('WORKFLOW_ROLLBACK_ACCOUNTING_CLOSE'),
    ('WORKFLOW_RUN_SCHEDULER')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET
    is_granted = EXCLUDED.is_granted,
    created_by = COALESCE(permissions_function_rights.created_by, EXCLUDED.created_by);

-- Optional operational groups reflecting the workflow model lanes.
INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES
    (
        'b0000000-0000-0000-0000-000000000010',
        'Workflow Operator',
        'Operators who perform investment day-start and day-end workflow actions',
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'b0000000-0000-0000-0000-000000000011',
        'Fund Manager',
        'Fund managers who approve reviewed investment activity before transaction closing',
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO UPDATE
SET
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('WORKFLOW_VIEW'),
    ('WORKFLOW_OPEN_DAY'),
    ('WORKFLOW_CLOSE_TRANSACTIONS'),
    ('WORKFLOW_CLOSE_ACCOUNTING'),
    ('WORKFLOW_RUN_SCHEDULER')
) AS p(code)
WHERE g.name = 'Workflow Operator'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('WORKFLOW_VIEW'),
    ('WORKFLOW_APPROVE'),
    ('WORKFLOW_CANCEL_APPROVAL')
) AS p(code)
WHERE g.name = 'Fund Manager'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

WITH seed_days (
    id,
    contract_id,
    business_date,
    current_state,
    opened_at,
    opened_by,
    manager_approved_at,
    manager_approved_by,
    transactions_locked_at,
    transaction_closed_at,
    transaction_closed_by,
    accounting_closed_at,
    accounting_closed_by,
    version,
    created_at,
    updated_at,
    created_by,
    updated_by
) AS (
    VALUES
        (
            'd0000000-0000-0000-0000-000000000001'::uuid,
            'c0000000-0000-0000-0000-000000000001'::uuid,
            DATE '2026-04-24',
            'DAY_OPEN',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::uuid,
            1,
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            'd0000000-0000-0000-0000-000000000002'::uuid,
            'c0000000-0000-0000-0000-000000000002'::uuid,
            DATE '2026-04-24',
            'MANAGER_APPROVED',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::uuid,
            2,
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            'd0000000-0000-0000-0000-000000000003'::uuid,
            'c0000000-0000-0000-0000-000000000003'::uuid,
            DATE '2026-04-24',
            'TRANSACTION_CLOSED',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            TIMESTAMPTZ '2026-04-24 10:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NULL::timestamptz,
            NULL::uuid,
            3,
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            TIMESTAMPTZ '2026-04-24 10:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            'd0000000-0000-0000-0000-000000000004'::uuid,
            'c0000000-0000-0000-0000-000000000004'::uuid,
            DATE '2026-04-24',
            'ACCOUNTING_CLOSED',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            TIMESTAMPTZ '2026-04-24 10:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            TIMESTAMPTZ '2026-04-24 11:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            4,
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            TIMESTAMPTZ '2026-04-24 11:30:00+00',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        )
)
INSERT INTO workflow__day_states (
    id,
    contract_id,
    business_date,
    current_state,
    opened_at,
    opened_by,
    manager_approved_at,
    manager_approved_by,
    transactions_locked_at,
    transaction_closed_at,
    transaction_closed_by,
    accounting_closed_at,
    accounting_closed_by,
    pending_reclose,
    reclose_count,
    version,
    created_at,
    updated_at,
    created_by,
    updated_by
)
SELECT
    id,
    contract_id,
    business_date,
    current_state,
    opened_at,
    opened_by,
    manager_approved_at,
    manager_approved_by,
    transactions_locked_at,
    transaction_closed_at,
    transaction_closed_by,
    accounting_closed_at,
    accounting_closed_by,
    false,
    0,
    version,
    created_at,
    updated_at,
    created_by,
    updated_by
FROM seed_days
ON CONFLICT (contract_id, business_date) DO UPDATE
SET
    current_state = EXCLUDED.current_state,
    opened_at = EXCLUDED.opened_at,
    opened_by = EXCLUDED.opened_by,
    manager_approved_at = EXCLUDED.manager_approved_at,
    manager_approved_by = EXCLUDED.manager_approved_by,
    transactions_locked_at = EXCLUDED.transactions_locked_at,
    transaction_closed_at = EXCLUDED.transaction_closed_at,
    transaction_closed_by = EXCLUDED.transaction_closed_by,
    accounting_closed_at = EXCLUDED.accounting_closed_at,
    accounting_closed_by = EXCLUDED.accounting_closed_by,
    pending_reclose = EXCLUDED.pending_reclose,
    reclose_count = EXCLUDED.reclose_count,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at,
    updated_by = EXCLUDED.updated_by;

WITH day_lookup AS (
    SELECT id AS workflow_day_id, contract_id, business_date
    FROM workflow__day_states
    WHERE business_date = DATE '2026-04-24'
      AND contract_id IN (
          'c0000000-0000-0000-0000-000000000001'::uuid,
          'c0000000-0000-0000-0000-000000000002'::uuid,
          'c0000000-0000-0000-0000-000000000003'::uuid,
          'c0000000-0000-0000-0000-000000000004'::uuid
      )
),
seed_transitions (
    id,
    contract_id,
    from_state,
    to_state,
    action,
    occurred_at,
    metadata
) AS (
    VALUES
        (
            'e0000000-0000-0000-0000-000000000101'::uuid,
            'c0000000-0000-0000-0000-000000000001'::uuid,
            'NOT_STARTED',
            'DAY_OPEN',
            'OPEN_DAY',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Investment Day-start","modelStepThai":"ลงทุนเริ่ม","investmentWorkbench":["Analysis Report","Investment Decisions","Save","Investment Execution","Investment Review","Confirmation and Settlement","Review"]}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000201'::uuid,
            'c0000000-0000-0000-0000-000000000002'::uuid,
            'NOT_STARTED',
            'DAY_OPEN',
            'OPEN_DAY',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Investment Day-start","modelStepThai":"ลงทุนเริ่ม"}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000202'::uuid,
            'c0000000-0000-0000-0000-000000000002'::uuid,
            'DAY_OPEN',
            'MANAGER_APPROVED',
            'APPROVE',
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Supervisor Approval","modelStepThai":"主管放行","transactionCount":3,"hasPendingUnreviewed":false,"zeroTransactionAttestation":false}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000301'::uuid,
            'c0000000-0000-0000-0000-000000000003'::uuid,
            'NOT_STARTED',
            'DAY_OPEN',
            'OPEN_DAY',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Investment Day-start","modelStepThai":"ลงทุนเริ่ม"}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000302'::uuid,
            'c0000000-0000-0000-0000-000000000003'::uuid,
            'DAY_OPEN',
            'MANAGER_APPROVED',
            'APPROVE',
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Supervisor Approval","modelStepThai":"主管放行","transactionCount":4,"hasPendingUnreviewed":false,"zeroTransactionAttestation":false}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000303'::uuid,
            'c0000000-0000-0000-0000-000000000003'::uuid,
            'MANAGER_APPROVED',
            'TRANSACTION_CLOSED',
            'CLOSE_TRANSACTIONS',
            TIMESTAMPTZ '2026-04-24 10:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Transaction Day-end","modelStepThai":"交易關帳","postTradeGate":"cleared"}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000401'::uuid,
            'c0000000-0000-0000-0000-000000000004'::uuid,
            'NOT_STARTED',
            'DAY_OPEN',
            'OPEN_DAY',
            TIMESTAMPTZ '2026-04-24 01:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Investment Day-start","modelStepThai":"ลงทุนเริ่ม"}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000402'::uuid,
            'c0000000-0000-0000-0000-000000000004'::uuid,
            'DAY_OPEN',
            'MANAGER_APPROVED',
            'APPROVE',
            TIMESTAMPTZ '2026-04-24 09:00:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Supervisor Approval","modelStepThai":"主管放行","transactionCount":5,"hasPendingUnreviewed":false,"zeroTransactionAttestation":false}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000403'::uuid,
            'c0000000-0000-0000-0000-000000000004'::uuid,
            'MANAGER_APPROVED',
            'TRANSACTION_CLOSED',
            'CLOSE_TRANSACTIONS',
            TIMESTAMPTZ '2026-04-24 10:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Transaction Day-end","modelStepThai":"交易關帳","postTradeGate":"cleared"}'::jsonb
        ),
        (
            'e0000000-0000-0000-0000-000000000404'::uuid,
            'c0000000-0000-0000-0000-000000000004'::uuid,
            'TRANSACTION_CLOSED',
            'ACCOUNTING_CLOSED',
            'CLOSE_ACCOUNTING',
            TIMESTAMPTZ '2026-04-24 11:30:00+00',
            '{"workflowModel":"IMS investment workflow","modelStep":"Accounting Day-end","modelStepThai":"會計關帳","navPostback":"seeded"}'::jsonb
        )
)
INSERT INTO workflow__transition_log (
    id,
    workflow_day_id,
    contract_id,
    business_date,
    from_state,
    to_state,
    action,
    actor_id,
    actor_type,
    actor_username,
    reason,
    metadata,
    occurred_at,
    request_id
)
SELECT
    st.id,
    dl.workflow_day_id,
    st.contract_id,
    dl.business_date,
    st.from_state,
    st.to_state,
    st.action,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'SYSTEM',
    'seed',
    NULL,
    st.metadata,
    st.occurred_at,
    'seed-workflow-model'
FROM seed_transitions st
JOIN day_lookup dl ON dl.contract_id = st.contract_id
ON CONFLICT (id) DO NOTHING;

WITH approved_days AS (
    SELECT id AS workflow_day_id, contract_id, business_date
    FROM workflow__day_states
    WHERE business_date = DATE '2026-04-24'
      AND contract_id IN (
          'c0000000-0000-0000-0000-000000000002'::uuid,
          'c0000000-0000-0000-0000-000000000003'::uuid,
          'c0000000-0000-0000-0000-000000000004'::uuid
      )
)
INSERT INTO workflow__approval_records (
    id,
    workflow_day_id,
    contract_id,
    business_date,
    approver_id,
    approver_username,
    approver_role,
    approval_status,
    is_zero_transaction,
    attestation_reason,
    approved_at,
    notes
)
SELECT
    CASE contract_id
        WHEN 'c0000000-0000-0000-0000-000000000002'::uuid THEN 'f0000000-0000-0000-0000-000000000002'::uuid
        WHEN 'c0000000-0000-0000-0000-000000000003'::uuid THEN 'f0000000-0000-0000-0000-000000000003'::uuid
        ELSE 'f0000000-0000-0000-0000-000000000004'::uuid
    END,
    workflow_day_id,
    contract_id,
    business_date,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'seed',
    'Fund Manager',
    'APPROVED',
    false,
    NULL,
    TIMESTAMPTZ '2026-04-24 09:00:00+00',
    'Seeded from the IMS workflow model: supervisor approval after investment review.'
FROM approved_days
ON CONFLICT (id) DO NOTHING;

COMMIT;
