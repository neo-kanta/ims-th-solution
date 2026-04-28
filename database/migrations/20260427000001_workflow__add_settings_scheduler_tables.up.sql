-- =============================================================================
-- Workflow Management Module - Phase 1 Settings and Scheduler Tables
-- =============================================================================
-- This migration adds configuration and audit tables for:
--   - working window settings, e.g. 08:30-17:30 Asia/Bangkok
--   - scheduler rules for automatic day start / day end checks
--   - scheduler tick audit records and per-contract/rule run items
--   - temporary active-contract bridge for scheduler scope
--   - manager / supervisor / high-level control decisions, including rejection
--
-- Runtime transitions still belong to workflow__day_states and
-- workflow__transition_log. The scheduler must call the application command
-- layer instead of writing runtime state directly.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Day settings: working window and automation behavior
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_day_settings_scope ON workflow__day_settings (scope_type, scope_id);
CREATE INDEX idx_wf_day_settings_active_eff ON workflow__day_settings (is_active, effective_from, effective_to);

CREATE TRIGGER trg_wf_day_settings_updated_at
    BEFORE UPDATE ON workflow__day_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE workflow__day_settings IS 'Configures the workflow working window and automation behavior.';
COMMENT ON COLUMN workflow__day_settings.work_start_time IS 'Local time after which automatic OPEN_DAY may run.';
COMMENT ON COLUMN workflow__day_settings.work_end_time IS 'Local time after which automatic END_DAY logic may run.';
COMMENT ON COLUMN workflow__day_settings.block_on_rejection IS 'When true, rejection decisions block workflow and investment process actions.';

-- ---------------------------------------------------------------------------
-- 2. Temporary scheduler contract bridge
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_scheduler_contracts_active_eff
    ON workflow__scheduler_contracts (is_active, effective_from, effective_to);
CREATE INDEX idx_wf_scheduler_contracts_fund
    ON workflow__scheduler_contracts (fund_id);

CREATE TRIGGER trg_wf_scheduler_contracts_updated_at
    BEFORE UPDATE ON workflow__scheduler_contracts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE workflow__scheduler_contracts IS
    'Temporary active contract bridge for the workflow scheduler until an authoritative contract/fund master module exists.';
COMMENT ON COLUMN workflow__scheduler_contracts.effective_to IS
    'NULL means the contract remains active after effective_from.';

-- ---------------------------------------------------------------------------
-- 3. Schedule rules: specific automated actions checked by the hourly job
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_schedule_rules_setting ON workflow__schedule_rules (day_setting_id);
CREATE INDEX idx_wf_schedule_rules_action_enabled ON workflow__schedule_rules (action, is_enabled, trigger_time_local);
CREATE INDEX idx_wf_schedule_rules_scope ON workflow__schedule_rules (scope_type, scope_id);
CREATE INDEX idx_wf_schedule_rules_priority ON workflow__schedule_rules (priority);

CREATE TRIGGER trg_wf_schedule_rules_updated_at
    BEFORE UPDATE ON workflow__schedule_rules
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE workflow__schedule_rules IS 'Defines automatic workflow actions checked by the Go scheduler.';

-- ---------------------------------------------------------------------------
-- 4. Scheduler runs: one row per hourly scheduler tick / execution attempt
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_scheduler_runs_name_date ON workflow__scheduler_runs (scheduler_name, business_date);
CREATE INDEX idx_wf_scheduler_runs_status ON workflow__scheduler_runs (status);
CREATE INDEX idx_wf_scheduler_runs_started_at ON workflow__scheduler_runs (started_at DESC);

COMMENT ON TABLE workflow__scheduler_runs IS
    'Tick-level audit header for every workflow scheduler execution attempt. Rule/action detail is stored in workflow__scheduler_run_items and summary JSON.';

-- ---------------------------------------------------------------------------
-- 5. Scheduler run items: per-contract/rule result for each scheduler tick
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_scheduler_items_run ON workflow__scheduler_run_items (scheduler_run_id);
CREATE INDEX idx_wf_scheduler_items_contract_date ON workflow__scheduler_run_items (contract_id, business_date);
CREATE INDEX idx_wf_scheduler_items_status ON workflow__scheduler_run_items (status);

COMMENT ON TABLE workflow__scheduler_run_items IS 'Per-contract and per-rule outcome for a workflow scheduler tick.';

-- ---------------------------------------------------------------------------
-- 6. Control decisions: manager/high-level approval or rejection guard
-- ---------------------------------------------------------------------------
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

CREATE INDEX idx_wf_control_decisions_day ON workflow__control_decisions (workflow_day_id);
CREATE INDEX idx_wf_control_decisions_contract_date ON workflow__control_decisions (contract_id, business_date);
CREATE INDEX idx_wf_control_decisions_scope_status ON workflow__control_decisions (decision_scope, decision_status);
CREATE INDEX idx_wf_control_decisions_blocking ON workflow__control_decisions (contract_id, business_date, decision_scope)
WHERE blocks_work = true AND decision_status = 'REJECTED' AND superseded_at IS NULL;

CREATE TRIGGER trg_wf_control_decisions_updated_at
    BEFORE UPDATE ON workflow__control_decisions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE workflow__control_decisions IS 'Manager/supervisor/high-level decisions that can approve or reject workflow/investment work.';
COMMENT ON COLUMN workflow__control_decisions.blocks_work IS 'When true with REJECTED status, all workflow/investment actions in scope must be blocked.';
