CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE iam_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_on_leave BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

CREATE INDEX idx_iam_users_username ON iam_users(username);
CREATE INDEX idx_iam_users_is_active ON iam_users(is_active);

COMMENT ON TABLE iam_users IS 'User accounts for the IMS system';
COMMENT ON COLUMN iam_users.is_on_leave IS 'When true, user is on approved leave and restricted from normal operations';
