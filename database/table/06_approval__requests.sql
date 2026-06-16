-- Table: approval__requests
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__requests (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    request_number       VARCHAR(40)  NOT NULL,
    process_type         VARCHAR(40)  NOT NULL,
    process_config_id    UUID         REFERENCES approval__process_configs(id) ON DELETE RESTRICT,
    subject_type         VARCHAR(40)  NOT NULL,
    subject_id           UUID         NOT NULL,
    subject_title        VARCHAR(255) NOT NULL DEFAULT '',
    subject_reference    VARCHAR(120) NOT NULL DEFAULT '',
    contract_id          UUID,
    portfolio_id         UUID,
    submitter_id         UUID         NOT NULL REFERENCES iam_users(id) ON DELETE RESTRICT,
    submitted_at         TIMESTAMPTZ,
    current_stage_number INTEGER      NOT NULL DEFAULT 1,
    status               VARCHAR(20)  NOT NULL DEFAULT 'DRAFT',
    final_decision_by    UUID         REFERENCES iam_users(id),
    final_decision_at    TIMESTAMPTZ,
    rejection_reason     TEXT         NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_request_number UNIQUE (request_number),
    CONSTRAINT chk_approval_request_status CHECK (status IN (
        'DRAFT','SUBMITTED','PENDING_APPROVAL','APPROVED','REJECTED','CANCELLED','WITHDRAWN'
    ))
);
