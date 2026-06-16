-- Table: investment__fund_categories
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__fund_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    asset_class_id UUID REFERENCES investment__asset_classes (id) ON DELETE RESTRICT,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_fund_categories_code UNIQUE (code)
);
