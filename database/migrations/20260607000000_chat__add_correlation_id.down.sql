DROP INDEX IF EXISTS idx_chat_tool_invocations_correlation;
DROP INDEX IF EXISTS idx_chat_messages_correlation;

ALTER TABLE chat_tool_invocations DROP COLUMN IF EXISTS correlation_id;
ALTER TABLE chat_messages DROP COLUMN IF EXISTS correlation_id;
