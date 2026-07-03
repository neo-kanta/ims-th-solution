-- Table: permission_user_role_assignments
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_user_role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES permission_roles(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    assigned_by UUID REFERENCES iam_users(id),
    approved_by UUID REFERENCES iam_users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_user_role_assignments_status
        CHECK (status IN ('PENDING','APPROVED','REJECTED','REVOKED','EXPIRED')),
    CONSTRAINT chk_permission_user_role_assignments_effective_window
        CHECK (effective_to IS NULL OR effective_from IS NULL OR effective_to > effective_from)
);
