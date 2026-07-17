-- =============================================================================
-- Integration module — dashboard view permission seed
-- =============================================================================
-- Grants INTEGRATION_DASHBOARD_VIEW to every seeded group. The Go permission
-- catalog (cmd/seed) upserts the function definition first; this file wires
-- the grant. Unlike admin-only module permissions, the personal dashboard
-- (task feed, workflow states, and the AUM/P&L valuation summary cards) is a
-- general landing page every authenticated user needs, not an admin-only
-- surface — so this is granted to all groups rather than just Admin.
-- Idempotent.
-- =============================================================================

BEGIN;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('INTEGRATION_DASHBOARD_VIEW')
) AS p(code)
ON CONFLICT (group_id, permission_code) DO UPDATE
    SET is_granted = EXCLUDED.is_granted;

COMMIT;
