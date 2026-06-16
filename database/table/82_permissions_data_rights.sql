-- Table: permissions_data_rights
-- Source: 20260301000002_permissions__create_tables.up.sql
CREATE TABLE IF NOT EXISTS permissions_data_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    contract_id VARCHAR(50) NOT NULL,
    is_granted BOOLEAN NOT NULL DEFAULT true,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by UUID REFERENCES iam_users(id),
    UNIQUE(user_id, contract_id)
);
