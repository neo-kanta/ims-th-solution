-- =============================================================================
-- Green function-permission parity with Ben (development/test only, idempotent)
-- =============================================================================
-- Formerly database/seeds/021_green_ben_workflow_permission_parity_seed.sql.
-- Green keeps her existing GLOBAL-TECH and BBL-EQUITY data scopes and existing
-- approval reviewer/supervisor memberships. This seed assigns only the same
-- function-permission groups Ben uses for the requested workflows. This file
-- is only executed when APP_ENV is development or test (see
-- backend/cmd/seed/sql_seeds.go); it must never reach production.
-- =============================================================================

BEGIN;

-- These are function-only roles. Fail closed if an environment has attached
-- group-owned data scope, because membership must not expand Green beyond her
-- existing GLOBAL-TECH and BBL-EQUITY grants.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM permission_data_rights data_right
        WHERE data_right.subject_type = 'GROUP'
          AND data_right.subject_id IN (
              'b0000000-0000-0000-0000-000000000011'::uuid,
              'b0000000-0000-0000-0000-000000000040'::uuid,
              'b0000000-0000-0000-0000-000000000041'::uuid,
              'b0000000-0000-0000-0000-000000000042'::uuid
          )
    ) THEN
        RAISE EXCEPTION 'cannot seed Green function parity because a source group owns data permissions';
    END IF;
END
$$;

-- NOTE: the membership id is intentionally DB-generated (column default), NOT a
-- fixed value. Earlier this seed used the fixed ids b1000000-…051..054, which
-- are exactly the ids forward migration 20260725000001 deletes as
-- migration-owned artifacts — so a re-run of that migration on a seeded dev DB
-- would have deleted green's legitimately-seeded demo memberships. Using
-- generated ids (like the ben seeds) keeps demo data disjoint from the
-- migration-owned id space and keeps that migration's "never by application
-- code" invariant true.
INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
SELECT
    green_user.id,
    permission_group.id,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM iam_users green_user
CROSS JOIN (VALUES
    ('b0000000-0000-0000-0000-000000000011'::uuid, 'Fund Manager'),
    ('b0000000-0000-0000-0000-000000000040'::uuid, 'Investment Decision Operator'),
    ('b0000000-0000-0000-0000-000000000041'::uuid, 'Investment Decision Approver'),
    ('b0000000-0000-0000-0000-000000000042'::uuid, 'Investment Operation Page Access')
) AS expected(group_id, group_name)
JOIN permissions_groups permission_group
  ON permission_group.id = expected.group_id
 AND permission_group.name = expected.group_name
 AND permission_group.is_active = true
 AND permission_group.deleted_at IS NULL
WHERE green_user.id = 'a0000000-0000-0000-0000-000000000011'::uuid
  AND lower(green_user.username) = 'green'
  AND green_user.is_active = true
  AND green_user.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM permissions_accounts_groups ben_membership
      WHERE ben_membership.user_id = 'a0000000-0000-0000-0000-000000000010'::uuid
        AND ben_membership.group_id = permission_group.id
  )
ON CONFLICT (user_id, group_id) DO NOTHING;

UPDATE permissions_groups
SET description = 'Read prerequisites for explicitly assigned users of the investment operation directory and OP-01 decision form.',
    updated_by  = 'a0000000-0000-0000-0000-000000000001'::uuid,
    updated_at  = NOW()
WHERE id = 'b0000000-0000-0000-0000-000000000042'::uuid
  AND name = 'Investment Operation Page Access';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM iam_users
        WHERE id = 'a0000000-0000-0000-0000-000000000011'::uuid
          AND lower(username) = 'green'
          AND is_active = true
          AND deleted_at IS NULL
    ) AND 4 <> (
        SELECT COUNT(*)
        FROM permissions_accounts_groups
        WHERE user_id = 'a0000000-0000-0000-0000-000000000011'::uuid
          AND group_id IN (
              'b0000000-0000-0000-0000-000000000011'::uuid,
              'b0000000-0000-0000-0000-000000000040'::uuid,
              'b0000000-0000-0000-0000-000000000041'::uuid,
              'b0000000-0000-0000-0000-000000000042'::uuid
          )
    ) THEN
        RAISE EXCEPTION 'failed to seed Green function-permission parity with Ben';
    END IF;
END
$$;

COMMIT;
