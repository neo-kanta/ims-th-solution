-- =============================================================================
-- Ben investment-decision approver seed (idempotent)
-- =============================================================================
-- Migrations run before development users are seeded. Migration
-- 20260716000001 creates this dedicated least-privilege role and grants its
-- permissions; this seed ensures the role and Ben membership also exist after
-- a fresh migrate + seed or a later idempotent seed rerun.
-- =============================================================================

BEGIN;

INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000041',
    'Investment Decision Approver',
    'Assigned investment-decision reviewers authorised to use the approval inbox and approve or reject tasks assigned to them.',
    true,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO UPDATE
SET description = EXCLUDED.description,
    is_active   = EXCLUDED.is_active,
    updated_by  = EXCLUDED.updated_by,
    updated_at  = NOW();

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT permission_group.id, permission_code.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups permission_group
CROSS JOIN (VALUES
    ('APPROVAL_VIEW_INBOX'),
    ('APPROVAL_VIEW_REQUEST'),
    ('APPROVAL_APPROVE'),
    ('APPROVAL_REJECT'),
    ('INVESTMENT_DECISION_APPROVE')
) AS permission_code(code)
WHERE permission_group.name = 'Investment Decision Approver'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
SELECT
    'a0000000-0000-0000-0000-000000000010'::uuid,
    permission_group.id,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups permission_group
WHERE permission_group.name = 'Investment Decision Approver'
ON CONFLICT (user_id, group_id) DO NOTHING;

COMMIT;
