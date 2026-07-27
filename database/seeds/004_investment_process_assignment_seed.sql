-- =============================================================================
-- Investment process step catalog seed (reference — always runs)
-- =============================================================================
-- Seeds the four investment workflow process-step definitions
-- (ANALYSIS_REPORT, INVESTMENT_DECISION, INVESTMENT_EXECUTION,
-- INVESTMENT_REVIEW). This is reference/catalog data the application's
-- process-assignment model relies on in every environment.
--
-- This file previously also created development users (ben/green/neo) and
-- assigned them to demo process groups (GROUP_A/GROUP_B) with step
-- assignments. That named-demo-identity content has moved to
-- database/seeds/demo/001_investment_process_assignment_demo_seed.sql, which
-- only runs in development/test (see backend/cmd/seed). Reference process
-- step definitions (this file) are unaffected and always run; who is actually
-- assigned to execute each step is an operator action in production.
-- =============================================================================

BEGIN;

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

COMMIT;
