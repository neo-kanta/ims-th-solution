-- Table: permission_change_request_labels
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_change_request_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES permission_labels(id) ON DELETE CASCADE,
    added_by UUID REFERENCES iam_users(id),
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_permission_change_request_labels UNIQUE (request_id, label_id)
);
