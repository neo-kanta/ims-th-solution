-- Table: workflow__control_decisions
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__control_decisions (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    workflow_day_id     UUID        REFERENCES workflow__day_states(id) ON DELETE CASCADE,
    contract_id         UUID        NOT NULL,
    business_date       DATE        NOT NULL,

    decision_scope      VARCHAR(40) NOT NULL,
    process_step        VARCHAR(60),
    decision_level      VARCHAR(30) NOT NULL,
    decision_status     VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    blocks_work         BOOLEAN     NOT NULL DEFAULT false,

    requested_by        UUID        REFERENCES iam_users(id),
    requested_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_by          UUID        REFERENCES iam_users(id),
    decided_at          TIMESTAMPTZ,

    reason              TEXT,
    notes               TEXT,
    metadata            JSONB       NOT NULL DEFAULT '{}',

    superseded_at       TIMESTAMPTZ,
    superseded_by       UUID        REFERENCES iam_users(id),

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_wf_control_decision_scope
        CHECK (decision_scope IN ('DAY_START', 'INTRADAY', 'DAY_END', 'INVESTMENT_PROCESS')),

    CONSTRAINT chk_wf_control_decision_process_step
        CHECK (
            decision_scope = 'INVESTMENT_PROCESS'
            OR process_step IS NULL
        ),

    CONSTRAINT chk_wf_control_decision_level
        CHECK (decision_level IN ('MANAGER', 'SUPERVISOR', 'HIGH_LEVEL', 'SYSTEM')),

    CONSTRAINT chk_wf_control_decision_status
        CHECK (decision_status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')),

    CONSTRAINT chk_wf_control_decision_decided
        CHECK (
            (decision_status IN ('PENDING', 'CANCELLED') AND decided_at IS NULL)
            OR (decision_status IN ('APPROVED', 'REJECTED') AND decided_at IS NOT NULL)
        ),

    CONSTRAINT chk_wf_control_decision_superseded
        CHECK (
            (superseded_at IS NULL AND superseded_by IS NULL)
            OR (superseded_at IS NOT NULL)
        )
);
