-- =============================================================================
-- Ben workflow-approver seed
-- =============================================================================
-- Grants Ben (username: ben, id: a0000000-...0010) the ability to perform
-- Manager Approval on the workflow day-state machine.
--
-- Two records are required:
--   1. permissions_accounts_groups — adds Ben to the "Fund Manager" group
--      (id: b0000000-...0011), which already holds WORKFLOW_APPROVE and
--      WORKFLOW_VIEW in permissions_function_rights. This is the only way to
--      grant a system-level permission code to a user (permissions_function_rights
--      is group-scoped; the newer permission_function_rights table FKs to
--      permission_function_definitions which only contains permissions-module
--      screen-level codes, not workflow codes).
--
--   2. workflow__approval_settings — makes Ben a configured approver for
--      MANAGER_APPROVE so the backend's checkDailyPermission passes.
--      (Admin users bypass this check; non-admins must be listed here.)
--
-- Idempotent: safe to run multiple times.
-- Depends on: 002 (Fund Manager group), 004 (Ben user).
-- =============================================================================

BEGIN;

-- 1. Add Ben to the Fund Manager group so he receives WORKFLOW_APPROVE
--    and WORKFLOW_VIEW at login.
INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
VALUES (
    'a0000000-0000-0000-0000-000000000010',   -- ben
    'b0000000-0000-0000-0000-000000000011',   -- Fund Manager
    'a0000000-0000-0000-0000-000000000001'    -- admin
)
ON CONFLICT (user_id, group_id) DO NOTHING;

-- 2. Register Ben as a configured approver for the MANAGER_APPROVE operation.
--    The backend's checkDailyPermission matches approver_account_code against
--    claims.Username (which equals 'ben').
--    workflow__approval_settings has no unique constraint; guard with EXISTS.
INSERT INTO workflow__approval_settings (
    operation_type,
    approval_mode,
    approver_account_code,
    approver_username,
    approver_role,
    is_active,
    updated_by
)
SELECT
    'MANAGER_APPROVE',
    'ANY_OF',
    'ben',
    'Ben',
    'Manager',
    true,
    'admin'
WHERE NOT EXISTS (
    SELECT 1
    FROM workflow__approval_settings
    WHERE operation_type = 'MANAGER_APPROVE'
      AND approver_account_code = 'ben'
);

COMMIT;
