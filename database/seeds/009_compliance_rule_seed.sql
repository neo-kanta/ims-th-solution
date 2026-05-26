-- =============================================================================
-- Compliance module - baseline active pre-trade rules
-- =============================================================================
-- Provides a non-empty GLOBAL rule set for investment transaction posting and
-- simulation. Idempotent: instances upsert by unique name, versions are
-- immutable, and bindings use stable UUID primary keys.
-- =============================================================================

BEGIN;

WITH inst AS (
    INSERT INTO compliance_rule_instances (
        id, rule_type_id, name, description, current_version,
        is_active, effective_from, effective_to, created_by
    ) VALUES (
        'c1000000-0000-0000-0000-000000000001',
        'cash.availability',
        'GLOBAL - Cash availability',
        'Blocks buy orders that exceed available portfolio cash.',
        1,
        true,
        '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    )
    ON CONFLICT (name) DO UPDATE SET
        rule_type_id    = EXCLUDED.rule_type_id,
        description     = EXCLUDED.description,
        current_version = 1,
        is_active       = true,
        effective_from  = EXCLUDED.effective_from,
        effective_to    = NULL,
        updated_at      = NOW()
    RETURNING id
)
INSERT INTO compliance_rule_instance_versions (
    id, rule_instance_id, version_number, parameters, change_note, created_by
)
SELECT
    'c2000000-0000-0000-0000-000000000001',
    id,
    1,
    '{"min_cash_buffer_pct": 0}'::jsonb,
    'Seed baseline cash availability rule',
    'a0000000-0000-0000-0000-000000000001'
FROM inst
WHERE NOT EXISTS (
    SELECT 1
    FROM compliance_rule_instance_versions existing
    WHERE existing.rule_instance_id = inst.id
      AND existing.version_number = 1
);

WITH inst AS (
    SELECT id FROM compliance_rule_instances WHERE name = 'GLOBAL - Cash availability'
)
INSERT INTO compliance_rule_bindings (
    id, rule_instance_id, rule_set_id, scope_type, scope_id,
    severity, priority, effective_from, effective_to,
    is_active, created_by
)
SELECT
    'c3000000-0000-0000-0000-000000000001',
    id,
    NULL,
    'GLOBAL',
    NULL,
    'BLOCK',
    10,
    '2026-01-01',
    NULL,
    true,
    'a0000000-0000-0000-0000-000000000001'
FROM inst
ON CONFLICT (id) DO UPDATE SET
    rule_instance_id = EXCLUDED.rule_instance_id,
    scope_type       = EXCLUDED.scope_type,
    scope_id         = EXCLUDED.scope_id,
    severity         = EXCLUDED.severity,
    priority         = EXCLUDED.priority,
    effective_from   = EXCLUDED.effective_from,
    effective_to     = NULL,
    is_active        = true,
    updated_at       = NOW();

WITH inst AS (
    INSERT INTO compliance_rule_instances (
        id, rule_type_id, name, description, current_version,
        is_active, effective_from, effective_to, created_by
    ) VALUES (
        'c1000000-0000-0000-0000-000000000002',
        'quantity.min_trading_unit',
        'GLOBAL - Minimum trading unit',
        'Blocks orders whose quantity is not a valid board-lot multiple.',
        1,
        true,
        '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    )
    ON CONFLICT (name) DO UPDATE SET
        rule_type_id    = EXCLUDED.rule_type_id,
        description     = EXCLUDED.description,
        current_version = 1,
        is_active       = true,
        effective_from  = EXCLUDED.effective_from,
        effective_to    = NULL,
        updated_at      = NOW()
    RETURNING id
)
INSERT INTO compliance_rule_instance_versions (
    id, rule_instance_id, version_number, parameters, change_note, created_by
)
SELECT
    'c2000000-0000-0000-0000-000000000002',
    id,
    1,
    '{"default_lot_size": 1, "overrides": {}}'::jsonb,
    'Seed baseline minimum trading unit rule',
    'a0000000-0000-0000-0000-000000000001'
FROM inst
WHERE NOT EXISTS (
    SELECT 1
    FROM compliance_rule_instance_versions existing
    WHERE existing.rule_instance_id = inst.id
      AND existing.version_number = 1
);

WITH inst AS (
    SELECT id FROM compliance_rule_instances WHERE name = 'GLOBAL - Minimum trading unit'
)
INSERT INTO compliance_rule_bindings (
    id, rule_instance_id, rule_set_id, scope_type, scope_id,
    severity, priority, effective_from, effective_to,
    is_active, created_by
)
SELECT
    'c3000000-0000-0000-0000-000000000002',
    id,
    NULL,
    'GLOBAL',
    NULL,
    'BLOCK',
    20,
    '2026-01-01',
    NULL,
    true,
    'a0000000-0000-0000-0000-000000000001'
FROM inst
ON CONFLICT (id) DO UPDATE SET
    rule_instance_id = EXCLUDED.rule_instance_id,
    scope_type       = EXCLUDED.scope_type,
    scope_id         = EXCLUDED.scope_id,
    severity         = EXCLUDED.severity,
    priority         = EXCLUDED.priority,
    effective_from   = EXCLUDED.effective_from,
    effective_to     = NULL,
    is_active        = true,
    updated_at       = NOW();

WITH inst AS (
    INSERT INTO compliance_rule_instances (
        id, rule_type_id, name, description, current_version,
        is_active, effective_from, effective_to, created_by
    ) VALUES (
        'c1000000-0000-0000-0000-000000000003',
        'amount.minimum_trade',
        'GLOBAL - Minimum trade amount',
        'Blocks orders below the minimum notional trade amount.',
        1,
        true,
        '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    )
    ON CONFLICT (name) DO UPDATE SET
        rule_type_id    = EXCLUDED.rule_type_id,
        description     = EXCLUDED.description,
        current_version = 1,
        is_active       = true,
        effective_from  = EXCLUDED.effective_from,
        effective_to    = NULL,
        updated_at      = NOW()
    RETURNING id
)
INSERT INTO compliance_rule_instance_versions (
    id, rule_instance_id, version_number, parameters, change_note, created_by
)
SELECT
    'c2000000-0000-0000-0000-000000000003',
    id,
    1,
    '{"min_amount": 1000, "currency": "THB"}'::jsonb,
    'Seed baseline minimum trade amount rule',
    'a0000000-0000-0000-0000-000000000001'
FROM inst
WHERE NOT EXISTS (
    SELECT 1
    FROM compliance_rule_instance_versions existing
    WHERE existing.rule_instance_id = inst.id
      AND existing.version_number = 1
);

WITH inst AS (
    SELECT id FROM compliance_rule_instances WHERE name = 'GLOBAL - Minimum trade amount'
)
INSERT INTO compliance_rule_bindings (
    id, rule_instance_id, rule_set_id, scope_type, scope_id,
    severity, priority, effective_from, effective_to,
    is_active, created_by
)
SELECT
    'c3000000-0000-0000-0000-000000000003',
    id,
    NULL,
    'GLOBAL',
    NULL,
    'BLOCK',
    30,
    '2026-01-01',
    NULL,
    true,
    'a0000000-0000-0000-0000-000000000001'
FROM inst
ON CONFLICT (id) DO UPDATE SET
    rule_instance_id = EXCLUDED.rule_instance_id,
    scope_type       = EXCLUDED.scope_type,
    scope_id         = EXCLUDED.scope_id,
    severity         = EXCLUDED.severity,
    priority         = EXCLUDED.priority,
    effective_from   = EXCLUDED.effective_from,
    effective_to     = NULL,
    is_active        = true,
    updated_at       = NOW();

WITH inst AS (
    INSERT INTO compliance_rule_instances (
        id, rule_type_id, name, description, current_version,
        is_active, effective_from, effective_to, created_by
    ) VALUES (
        'c1000000-0000-0000-0000-000000000004',
        'ratio.sector_exposure',
        'GLOBAL - Energy sector exposure',
        'Blocks trades that would push ENERGY exposure above the configured NAV percentage.',
        1,
        true,
        '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    )
    ON CONFLICT (name) DO UPDATE SET
        rule_type_id    = EXCLUDED.rule_type_id,
        description     = EXCLUDED.description,
        current_version = 1,
        is_active       = true,
        effective_from  = EXCLUDED.effective_from,
        effective_to    = NULL,
        updated_at      = NOW()
    RETURNING id
)
INSERT INTO compliance_rule_instance_versions (
    id, rule_instance_id, version_number, parameters, change_note, created_by
)
SELECT
    'c2000000-0000-0000-0000-000000000004',
    id,
    1,
    '{"sector": "ENERGY", "max_pct": 35}'::jsonb,
    'Seed baseline sector exposure rule',
    'a0000000-0000-0000-0000-000000000001'
FROM inst
WHERE NOT EXISTS (
    SELECT 1
    FROM compliance_rule_instance_versions existing
    WHERE existing.rule_instance_id = inst.id
      AND existing.version_number = 1
);

WITH inst AS (
    SELECT id FROM compliance_rule_instances WHERE name = 'GLOBAL - Energy sector exposure'
)
INSERT INTO compliance_rule_bindings (
    id, rule_instance_id, rule_set_id, scope_type, scope_id,
    severity, priority, effective_from, effective_to,
    is_active, created_by
)
SELECT
    'c3000000-0000-0000-0000-000000000004',
    id,
    NULL,
    'GLOBAL',
    NULL,
    'BLOCK',
    40,
    '2026-01-01',
    NULL,
    true,
    'a0000000-0000-0000-0000-000000000001'
FROM inst
ON CONFLICT (id) DO UPDATE SET
    rule_instance_id = EXCLUDED.rule_instance_id,
    scope_type       = EXCLUDED.scope_type,
    scope_id         = EXCLUDED.scope_id,
    severity         = EXCLUDED.severity,
    priority         = EXCLUDED.priority,
    effective_from   = EXCLUDED.effective_from,
    effective_to     = NULL,
    is_active        = true,
    updated_at       = NOW();

COMMIT;
