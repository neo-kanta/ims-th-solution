-- IAM Audit Events table — immutable append-only log.
-- Trigger prevents UPDATE and DELETE to ensure tamper-resistance.

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

CREATE INDEX idx_iam_audit_events_actor ON iam_audit_events(actor_id);
CREATE INDEX idx_iam_audit_events_type ON iam_audit_events(event_type);
CREATE INDEX idx_iam_audit_events_created ON iam_audit_events(created_at);
CREATE INDEX idx_iam_audit_events_target ON iam_audit_events(target_type, target_id);

-- Prevent modification of audit events (immutable)
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit events are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_iam_audit_events_no_update
    BEFORE UPDATE ON iam_audit_events
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_modification();

CREATE TRIGGER trg_iam_audit_events_no_delete
    BEFORE DELETE ON iam_audit_events
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_modification();

COMMENT ON TABLE iam_audit_events IS 'Immutable audit trail for IAM actions (login, logout, password change, etc.)';
COMMENT ON COLUMN iam_audit_events.event_type IS 'Event type: LOGIN_SUCCESS, LOGIN_FAILURE, LOGOUT, PASSWORD_CHANGE, TOKEN_REFRESH, ACCOUNT_LOCKED, etc.';
COMMENT ON COLUMN iam_audit_events.metadata IS 'Additional event-specific data as JSON';
