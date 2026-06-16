-- Table: permission_role_assignment_policies
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_role_assignment_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grantor_role_id UUID NOT NULL REFERENCES permission_roles(id) ON DELETE CASCADE,
    max_target_priority_exclusive INTEGER NOT NULL,
    assignment_scope VARCHAR(80) NOT NULL,
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    can_request BOOLEAN NOT NULL DEFAULT true,
    can_direct_merge BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_role_assignment_policies_priority
        CHECK (max_target_priority_exclusive >= 0),
    CONSTRAINT chk_permission_role_assignment_policies_approvals
        CHECK (min_approvals_required >= 1)
);
