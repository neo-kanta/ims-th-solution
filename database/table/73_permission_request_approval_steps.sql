-- Table: permission_request_approval_steps
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_request_approval_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    workflow_setting_id UUID NOT NULL REFERENCES approval_workflow_settings(id) ON DELETE RESTRICT,
    step_no INTEGER NOT NULL,
    step_name VARCHAR(180) NOT NULL,
    approval_mode VARCHAR(40) NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'NOT_STARTED',
    min_approvals_required INTEGER NOT NULL DEFAULT 1,
    approvals_received INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_request_approval_steps_request_step UNIQUE (request_id, step_no),
    CONSTRAINT chk_permission_request_approval_steps_status CHECK (status IN (
        'NOT_STARTED','PENDING','APPROVED','CHANGES_REQUESTED','REJECTED','SKIPPED'
    )),
    CONSTRAINT chk_permission_request_approval_steps_counts CHECK (
        min_approvals_required >= 1 AND approvals_received >= 0
    )
);
