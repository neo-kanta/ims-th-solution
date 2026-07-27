-- =============================================================================
-- Approval module — permission seed
-- =============================================================================
-- Grants the full APPROVAL_* function permission set to the Admin group.
-- Idempotent. The Go permission catalog (cmd/seed) upserts the function
-- definitions first; this file wires the grants to the Admin group.
-- =============================================================================

BEGIN;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('APPROVAL_VIEW_INBOX'),
    ('APPROVAL_VIEW_REQUEST'),
    ('APPROVAL_SUBMIT'),
    ('APPROVAL_APPROVE'),
    ('APPROVAL_REJECT'),
    ('APPROVAL_WITHDRAW'),
    ('APPROVAL_CANCEL'),
    ('APPROVAL_REVOKE'),
    ('APPROVAL_AUDIT_VIEW'),
    ('APPROVAL_SYNC_VIEW'),
    ('APPROVAL_SYNC_RETRY'),
    ('APPROVAL_CONFIG_VIEW'),
    ('APPROVAL_GROUP_MANAGE'),
    ('APPROVAL_TEAM_MANAGE'),
    ('APPROVAL_PROCESS_MANAGE')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted,
    created_by = COALESCE(permissions_function_rights.created_by, EXCLUDED.created_by);

COMMIT;
