-- =============================================================================
-- Approval module — process wiring seed (reference — always runs)
-- =============================================================================
-- Seeds the approval groups (role shells), the two default process
-- configurations, and their stages so INVESTMENT_ANALYSIS_REPORT and
-- INVESTMENT_DECISION submissions have a reachable approval process in every
-- environment, including production. The admin bootstrap account is also
-- registered as an INVESTMENT_SUPERVISORS member so production has at least
-- one eligible final-stage approver out of the box; assigning real
-- reviewers/supervisors to these groups afterward is an operator action
-- performed through the application.
--
-- Named demo users (ben, green) were previously assigned as members of these
-- groups directly in this file, along with a demo delegation from ben to
-- green. That named-demo-identity content has moved to
-- database/seeds/demo/002_approval_demo_seed.sql, which only runs in
-- development/test (see backend/cmd/seed). Without it, in production these
-- groups exist with only the admin supervisor membership until an
-- administrator assigns real reviewers/supervisors.
--
-- Groups:
--   FUND_MANAGER_REVIEWERS : (no reference members; assign real reviewers)
--   INVESTMENT_SUPERVISORS : admin (priority 1)
--   TRADING_SUPERVISORS    : (no reference members; assign real supervisors)
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
    ('a9000000-0000-0000-0000-000000000001', 'FUND_MANAGER_REVIEWERS', 'Fund Manager Reviewers', 'Reviewer group for INVESTMENT_ANALYSIS_REPORT / INVESTMENT_DECISION stage 1; assign real reviewers via the application.', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9000000-0000-0000-0000-000000000002', 'INVESTMENT_SUPERVISORS', 'Investment Supervisors', 'Supervisor group for INVESTMENT_ANALYSIS_REPORT / INVESTMENT_DECISION final stage.', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9000000-0000-0000-0000-000000000003', 'TRADING_SUPERVISORS', 'Trading Supervisors', 'Supervisor group reserved for trading approval processes; assign real supervisors via the application.', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_code) DO NOTHING;

-- ── Group members (APPROVED + active so they are eligible approvers) ──────────
-- Only the admin bootstrap account is registered here so production has a
-- working final-stage approver. Real reviewer/supervisor membership for
-- FUND_MANAGER_REVIEWERS and TRADING_SUPERVISORS is an operator action.
INSERT INTO approval__group_members (id, group_id, user_id, priority_order, member_type, status, is_active, created_by)
VALUES
    ('a9100000-0000-0000-0000-000000000003', 'a9000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001', 1, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001')
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
