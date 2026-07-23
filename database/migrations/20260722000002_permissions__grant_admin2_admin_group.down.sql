-- Revoke only the Admin-group membership owned by the matching up migration.
-- admin2 retains the narrower Security Approver group and SYSADMIN_SECURITY role.

BEGIN;

DELETE FROM permissions_accounts_groups
WHERE id = 'b1000000-0000-0000-0000-000000000003'::uuid
  AND user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
  AND group_id = 'b0000000-0000-0000-0000-000000000001'::uuid;

INSERT INTO audit_logs (
    id, actor_user_id, action, module, entity_type, entity_id,
    before_json, after_json
)
VALUES (
    'd1000000-0000-0000-0000-000000000004'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'BOOTSTRAP_ADMIN_GROUP_REVOKED',
    'permissions',
    'iam_user',
    'a0000000-0000-0000-0000-000000000002',
    '{"group":"Admin","authority":"full group-derived administrator permissions","source":"migration 20260722000002"}'::jsonb,
    '{}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

COMMIT;
