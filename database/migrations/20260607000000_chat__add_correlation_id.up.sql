-- Add a correlation_id to chat message and tool-invocation rows so one chat
-- turn can be traced end-to-end: HTTP request id -> chat turn -> audit event ->
-- tool invocation -> logs. Nullable (older rows have none). Adding a column is
-- DDL and does not trip the append-only UPDATE-blocking triggers.

ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(64);
ALTER TABLE chat_tool_invocations ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_chat_messages_correlation ON chat_messages(correlation_id);
CREATE INDEX IF NOT EXISTS idx_chat_tool_invocations_correlation ON chat_tool_invocations(correlation_id);

COMMENT ON COLUMN chat_messages.correlation_id IS 'Request correlation id (chi RequestID) of the turn that produced this message; ties to iam_audit_events metadata.';
COMMENT ON COLUMN chat_tool_invocations.correlation_id IS 'Request correlation id of the turn that issued this tool call; ties to iam_audit_events metadata.';
