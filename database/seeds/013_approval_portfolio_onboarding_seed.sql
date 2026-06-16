-- =============================================================================
-- Approval module — PORTFOLIO_ONBOARDING default process config (idempotent)
-- =============================================================================
-- Seeds the default two-stage PORTFOLIO_ONBOARDING approval process so that
-- any portfolio submitted for onboarding (status PENDING_APPROVAL) has a
-- reachable process config without requiring manual admin setup.
--
-- Process (COMPANY-global, contract_id NULL):
--   PORTFOLIO_ONBOARDING
--     stage 1  GROUP_ANY    FUND_MANAGER_REVIEWERS  (reviewer sign-off)
--     stage 2  GROUP_ANY    INVESTMENT_SUPERVISORS   (supervisor final approval)
--
-- Groups referenced here are seeded by 012_approval_demo_seed.sql — run
-- that seed before this one.  The seed is safe to re-run (ON CONFLICT DO NOTHING).
-- =============================================================================

BEGIN;

-- ── Process config ────────────────────────────────────────────────────────────
INSERT INTO approval__process_configs
    (id, process_code, process_name, process_type, contract_type, contract_id, effective_date,
     is_active, group_approval_enabled, require_team_approval, created_by)
VALUES
    ('a9200000-0000-0000-0000-000000000003',
     'PROC_PORTFOLIO_ONBOARDING_DEFAULT',
     'Portfolio Onboarding Approval (default)',
     'PORTFOLIO_ONBOARDING',
     'COMPANY',
     NULL,
     CURRENT_DATE - 1,
     true, true, false,
     'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (process_code) DO NOTHING;

-- ── Stages ────────────────────────────────────────────────────────────────────
INSERT INTO approval__process_stages
    (id, process_config_id, stage_number, stage_name,
     approver_mode, approval_group_id, required_approval_count, is_final_stage, reject_policy)
VALUES
    -- Stage 1: Fund manager reviewer validates the portfolio details
    ('a9300000-0000-0000-0000-000000000005',
     'a9200000-0000-0000-0000-000000000003',
     1, 'Reviewer sign-off',
     'GROUP_ANY', 'a9000000-0000-0000-0000-000000000001',
     1, false, 'STOP'),
    -- Stage 2: Investment supervisor gives final approval / rejection
    ('a9300000-0000-0000-0000-000000000006',
     'a9200000-0000-0000-0000-000000000003',
     2, 'Supervisor final approval',
     'GROUP_ANY', 'a9000000-0000-0000-0000-000000000002',
     1, true, 'STOP')
ON CONFLICT (process_config_id, stage_number) DO NOTHING;

COMMIT;
