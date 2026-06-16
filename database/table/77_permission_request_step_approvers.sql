-- Table: permission_request_step_approvers
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_request_step_approvers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    approval_step_id UUID NOT NULL REFERENCES permission_request_approval_steps(id) ON DELETE CASCADE,
    approver_user_id UUID REFERENCES iam_users(id),
    approver_role_code VARCHAR(80),
    approval_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
    approval_comment TEXT,
    is_delegated BOOLEAN NOT NULL DEFAULT false,
    delegated_from_user_id UUID REFERENCES iam_users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_permission_request_step_approvers_status CHECK (
        approval_status IN ('PENDING','APPROVED','CHANGES_REQUESTED','REJECTED')
    )
);
