-- Table: workflow__schedule_rules
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__schedule_rules (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    day_setting_id              UUID        NOT NULL REFERENCES workflow__day_settings(id) ON DELETE CASCADE,

    name                        VARCHAR(120) NOT NULL,
    description                 TEXT,

    scope_type                  VARCHAR(30) NOT NULL DEFAULT 'GLOBAL',
    scope_id                    UUID,

    action                      VARCHAR(40) NOT NULL,
    trigger_time_local          TIME        NOT NULL,
    timezone                    VARCHAR(64) NOT NULL DEFAULT 'Asia/Bangkok',
    days_of_week                SMALLINT[]  NOT NULL DEFAULT ARRAY[1,2,3,4,5]::SMALLINT[],
    skip_holidays               BOOLEAN     NOT NULL DEFAULT true,

    is_enabled                  BOOLEAN     NOT NULL DEFAULT true,
    priority                    INTEGER     NOT NULL DEFAULT 100,
    effective_from              DATE        NOT NULL,
    effective_to                DATE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by                  UUID        REFERENCES iam_users(id),
    updated_by                  UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_wf_schedule_rule_name UNIQUE (name),

    CONSTRAINT chk_wf_schedule_rule_scope
        CHECK (scope_type IN ('GLOBAL', 'CONTRACT', 'FUND_GROUP')),

    CONSTRAINT chk_wf_schedule_rule_scope_id
        CHECK (
            (scope_type = 'GLOBAL' AND scope_id IS NULL)
            OR (scope_type <> 'GLOBAL' AND scope_id IS NOT NULL)
        ),

    CONSTRAINT chk_wf_schedule_rule_action
        CHECK (action IN ('OPEN_DAY', 'END_DAY')),

    CONSTRAINT chk_wf_schedule_rule_days
        CHECK (
            array_length(days_of_week, 1) IS NOT NULL
            AND days_of_week <@ ARRAY[1,2,3,4,5,6,7]::SMALLINT[]
        ),

    CONSTRAINT chk_wf_schedule_rule_priority
        CHECK (priority >= 0),

    CONSTRAINT chk_wf_schedule_rule_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
