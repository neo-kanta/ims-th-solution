-- Seed admin user and Admin group with full permissions.
-- Password: admin123 (bcrypt hash, cost 12)
-- IMPORTANT: Change this password immediately in production!
-- force_password_change is true to enforce first-login password reset.

-- Insert admin user
INSERT INTO iam_users (id, username, display_name, email, password_hash, is_active, force_password_change)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    'System Administrator',
    'admin@ims.local',
    '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C',
    true,
    true
) ON CONFLICT ON CONSTRAINT uq_iam_users_username DO NOTHING;

-- Insert Admin group
INSERT INTO permissions_groups (id, name, description, is_active)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Admin',
    'System administrators with full access to all functions and data',
    true
) ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO NOTHING;

-- Assign admin user to Admin group
INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
) ON CONFLICT (user_id, group_id) DO NOTHING;

-- Grant core function permissions to Admin group
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
VALUES
    -- IAM permissions
    ('b0000000-0000-0000-0000-000000000001', 'IAM_USER_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'IAM_USER_CREATE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'IAM_USER_UPDATE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'IAM_USER_DEACTIVATE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'IAM_AUDIT_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Workflow permissions
    ('b0000000-0000-0000-0000-000000000001', 'WORKFLOW_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'WORKFLOW_EXECUTE', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Investment permissions
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_CREATE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_APPROVE', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Leave permissions
    ('b0000000-0000-0000-0000-000000000001', 'LEAVE_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'LEAVE_MANAGE', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Approval permissions
    ('b0000000-0000-0000-0000-000000000001', 'APPROVAL_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'APPROVAL_CONFIG', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Permission management
    ('b0000000-0000-0000-0000-000000000001', 'PERMISSIONS_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'PERMISSIONS_MANAGE', true, 'a0000000-0000-0000-0000-000000000001'),
    -- Other
    ('b0000000-0000-0000-0000-000000000001', 'NOTIFICATION_CONFIG', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'AUDIT_VIEW', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_id, permission_code) DO NOTHING;
