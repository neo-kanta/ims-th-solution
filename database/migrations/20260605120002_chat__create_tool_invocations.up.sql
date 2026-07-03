-- chat_tool_invocations is the provenance DETAIL store. The central audit
-- table (iam_audit_events) holds the trail spine; each CHAT_TOOL_CALL audit
-- entry carries this row's id (tool_invocation_id) plus session_id +
-- message_id, so an auditor can pivot from the canonical trail to the full
-- tool call + raw result here. Append-only, like chat_messages.

CREATE TABLE IF NOT EXISTS chat_tool_invocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    message_id UUID REFERENCES chat_messages(id) ON DELETE CASCADE,
    tool_call_id VARCHAR(128) NOT NULL,          -- provider-assigned id (e.g. toolu_...)
    tool_name VARCHAR(128) NOT NULL,
    mcp_server VARCHAR(128) NOT NULL,
    arguments JSONB NOT NULL DEFAULT '{}'::jsonb, -- model-supplied args (never contains the auth token)
    raw_result JSONB,                             -- verbatim tool output (provenance source of record)
    is_error BOOLEAN NOT NULL DEFAULT FALSE,
    state VARCHAR(32) NOT NULL,                   -- requested | succeeded | failed | denied
    error TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chat_tool_invocations_session ON chat_tool_invocations(session_id, started_at);
CREATE INDEX IF NOT EXISTS idx_chat_tool_invocations_message ON chat_tool_invocations(message_id);

CREATE OR REPLACE FUNCTION prevent_chat_tool_invocation_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Chat tool invocations are append-only and cannot be modified';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_chat_tool_invocations_no_update
    BEFORE UPDATE ON chat_tool_invocations
    FOR EACH ROW
    EXECUTE FUNCTION prevent_chat_tool_invocation_update();

COMMENT ON TABLE chat_tool_invocations IS 'Append-only provenance detail for chat tool calls. Pivot target from iam_audit_events CHAT_TOOL_CALL entries.';
COMMENT ON COLUMN chat_tool_invocations.raw_result IS 'Verbatim MCP tool output. The source of record for every figure the assistant cites.';
COMMENT ON COLUMN chat_tool_invocations.arguments IS 'Model-supplied tool arguments. The auth token travels via MCP _meta and is never stored here.';
