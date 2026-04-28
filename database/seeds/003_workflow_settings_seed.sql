-- =============================================================================
-- Workflow settings seed: default Thai working window and scheduler rules
-- =============================================================================

BEGIN;

INSERT INTO workflow__day_settings (
    id,
    name,
    description,
    scope_type,
    scope_id,
    timezone,
    work_start_time,
    work_end_time,
    scheduler_interval_minutes,
    auto_start_enabled,
    auto_end_enabled,
    skip_non_business_days,
    requires_manager_approval,
    allow_high_level_override,
    block_on_rejection,
    is_active,
    effective_from,
    effective_to,
    created_by,
    updated_by
)
VALUES (
    '77000000-0000-0000-0000-000000000001',
    'Default Thai Business Day',
    'Default workflow setting: scheduler checks hourly, opens after 08:30, ends after 17:30, and blocks work on manager/high-level rejection.',
    'GLOBAL',
    NULL,
    'Asia/Bangkok',
    TIME '08:30',
    TIME '17:30',
    60,
    true,
    true,
    true,
    true,
    true,
    true,
    true,
    DATE '2026-01-01',
    NULL,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT ON CONSTRAINT uq_wf_day_settings_name DO UPDATE
SET
    description = EXCLUDED.description,
    timezone = EXCLUDED.timezone,
    work_start_time = EXCLUDED.work_start_time,
    work_end_time = EXCLUDED.work_end_time,
    scheduler_interval_minutes = EXCLUDED.scheduler_interval_minutes,
    auto_start_enabled = EXCLUDED.auto_start_enabled,
    auto_end_enabled = EXCLUDED.auto_end_enabled,
    skip_non_business_days = EXCLUDED.skip_non_business_days,
    requires_manager_approval = EXCLUDED.requires_manager_approval,
    allow_high_level_override = EXCLUDED.allow_high_level_override,
    block_on_rejection = EXCLUDED.block_on_rejection,
    is_active = EXCLUDED.is_active,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

INSERT INTO workflow__schedule_rules (
    id,
    day_setting_id,
    name,
    description,
    scope_type,
    scope_id,
    action,
    trigger_time_local,
    timezone,
    days_of_week,
    skip_holidays,
    is_enabled,
    priority,
    effective_from,
    effective_to,
    created_by,
    updated_by
)
VALUES
    (
        '77000000-0000-0000-0000-000000000010',
        '77000000-0000-0000-0000-000000000001',
        'Default Open Day Rule',
        'Hourly scheduler may open workflow days from 08:30 Asia/Bangkok on Thai business days.',
        'GLOBAL',
        NULL,
        'OPEN_DAY',
        TIME '08:30',
        'Asia/Bangkok',
        ARRAY[1,2,3,4,5]::SMALLINT[],
        true,
        true,
        100,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '77000000-0000-0000-0000-000000000011',
        '77000000-0000-0000-0000-000000000001',
        'Default End Day Rule',
        'Hourly scheduler may end workflow days from 17:30 Asia/Bangkok on Thai business days.',
        'GLOBAL',
        NULL,
        'END_DAY',
        TIME '17:30',
        'Asia/Bangkok',
        ARRAY[1,2,3,4,5]::SMALLINT[],
        true,
        true,
        100,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_wf_schedule_rule_name DO UPDATE
SET
    description = EXCLUDED.description,
    action = EXCLUDED.action,
    trigger_time_local = EXCLUDED.trigger_time_local,
    timezone = EXCLUDED.timezone,
    days_of_week = EXCLUDED.days_of_week,
    skip_holidays = EXCLUDED.skip_holidays,
    is_enabled = EXCLUDED.is_enabled,
    priority = EXCLUDED.priority,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

COMMIT;
