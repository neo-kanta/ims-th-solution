-- =============================================================================
-- Permission Management + Approval Workflow seed
-- =============================================================================
-- Idempotent seed for role hierarchy, approval workflow settings, labels,
-- notification templates, and the permission-management function catalog.
-- =============================================================================

BEGIN;

INSERT INTO permission_roles (
    role_code, role_name, department, role_category, priority_rank, assignment_scope,
    can_request_role_assignment, can_approve_role_assignment, is_high_risk,
    is_active, description
)
VALUES
    ('CIO', 'Chief Investment Officer', 'Investment', 'Executive', 100, 'ALL_BUSINESS', true, true, true, true, 'Highest permission-management approval authority.'),
    ('HEAD_INVESTMENT', 'Head of Investment', 'Investment', 'Business', 90, 'INVESTMENT', true, true, true, true, 'Investment department approver.'),
    ('COMPLIANCE_OFFICER', 'Compliance Officer', 'Compliance', 'Control', 85, 'COMPLIANCE', true, true, true, true, 'Compliance control approver.'),
    ('RISK_MANAGER', 'Risk Manager', 'Risk', 'Control', 85, 'RISK', true, true, true, true, 'Risk control approver.'),
    ('SYSADMIN_SECURITY', 'Security System Administrator', 'Technology', 'Technical', 85, 'TECHNICAL_ONLY', true, true, true, true, 'Technical security administrator.'),
    ('TRADING_SUPERVISOR', 'Trading Supervisor', 'Trading', 'Business', 80, 'TRADING', true, true, true, true, 'Trading approval authority.'),
    ('NAV_OFFICER', 'NAV Officer', 'Accounting', 'Operations', 80, 'ACCOUNTING', true, true, true, true, 'Accounting and NAV approval authority.'),
    ('PORTFOLIO_MANAGER', 'Portfolio Manager', 'Investment', 'Business', 75, 'INVESTMENT', true, true, false, true, 'Portfolio manager with investment scope.'),
    ('FUND_ACCOUNTANT', 'Fund Accountant', 'Accounting', 'Operations', 70, 'ACCOUNTING', true, false, false, true, 'Fund accounting operator.'),
    ('INVESTMENT_OPS_OFFICER', 'Investment Operations Officer', 'Operations', 'Operations', 70, 'OPERATIONS', true, false, false, true, 'Investment operations operator.'),
    ('MIDDLE_OFFICE_OFFICER', 'Middle Office Officer', 'Operations', 'Operations', 70, 'OPERATIONS', true, false, false, true, 'Middle office operator.'),
    ('PRODUCT_MANAGER', 'Product Manager', 'Product', 'Business', 65, 'PRODUCT', true, false, false, true, 'Product management operator.'),
    ('INTERNAL_AUDITOR', 'Internal Auditor', 'Audit', 'Control', 65, 'AUDIT_READ_ONLY', false, false, true, true, 'Read-only internal audit role.'),
    ('TRADER', 'Trader', 'Trading', 'Business', 60, 'TRADING', false, false, false, true, 'Trading operator.'),
    ('RESEARCH_ANALYST', 'Research Analyst', 'Investment', 'Business', 60, 'INVESTMENT', false, false, false, true, 'Investment research analyst.'),
    ('ASSISTANT_PM', 'Assistant Portfolio Manager', 'Investment', 'Business', 55, 'INVESTMENT', false, false, false, true, 'Assistant portfolio manager.'),
    ('TRANSFER_AGENT_OFFICER', 'Transfer Agent Officer', 'Transfer Agency', 'Operations', 55, 'TA', false, false, false, true, 'Transfer agency operator.'),
    ('DATA_ENGINEER', 'Data Engineer', 'Technology', 'Technical', 50, 'TECHNICAL_ONLY', false, false, false, true, 'Data engineering operator.'),
    ('IT_APP_SUPPORT', 'IT Application Support', 'Technology', 'Technical', 50, 'TECHNICAL_ONLY', false, false, false, true, 'Application support operator.'),
    ('SALES_RM', 'Sales Relationship Manager', 'Sales', 'Business', 45, 'SALES', false, false, false, true, 'Sales relationship manager.')
