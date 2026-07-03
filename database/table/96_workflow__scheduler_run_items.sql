-- Table: workflow__scheduler_run_items
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__scheduler_run_items (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    scheduler_run_id    UUID        NOT NULL REFERENCES workflow__scheduler_runs(id) ON DELETE CASCADE,
    rule_id             UUID        REFERENCES workflow__schedule_rules(id) ON DELETE SET NULL,

    contract_id         UUID        NOT NULL,
    business_date       DATE        NOT NULL,
    action              VARCHAR(40) NOT NULL,
    status              VARCHAR(30) NOT NULL,

    skip_reason         VARCHAR(80),
    error_message       TEXT,

    workflow_day_id     UUID        REFERENCES workflow__day_states(id) ON DELETE SET NULL,
    transition_id       UUID        REFERENCES workflow__transition_log(id) ON DELETE SET NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_wf_scheduler_item_action
        CHECK (action IN ('OPEN_DAY', 'END_DAY')),

    CONSTRAINT chk_wf_scheduler_item_status
        CHECK (status IN ('SUCCESS', 'SKIPPED', 'FAILED')),

    CONSTRAINT uq_wf_scheduler_item_contract_action
        UNIQUE (scheduler_run_id, rule_id, contract_id, action)
);
