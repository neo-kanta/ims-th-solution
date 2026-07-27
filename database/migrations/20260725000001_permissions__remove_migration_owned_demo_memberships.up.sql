-- =============================================================================
-- Forward cleanup: remove migration-owned demo-user group memberships
-- =============================================================================
-- Earlier versions of migrations 20260716000001, 20260716000002, and
-- 20260716000003 conditionally assigned the named demo users "ben" and
-- "green" to production role catalogs (Investment Decision Approver,
-- Investment Operation Page Access, Fund Manager, Investment Decision
-- Operator) whenever those users already existed in the target database at
-- migration-apply time. Those migrations have since been corrected to never
-- perform named-identity assignment (see their current up.sql headers), but
-- a database that was migrated BEFORE that correction may already contain the
-- six membership rows those conditional INSERTs created.
--
-- This migration removes exactly those six migration-owned rows, identified
-- by their fixed primary-key ids (assigned only by the migrations above, never
-- by application code or an administrator), and nothing else:
--
--   id                                     user   group
--   b1000000-0000-0000-0000-000000000041   ben    Investment Decision Approver (b0000000-...-041)
--   b1000000-0000-0000-0000-000000000042   ben    Investment Operation Page Access (b0000000-...-042)
--   b1000000-0000-0000-0000-000000000051   green  Fund Manager (b0000000-...-011)
--   b1000000-0000-0000-0000-000000000052   green  Investment Decision Operator (b0000000-...-040)
--   b1000000-0000-0000-0000-000000000053   green  Investment Decision Approver (b0000000-...-041)
--   b1000000-0000-0000-0000-000000000054   green  Investment Operation Page Access (b0000000-...-042)
--
-- `permissions_accounts_groups.id` is the table's primary key, so matching on
-- id alone is already exact; the user_id/group_id columns are included in the
-- WHERE clause purely as a documented double-check, not because a collision is
-- possible. Any membership an administrator or the application created for
-- ben/green with a DIFFERENT id (i.e. a deliberate, real assignment made after
-- deployment) is untouched. Seed/admin-bootstrap memberships (e.g. the admin
-- user's own group memberships) use unrelated ids and are never touched by
-- this statement.
--
-- On a fresh database that only ever ran the corrected migrations (or that
-- never had ben/green accounts), none of these six ids exist and this
-- statement deletes zero rows — safe no-op.
-- =============================================================================

BEGIN;

DELETE FROM permissions_accounts_groups account_group
USING (VALUES
    ('b1000000-0000-0000-0000-000000000041'::uuid, 'a0000000-0000-0000-0000-000000000010'::uuid, 'b0000000-0000-0000-0000-000000000041'::uuid),
    ('b1000000-0000-0000-0000-000000000042'::uuid, 'a0000000-0000-0000-0000-000000000010'::uuid, 'b0000000-0000-0000-0000-000000000042'::uuid),
    ('b1000000-0000-0000-0000-000000000051'::uuid, 'a0000000-0000-0000-0000-000000000011'::uuid, 'b0000000-0000-0000-0000-000000000011'::uuid),
    ('b1000000-0000-0000-0000-000000000052'::uuid, 'a0000000-0000-0000-0000-000000000011'::uuid, 'b0000000-0000-0000-0000-000000000040'::uuid),
    ('b1000000-0000-0000-0000-000000000053'::uuid, 'a0000000-0000-0000-0000-000000000011'::uuid, 'b0000000-0000-0000-0000-000000000041'::uuid),
    ('b1000000-0000-0000-0000-000000000054'::uuid, 'a0000000-0000-0000-0000-000000000011'::uuid, 'b0000000-0000-0000-0000-000000000042'::uuid)
) AS owned(membership_id, expected_user_id, expected_group_id)
WHERE account_group.id = owned.membership_id
  AND account_group.user_id = owned.expected_user_id
  AND account_group.group_id = owned.expected_group_id;

COMMIT;
