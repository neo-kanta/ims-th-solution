-- Table: iam_audit_events
-- Source: 20260301000004_iam__create_audit_events.up.sql
CREATE TABLE IF NOT EXISTS iam_audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES iam_users(id),
    event_type VARCHAR(100) NOT NULL,
    target_type VARCHAR(100),
    target_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
