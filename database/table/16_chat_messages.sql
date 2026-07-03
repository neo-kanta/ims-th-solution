-- Table: chat_messages
-- Source: 20260605120001_chat__create_messages.up.sql
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL CHECK (role IN ('system','user','assistant','tool')),
    content TEXT NOT NULL,
    provenance_map JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_provider_payload JSONB,
    correlation_id VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
