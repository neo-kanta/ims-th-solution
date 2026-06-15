-- Table: workflow__day_settings
-- Source: 20260427000001_workflow__add_settings_scheduler_tables.up.sql
CREATE TABLE workflow__day_settings (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    name                        VARCHAR(120) NOT NULL,
    description                 TEXT,

    -- Scope allows a global default first, then contract/fund overrides later.
    scope_type                  VARCHAR(30) NOT NULL DEFAULT 'GLOBAL',
    scope_id                    UUID,

    timezone                    VARCHAR(64) NOT NULL DEFAULT 'Asia/Bangkok',
    work_start_time             TIME        NOT NULL,
    work_end_time               TIME        NOT NULL,
    scheduler_interval_minutes  SMALLINT    NOT NULL DEFAULT 60,

    auto_start_enabled          BOOLEAN     NOT NULL DEFAULT true,
    auto_end_enabled            BOOLEAN     NOT NULL DEFAULT true,
    skip_non_business_days      BOOLEAN     NOT NULL DEFAULT true,

    -- Manager approval / rejection controls.
    requires_manager_approval   BOOLEAN     NOT NULL DEFAULT true,
    allow_high_level_override   BOOLEAN     NOT NULL DEFAULT true,
    block_on_rejection          BOOLEAN     NOT NULL DEFAULT true,

    is_active                   BOOLEAN     NOT NULL DEFAULT true,
    effective_from              DATE        NOT NULL,
    effective_to                DATE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by                  UUID        REFERENCES iam_users(id),
    updated_by                  UUID        REFERENCES iam_users(id),

    CONSTRAINT uq_wf_day_settings_name UNIQUE (name),

    CONSTRAINT chk_wf_day_settings_scope
        CHECK (scope_type IN ('GLOBAL', 'CONTRACT', 'FUND_GROUP')),

    CONSTRAINT chk_wf_day_settings_scope_id
        CHECK (
            (scope_type = 'GLOBAL' AND scope_id IS NULL)
            OR (scope_type <> 'GLOBAL' AND scope_id IS NOT NULL)
        ),

    CONSTRAINT chk_wf_day_settings_work_window
        CHECK (work_start_time < work_end_time),

    CONSTRAINT chk_wf_day_settings_interval
        CHECK (scheduler_interval_minutes > 0 AND scheduler_interval_minutes <= 1440),

    CONSTRAINT chk_wf_day_settings_effective_range
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);
