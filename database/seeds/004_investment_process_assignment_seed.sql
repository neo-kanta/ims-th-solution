-- =============================================================================
-- Investment process assignment seed
-- =============================================================================
-- Seeds:
--   Group A: Ben + Green
--   Group B: Neo
-- Both groups can execute:
--   Analysis Report, Investment Decisions, Investment Execution, Investment Review
-- =============================================================================

BEGIN;

-- Development users for process-assignment testing.
-- Password hash is the same dev hash used by the admin seed; force password
-- change remains true so these are not production-ready credentials.
INSERT INTO iam_users (
    id,
    username,
    display_name,
    email,
    password_hash,
    is_active,
    force_password_change,
    created_by,
    updated_by
)
VALUES
    (
        'a0000000-0000-0000-0000-000000000010',
        'ben',
        'Ben',
        'ben@ims.local',
        '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C',
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'a0000000-0000-0000-0000-000000000011',
        'green',
        'Green',
        'green@ims.local',
        '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C',
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'a0000000-0000-0000-0000-000000000012',
        'neo',
        'Neo',
        'neo@ims.local',
        '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C',
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_iam_users_username DO UPDATE
SET
    display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    is_active = EXCLUDED.is_active,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

INSERT INTO investment__process_steps (
    id,
    step_key,
    name,
    description,
    sequence_no,
    requires_workflow_open,
    blocks_after_rejection,
    is_active,
    created_by,
    updated_by
)
VALUES
    (
        '88000000-0000-0000-0000-000000000001',
        'ANALYSIS_REPORT',
        'Analysis Report',
        'Research analyst creates analysis report and recommendation.',
        1,
        true,
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000002',
        'INVESTMENT_DECISION',
        'Investment Decisions',
        'Fund manager records investment decision linked to analysis.',
        2,
        true,
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000003',
        'INVESTMENT_EXECUTION',
        'Investment Execution',
        'Trader or authorized operator executes approved investment order.',
        3,
        true,
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000004',
        'INVESTMENT_REVIEW',
        'Investment Review',
        'Authorized reviewer checks completed investment activity.',
        4,
        true,
        true,
        true,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_investment_process_steps_key DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    sequence_no = EXCLUDED.sequence_no,
    requires_workflow_open = EXCLUDED.requires_workflow_open,
    blocks_after_rejection = EXCLUDED.blocks_after_rejection,
    is_active = EXCLUDED.is_active,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

INSERT INTO investment__process_groups (
    id,
    group_key,
    name,
    description,
    is_active,
    effective_from,
    effective_to,
    created_by,
    updated_by
)
VALUES
    (
        '88000000-0000-0000-0000-000000000101',
        'GROUP_A',
        'Group A',
        'Ben and Green can execute all investment process steps.',
        true,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000102',
        'GROUP_B',
        'Group B',
        'Neo can execute all investment process steps.',
        true,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_investment_process_groups_key DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

INSERT INTO investment__process_group_members (
    id,
    group_id,
    user_id,
    member_role,
    is_active,
    effective_from,
    effective_to,
    assigned_by
)
VALUES
    (
        '88000000-0000-0000-0000-000000000201',
        '88000000-0000-0000-0000-000000000101',
        'a0000000-0000-0000-0000-000000000010',
        'MEMBER',
        true,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000202',
        '88000000-0000-0000-0000-000000000101',
        'a0000000-0000-0000-0000-000000000011',
        'MEMBER',
        true,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        '88000000-0000-0000-0000-000000000203',
        '88000000-0000-0000-0000-000000000102',
        'a0000000-0000-0000-0000-000000000012',
        'LEAD',
        true,
        DATE '2026-01-01',
        NULL,
        'a0000000-0000-0000-0000-000000000001'
    )
ON CONFLICT ON CONSTRAINT uq_investment_process_group_member DO UPDATE
SET
    member_role = EXCLUDED.member_role,
    is_active = EXCLUDED.is_active,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    assigned_by = EXCLUDED.assigned_by;

INSERT INTO investment__process_step_assignments (
    id,
    process_step_id,
    assignment_type,
    process_group_id,
    user_id,
    scope_type,
    scope_id,
    can_execute,
    is_active,
    priority,
    effective_from,
    effective_to,
    created_by,
    updated_by
)
SELECT
    CASE
        WHEN s.step_key = 'ANALYSIS_REPORT' AND g.group_key = 'GROUP_A'
            THEN '88000000-0000-0000-0000-000000000301'::uuid
        WHEN s.step_key = 'INVESTMENT_DECISION' AND g.group_key = 'GROUP_A'
            THEN '88000000-0000-0000-0000-000000000302'::uuid
        WHEN s.step_key = 'INVESTMENT_EXECUTION' AND g.group_key = 'GROUP_A'
            THEN '88000000-0000-0000-0000-000000000303'::uuid
        WHEN s.step_key = 'INVESTMENT_REVIEW' AND g.group_key = 'GROUP_A'
            THEN '88000000-0000-0000-0000-000000000304'::uuid
        WHEN s.step_key = 'ANALYSIS_REPORT' AND g.group_key = 'GROUP_B'
            THEN '88000000-0000-0000-0000-000000000305'::uuid
        WHEN s.step_key = 'INVESTMENT_DECISION' AND g.group_key = 'GROUP_B'
            THEN '88000000-0000-0000-0000-000000000306'::uuid
        WHEN s.step_key = 'INVESTMENT_EXECUTION' AND g.group_key = 'GROUP_B'
            THEN '88000000-0000-0000-0000-000000000307'::uuid
        ELSE '88000000-0000-0000-0000-000000000308'::uuid
    END,
    s.id,
    'GROUP',
    g.id,
    NULL,
    'GLOBAL',
    NULL,
    true,
    true,
    100,
    DATE '2026-01-01',
    NULL,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
FROM investment__process_steps s
CROSS JOIN investment__process_groups g
WHERE s.step_key IN (
    'ANALYSIS_REPORT',
    'INVESTMENT_DECISION',
    'INVESTMENT_EXECUTION',
    'INVESTMENT_REVIEW'
)
AND g.group_key IN ('GROUP_A', 'GROUP_B')
ON CONFLICT (id) DO UPDATE
SET
    can_execute = EXCLUDED.can_execute,
    is_active = EXCLUDED.is_active,
    priority = EXCLUDED.priority,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    updated_by = EXCLUDED.updated_by,
    updated_at = NOW();

COMMIT;
