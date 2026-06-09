-- =============================================================================
-- Demo seed — Compliance rule instance + check records + breaches
-- =============================================================================
-- Adds a "single-issuer concentration" rule instance, then plants one BLOCK-
-- severity OPEN breach (BBL-EQUITY / KBANK at 11.13%) plus three WARN-severity
-- OPEN warnings so the "Open Breaches" KPI shows 1 / 3 warnings, matching the
-- product cockpit screenshot.
--
-- Implementation notes:
--   * compliance_check_records is append-only (REVOKE UPDATE/DELETE from
--     PUBLIC at migration time). We rely on stable UUIDs + ON CONFLICT (id)
--     DO NOTHING so re-runs are safe.
--   * Breaches reference a check_record by FK, so the check_record must be
--     inserted first within the same seed file.
--   * rule_type_id is a synthetic string for the demo; we are not registering
--     a runtime evaluator here — these rows are historical records.
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Demo rule instance + version: concentration.single_issuer (max 10% NAV)
-- ---------------------------------------------------------------------------
INSERT INTO compliance_rule_instances (
    id, rule_type_id, name, description, current_version,
    is_active, effective_from, effective_to, created_by
)
VALUES (
    'c1000000-0000-0000-0000-000000000010',
    'concentration.single_issuer',
    'DEMO - Single-issuer concentration',
    'Warns when a single issuer exceeds 10% of fund NAV; blocks above 12%.',
    1,
    true,
    DATE '2026-01-01',
    NULL,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (name) DO UPDATE SET
    rule_type_id    = EXCLUDED.rule_type_id,
    description     = EXCLUDED.description,
    current_version = EXCLUDED.current_version,
    is_active       = EXCLUDED.is_active,
    updated_at      = NOW();

INSERT INTO compliance_rule_instance_versions (
    id, rule_instance_id, version_number, parameters, change_note, created_by
)
SELECT 'c2000000-0000-0000-0000-000000000010'::uuid,
       'c1000000-0000-0000-0000-000000000010'::uuid,
       1,
       -- The evaluator uses `max_pct` for the issuer ceiling; warn/block split
       -- isn't yet supported by the SPI so we cap at 10% as the single limit.
       '{"max_pct": 10}'::jsonb,
       'Demo seed baseline for single-issuer concentration',
       'a0000000-0000-0000-0000-000000000001'::uuid
WHERE NOT EXISTS (
    SELECT 1 FROM compliance_rule_instance_versions
    WHERE rule_instance_id = 'c1000000-0000-0000-0000-000000000010'::uuid
      AND version_number   = 1
);

INSERT INTO compliance_rule_bindings (
    id, rule_instance_id, rule_set_id, scope_type, scope_id,
    severity, priority, effective_from, effective_to,
    is_active, created_by
)
VALUES (
    'c3000000-0000-0000-0000-000000000010',
    'c1000000-0000-0000-0000-000000000010',
    NULL,
    'GLOBAL',
    NULL,
    'WARN',
    50,
    DATE '2026-01-01',
    NULL,
    true,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (id) DO UPDATE SET
    severity   = EXCLUDED.severity,
    is_active  = EXCLUDED.is_active,
    updated_at = NOW();

-- ---------------------------------------------------------------------------
-- 2. Compliance check records (append-only; one per finding)
-- ---------------------------------------------------------------------------
INSERT INTO compliance_check_records (
    id, check_group_id, timing, order_id, portfolio_id, contract_id,
    ticker, rule_type_id, rule_instance_id, rule_instance_version,
    parameter_snapshot, verdict, effective_severity, final_verdict,
    evidence, message, data_snapshot_hash, eval_duration_ms, checked_by,
    business_date, checked_at
)
VALUES
    -- BBL-EQUITY / KBANK at 11.127% of NAV — BLOCK
    ('d000c000-0000-0000-0000-000000000001',
        'd000c000-0000-0000-0000-000000000fff',
        'POST_TRADE', NULL,
        'd0002000-0000-0000-0000-000000000002'::uuid,
        'd0001000-0000-0000-0000-000000000002'::uuid,
        'KBANK', 'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid, 1,
        '{"warn_pct":10,"block_pct":12}'::jsonb,
        'BLOCK', 'BLOCK', 'BLOCK',
        '{"issuer":"KBANK","mv":142920000,"aum":1284500000,"pct":11.127}'::jsonb,
        'Single-issuer concentration: KBANK at 11.13% of NAV (limit 10%).',
        'demo-seed-hash-001', 4, 'system',
        CURRENT_DATE, NOW() - INTERVAL '25 minutes'),

    -- BBL-EQUITY / SCB at 9.5% — WARN (approaching)
    ('d000c000-0000-0000-0000-000000000002',
        'd000c000-0000-0000-0000-000000000fff',
        'POST_TRADE', NULL,
        'd0002000-0000-0000-0000-000000000002'::uuid,
        'd0001000-0000-0000-0000-000000000002'::uuid,
        'SCB', 'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid, 1,
        '{"warn_pct":10,"block_pct":12}'::jsonb,
        'WARN', 'WARN', 'WARN',
        '{"issuer":"SCB","mv":226525000,"aum":1284500000,"pct":17.63}'::jsonb,
        'Single-issuer concentration: SCB at 17.63% of NAV approaching limit.',
        'demo-seed-hash-002', 3, 'system',
        CURRENT_DATE, NOW() - INTERVAL '24 minutes'),

    -- KTB-BALANCED / AOT 8.8% of fund AUM — WARN
    ('d000c000-0000-0000-0000-000000000003',
        'd000c000-0000-0000-0000-000000000ffe',
        'POST_TRADE', NULL,
        'd0002000-0000-0000-0000-000000000004'::uuid,
        'd0001000-0000-0000-0000-000000000004'::uuid,
        'AOT', 'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid, 1,
        '{"warn_pct":10,"block_pct":12}'::jsonb,
        'WARN', 'WARN', 'WARN',
        '{"issuer":"AOT","mv":158750000,"aum":986210805,"pct":16.10}'::jsonb,
        'Single-issuer concentration: AOT at 16.10% of fund NAV.',
        'demo-seed-hash-003', 3, 'system',
        CURRENT_DATE, NOW() - INTERVAL '17 minutes'),

    -- GLOBAL-TECH stale provider feed — WARN (using sector exposure rule)
    ('d000c000-0000-0000-0000-000000000004',
        'd000c000-0000-0000-0000-000000000ffd',
        'PERIODIC', NULL,
        'd0002000-0000-0000-0000-000000000006'::uuid,
        'd0001000-0000-0000-0000-000000000005'::uuid,
        '', 'ratio.sector_exposure',
        'c1000000-0000-0000-0000-000000000004'::uuid, 1,
        '{"sector":"INFO_TECH","max_pct":35}'::jsonb,
        'WARN', 'WARN', 'WARN',
        '{"sector":"INFO_TECH","pct":85.0,"stale_inputs":true,"feed_offline_hours":72}'::jsonb,
        'Tech sector concentration 85% combined with stale provider feed (>72h).',
        'demo-seed-hash-004', 5, 'system',
        CURRENT_DATE - 3, NOW() - INTERVAL '3 days')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. Open breaches — visible on the My Funds cockpit
-- ---------------------------------------------------------------------------
INSERT INTO compliance_breaches (
    id, check_record_id, check_group_id, portfolio_id, contract_id,
    rule_type_id, rule_instance_id, severity, verdict, status,
    evidence, message, business_date
)
VALUES
    -- BR-117: KBANK single-issuer 11.13% — BLOCK
    ('d000d000-0000-0000-0000-000000000001',
        'd000c000-0000-0000-0000-000000000001'::uuid,
        'd000c000-0000-0000-0000-000000000fff'::uuid,
        'd0002000-0000-0000-0000-000000000002'::uuid,
        'd0001000-0000-0000-0000-000000000002'::uuid,
        'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid,
        'BLOCK', 'BLOCK', 'OPEN',
        '{"issuer":"KBANK","pct":11.127,"limit":10,"breach_ref":"BR-117"}'::jsonb,
        'KBANK 11.2% single-issuer concentration (limit 10%).',
        CURRENT_DATE),

    -- WARN: SCB approaching single-issuer limit
    ('d000d000-0000-0000-0000-000000000002',
        'd000c000-0000-0000-0000-000000000002'::uuid,
        'd000c000-0000-0000-0000-000000000fff'::uuid,
        'd0002000-0000-0000-0000-000000000002'::uuid,
        'd0001000-0000-0000-0000-000000000002'::uuid,
        'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid,
        'WARN', 'WARN', 'OPEN',
        '{"issuer":"SCB","pct":17.63,"limit":10,"breach_ref":"BR-118"}'::jsonb,
        'SCB single-issuer concentration approaching block threshold.',
        CURRENT_DATE),

    -- WARN: AOT 16.10% of KTB-BALANCED fund NAV
    ('d000d000-0000-0000-0000-000000000003',
        'd000c000-0000-0000-0000-000000000003'::uuid,
        'd000c000-0000-0000-0000-000000000ffe'::uuid,
        'd0002000-0000-0000-0000-000000000004'::uuid,
        'd0001000-0000-0000-0000-000000000004'::uuid,
        'concentration.single_issuer',
        'c1000000-0000-0000-0000-000000000010'::uuid,
        'WARN', 'WARN', 'OPEN',
        '{"issuer":"AOT","pct":16.10,"limit":10,"breach_ref":"BR-119"}'::jsonb,
        'AOT single-issuer concentration 16.10% in KTB-BALANCED.',
        CURRENT_DATE),

    -- WARN: GLOBAL-TECH stale + sector concentration
    ('d000d000-0000-0000-0000-000000000004',
        'd000c000-0000-0000-0000-000000000004'::uuid,
        'd000c000-0000-0000-0000-000000000ffd'::uuid,
        'd0002000-0000-0000-0000-000000000006'::uuid,
        'd0001000-0000-0000-0000-000000000005'::uuid,
        'ratio.sector_exposure',
        'c1000000-0000-0000-0000-000000000004'::uuid,
        'WARN', 'WARN', 'OPEN',
        '{"sector":"INFO_TECH","pct":85.0,"stale":true,"breach_ref":"BR-120"}'::jsonb,
        'INFO_TECH exposure 85% with stale provider feed (>72h).',
        CURRENT_DATE - 3)
ON CONFLICT (id) DO NOTHING;

COMMIT;
