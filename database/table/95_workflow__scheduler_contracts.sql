-- Table: workflow__scheduler_contracts
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__scheduler_contracts (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    contract_id         UUID        NOT NULL,
    contract_code       VARCHAR(80) NOT NULL,
    contract_name       VARCHAR(255),

    fund_id             UUID,
    fund_code           VARCHAR(80),

    is_active           BOOLEAN     NOT NULL DEFAULT true,
    effective_from      DATE        NOT NULL,
    effective_to        DATE,

    source              VARCHAR(60) NOT NULL DEFAULT 'TEMP_WORKFLOW_BRIDGE',
    metadata            JSONB       NOT NULL DEFAULT '{}',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID        REFERENCES iam_users(id),
    updated_by          UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_wf_scheduler_contracts_contract UNIQUE (contract_id),

    CONSTRAINT chk_wf_scheduler_contracts_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
