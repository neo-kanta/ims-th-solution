-- Table: investment__sectors
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__sectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    parent_id UUID REFERENCES investment__sectors (id) ON DELETE RESTRICT,
    level SMALLINT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_sectors_parent_code UNIQUE (parent_id, code),
    CONSTRAINT chk_inv_sectors_level CHECK (level BETWEEN 1 AND 4)
);
