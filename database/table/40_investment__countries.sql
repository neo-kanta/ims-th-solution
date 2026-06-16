-- Table: investment__countries
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__countries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    iso_code CHAR(2) NOT NULL,
    name VARCHAR(120) NOT NULL,
    region_id UUID REFERENCES investment__regions (id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_countries_iso UNIQUE (iso_code),
    CONSTRAINT chk_inv_countries_iso_format CHECK (iso_code ~ '^[A-Z]{2}$')
);
