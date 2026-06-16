-- Table: audit_logs
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES iam_users(id),
    action VARCHAR(120) NOT NULL,
    module VARCHAR(80) NOT NULL,
    entity_type VARCHAR(120),
    entity_id VARCHAR(160),
    before_json JSONB NOT NULL DEFAULT '{}',
    after_json JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    correlation_id VARCHAR(160),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