ON CONFLICT (role_code) DO UPDATE SET
    role_name = EXCLUDED.role_name,
    department = EXCLUDED.department,
    role_category = EXCLUDED.role_category,
    priority_rank = EXCLUDED.priority_rank,
    assignment_scope = EXCLUDED.assignment_scope,
    can_request_role_assignment = EXCLUDED.can_request_role_assignment,
    can_approve_role_assignment = EXCLUDED.can_approve_role_assignment,
    is_high_risk = EXCLUDED.is_high_risk,
    is_active = EXCLUDED.is_active,
    description = EXCLUDED.description,
    updated_at = NOW();

INSERT INTO permission_role_assignment_policies (
    grantor_role_id, max_target_priority_exclusive, assignment_scope,
    min_approvals_required, can_request, can_direct_merge
)
SELECT id, priority_rank, assignment_scope, 1, can_request_role_assignment, false
FROM permission_roles
WHERE can_request_role_assignment = true
ON CONFLICT DO NOTHING;

-- Give the seeded admin user enough bootstrap authority to exercise the
-- approval workflow. This is still auditable and future changes must use
-- permission change requests.
INSERT INTO permission_user_role_assignments (
    user_id, role_id, status, assigned_by, approved_by, approved_at, effective_from
)
SELECT
    'a0000000-0000-0000-0000-000000000001'::uuid,
    r.id,
    'APPROVED',
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    NOW(),
    NOW()
FROM permission_roles r
WHERE r.role_code IN ('CIO', 'SYSADMIN_SECURITY')
ON CONFLICT DO NOTHING;

INSERT INTO permission_function_definitions (code, module, screen, action, name, description)
VALUES
    ('permission.users.view', 'permission', 'users', 'view', 'View users', 'View permission-management user accounts.'),
    ('permission.users.create', 'permission', 'users', 'create', 'Create users', 'Create user accounts through permission workflow.'),
    ('permission.users.edit', 'permission', 'users', 'edit', 'Edit users', 'Edit user account metadata.'),
    ('permission.users.lock', 'permission', 'users', 'lock', 'Lock users', 'Lock user accounts.'),
    ('permission.users.unlock', 'permission', 'users', 'unlock', 'Unlock users', 'Unlock user accounts.'),
    ('permission.users.activate', 'permission', 'users', 'activate', 'Activate users', 'Activate user accounts.'),
    ('permission.users.deactivate', 'permission', 'users', 'deactivate', 'Deactivate users', 'Deactivate user accounts.'),
    ('permission.users.reset_password', 'permission', 'users', 'reset_password', 'Reset passwords', 'Reset user passwords.'),
    ('permission.groups.view', 'permission', 'groups', 'view', 'View groups', 'View permission groups.'),
    ('permission.groups.create', 'permission', 'groups', 'create', 'Create groups', 'Create permission groups.'),
    ('permission.groups.edit', 'permission', 'groups', 'edit', 'Edit groups', 'Edit permission groups.'),
    ('permission.groups.approve', 'permission', 'groups', 'approve', 'Approve groups', 'Approve group changes.'),
    ('permission.membership.assign', 'permission', 'membership', 'assign', 'Assign membership', 'Request account and group mapping changes.'),
    ('permission.membership.remove', 'permission', 'membership', 'remove', 'Remove membership', 'Request group membership removal.'),
    ('permission.membership.approve', 'permission', 'membership', 'approve', 'Approve membership', 'Approve group membership changes.'),
    ('permission.roles.view', 'permission', 'roles', 'view', 'View roles', 'View role hierarchy.'),
    ('permission.roles.assign', 'permission', 'roles', 'assign', 'Assign roles', 'Request role assignment changes.'),
    ('permission.roles.approve', 'permission', 'roles', 'approve', 'Approve roles', 'Approve role assignment changes.'),
    ('permission.function_rights.view', 'permission', 'function_rights', 'view', 'View function rights', 'View function permissions.'),
    ('permission.function_rights.edit', 'permission', 'function_rights', 'edit', 'Edit function rights', 'Request function-permission changes.'),
    ('permission.function_rights.approve', 'permission', 'function_rights', 'approve', 'Approve function rights', 'Approve function-permission changes.'),
    ('permission.data_rights.view', 'permission', 'data_rights', 'view', 'View data rights', 'View fund, contract, and portfolio permissions.'),
    ('permission.data_rights.edit', 'permission', 'data_rights', 'edit', 'Edit data rights', 'Request data-permission changes.'),
    ('permission.data_rights.approve', 'permission', 'data_rights', 'approve', 'Approve data rights', 'Approve data-permission changes.'),
    ('permission.change_request.create', 'permission', 'change_request', 'create', 'Create change requests', 'Create draft permission change requests.'),
    ('permission.change_request.submit', 'permission', 'change_request', 'submit', 'Submit change requests', 'Submit permission requests for approval.'),
    ('permission.change_request.review', 'permission', 'change_request', 'review', 'Review change requests', 'Review permission requests.'),
    ('permission.change_request.approve', 'permission', 'change_request', 'approve', 'Approve change requests', 'Approve permission request steps.'),
    ('permission.change_request.merge', 'permission', 'change_request', 'merge', 'Merge change requests', 'Apply approved permission changes.'),
    ('permission.change_request.reject', 'permission', 'change_request', 'reject', 'Reject change requests', 'Reject permission requests.'),
    ('permission.change_request.close', 'permission', 'change_request', 'close', 'Close change requests', 'Close permission requests.'),
    ('permission.change_request.cancel', 'permission', 'change_request', 'cancel', 'Cancel change requests', 'Cancel permission requests.'),
    ('permission.audit.view', 'permission', 'audit', 'view', 'View audit logs', 'View permission audit logs.'),
    ('permission.audit.export', 'permission', 'audit', 'export', 'Export audit logs', 'Export permission audit logs.'),
    ('permission.notification.view', 'permission', 'notification', 'view', 'View notification settings', 'View permission notification settings.'),
    ('permission.notification.edit', 'permission', 'notification', 'edit', 'Edit notification settings', 'Edit permission notification settings.')
