-- Table: permission_roles
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_code VARCHAR(80) NOT NULL,
    role_name VARCHAR(160) NOT NULL,
    department VARCHAR(120),
    role_category VARCHAR(80),
    priority_rank INTEGER NOT NULL,
    assignment_scope VARCHAR(80) NOT NULL,
    can_request_role_assignment BOOLEAN NOT NULL DEFAULT false,
    can_approve_role_assignment BOOLEAN NOT NULL DEFAULT false,
    is_high_risk BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_roles_code UNIQUE (role_code),
    CONSTRAINT chk_permission_roles_priority CHECK (priority_rank >= 0)
);
