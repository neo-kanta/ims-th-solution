-- Table: permission_request_comments
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_request_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES iam_users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
