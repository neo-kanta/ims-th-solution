-- Seed admin user and Admin group with full permissions.
-- Password: admin123 (bcrypt hash, cost 12)
-- IMPORTANT: Change this password in production!

-- Insert admin user
INSERT INTO iam_users (id, username, display_name, email, password_hash, is_active, is_on_leave)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin',
    'System Administrator',
    'admin@ims.local',
    '$2a$12$LJ3m4zCGq5N9B0BaC2K8kOqE1WVkzIq3z5R8jQxH1KjMqNhKDXHXi',
    true,
    false
) ON CONFLICT (username) DO NOTHING;

-- Insert Admin group
INSERT INTO permissions_groups (id, name, description, is_active)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Admin',
    'System administrators with full access to all functions and data',
    true
) ON CONFLICT (name) DO NOTHING;

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
    ('b0000000-0000-0000-0000-000000000001', 'WORKFLOW_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'WORKFLOW_EXECUTE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_CREATE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'INVESTMENT_APPROVE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'LEAVE_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'LEAVE_MANAGE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'APPROVAL_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'APPROVAL_CONFIG', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'PERMISSIONS_VIEW', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'PERMISSIONS_MANAGE', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'NOTIFICATION_CONFIG', true, 'a0000000-0000-0000-0000-000000000001'),
    ('b0000000-0000-0000-0000-000000000001', 'AUDIT_VIEW', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_id, permission_code) DO NOTHING;
