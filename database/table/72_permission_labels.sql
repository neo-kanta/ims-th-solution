-- Table: permission_labels
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label_code VARCHAR(100) NOT NULL,
    label_name VARCHAR(140) NOT NULL,
    label_type VARCHAR(40) NOT NULL,
    color VARCHAR(20) NOT NULL DEFAULT '#6e7781',
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_labels_code UNIQUE (label_code)
);
