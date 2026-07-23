-- Bootstrap a second named security approver for maker-checker permission
-- requests without granting full application-administrator authority.
--
-- Development-only initial password: admin123
-- The account is forced to change this password at first login. The backend
-- bootstrap guard refuses non-development startup while this hash remains.

BEGIN;

DO $$
BEGIN
    -- A disabled row with this sentinel is owned by the matching down
    -- migration and may be safely reactivated. Any other collision fails
    -- closed so an existing human identity is never hijacked.
    IF EXISTS (
        SELECT 1
        FROM iam_users
        WHERE id = 'a0000000-0000-0000-0000-000000000002'::uuid
           OR lower(username) = 'admin2'
    ) AND NOT EXISTS (
        SELECT 1
        FROM iam_users
        WHERE id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND username = 'admin2'
          AND password_hash = '!admin2-bootstrap-rolled-back'
          AND is_active = false
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'admin2 bootstrap identity conflicts with an existing IAM user';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM permissions_groups
        WHERE id = 'b0000000-0000-0000-0000-000000000050'::uuid
           OR name = 'Security Approver'
    ) AND NOT EXISTS (
        SELECT 1
        FROM permissions_groups
        WHERE id = 'b0000000-0000-0000-0000-000000000050'::uuid
          AND name = 'Security Approver'
          AND description = '!admin2-bootstrap-rolled-back'
          AND is_active = false
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Security Approver bootstrap group conflicts with an existing permission group';
    END IF;

    IF EXISTS (
        SELECT 1 FROM permissions_accounts_groups
        WHERE id = 'b1000000-0000-0000-0000-000000000002'::uuid
    ) THEN
        RAISE EXCEPTION 'admin2 Security Approver membership id is already in use';
    END IF;

    IF EXISTS (
        SELECT 1 FROM permission_user_role_assignments
        WHERE id = 'c1000000-0000-0000-0000-000000000002'::uuid
    ) THEN
        RAISE EXCEPTION 'admin2 security-role assignment id is already in use';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM permission_roles
        WHERE role_code = 'SYSADMIN_SECURITY'
          AND (is_active = false OR can_approve_role_assignment = false)
    ) THEN
        RAISE EXCEPTION 'SYSADMIN_SECURITY role is inactive or cannot approve role assignments';
    END IF;
END
$$;

INSERT INTO iam_users (
    id, username, display_name, email, password_hash, is_active,
    force_password_change, created_by, updated_by
)
VALUES (
    'a0000000-0000-0000-0000-000000000002'::uuid,
    'admin2',
    'Secondary Security Approver',
    'admin2@ims.local',
    '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C',
    true,
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid
)
ON CONFLICT ON CONSTRAINT uq_iam_users_username DO UPDATE SET
    display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    is_active = true,
    force_password_change = true,
    failed_login_attempts = 0,
    locked_until = NULL,
    deleted_at = NULL,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW()
WHERE iam_users.id = EXCLUDED.id
  AND iam_users.password_hash = '!admin2-bootstrap-rolled-back';

INSERT INTO permissions_groups (
    id, name, description, is_active, created_by, updated_by
)
VALUES (
    'b0000000-0000-0000-0000-000000000050'::uuid,
    'Security Approver',
    'Independent reviewers for permission change requests; no merge or application-administration authority.',
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid
)
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO UPDATE SET
    description = EXCLUDED.description,
    is_active = true,
    deleted_at = NULL,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW()
WHERE permissions_groups.id = EXCLUDED.id
  AND permissions_groups.description = '!admin2-bootstrap-rolled-back';

-- These definitions are also maintained by seed 008. Keeping the canonical
-- definitions here makes migrate-only installations operational.
INSERT INTO permission_function_definitions (
    code, module, screen, action, name, description
)
VALUES
    (
        'permission.change_request.review', 'permission', 'change_request',
        'review', 'Review change requests', 'Review permission requests.'
    ),
    (
        'permission.change_request.approve', 'permission', 'change_request',
        'approve', 'Approve change requests', 'Approve permission request steps.'
    )
ON CONFLICT (code) DO NOTHING;

