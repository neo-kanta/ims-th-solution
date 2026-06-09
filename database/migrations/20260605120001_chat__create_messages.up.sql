-- chat_messages is append-only conversation transcript material. UPDATE is
-- blocked at the row level; DELETE is permitted only via ON DELETE CASCADE
-- from chat_sessions so user-initiated session deletion (Slice D) can
-- proceed. The central audit log retains every turn event independently —
-- those rows are immutable in audit's own table — so deleting a session
-- never erases the audit trail.

CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL CHECK (role IN ('system','user','assistant','tool')),
    content TEXT NOT NULL,
    provenance_map JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_provider_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_session_created
    ON chat_messages(session_id, created_at);

CREATE OR REPLACE FUNCTION prevent_chat_message_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Chat messages are append-only and cannot be modified';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_chat_messages_no_update
    BEFORE UPDATE ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION prevent_chat_message_update();

COMMENT ON TABLE chat_messages IS 'Append-only transcript of one chat session. UPDATE blocked by trigger; DELETE only via session cascade.';
COMMENT ON COLUMN chat_messages.provenance_map IS 'Slice D figure-to-tool-result links. {} until Slice D wires the guardrail.';
COMMENT ON COLUMN chat_messages.raw_provider_payload IS 'Optional accumulated provider response for audit. NULL in Slice A.';
