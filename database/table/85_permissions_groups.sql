-- Table: permissions_groups
-- Source: 20260301000002_permissions__create_tables.up.sql
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
