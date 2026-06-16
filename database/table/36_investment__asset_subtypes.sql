-- Table: investment__asset_subtypes
-- Source: 20260428000001_investment__create_reference_tables.up.sql
CREATE TABLE investment__asset_subtypes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    asset_class_id UUID NOT NULL REFERENCES investment__asset_classes (id) ON DELETE RESTRICT,
    code VARCHAR(40) NOT NULL,
    name VARCHAR(120) NOT NULL,
    description TEXT,
    display_order SMALLINT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES iam_users (id),
    updated_by UUID REFERENCES iam_users (id),
    CONSTRAINT uq_inv_asset_subtypes_class_code UNIQUE (asset_class_id, code)
);
