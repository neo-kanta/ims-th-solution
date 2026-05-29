-- =============================================================================
-- Approval module — DEV/DEMO seed (idempotent)
-- =============================================================================
-- Seeds approval groups, members and two process configurations so the approval
-- workflow is usable out of the box in development. DEV ONLY — references the
-- dev users created by 004_investment_process_assignment_seed.sql
-- (admin / ben / green) and is safe to re-run.
--
-- Groups:
--   FUND_MANAGER_REVIEWERS : ben (priority 1), green (priority 2)
--   INVESTMENT_SUPERVISORS : admin (priority 1), green (priority 2)
--   TRADING_SUPERVISORS    : green (priority 1)
--
-- Processes (COMPANY-global, contract_id NULL → no per-fund data-scope needed):
--   INVESTMENT_ANALYSIS_REPORT
--     stage 1 GROUP_PRIORITY  FUND_MANAGER_REVIEWERS
--     stage 2 GROUP_ANY       INVESTMENT_SUPERVISORS (final)
--   INVESTMENT_DECISION
--     stage 1 GROUP_ANY       FUND_MANAGER_REVIEWERS
--     stage 2 GROUP_ANY       INVESTMENT_SUPERVISORS (final)
-- =============================================================================

BEGIN;

-- ── Groups ───────────────────────────────────────────────────────────────────
INSERT INTO approval__groups (id, group_code, group_name, remarks, is_active, created_by)
VALUES
    ('a9000000-0000-0000-0000-000000000001', 'FUND_MANAGER_REVIEWERS', 'Fund Manager Reviewers', 'Dev seed group', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9000000-0000-0000-0000-000000000002', 'INVESTMENT_SUPERVISORS', 'Investment Supervisors', 'Dev seed group', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9000000-0000-0000-0000-000000000003', 'TRADING_SUPERVISORS', 'Trading Supervisors', 'Dev seed group', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_code) DO NOTHING;

-- ── Group members (APPROVED + active so they are eligible approvers) ──────────
INSERT INTO approval__group_members (id, group_id, user_id, priority_order, member_type, status, is_active, created_by)
VALUES
    -- FUND_MANAGER_REVIEWERS
    ('a9100000-0000-0000-0000-000000000001', 'a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000010', 1, 'MEMBER',     'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9100000-0000-0000-0000-000000000002', 'a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000011', 2, 'MEMBER',     'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    -- INVESTMENT_SUPERVISORS
    ('a9100000-0000-0000-0000-000000000003', 'a9000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 1, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9100000-0000-0000-0000-000000000004', 'a9000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000011', 2, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    -- TRADING_SUPERVISORS
    ('a9100000-0000-0000-0000-000000000005', 'a9000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000011', 1, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_id, user_id) DO NOTHING;

-- ── Process config: INVESTMENT_ANALYSIS_REPORT ───────────────────────────────
INSERT INTO approval__process_configs
    (id, process_code, process_name, process_type, contract_type, contract_id, effective_date,
     is_active, group_approval_enabled, require_team_approval, created_by)
VALUES
    ('a9200000-0000-0000-0000-000000000001', 'PROC_ANALYSIS_DEFAULT', 'Analysis Report Approval (default)',
     'INVESTMENT_ANALYSIS_REPORT', 'COMPANY', NULL, CURRENT_DATE - 1,
     true, true, false, 'a0000000-0000-0000-0000-000000000001'),
    ('a9200000-0000-0000-0000-000000000002', 'PROC_DECISION_DEFAULT', 'Investment Decision Approval (default)',
     'INVESTMENT_DECISION', 'COMPANY', NULL, CURRENT_DATE - 1,
     true, true, false, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (process_code) DO NOTHING;

-- ── Stages ───────────────────────────────────────────────────────────────────
INSERT INTO approval__process_stages
    (id, process_config_id, stage_number, stage_name, approver_mode, approval_group_id, required_approval_count, is_final_stage, reject_policy)
VALUES
    -- Analysis report
    ('a9300000-0000-0000-0000-000000000001', 'a9200000-0000-0000-0000-000000000001', 1, 'Reviewer sign-off',  'GROUP_PRIORITY', 'a9000000-0000-0000-0000-000000000001', 1, false, 'STOP'),
    ('a9300000-0000-0000-0000-000000000002', 'a9200000-0000-0000-0000-000000000001', 2, 'Supervisor sign-off','GROUP_ANY',      'a9000000-0000-0000-0000-000000000002', 1, true,  'STOP'),
    -- Investment decision
    ('a9300000-0000-0000-0000-000000000003', 'a9200000-0000-0000-0000-000000000002', 1, 'Reviewer sign-off',  'GROUP_ANY', 'a9000000-0000-0000-0000-000000000001', 1, false, 'STOP'),
    ('a9300000-0000-0000-0000-000000000004', 'a9200000-0000-0000-0000-000000000002', 2, 'Supervisor sign-off','GROUP_ANY', 'a9000000-0000-0000-0000-000000000002', 1, true,  'STOP')
ON CONFLICT (process_config_id, stage_number) DO NOTHING;

COMMIT;
