-- =============================================================================
-- Notification module — email permission seed
-- =============================================================================
-- Grants all notification email admin permission codes to the Admin group.
-- Idempotent. The Go permission catalog (cmd/seed) upserts the function
-- definitions first; this file wires the grants to the Admin group.
-- =============================================================================

BEGIN;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('NOTIFICATION_CONFIG'),
    ('NOTIFICATION_VIEW'),
    ('NOTIFICATION_EDIT'),
    ('NOTIFICATION_RETRY'),
    ('NOTIFICATION_TEST'),
    ('NOTIFICATION_HEALTH')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
    SET is_granted = EXCLUDED.is_granted;

COMMIT;
