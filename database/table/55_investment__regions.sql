-- Table: investment__regions
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(20) NOT NULL,
    name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_regions_code UNIQUE (code)
);
