DROP TRIGGER IF EXISTS trg_chat_messages_no_update ON chat_messages;
DROP FUNCTION IF EXISTS prevent_chat_message_update();
DROP INDEX IF EXISTS idx_chat_messages_session_created;
DROP TABLE IF EXISTS chat_messages;