INSERT INTO permission_function_rights (
    subject_type, subject_id, permission_code,
    can_view, can_search, can_add, can_edit, can_delete, can_approve,
    can_revoke_approval, can_export, can_configure, approved_at, approved_by
)
VALUES
    (
        'GROUP', 'b0000000-0000-0000-0000-000000000050'::uuid,
        'permission.change_request.review',
        true, false, false, false, false, false, false, false, false,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid
    ),
    (
        'GROUP', 'b0000000-0000-0000-0000-000000000050'::uuid,
        'permission.change_request.approve',
        true, false, false, false, false, true, false, false, false,
        NOW(), 'a0000000-0000-0000-0000-000000000001'::uuid
    )
ON CONFLICT (subject_type, subject_id, permission_code) DO UPDATE SET
    can_view = EXCLUDED.can_view,
    can_search = EXCLUDED.can_search,
    can_add = EXCLUDED.can_add,
    can_edit = EXCLUDED.can_edit,
    can_delete = EXCLUDED.can_delete,
    can_approve = EXCLUDED.can_approve,
    can_revoke_approval = EXCLUDED.can_revoke_approval,
    can_export = EXCLUDED.can_export,
    can_configure = EXCLUDED.can_configure,
    approved_at = EXCLUDED.approved_at,
    approved_by = EXCLUDED.approved_by,
    updated_at = NOW();

INSERT INTO permissions_accounts_groups (id, user_id, group_id, assigned_by)
VALUES (
    'b1000000-0000-0000-0000-000000000002'::uuid,
    'a0000000-0000-0000-0000-000000000002'::uuid,
    'b0000000-0000-0000-0000-000000000050'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid
);

INSERT INTO permission_roles (
    role_code, role_name, department, role_category, priority_rank,
    assignment_scope, can_request_role_assignment,
    can_approve_role_assignment, is_high_risk, is_active, description
)
VALUES (
    'SYSADMIN_SECURITY', 'Security System Administrator', 'Technology',
    'Technical', 85, 'TECHNICAL_ONLY', true, true, true, true,
    'Technical security administrator.'
)
ON CONFLICT (role_code) DO NOTHING;

INSERT INTO permission_user_role_assignments (
    id, user_id, role_id, status, assigned_by, approved_by,
    approved_at, effective_from
)
SELECT
    'c1000000-0000-0000-0000-000000000002'::uuid,
    'a0000000-0000-0000-0000-000000000002'::uuid,
    role.id,
    'APPROVED',
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    NOW(),
    NOW()
FROM permission_roles role
WHERE role.role_code = 'SYSADMIN_SECURITY'
  AND role.is_active = true;

DO $$
BEGIN
    IF 2 <> (
        SELECT COUNT(*)
        FROM iam_users admin2_user
        JOIN permissions_accounts_groups account_group
          ON account_group.user_id = admin2_user.id
        JOIN permissions_groups permission_group
          ON permission_group.id = account_group.group_id
         AND permission_group.name = 'Security Approver'
         AND permission_group.is_active = true
         AND permission_group.deleted_at IS NULL
        JOIN permission_function_rights function_right
          ON function_right.subject_type = 'GROUP'
         AND function_right.subject_id = permission_group.id
        WHERE admin2_user.id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND admin2_user.username = 'admin2'
          AND admin2_user.is_active = true
          AND admin2_user.deleted_at IS NULL
          AND function_right.permission_code IN (
              'permission.change_request.review',
              'permission.change_request.approve'
          )
          AND (function_right.can_view OR function_right.can_approve)
    ) OR NOT EXISTS (
        SELECT 1
        FROM permission_user_role_assignments role_assignment
        JOIN permission_roles role ON role.id = role_assignment.role_id
        WHERE role_assignment.user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND role_assignment.status = 'APPROVED'
          AND role_assignment.effective_to IS NULL
          AND role.role_code = 'SYSADMIN_SECURITY'
          AND role.is_active = true
          AND role.can_approve_role_assignment = true
    ) THEN
        RAISE EXCEPTION 'failed to bootstrap admin2 as an active security approver';
    END IF;
END
$$;

COMMIT;
