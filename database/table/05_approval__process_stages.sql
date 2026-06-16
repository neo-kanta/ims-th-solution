-- Table: approval__process_stages
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__process_stages (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    process_config_id       UUID         NOT NULL REFERENCES approval__process_configs(id) ON DELETE CASCADE,
    stage_number            INTEGER      NOT NULL,
    stage_name              VARCHAR(160) NOT NULL DEFAULT '',
    approver_mode           VARCHAR(20)  NOT NULL,
    approver_user_id        UUID         REFERENCES iam_users(id),
    approval_group_id       UUID         REFERENCES approval__groups(id),
    required_approval_count INTEGER      NOT NULL DEFAULT 1,
    is_final_stage          BOOLEAN      NOT NULL DEFAULT false,
    reject_policy           VARCHAR(20)  NOT NULL DEFAULT 'STOP',
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_process_stage UNIQUE (process_config_id, stage_number),
    CONSTRAINT chk_approval_stage_number CHECK (stage_number BETWEEN 1 AND 3),
    CONSTRAINT chk_approval_stage_mode CHECK (approver_mode IN (
        'SINGLE_USER','GROUP_PRIORITY','GROUP_ANY','TEAM_MINIMUM'
    )),
    CONSTRAINT chk_approval_stage_reject_policy CHECK (reject_policy IN ('STOP')),
    CONSTRAINT chk_approval_stage_required_count CHECK (required_approval_count >= 1)
);
