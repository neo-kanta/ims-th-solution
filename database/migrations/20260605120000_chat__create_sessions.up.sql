-- chat_sessions holds one row per conversation. Provider + model are
-- captured at session-create time so audit and history reflect what was
-- actually used (provider switches at runtime via LLM_PROVIDER apply only
-- to new sessions).

CREATE TABLE IF NOT EXISTS chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id),
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_user ON chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated ON chat_sessions(updated_at DESC);

COMMENT ON TABLE chat_sessions IS 'One row per conversation. Owned by user_id; provider/model captured at create time.';
COMMENT ON COLUMN chat_sessions.provider IS 'Canonical LLM provider id: anthropic | openai | gemini | openai_compatible.';
