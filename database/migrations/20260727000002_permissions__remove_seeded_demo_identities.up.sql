-- =============================================================================
-- Forward cleanup: remove seed-created demo-identity access (ben/green/neo)
-- =============================================================================
-- IMS-MERGE-BLOCKERS G1 residual-access finding (owner decision D5,
-- 2026-07-27). Migration 20260725000001 removed the six migration-owned
-- ben/green group memberships (fixed ids b1000000-...041/042/051/052/053/054),
-- created by an earlier, now-corrected version of migrations 20260716000001/
-- 2/3. That migration does NOT remove a SEPARATE, wider set of grants that the
-- legacy always-run seed files (former 004/012/013/014/018/019/020/021/005,
-- now database/seeds/demo/001-009) created every time `cmd/seed` ran, before
-- this program's demo/production seed split existed. Those seed-created rows
-- carry DB-generated ids (not the fixed b1000000-... ids), so the earlier
-- migration's exact-id match cannot see them. A production database that ever
-- ran `cmd/seed` before the split may still hold every one of these rows.
--
-- This migration removes exactly the demo-identity access those legacy seeds
-- created, and nothing else:
--
--   1. investment__process_groups GROUP_A / GROUP_B (ids 88000000-...101/102,
--      fixed by the seed) — CASCADE-deletes their
--      investment__process_group_members (ben/green/neo) and
--      investment__process_step_assignments rows. These two groups exist only
--      to demonstrate process-step execution for ben/green/neo; their own
--      description text names those users explicitly, so no legitimate
--      production admin could have created them independently.
--   2. approval__group_members rows for ben/green in FUND_MANAGER_REVIEWERS /
--      INVESTMENT_SUPERVISORS / TRADING_SUPERVISORS, matched by the exact
--      UNIQUE(group_id, user_id) tuple (constraint uq_approval_group_members)
--      on the seed's own fixed group ids and the demo users' fixed ids — not
--      by the row's own generated id, so a row that a real re-run of the seed
--      (or any process) created under a different id for the same
--      (group, user) pair is still caught. The approval GROUPS themselves
--      (a9000000-...0001/0002/0003) are reference catalog data seeded by the
--      always-run 012_approval_process_seed.sql and may carry real production
--      members; they are never touched.
--   3. approval__delegations row a9400000-...0001 (ben -> green wildcard demo
--      delegation).
--   4. permissions_accounts_groups rows for ben/green in the four named
--      catalog groups (Fund Manager b0000000-...011, Investment Decision
--      Operator b0000000-...040, Investment Decision Approver
--      b0000000-...041, Investment Operation Page Access b0000000-...042),
--      matched by the table's UNIQUE(user_id, group_id) constraint on the
--      demo users' FIXED ids — exact, not a guess. The catalog GROUPS
--      themselves are production role catalog and are never touched; only
--      ben/green's membership in them is removed.
--   5. workflow__approval_settings row registering ben as a MANAGER_APPROVE
--      approver.
--   6. permissions_data_rights rows granting ben/green/neo per-fund data-scope
--      access (database/seeds/zz_demo/06_demo_data_permissions.sql: ben ->
--      SCB-FIXED manage + TH-GOV-LTF watch; green -> GLOBAL-TECH manage +
--      BBL-EQUITY watch; neo -> KTB-BALANCED + GLOBAL-TECH watch), matched by
--      the seed's own six fixed ids (d000e000-...0010/0011/0020/0021/0030/
--      0031). This table is live production code — both
--      permissions/infrastructure/persistence's effectiveContractScopes and
--      iam's login-time GetUserDataPermissions read it — so leaving it
--      untouched would have left a real, exploitable data-scope grant intact
--      the instant anyone reactivated one of these accounts. The admin
--      wildcard row (d000e000-...0001, contract_id='*') is a DIFFERENT id and
--      is never touched.
--   7. Finally, the three demo iam_users rows (ben/green/neo, ids
--      a0000000-...010/011/012) are DEACTIVATED (is_active = false), not
--      hard-deleted. A hard DELETE risks failing against FK-constrained
--      history a real environment may have accumulated for these accounts
--      (sessions, audit events, approval requests they submitted, etc.);
--      deactivation immediately and unconditionally revokes their ability to
--      authenticate or act, which is the actual security requirement here,
--      without any cascade risk. Their historical audit trail (who did what,
--      when) is deliberately preserved, matching this program's "preserve an
--      immutable, attributable audit trail" rule (docs/MANAGER/MEMORY.md).
--
-- Every step targets rows by the seed's own fixed ids (steps 1, 3, 5, 6) or by
-- an exact UNIQUE-constrained tuple on the demo users'/groups' fixed ids
-- (steps 2, 4) — so this is FK-safe (children deleted via CASCADE or before
-- parents) and a true no-op on any database that never ran the legacy demo
-- seeds.
--
-- Independent review (2026-07-27) of an earlier version of this file found
-- one P0 (permissions_data_rights was missing entirely — now step 6 above)
-- and one P1 (step 2 matched by row id only, inconsistent with step 4's
-- safer tuple match — now fixed to match the same pattern). Both are
-- corrected in this version.
-- =============================================================================

