-- =============================================================================
-- Approval module — PORTFOLIO_CASH_TRANSACTION process config seed (idempotent)
-- =============================================================================
-- Seeds the process configuration and single approver stage for
-- CASH_TRANSACTION subjects. Without this seed, any call to
-- SubmitForApproval("PORTFOLIO_CASH_TRANSACTION") fails to resolve a config.
--
-- Process (COMPANY-global, contract_id NULL):
--   PORTFOLIO_CASH_TRANSACTION
--     stage 1 GROUP_ANY  FUND_MANAGER_REVIEWERS (final)
--
-- Design choice (per spec §4.2): a single GROUP_ANY stage routed to the same
-- reviewers group (FUND_MANAGER_REVIEWERS = ben, green) that governs
-- INVESTMENT_DECISION, and COMPANY contract-type so it resolves for BOTH
-- fund-bound and fund-less (fund-optional) portfolios — a per-fund contract
-- config would never match a fund-less portfolio's submission.
--
-- The FUND_MANAGER_REVIEWERS group referenced here is defined by the
-- always-run reference seed 012_approval_process_seed.sql (its production
-- membership, if any, is provisioned separately by an administrator or a
-- dev/test-only demo seed — this file only wires the process config to the
-- group id, it does not grant membership). Safe to re-run
-- (ON CONFLICT DO NOTHING). This is a PRODUCTION reference seed: without it,
-- no LIVE cash-movement submission can resolve an approval config in any
-- environment, including production. `created_by` references the admin user
-- from 001_initial_seed.sql purely as the audit-trail actor, same as every
-- other reference seed in this directory.
-- =============================================================================

BEGIN;

-- ── Process config: PORTFOLIO_CASH_TRANSACTION ───────────────────────────────
INSERT INTO approval__process_configs
    (id, process_code, process_name, process_type, contract_type, contract_id, effective_date,
     is_active, group_approval_enabled, require_team_approval, created_by)
VALUES (
    'a9200000-0000-0000-0000-000000000005',
    'PROC_CASH_TRANSACTION_DEFAULT',
    'Portfolio Cash Transaction Approval (default)',
    'PORTFOLIO_CASH_TRANSACTION', 'COMPANY', NULL, CURRENT_DATE - 1,
    true, true, false,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (process_code) DO NOTHING;

-- ── Stage ─────────────────────────────────────────────────────────────────────
INSERT INTO approval__process_stages
    (id, process_config_id, stage_number, stage_name, approver_mode, approval_group_id, required_approval_count, is_final_stage, reject_policy)
VALUES (
    'a9300000-0000-0000-0000-000000000008',
    'a9200000-0000-0000-0000-000000000005',
    1, 'Cash movement sign-off', 'GROUP_ANY',
    'a9000000-0000-0000-0000-000000000001',  -- FUND_MANAGER_REVIEWERS (ben, green)
    1, true, 'STOP'
)
ON CONFLICT (process_config_id, stage_number) DO NOTHING;

COMMIT;
