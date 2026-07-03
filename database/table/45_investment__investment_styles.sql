-- Table: investment__investment_styles
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__investment_styles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_styles_code UNIQUE (code)
);
