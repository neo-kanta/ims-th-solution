-- Table: permissions_function_rights
-- Source: 20260301000002_permissions__create_tables.up.sql
CREATE TABLE IF NOT EXISTS permissions_function_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES permissions_groups(id) ON DELETE CASCADE,
    permission_code VARCHAR(100) NOT NULL,
    is_granted BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users(id),
    UNIQUE(group_id, permission_code),
    CONSTRAINT fk_permissions_function_rights_definition
        FOREIGN KEY (permission_code)
        REFERENCES permissions_function_definitions (code)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);
