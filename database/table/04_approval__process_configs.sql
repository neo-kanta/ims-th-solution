-- Table: approval__process_configs
-- Source: 20260529000001_approval__create_tables.up.sql
CREATE TABLE approval__process_configs (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    process_code           VARCHAR(60)  NOT NULL,
    process_name           VARCHAR(160) NOT NULL,
    process_type           VARCHAR(40)  NOT NULL,
    contract_type          VARCHAR(20)  NOT NULL DEFAULT 'COMPANY',
    contract_id            UUID,
    effective_date         DATE         NOT NULL DEFAULT CURRENT_DATE,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    group_approval_enabled BOOLEAN      NOT NULL DEFAULT false,
    require_team_approval   BOOLEAN     NOT NULL DEFAULT false,
    created_by             UUID         REFERENCES iam_users(id),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by             UUID         REFERENCES iam_users(id),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_approval_process_code UNIQUE (process_code),
    CONSTRAINT chk_approval_process_type CHECK (process_type IN (
        'INVESTMENT_ANALYSIS_REPORT','INVESTMENT_DECISION','INVESTMENT_CANCELLATION',
        'WORKFLOW_OPERATION','LEAVE_REQUEST','LEAVE_CANCELLATION','DELEGATION_REQUEST',
        'PORTFOLIO_ONBOARDING','COMPLIANCE_RELEASE'
    )),
    CONSTRAINT chk_approval_contract_type CHECK (contract_type IN ('FUND','DISCRETIONARY','COMPANY'))
);
