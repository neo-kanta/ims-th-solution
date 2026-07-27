-- =============================================================================
-- Approval module — DEV/DEMO seed (development/test only)
-- =============================================================================
-- Extracted from the former database/seeds/012_approval_demo_seed.sql.
-- Assigns the named demo users (ben, green) as members of the approval groups
-- seeded by the always-run database/seeds/012_approval_process_seed.sql
-- (formerly 012_approval_demo_seed.sql, now reference-only), and seeds one demo
-- delegation, so the approval workflow is usable out of the box in
-- development. This file is only executed when APP_ENV is development or test
-- (see backend/cmd/seed/sql_seeds.go); it must never reach production.
--
-- Groups (defined by the reference seed; membership added here):
--   FUND_MANAGER_REVIEWERS : ben (priority 1), green (priority 2)
--   INVESTMENT_SUPERVISORS : green (priority 2) — admin (priority 1) is
--                            already a reference member
--   TRADING_SUPERVISORS    : green (priority 1)
--
-- Delegation:
--   ben → green, wildcard (all contracts), 30 days from seed date.
--   Demonstrates the delegate proxy path: green can approve tasks assigned to
--   ben, producing a DELEGATED signature that records both actors.
-- =============================================================================

BEGIN;

-- ── Group members (APPROVED + active so they are eligible approvers) ──────────
INSERT INTO approval__group_members (id, group_id, user_id, priority_order, member_type, status, is_active, created_by)
VALUES
    -- FUND_MANAGER_REVIEWERS
    ('a9100000-0000-0000-0000-000000000001', 'a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000010', 1, 'MEMBER',     'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    ('a9100000-0000-0000-0000-000000000002', 'a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000011', 2, 'MEMBER',     'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    -- INVESTMENT_SUPERVISORS (admin priority 1 is a reference member seeded elsewhere)
    ('a9100000-0000-0000-0000-000000000004', 'a9000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000011', 2, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001'),
    -- TRADING_SUPERVISORS
    ('a9100000-0000-0000-0000-000000000005', 'a9000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000011', 1, 'SUPERVISOR', 'APPROVED', true, 'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (group_id, user_id) DO NOTHING;

-- ── Delegation case ──────────────────────────────────────────────────────────
-- Demo: ben delegates to green for 30 days from seed date.
-- This lets green act as a proxy approver on any task assigned to ben,
-- producing a DELEGATED signature that records both actors.
INSERT INTO approval__delegations
    (id, from_user_id, to_user_id, contract_id, active_from, active_until, is_active, remarks, created_by)
VALUES
    ('a9400000-0000-0000-0000-000000000001',
     'a0000000-0000-0000-0000-000000000010',  -- ben (from)
     'a0000000-0000-0000-0000-000000000011',  -- green (to / delegate)
     NULL,                                    -- wildcard: applies to all contracts
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP + INTERVAL '30 days',
     true,
     'Demo delegation: ben is on leave — green acts as proxy for all approvals.',
     'a0000000-0000-0000-0000-000000000001')
ON CONFLICT (id) DO NOTHING;

COMMIT;
