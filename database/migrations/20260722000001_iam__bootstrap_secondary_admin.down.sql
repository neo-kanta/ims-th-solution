-- Revoke admin2 authority while preserving the IAM identity for immutable
-- approval and audit references that may point to it.

BEGIN;

DELETE FROM permission_user_role_assignments role_assignment
USING permission_roles role
WHERE role_assignment.id = 'c1000000-0000-0000-0000-000000000002'::uuid
  AND role_assignment.user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
  AND role.id = role_assignment.role_id
  AND role.role_code = 'SYSADMIN_SECURITY';

DELETE FROM permissions_accounts_groups
WHERE id = 'b1000000-0000-0000-0000-000000000002'::uuid
  AND user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
  AND group_id = 'b0000000-0000-0000-0000-000000000050'::uuid;

UPDATE permissions_groups
SET description = '!admin2-bootstrap-rolled-back',
    is_active = false,
    deleted_at = COALESCE(deleted_at, NOW()),
    updated_by = 'a0000000-0000-0000-0000-000000000001'::uuid,
    updated_at = NOW()
WHERE id = 'b0000000-0000-0000-0000-000000000050'::uuid
  AND name = 'Security Approver';

UPDATE iam_users
SET password_hash = '!admin2-bootstrap-rolled-back',
    is_active = false,
    force_password_change = false,
    locked_until = NULL,
    deleted_at = COALESCE(deleted_at, NOW()),
    updated_by = 'a0000000-0000-0000-0000-000000000001'::uuid,
    updated_at = NOW()
WHERE id = 'a0000000-0000-0000-0000-000000000002'::uuid
  AND username = 'admin2';

COMMIT;
