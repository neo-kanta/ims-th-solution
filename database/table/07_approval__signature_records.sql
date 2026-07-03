-- Table: approval__signature_records
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__signature_records (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id UUID         NOT NULL REFERENCES approval__requests(id) ON DELETE CASCADE,
    stage_number        INTEGER      NOT NULL,
    signer_user_id      UUID         NOT NULL REFERENCES iam_users(id),
    signer_display_name VARCHAR(255) NOT NULL DEFAULT '',
    signer_title        VARCHAR(160),
    signed_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_proxy_signature  BOOLEAN      NOT NULL DEFAULT false,
    proxy_for_user_id   UUID         REFERENCES iam_users(id),
    signature_label     VARCHAR(20)  NOT NULL DEFAULT 'NORMAL',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_approval_signature_label CHECK (signature_label IN ('NORMAL','DELEGATED'))
);