ON CONFLICT (code) DO UPDATE SET
    module = EXCLUDED.module,
    screen = EXCLUDED.screen,
    action = EXCLUDED.action,
    name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO permission_function_rights (
    subject_type, subject_id, permission_code,
    can_view, can_search, can_add, can_edit, can_delete, can_approve,
    can_revoke_approval, can_export, can_configure, approved_at, approved_by
)
SELECT
    'GROUP',
    g.id,
    d.code,
    true, true, true, true, true, true, true, true, true,
    NOW(),
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN permission_function_definitions d
WHERE g.name = 'Admin'
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

INSERT INTO approval_workflow_settings (
    setting_code, setting_name, module, request_type, risk_level, is_active,
    is_default, sequential_approval, allow_creator_approval,
    failed_check_blocks_submit, failed_check_blocks_merge, description
)
VALUES
    ('PERMISSION_LOW_RISK_DEFAULT', 'Permission low-risk default', 'permission', 'PERMISSION_CHANGE', 'LOW', true, true, true, false, true, true, 'Default low-risk permission workflow.'),
    ('PERMISSION_MEDIUM_RISK_INVESTMENT', 'Permission medium-risk investment', 'permission', 'ROLE_ASSIGNMENT', 'MEDIUM', true, false, true, false, true, true, 'Medium-risk investment role and permission workflow.'),
    ('PERMISSION_MEDIUM_RISK_TRADING', 'Permission medium-risk trading', 'permission', 'TRADING_PERMISSION', 'MEDIUM', true, false, true, false, true, true, 'Medium-risk trading workflow.'),
    ('PERMISSION_HIGH_RISK_ADMIN', 'Permission high-risk admin', 'permission', 'ADMIN_PERMISSION', 'HIGH', true, true, true, false, true, true, 'High-risk administrative permission workflow.'),
    ('PERMISSION_HIGH_RISK_AUDIT', 'Permission high-risk audit', 'permission', 'AUDIT_PERMISSION', 'HIGH', true, false, true, false, true, true, 'High-risk audit permission workflow.'),
    ('PERMISSION_DATA_RIGHTS_CHANGE', 'Permission data-rights change', 'permission', 'DATA_RIGHTS_CHANGE', 'MEDIUM', true, false, true, false, true, true, 'Fund, contract, and portfolio data-rights workflow.')
ON CONFLICT (setting_code) DO UPDATE SET
    setting_name = EXCLUDED.setting_name,
    module = EXCLUDED.module,
    request_type = EXCLUDED.request_type,
    risk_level = EXCLUDED.risk_level,
    is_active = EXCLUDED.is_active,
    is_default = EXCLUDED.is_default,
    sequential_approval = EXCLUDED.sequential_approval,
    allow_creator_approval = EXCLUDED.allow_creator_approval,
    failed_check_blocks_submit = EXCLUDED.failed_check_blocks_submit,
    failed_check_blocks_merge = EXCLUDED.failed_check_blocks_merge,
    description = EXCLUDED.description,
    updated_at = NOW();

