-- Table: notification_settings
-- Source: 20260521000001_permissions__approval_workflow.up.sql
CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_code VARCHAR(100) NOT NULL,
    channel VARCHAR(40) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    target_scope VARCHAR(80) NOT NULL DEFAULT 'APPROVERS',
    template_subject VARCHAR(255) NOT NULL DEFAULT '',
    template_body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_notification_settings_event_channel UNIQUE (event_code, channel)
);
