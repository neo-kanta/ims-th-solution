-- Table: chat_tool_invocations
-- Source: 20260605120002_chat__create_tool_invocations.up.sql
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
