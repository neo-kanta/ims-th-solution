-- =============================================================================
-- Demo seed - global Workflow day states + transition log
-- =============================================================================
-- Workflow is global per business_date. This demo therefore spreads example
-- states across adjacent dates instead of creating multiple contract-scoped
-- rows for the same date.
--
--   CURRENT_DATE     -> INVESTMENT_DAY_STARTED
--   CURRENT_DATE - 1 -> MANAGER_APPROVED_END_OF_DAY
--   CURRENT_DATE - 2 -> TRANSACTION_CLOSED
--   CURRENT_DATE - 3 -> ACCOUNTING_CLOSED
--
-- contract_id is stored as a fixed legacy placeholder only because current Go
-- entities still scan it as a UUID. It is not used as a workflow key.
-- =============================================================================

BEGIN;

WITH seed_days (
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
    created_by,
    updated_by
) AS (
    VALUES
        (
            CURRENT_DATE,
            'INVESTMENT_DAY_STARTED',
            NOW() - INTERVAL '30 minutes',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::uuid,
            1,
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            CURRENT_DATE - 1,
            'MANAGER_APPROVED_END_OF_DAY',
            NOW() - INTERVAL '1 day 4 hours',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NOW() - INTERVAL '1 day 20 minutes',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NOW() - INTERVAL '1 day 20 minutes',
            NULL::timestamptz,
            NULL::uuid,
            NULL::timestamptz,
            NULL::uuid,
            2,
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            CURRENT_DATE - 2,
            'TRANSACTION_CLOSED',
            NOW() - INTERVAL '2 days 5 hours',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NOW() - INTERVAL '2 days 2 hours',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NOW() - INTERVAL '2 days 2 hours',
            NOW() - INTERVAL '2 days 15 minutes',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            NULL::timestamptz,
            NULL::uuid,
            3,
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000001'::uuid
        ),
        (
            CURRENT_DATE - 3,
            'ACCOUNTING_CLOSED',
            NOW() - INTERVAL '3 days 8 hours',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            NOW() - INTERVAL '3 days 5 hours',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            NOW() - INTERVAL '3 days 5 hours',
            NOW() - INTERVAL '3 days 4 hours',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            NOW() - INTERVAL '3 days 2 hours',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            4,
            'a0000000-0000-0000-0000-000000000010'::uuid,
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
    created_by,
    updated_by
)
SELECT
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000000'::uuid,
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
    created_by,
    updated_by
FROM seed_days
ON CONFLICT (business_date) DO UPDATE SET
    contract_id              = EXCLUDED.contract_id,
    current_state            = EXCLUDED.current_state,
    opened_at                = EXCLUDED.opened_at,
    opened_by                = EXCLUDED.opened_by,
    manager_approved_at      = EXCLUDED.manager_approved_at,
    manager_approved_by      = EXCLUDED.manager_approved_by,
    transactions_locked_at   = EXCLUDED.transactions_locked_at,
    transaction_closed_at    = EXCLUDED.transaction_closed_at,
    transaction_closed_by    = EXCLUDED.transaction_closed_by,
    accounting_closed_at     = EXCLUDED.accounting_closed_at,
    accounting_closed_by     = EXCLUDED.accounting_closed_by,
    pending_reclose          = EXCLUDED.pending_reclose,
    reclose_count            = EXCLUDED.reclose_count,
    updated_at               = NOW(),
    updated_by               = EXCLUDED.updated_by,
    version                  = workflow__day_states.version + 1;

WITH day_lookup AS (
    SELECT id AS workflow_day_id, contract_id, business_date
    FROM workflow__day_states
    WHERE business_date BETWEEN CURRENT_DATE - 3 AND CURRENT_DATE
),
seed_transitions (
    request_id,
    business_date,
    from_state,
    to_state,
    action,
    actor_id,
    actor_username,
    reason,
    metadata,
    age
) AS (
    VALUES
        (
            'demo-seed-workflow-today-open',
            CURRENT_DATE,
            'NOT_STARTED',
            'INVESTMENT_DAY_STARTED',
            'OPEN_DAY',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '30 minutes'
        ),
        (
            'demo-seed-workflow-minus-1-open',
            CURRENT_DATE - 1,
            'NOT_STARTED',
            'INVESTMENT_DAY_STARTED',
            'OPEN_DAY',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '1 day 4 hours'
        ),
        (
            'demo-seed-workflow-minus-1-approve',
            CURRENT_DATE - 1,
            'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED_END_OF_DAY',
            'APPROVE',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            'EOD decisions reviewed.',
            '{"zeroTransactionAttestation":false}'::jsonb,
            INTERVAL '1 day 20 minutes'
        ),
        (
            'demo-seed-workflow-minus-2-open',
            CURRENT_DATE - 2,
            'NOT_STARTED',
            'INVESTMENT_DAY_STARTED',
            'OPEN_DAY',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '2 days 5 hours'
        ),
        (
            'demo-seed-workflow-minus-2-approve',
            CURRENT_DATE - 2,
            'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED_END_OF_DAY',
            'APPROVE',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            'Quarterly rebalance approved.',
            '{}'::jsonb,
            INTERVAL '2 days 2 hours'
        ),
        (
            'demo-seed-workflow-minus-2-close-transactions',
            CURRENT_DATE - 2,
            'MANAGER_APPROVED_END_OF_DAY',
            'TRANSACTION_CLOSED',
            'CLOSE_TRANSACTIONS',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            NULL::text,
            '{"ordersInFlight":2}'::jsonb,
            INTERVAL '2 days 15 minutes'
        ),
        (
            'demo-seed-workflow-minus-3-open',
            CURRENT_DATE - 3,
            'NOT_STARTED',
            'INVESTMENT_DAY_STARTED',
            'OPEN_DAY',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            'ben',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '3 days 8 hours'
        ),
        (
            'demo-seed-workflow-minus-3-approve',
            CURRENT_DATE - 3,
            'INVESTMENT_DAY_STARTED',
            'MANAGER_APPROVED_END_OF_DAY',
            'APPROVE',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            'ben',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '3 days 5 hours'
        ),
        (
            'demo-seed-workflow-minus-3-close-transactions',
            CURRENT_DATE - 3,
            'MANAGER_APPROVED_END_OF_DAY',
            'TRANSACTION_CLOSED',
            'CLOSE_TRANSACTIONS',
            'a0000000-0000-0000-0000-000000000010'::uuid,
            'ben',
            NULL::text,
            '{}'::jsonb,
            INTERVAL '3 days 4 hours'
        ),
        (
            'demo-seed-workflow-minus-3-close-accounting',
            CURRENT_DATE - 3,
            'TRANSACTION_CLOSED',
            'ACCOUNTING_CLOSED',
            'CLOSE_ACCOUNTING',
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'admin',
            NULL::text,
            '{"irgPostTrade":"PASS"}'::jsonb,
            INTERVAL '3 days 2 hours'
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
    (
        SUBSTR(MD5(st.request_id || ':' || st.business_date::text), 1, 8) || '-' ||
        SUBSTR(MD5(st.request_id || ':' || st.business_date::text), 9, 4) || '-' ||
        SUBSTR(MD5(st.request_id || ':' || st.business_date::text), 13, 4) || '-' ||
        SUBSTR(MD5(st.request_id || ':' || st.business_date::text), 17, 4) || '-' ||
        SUBSTR(MD5(st.request_id || ':' || st.business_date::text), 21, 12)
    )::uuid,
    dl.workflow_day_id,
    dl.contract_id,
    dl.business_date,
    st.from_state,
    st.to_state,
    st.action,
    st.actor_id,
    'HUMAN',
    st.actor_username,
    st.reason,
    st.metadata,
    NOW() - st.age,
    st.request_id
FROM seed_transitions st
JOIN day_lookup dl ON dl.business_date = st.business_date
ON CONFLICT (id) DO UPDATE
SET
    workflow_day_id = EXCLUDED.workflow_day_id,
    contract_id = EXCLUDED.contract_id,
    business_date = EXCLUDED.business_date,
    from_state = EXCLUDED.from_state,
    to_state = EXCLUDED.to_state,
    action = EXCLUDED.action,
    actor_id = EXCLUDED.actor_id,
    actor_type = EXCLUDED.actor_type,
    actor_username = EXCLUDED.actor_username,
    reason = EXCLUDED.reason,
    metadata = EXCLUDED.metadata,
    occurred_at = EXCLUDED.occurred_at,
    request_id = EXCLUDED.request_id;

COMMIT;
