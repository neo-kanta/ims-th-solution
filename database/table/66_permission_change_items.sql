-- Table: permission_change_items
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE permission_change_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES permission_change_requests(id) ON DELETE CASCADE,
    item_type VARCHAR(60) NOT NULL,
    target_table VARCHAR(100),
    target_id VARCHAR(120),
    action_type VARCHAR(60) NOT NULL,
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
