-- Table: workflow__scheduler_runs
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__scheduler_runs (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    scheduler_name      VARCHAR(100) NOT NULL,
    rule_id             UUID        REFERENCES workflow__schedule_rules(id) ON DELETE SET NULL,
    action              VARCHAR(40),

    business_date       DATE        NOT NULL,
    timezone            VARCHAR(64) NOT NULL DEFAULT 'Asia/Bangkok',

    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at         TIMESTAMPTZ,
    status              VARCHAR(30) NOT NULL DEFAULT 'RUNNING',

    triggered_by        VARCHAR(30) NOT NULL DEFAULT 'SYSTEM',
    locked_by           VARCHAR(255) NOT NULL DEFAULT '',

    summary             JSONB       NOT NULL DEFAULT '{}',
    error_message       TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_wf_scheduler_run_action
        CHECK (action IS NULL OR action IN ('OPEN_DAY', 'END_DAY')),

    CONSTRAINT chk_wf_scheduler_run_status
        CHECK (status IN ('RUNNING', 'SUCCESS', 'PARTIAL_SUCCESS', 'FAILED', 'SKIPPED')),

    CONSTRAINT chk_wf_scheduler_run_triggered_by
        CHECK (triggered_by IN ('SYSTEM', 'MANUAL')),

    CONSTRAINT chk_wf_scheduler_run_time_order
        CHECK (finished_at IS NULL OR finished_at >= started_at)
);
