DROP TRIGGER IF EXISTS trg_chat_tool_invocations_no_update ON chat_tool_invocations;
DROP FUNCTION IF EXISTS prevent_chat_tool_invocation_update();
DROP INDEX IF EXISTS idx_chat_tool_invocations_message;
DROP INDEX IF EXISTS idx_chat_tool_invocations_session;
DROP TABLE IF EXISTS chat_tool_invocations;
