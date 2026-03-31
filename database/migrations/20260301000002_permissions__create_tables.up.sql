-- Permissions Management tables
-- Groups, account-group mappings, function rights, and data rights

CREATE TABLE IF NOT EXISTS permissions_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users(id),
    updated_by UUID REFERENCES iam_users(id),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_permissions_groups_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS permissions_accounts_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES permissions_groups(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES iam_users(id),
    UNIQUE(user_id, group_id)
);

CREATE TABLE IF NOT EXISTS permissions_function_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES permissions_groups(id) ON DELETE CASCADE,
    permission_code VARCHAR(100) NOT NULL,
    is_granted BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users(id),
    UNIQUE(group_id, permission_code)
);

CREATE TABLE IF NOT EXISTS permissions_data_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    contract_id VARCHAR(50) NOT NULL,
    is_granted BOOLEAN NOT NULL DEFAULT true,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by UUID REFERENCES iam_users(id),
    UNIQUE(user_id, contract_id)
);

-- Indexes for common queries
CREATE INDEX idx_permissions_ag_user ON permissions_accounts_groups(user_id);
CREATE INDEX idx_permissions_ag_group ON permissions_accounts_groups(group_id);
CREATE INDEX idx_permissions_fr_group ON permissions_function_rights(group_id);
CREATE INDEX idx_permissions_dr_user ON permissions_data_rights(user_id);

-- Trigger: auto-update updated_at on groups
CREATE TRIGGER trg_permissions_groups_updated_at
    BEFORE UPDATE ON permissions_groups
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE permissions_groups IS 'Permission groups (Manager, Trader, Risk Control, Admin, etc.)';
COMMENT ON TABLE permissions_accounts_groups IS 'User-to-group assignments';
COMMENT ON TABLE permissions_function_rights IS 'Function-level permissions assigned to groups';
COMMENT ON TABLE permissions_data_rights IS 'Data-level (contract/fund) permissions assigned to users';
