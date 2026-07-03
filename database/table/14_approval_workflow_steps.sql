-- Table: approval_workflow_steps
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE approval_workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_setting_id UUID NOT NULL REFERENCES approval_workflow_settings(id) ON DELETE CASCADE,
    step_no INTEGER NOT NULL,
    step_name VARCHAR(180) NOT NULL,
    step_description TEXT,
    approval_mode VARCHAR(40) NOT NULL DEFAULT 'SINGLE',
    approver_type VARCHAR(40) NOT NULL DEFAULT 'ROLE',
    required_role_code VARCHAR(80),
    required_group_id UUID REFERENCES permissions_groups(id),
    required_user_id UUID REFERENCES iam_users(id),
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    allow_delegation BOOLEAN NOT NULL DEFAULT false,
    is_required BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_workflow_steps_setting_step UNIQUE (workflow_setting_id, step_no),
    CONSTRAINT chk_approval_workflow_steps_mode CHECK (approval_mode IN (
        'SINGLE','ANY_OF_GROUP','ALL_OF_GROUP','QUORUM','ROLE_BASED','USER_BASED'
    )),
    CONSTRAINT chk_approval_workflow_steps_approver_type CHECK (approver_type IN (
        'ROLE','GROUP','USER','REQUEST_TARGET_OWNER'
    )),
    CONSTRAINT chk_approval_workflow_steps_min CHECK (min_approvals_required >= 1)
);
