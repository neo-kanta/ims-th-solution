-- =============================================================================
-- Demo seed — Data-scope permissions for demo users
-- =============================================================================
-- Grants per-fund visibility so the My Funds cockpit shows different content
-- depending on which demo account is logged in:
--
--   admin → "*" wildcard (every fund visible)         — used by the cockpit demo
--   ben   → SCB-FIXED (manages) + TH-GOV-LTF (watch)  — fixed-income operator
--   green → GLOBAL-TECH (manages) + BBL-EQUITY (watch)— global equity operator
--   neo   → KTB-BALANCED (watch) + GLOBAL-TECH (watch)— balanced/research view
--
-- The accessibleFundIDs helper in backend/internal/investment/transport/handler/
-- treats "*" as "no filter" so admin sees the full list.
-- =============================================================================

BEGIN;

INSERT INTO permissions_data_rights (
    id, user_id, contract_id, is_granted, granted_at, granted_by
)
VALUES
    -- admin → wildcard
    ('d000e000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'::uuid, '*', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),

    -- ben — SCB-FIXED (manager) + TH-GOV-LTF (watch)
    ('d000e000-0000-0000-0000-000000000010',
        'a0000000-0000-0000-0000-000000000010'::uuid,
        'd0001000-0000-0000-0000-000000000003', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('d000e000-0000-0000-0000-000000000011',
        'a0000000-0000-0000-0000-000000000010'::uuid,
        'd0001000-0000-0000-0000-000000000001', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),

    -- green — GLOBAL-TECH (manager) + BBL-EQUITY (watch)
    ('d000e000-0000-0000-0000-000000000020',
        'a0000000-0000-0000-0000-000000000011'::uuid,
        'd0001000-0000-0000-0000-000000000005', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('d000e000-0000-0000-0000-000000000021',
        'a0000000-0000-0000-0000-000000000011'::uuid,
        'd0001000-0000-0000-0000-000000000002', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),

    -- neo — KTB-BALANCED + GLOBAL-TECH (research/watch)
    ('d000e000-0000-0000-0000-000000000030',
        'a0000000-0000-0000-0000-000000000012'::uuid,
        'd0001000-0000-0000-0000-000000000004', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid),
    ('d000e000-0000-0000-0000-000000000031',
        'a0000000-0000-0000-0000-000000000012'::uuid,
        'd0001000-0000-0000-0000-000000000005', true,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid)
ON CONFLICT (user_id, contract_id) DO UPDATE SET
    is_granted = EXCLUDED.is_granted,
    granted_at = NOW(),
    granted_by = EXCLUDED.granted_by;

COMMIT;
