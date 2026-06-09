-- =============================================================================
-- Demo seed — Workflow day states + transition log
-- =============================================================================
-- Sets a representative spread of workflow states for CURRENT_DATE across the
-- five active demo contracts so the My Funds cockpit shows variety:
--
--   TH-GOV-LTF   → DAY_OPEN              (Analysis stage in the UI bar)
--   BBL-EQUITY   → MANAGER_APPROVED      (Decision committed; locked from PM)
--   SCB-FIXED    → ACCOUNTING_CLOSED     (Locked; "LOCKED" badge in screenshot)
--   KTB-BALANCED → TRANSACTION_CLOSED    (Execution complete)
--   GLOBAL-TECH  → DAY_OPEN              (Stale data; demo of mixed signals)
--
-- MMF-CASH intentionally has no day-state row to demonstrate the
-- "NOT_STARTED / day not yet opened" UI path.
--
-- Idempotency strategy (re-runnable on ANY calendar day):
--   * Day-state ids use gen_random_uuid() rather than fixed UUIDs. A fixed id
--     paired with a CURRENT_DATE business_date is NOT safe across days — on a
--     later day the natural key (contract_id, business_date) no longer matches
--     the ON CONFLICT target, so the fixed primary key collides with the row
--     written on the first run (SQLSTATE 23505 on *_pkey).
--   * ON CONFLICT (contract_id, business_date) DO UPDATE still resets the demo
--     state to the canonical values for the current business date and merges
--     cleanly with any row already opened for today.
--   * The transition log resolves workflow_day_id by natural key (contract +
--     today) and keeps stable ids with ON CONFLICT (id) DO NOTHING.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Day states for the five active demo funds (today)
-- ---------------------------------------------------------------------------
INSERT INTO workflow__day_states (
    id, contract_id, business_date, current_state,
    opened_at, opened_by,
    manager_approved_at, manager_approved_by, transactions_locked_at,
    transaction_closed_at, transaction_closed_by,
    accounting_closed_at, accounting_closed_by,
    pending_reclose, reclose_count, version,
    created_by, updated_by
)
VALUES
    -- TH-GOV-LTF — DAY_OPEN, opened 30 min ago by admin
    (gen_random_uuid(), 'd0001000-0000-0000-0000-000000000001'::uuid, CURRENT_DATE,
        'DAY_OPEN',
        NOW() - INTERVAL '30 minutes', 'a0000000-0000-0000-0000-000000000001'::uuid,
        NULL, NULL, NULL,
        NULL, NULL,
        NULL, NULL,
        false, 0, 1,
        'a0000000-0000-0000-0000-000000000001'::uuid,
        'a0000000-0000-0000-0000-000000000001'::uuid),

    -- BBL-EQUITY — MANAGER_APPROVED 20 min ago, locked
    (gen_random_uuid(), 'd0001000-0000-0000-0000-000000000002'::uuid, CURRENT_DATE,
        'MANAGER_APPROVED',
        NOW() - INTERVAL '4 hours',  'a0000000-0000-0000-0000-000000000001'::uuid,
        NOW() - INTERVAL '20 minutes', 'a0000000-0000-0000-0000-000000000001'::uuid,
        NOW() - INTERVAL '20 minutes',
        NULL, NULL,
        NULL, NULL,
        false, 0, 1,
        'a0000000-0000-0000-0000-000000000001'::uuid,
        'a0000000-0000-0000-0000-000000000001'::uuid),

    -- SCB-FIXED — ACCOUNTING_CLOSED 2 hours ago by ben (manager)
    (gen_random_uuid(), 'd0001000-0000-0000-0000-000000000003'::uuid, CURRENT_DATE,
        'ACCOUNTING_CLOSED',
        NOW() - INTERVAL '8 hours', 'a0000000-0000-0000-0000-000000000010'::uuid,
        NOW() - INTERVAL '5 hours', 'a0000000-0000-0000-0000-000000000010'::uuid,
        NOW() - INTERVAL '5 hours',
        NOW() - INTERVAL '4 hours', 'a0000000-0000-0000-0000-000000000010'::uuid,
        NOW() - INTERVAL '2 hours', 'a0000000-0000-0000-0000-000000000001'::uuid,
        false, 0, 1,
        'a0000000-0000-0000-0000-000000000010'::uuid,
        'a0000000-0000-0000-0000-000000000001'::uuid),

    -- KTB-BALANCED — TRANSACTION_CLOSED 15 min ago
    (gen_random_uuid(), 'd0001000-0000-0000-0000-000000000004'::uuid, CURRENT_DATE,
        'TRANSACTION_CLOSED',
        NOW() - INTERVAL '5 hours', 'a0000000-0000-0000-0000-000000000001'::uuid,
        NOW() - INTERVAL '2 hours', 'a0000000-0000-0000-0000-000000000001'::uuid,
        NOW() - INTERVAL '2 hours',
        NOW() - INTERVAL '15 minutes', 'a0000000-0000-0000-0000-000000000001'::uuid,
        NULL, NULL,
        false, 0, 1,
        'a0000000-0000-0000-0000-000000000001'::uuid,
        'a0000000-0000-0000-0000-000000000001'::uuid),

    -- GLOBAL-TECH — DAY_OPEN 3 hours ago (stale provider feed in valuation)
    (gen_random_uuid(), 'd0001000-0000-0000-0000-000000000005'::uuid, CURRENT_DATE,
        'DAY_OPEN',
        NOW() - INTERVAL '3 hours', 'a0000000-0000-0000-0000-000000000011'::uuid,
        NULL, NULL, NULL,
        NULL, NULL,
        NULL, NULL,
        false, 0, 1,
        'a0000000-0000-0000-0000-000000000011'::uuid,
        'a0000000-0000-0000-0000-000000000011'::uuid)