BEGIN;

-- 1. Demo process groups (CASCADEs to members + step assignments).
DELETE FROM investment__process_groups
WHERE id IN (
    '88000000-0000-0000-0000-000000000101', -- GROUP_A (Ben + Green)
    '88000000-0000-0000-0000-000000000102'  -- GROUP_B (Neo)
);

-- 2. Approval group memberships seeded for ben/green, matched by the exact
--    UNIQUE(group_id, user_id) tuple (consistent with step 4's safer pattern).
DELETE FROM approval__group_members
WHERE (group_id, user_id) IN (
    ('a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000010'), -- ben  -> FUND_MANAGER_REVIEWERS
    ('a9000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000011'), -- green-> FUND_MANAGER_REVIEWERS
    ('a9000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000011'), -- green-> INVESTMENT_SUPERVISORS
    ('a9000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000011')  -- green-> TRADING_SUPERVISORS
);

-- 3. Ben -> Green demo delegation.
DELETE FROM approval__delegations
WHERE id = 'a9400000-0000-0000-0000-000000000001';

-- 4. Function-permission group memberships for ben/green (exact tuple match
--    on UNIQUE(user_id, group_id); demo user ids are fixed, catalog group ids
--    are fixed production role-catalog ids — never delete the groups).
DELETE FROM permissions_accounts_groups
WHERE (user_id, group_id) IN (
    ('a0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000011'), -- ben   -> Fund Manager
    ('a0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000040'), -- ben   -> Investment Decision Operator
    ('a0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000041'), -- ben   -> Investment Decision Approver
    ('a0000000-0000-0000-0000-000000000010', 'b0000000-0000-0000-0000-000000000042'), -- ben   -> Investment Operation Page Access
    ('a0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000011'), -- green -> Fund Manager
    ('a0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000040'), -- green -> Investment Decision Operator
    ('a0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000041'), -- green -> Investment Decision Approver
    ('a0000000-0000-0000-0000-000000000011', 'b0000000-0000-0000-0000-000000000042')  -- green -> Investment Operation Page Access
);

-- 5. Ben's workflow MANAGER_APPROVE approver registration.
DELETE FROM workflow__approval_settings
WHERE operation_type = 'MANAGER_APPROVE'
  AND approver_account_code = 'ben';

-- 6. Per-fund data-scope grants for ben/green/neo (fixed seed ids). The admin
--    wildcard row (d000e000-...0001) is a different id and is untouched.
DELETE FROM permissions_data_rights
WHERE id IN (
    'd000e000-0000-0000-0000-000000000010', -- ben   -> SCB-FIXED (manage)
    'd000e000-0000-0000-0000-000000000011', -- ben   -> TH-GOV-LTF (watch)
    'd000e000-0000-0000-0000-000000000020', -- green -> GLOBAL-TECH (manage)
    'd000e000-0000-0000-0000-000000000021', -- green -> BBL-EQUITY (watch)
    'd000e000-0000-0000-0000-000000000030', -- neo   -> KTB-BALANCED (watch)
    'd000e000-0000-0000-0000-000000000031'  -- neo   -> GLOBAL-TECH (watch)
);

-- 7. Deactivate the demo identities themselves (not a hard delete — see
--    header). Idempotent: a second run only re-affirms is_active = false.
UPDATE iam_users
   SET is_active = false,
       updated_by = 'a0000000-0000-0000-0000-000000000001', -- admin
       updated_at = NOW()
 WHERE id IN (
    'a0000000-0000-0000-0000-000000000010', -- ben
    'a0000000-0000-0000-0000-000000000011', -- green
    'a0000000-0000-0000-0000-000000000012'  -- neo
 )
   AND is_active = true;

COMMIT;