DELETE FROM approval_workflow_steps
WHERE workflow_setting_id IN (
    SELECT id FROM approval_workflow_settings
    WHERE setting_code IN (
        'PERMISSION_LOW_RISK_DEFAULT',
        'PERMISSION_MEDIUM_RISK_INVESTMENT',
        'PERMISSION_MEDIUM_RISK_TRADING',
        'PERMISSION_HIGH_RISK_ADMIN',
        'PERMISSION_HIGH_RISK_AUDIT',
        'PERMISSION_DATA_RIGHTS_CHANGE'
    )
);

INSERT INTO approval_workflow_steps (
    workflow_setting_id, step_no, step_name, step_description, approval_mode,
    approver_type, required_role_code, min_approvals_required, allow_delegation, is_required
)
SELECT s.id, v.step_no, v.step_name, v.step_description, v.approval_mode,
       v.approver_type, v.required_role_code, v.min_approvals_required,
       v.allow_delegation, true
FROM approval_workflow_settings s
JOIN (VALUES
    ('PERMISSION_LOW_RISK_DEFAULT', 1, 'SYSADMIN_SECURITY approval', 'Security administrator approval.', 'SINGLE', 'ROLE', 'SYSADMIN_SECURITY', 1, false),
    ('PERMISSION_MEDIUM_RISK_INVESTMENT', 1, 'HEAD_INVESTMENT approval', 'Investment head approval.', 'SINGLE', 'ROLE', 'HEAD_INVESTMENT', 1, false),
    ('PERMISSION_MEDIUM_RISK_INVESTMENT', 2, 'COMPLIANCE_OFFICER approval', 'Compliance approval.', 'SINGLE', 'ROLE', 'COMPLIANCE_OFFICER', 1, false),
    ('PERMISSION_MEDIUM_RISK_TRADING', 1, 'TRADING_SUPERVISOR approval', 'Trading supervisor approval.', 'SINGLE', 'ROLE', 'TRADING_SUPERVISOR', 1, false),
    ('PERMISSION_MEDIUM_RISK_TRADING', 2, 'COMPLIANCE_OFFICER approval', 'Compliance approval.', 'SINGLE', 'ROLE', 'COMPLIANCE_OFFICER', 1, false),
    ('PERMISSION_HIGH_RISK_ADMIN', 1, 'SYSADMIN_SECURITY approval', 'Security administrator approval.', 'SINGLE', 'ROLE', 'SYSADMIN_SECURITY', 1, false),
    ('PERMISSION_HIGH_RISK_ADMIN', 2, 'COMPLIANCE_OFFICER approval', 'Compliance approval.', 'SINGLE', 'ROLE', 'COMPLIANCE_OFFICER', 1, false),
    ('PERMISSION_HIGH_RISK_ADMIN', 3, 'CIO approval', 'CIO approval.', 'SINGLE', 'ROLE', 'CIO', 1, false),
    ('PERMISSION_HIGH_RISK_AUDIT', 1, 'COMPLIANCE_OFFICER approval', 'Compliance approval.', 'SINGLE', 'ROLE', 'COMPLIANCE_OFFICER', 1, false),
    ('PERMISSION_HIGH_RISK_AUDIT', 2, 'CIO approval', 'CIO approval.', 'SINGLE', 'ROLE', 'CIO', 1, false),
    ('PERMISSION_DATA_RIGHTS_CHANGE', 1, 'REQUEST_TARGET_OWNER approval', 'Target fund or portfolio owner approval.', 'SINGLE', 'REQUEST_TARGET_OWNER', NULL, 1, false),
    ('PERMISSION_DATA_RIGHTS_CHANGE', 2, 'COMPLIANCE_OFFICER approval', 'Compliance approval.', 'SINGLE', 'ROLE', 'COMPLIANCE_OFFICER', 1, false)
) AS v(setting_code, step_no, step_name, step_description, approval_mode, approver_type, required_role_code, min_approvals_required, allow_delegation)
    ON v.setting_code = s.setting_code;

