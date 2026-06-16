-- Table: permission_request_revisions
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_request_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    revision_no INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES iam_users(id),
    change_summary TEXT,
    snapshot_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_request_revisions_request_revision UNIQUE (request_id, revision_no)
);
