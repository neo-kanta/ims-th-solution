-- =============================================================================
-- Approval module — PORTFOLIO_ONBOARDING default process config (development/test only)
-- =============================================================================
-- Formerly database/seeds/013_approval_portfolio_onboarding_seed.sql. This
-- process config is explicitly deactivated below (Option A, deferred) and is
-- kept only as an inert schema-reference scaffold — it is never reached by a
-- live approval flow. It depends on the FUND_MANAGER_REVIEWERS and
-- INVESTMENT_SUPERVISORS approval groups, which are seeded by the always-run
-- database/seeds/012_approval_process_seed.sql (reference), so the FK to
-- approval__groups is satisfied in every environment. It is kept here as
-- demo/dev-only content, not because of an FK dependency on demo data, but
-- because a dead/deferred scaffold has no reason to exist in production and
-- this keeps the seed taxonomy simple: this file is only executed when
-- APP_ENV is development or test (see backend/cmd/seed/sql_seeds.go).
--
-- Seeds the default two-stage PORTFOLIO_ONBOARDING approval process so that
-- any portfolio submitted for onboarding (status PENDING_APPROVAL) has a
-- reachable process config without requiring manual admin setup, once this
-- feature is re-enabled.
--
-- Process (COMPANY-global, contract_id NULL):
--   PORTFOLIO_ONBOARDING
--     stage 1  GROUP_ANY    FUND_MANAGER_REVIEWERS  (reviewer sign-off)
--     stage 2  GROUP_ANY    INVESTMENT_SUPERVISORS   (supervisor final approval)
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

-- PORTFOLIO approval is deferred (Option A). The process config above is kept
-- for schema reference but is explicitly marked inactive so it is never reached
-- by a live approval flow. Re-enable once the PORTFOLIO subject access port and
-- validator are implemented (see docs/handoff/approval-subject-access-port.md).
UPDATE approval__process_configs
   SET is_active = false
 WHERE process_code = 'PROC_PORTFOLIO_ONBOARDING_DEFAULT';

COMMIT;