INSERT INTO permission_labels (label_code, label_name, label_type, color, description, is_system, is_active)
VALUES
    ('risk:low', 'Low risk', 'RISK', '#1a7f37', 'Low-risk permission change.', true, true),
    ('risk:medium', 'Medium risk', 'RISK', '#9a6700', 'Medium-risk permission change.', true, true),
    ('risk:high', 'High risk', 'RISK', '#cf222e', 'High-risk permission change.', true, true),
    ('risk:critical', 'Critical risk', 'RISK', '#8250df', 'Critical-risk permission change.', true, true),
    ('priority:low', 'Low priority', 'PRIORITY', '#6e7781', 'Low-priority request.', true, true),
    ('priority:normal', 'Normal priority', 'PRIORITY', '#0969da', 'Normal-priority request.', true, true),
    ('priority:high', 'High priority', 'PRIORITY', '#bc4c00', 'High-priority request.', true, true),
    ('priority:urgent', 'Urgent priority', 'PRIORITY', '#cf222e', 'Urgent request.', true, true),
    ('module:permission', 'Permission', 'MODULE', '#0969da', 'Permission module request.', true, true),
    ('module:workflow', 'Workflow', 'MODULE', '#8250df', 'Workflow module request.', true, true),
    ('module:investment', 'Investment', 'MODULE', '#1a7f37', 'Investment module request.', true, true),
    ('module:compliance', 'Compliance', 'MODULE', '#cf222e', 'Compliance module request.', true, true),
    ('module:audit', 'Audit', 'MODULE', '#6e7781', 'Audit module request.', true, true),
    ('compliance-warning', 'Compliance warning', 'COMPLIANCE', '#9a6700', 'Compliance warning present.', true, true),
    ('compliance-blocker', 'Compliance blocker', 'COMPLIANCE', '#cf222e', 'Compliance blocker present.', true, true),
    ('override-required', 'Override required', 'COMPLIANCE', '#bc4c00', 'Override required before merge.', true, true),
    ('awaiting-review', 'Awaiting review', 'STATUS', '#0969da', 'Submitted and waiting for reviewers.', true, true),
    ('changes-requested', 'Changes requested', 'STATUS', '#bc4c00', 'Changes requested by reviewer.', true, true),
    ('blocked', 'Blocked', 'STATUS', '#cf222e', 'Required check failed.', true, true),
    ('ready-to-merge', 'Ready to merge', 'STATUS', '#1a7f37', 'All required approvals and checks are complete.', true, true),
    ('scope:role-assignment', 'Role assignment', 'SCOPE', '#0969da', 'Role assignment scope.', true, true),
    ('scope:function-rights', 'Function rights', 'SCOPE', '#8250df', 'Function-rights scope.', true, true),
    ('scope:data-rights', 'Data rights', 'SCOPE', '#1a7f37', 'Data-rights scope.', true, true),
    ('scope:user-management', 'User management', 'SCOPE', '#6e7781', 'User-management scope.', true, true),
    ('scope:group-management', 'Group management', 'SCOPE', '#6e7781', 'Group-management scope.', true, true)
ON CONFLICT (label_code) DO UPDATE SET
    label_name = EXCLUDED.label_name,
    label_type = EXCLUDED.label_type,
    color = EXCLUDED.color,
    description = EXCLUDED.description,
    is_system = EXCLUDED.is_system,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

INSERT INTO notification_settings (event_code, channel, enabled, target_scope, template_subject, template_body)
VALUES
    ('permission.request.submitted', 'IN_APP', true, 'APPROVERS', 'Permission request submitted', 'A permission request is ready for review.'),
    ('permission.request.changes_requested', 'IN_APP', true, 'REQUESTER', 'Changes requested', 'A reviewer requested changes on your permission request.'),
    ('permission.request.approved', 'IN_APP', true, 'REQUESTER', 'Permission request approved', 'Your permission request has been approved.'),
    ('permission.request.merged', 'IN_APP', true, 'REQUESTER', 'Permission request merged', 'Approved permission changes have been applied.'),
    ('permission.request.check_failed', 'IN_APP', true, 'REQUESTER', 'Permission request blocked', 'A required validation check failed.')
ON CONFLICT (event_code, channel) DO UPDATE SET
    enabled = EXCLUDED.enabled,
    target_scope = EXCLUDED.target_scope,
    template_subject = EXCLUDED.template_subject,
    template_body = EXCLUDED.template_body,
    updated_at = NOW();

COMMIT;
