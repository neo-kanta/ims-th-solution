-- Table: permissions_accounts_groups
-- Source: 20260301000002_permissions__create_tables.up.sql
CREATE TABLE IF NOT EXISTS permissions_accounts_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES permissions_groups(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES iam_users(id),
    UNIQUE(user_id, group_id)
);
