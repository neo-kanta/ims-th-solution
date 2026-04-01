-- IAM Sessions table for server-side refresh token management.
-- Supports token rotation with family-based breach detection.

CREATE TABLE IF NOT EXISTS iam_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam_users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    token_family UUID NOT NULL,
    ip_address INET,
    user_agent TEXT,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at TIMESTAMPTZ
);

CREATE INDEX idx_iam_sessions_user_id ON iam_sessions(user_id);
CREATE INDEX idx_iam_sessions_token_family ON iam_sessions(token_family);
CREATE INDEX idx_iam_sessions_refresh_token ON iam_sessions(refresh_token_hash) WHERE is_revoked = false;
CREATE INDEX idx_iam_sessions_expires_at ON iam_sessions(expires_at);

COMMENT ON TABLE iam_sessions IS 'Server-side refresh token sessions with rotation tracking';
COMMENT ON COLUMN iam_sessions.refresh_token_hash IS 'SHA-256 hash of the refresh token (never store plaintext)';
COMMENT ON COLUMN iam_sessions.token_family IS 'Token family UUID; all rotated tokens share this — used for breach detection';
COMMENT ON COLUMN iam_sessions.rotated_at IS 'When this token was rotated (replaced by a new one)';