ON CONFLICT (contract_id, business_date) DO UPDATE SET
    current_state           = EXCLUDED.current_state,
    opened_at               = EXCLUDED.opened_at,
    opened_by               = EXCLUDED.opened_by,
    manager_approved_at     = EXCLUDED.manager_approved_at,
    manager_approved_by     = EXCLUDED.manager_approved_by,
    transactions_locked_at  = EXCLUDED.transactions_locked_at,
    transaction_closed_at   = EXCLUDED.transaction_closed_at,
    transaction_closed_by   = EXCLUDED.transaction_closed_by,
    accounting_closed_at    = EXCLUDED.accounting_closed_at,
    accounting_closed_by    = EXCLUDED.accounting_closed_by,
    updated_at              = NOW(),
    updated_by              = EXCLUDED.updated_by,
    version                 = workflow__day_states.version + 1;

-- ---------------------------------------------------------------------------
-- 2. Transition log — append-only audit entries for the activity feed.
--    workflow_day_id is resolved by natural key (contract + today) because the
--    day-state ids are now generated, not fixed.
-- ---------------------------------------------------------------------------
INSERT INTO workflow__transition_log (
    id, workflow_day_id, contract_id, business_date,
    from_state, to_state, action,
    actor_id, actor_type, actor_username,
    reason, metadata, occurred_at, request_id
)
SELECT v.id::uuid,
       (SELECT ds.id FROM workflow__day_states ds
         WHERE ds.contract_id = v.contract_id::uuid
           AND ds.business_date = CURRENT_DATE
         LIMIT 1),
       v.contract_id::uuid, CURRENT_DATE,
       v.from_state, v.to_state, v.action,
       v.actor_id::uuid, 'HUMAN', v.actor_username,
       v.reason, v.metadata::jsonb, NOW() - v.age::interval, v.request_id
FROM (VALUES
    -- TH-GOV-LTF open
    ('d000b000-0000-0000-0000-000000000001', 'd0001000-0000-0000-0000-000000000001',
        'NOT_STARTED', 'DAY_OPEN', 'open_day',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        NULL, '{}', '30 minutes', 'demo-seed-001'),

    -- BBL-EQUITY: open → approve
    ('d000b000-0000-0000-0000-000000000010', 'd0001000-0000-0000-0000-000000000002',
        'NOT_STARTED', 'DAY_OPEN', 'open_day',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        NULL, '{}', '4 hours', 'demo-seed-010'),
    ('d000b000-0000-0000-0000-000000000011', 'd0001000-0000-0000-0000-000000000002',
        'DAY_OPEN', 'MANAGER_APPROVED', 'manager_approve',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        'EOD decisions reviewed; KBANK single-issuer breach acknowledged.',
        '{"breach_acknowledged":"BR-117"}', '20 minutes', 'demo-seed-011'),

    -- SCB-FIXED: full close
    ('d000b000-0000-0000-0000-000000000020', 'd0001000-0000-0000-0000-000000000003',
        'NOT_STARTED', 'DAY_OPEN', 'open_day',
        'a0000000-0000-0000-0000-000000000010', 'ben',
        NULL, '{}', '8 hours', 'demo-seed-020'),
    ('d000b000-0000-0000-0000-000000000021', 'd0001000-0000-0000-0000-000000000003',
        'DAY_OPEN', 'MANAGER_APPROVED', 'manager_approve',
        'a0000000-0000-0000-0000-000000000010', 'ben',
        NULL, '{}', '5 hours', 'demo-seed-021'),
    ('d000b000-0000-0000-0000-000000000022', 'd0001000-0000-0000-0000-000000000003',
        'MANAGER_APPROVED', 'TRANSACTION_CLOSED', 'transaction_close',
        'a0000000-0000-0000-0000-000000000010', 'ben',
        NULL, '{}', '4 hours', 'demo-seed-022'),
    ('d000b000-0000-0000-0000-000000000023', 'd0001000-0000-0000-0000-000000000003',
        'TRANSACTION_CLOSED', 'ACCOUNTING_CLOSED', 'accounting_close',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        NULL, '{"irg_post_trade":"PASS"}', '2 hours', 'demo-seed-023'),

    -- KTB-BALANCED: open → approve → tx close
    ('d000b000-0000-0000-0000-000000000030', 'd0001000-0000-0000-0000-000000000004',
        'NOT_STARTED', 'DAY_OPEN', 'open_day',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        NULL, '{}', '5 hours', 'demo-seed-030'),
    ('d000b000-0000-0000-0000-000000000031', 'd0001000-0000-0000-0000-000000000004',
        'DAY_OPEN', 'MANAGER_APPROVED', 'manager_approve',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        'Quarterly rebalance approved.', '{}', '2 hours', 'demo-seed-031'),
    ('d000b000-0000-0000-0000-000000000032', 'd0001000-0000-0000-0000-000000000004',
        'MANAGER_APPROVED', 'TRANSACTION_CLOSED', 'transaction_close',
        'a0000000-0000-0000-0000-000000000001', 'admin',
        NULL, '{"orders_in_flight":2}', '15 minutes', 'demo-seed-032'),

    -- GLOBAL-TECH: just opened, stale prices
    ('d000b000-0000-0000-0000-000000000040', 'd0001000-0000-0000-0000-000000000005',
        'NOT_STARTED', 'DAY_OPEN', 'open_day',
        'a0000000-0000-0000-0000-000000000011', 'green',
        NULL, '{"stale_inputs":true}', '3 hours', 'demo-seed-040')
) AS v(id, contract_id, from_state, to_state, action, actor_id, actor_username, reason, metadata, age, request_id)
ON CONFLICT (id) DO NOTHING;

COMMIT;
