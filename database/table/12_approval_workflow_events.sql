-- Table: approval_workflow_events
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE approval_workflow_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    approval_step_id UUID REFERENCES permission_request_approval_steps(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES iam_users(id),
    event_type VARCHAR(80) NOT NULL,
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
