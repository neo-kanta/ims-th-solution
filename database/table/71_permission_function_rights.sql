-- Table: permission_function_rights
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_function_rights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type VARCHAR(20) NOT NULL,
    subject_id UUID NOT NULL,
    permission_code VARCHAR(140) NOT NULL REFERENCES permission_function_definitions(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    can_view BOOLEAN NOT NULL DEFAULT false,
    can_search BOOLEAN NOT NULL DEFAULT false,
    can_add BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    can_delete BOOLEAN NOT NULL DEFAULT false,
    can_approve BOOLEAN NOT NULL DEFAULT false,
    can_revoke_approval BOOLEAN NOT NULL DEFAULT false,
    can_export BOOLEAN NOT NULL DEFAULT false,
    can_configure BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by UUID REFERENCES iam_users(id),
    CONSTRAINT uq_permission_function_rights_subject_code UNIQUE (subject_type, subject_id, permission_code),
    CONSTRAINT chk_permission_function_rights_subject CHECK (subject_type IN ('USER','GROUP','ROLE'))
);
