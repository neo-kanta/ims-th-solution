-- =============================================================================
-- Investment Decision Operator — Ben membership (development/test only)
-- =============================================================================
-- Extracted from the former database/seeds/018_investment_decision_execution_permission_seed.sql.
-- Assigns the named demo user "ben" to the "Investment Decision Operator"
-- group (created by the always-run, reference
-- database/seeds/018_investment_decision_execution_permission_seed.sql) so
-- ben can exercise OP-01/OP-02/OP-03 in development. This file is only
-- executed when APP_ENV is development or test (see
-- backend/cmd/seed/sql_seeds.go); it must never reach production.
--
-- Depends on: demo/001 (ben user), 018 (Investment Decision Operator group).
-- =============================================================================

BEGIN;

INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
SELECT 'a0000000-0000-0000-0000-000000000010'::uuid, g.id, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
WHERE g.name = 'Investment Decision Operator'
ON CONFLICT (user_id, group_id) DO NOTHING;

COMMIT;
