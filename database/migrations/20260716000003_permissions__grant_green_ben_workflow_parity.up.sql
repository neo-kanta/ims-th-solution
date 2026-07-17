-- =============================================================================
-- Green function-permission parity with Ben
-- =============================================================================
-- Green is already an APPROVED member of the seeded approval reviewer and
-- supervisor groups. Assign the same function-permission groups Ben uses for
-- workflow state, OP-01/OP-02/OP-03, approval runtime, and Operation Page
-- reads. Data permissions remain Green's own scoped contracts.
--
-- No function rights are duplicated, no wildcard data access is introduced,
-- and approval process stages or maker-checker rules are not changed.
-- =============================================================================

BEGIN;

-- These identifiers are owned by this migration. Fail closed rather than
-- reusing a membership identifier that belongs to a different relationship.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM permissions_accounts_groups account_group
        JOIN (VALUES
            ('b1000000-0000-0000-0000-000000000051'::uuid, 'b0000000-0000-0000-0000-000000000011'::uuid),
            ('b1000000-0000-0000-0000-000000000052'::uuid, 'b0000000-0000-0000-0000-000000000040'::uuid),
            ('b1000000-0000-0000-0000-000000000053'::uuid, 'b0000000-0000-0000-0000-000000000041'::uuid),
            ('b1000000-0000-0000-0000-000000000054'::uuid, 'b0000000-0000-0000-0000-000000000042'::uuid)
        ) AS expected(membership_id, group_id)
          ON expected.membership_id = account_group.id
        WHERE account_group.user_id <> 'a0000000-0000-0000-0000-000000000011'::uuid
           OR account_group.group_id <> expected.group_id
    ) THEN
        RAISE EXCEPTION 'Green permission-parity membership id is already owned by another relationship';
    END IF;
END
$$;

-- Group membership can also contribute data permissions through the newer
-- subject-based permission model. These four groups are intended to be
-- function-only roles; refuse the copy if an administrator attached any
-- group-owned fund, contract, or portfolio scope in an upgraded database.
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
        RAISE EXCEPTION 'cannot establish Green function parity because a source group owns data permissions';
    END IF;
END
$$;

-- On an upgraded development database, require the exact Ben source account
-- and all four expected active groups before copying memberships. A fresh
-- migration-only database has no demo users yet and is completed by seed 021.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM iam_users
        WHERE id = 'a0000000-0000-0000-0000-000000000011'::uuid
          AND lower(username) = 'green'
          AND is_active = true
          AND deleted_at IS NULL
    ) AND (
        4 <> (
            SELECT COUNT(*)
            FROM permissions_groups permission_group
            JOIN (VALUES
                ('b0000000-0000-0000-0000-000000000011'::uuid, 'Fund Manager'),
                ('b0000000-0000-0000-0000-000000000040'::uuid, 'Investment Decision Operator'),
                ('b0000000-0000-0000-0000-000000000041'::uuid, 'Investment Decision Approver'),
                ('b0000000-0000-0000-0000-000000000042'::uuid, 'Investment Operation Page Access')
            ) AS expected(group_id, group_name)
              ON expected.group_id = permission_group.id
             AND expected.group_name = permission_group.name
            WHERE permission_group.is_active = true
              AND permission_group.deleted_at IS NULL
        )
        OR 4 <> (
            SELECT COUNT(*)
            FROM iam_users ben_user
            JOIN permissions_accounts_groups account_group
              ON account_group.user_id = ben_user.id
            WHERE ben_user.id = 'a0000000-0000-0000-0000-000000000010'::uuid
              AND lower(ben_user.username) = 'ben'
              AND ben_user.is_active = true
              AND ben_user.deleted_at IS NULL
              AND account_group.group_id IN (
                  'b0000000-0000-0000-0000-000000000011'::uuid,
                  'b0000000-0000-0000-0000-000000000040'::uuid,
                  'b0000000-0000-0000-0000-000000000041'::uuid,
                  'b0000000-0000-0000-0000-000000000042'::uuid
              )
        )
    ) THEN
        RAISE EXCEPTION 'cannot establish Green permission parity because Ben or an expected group is missing';
    END IF;
END
$$;

INSERT INTO permissions_accounts_groups (id, user_id, group_id, assigned_by)
SELECT
    expected.membership_id,
    green_user.id,
    permission_group.id,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM iam_users green_user
CROSS JOIN (VALUES
    ('b1000000-0000-0000-0000-000000000051'::uuid, 'b0000000-0000-0000-0000-000000000011'::uuid, 'Fund Manager'),
    ('b1000000-0000-0000-0000-000000000052'::uuid, 'b0000000-0000-0000-0000-000000000040'::uuid, 'Investment Decision Operator'),
    ('b1000000-0000-0000-0000-000000000053'::uuid, 'b0000000-0000-0000-0000-000000000041'::uuid, 'Investment Decision Approver'),
    ('b1000000-0000-0000-0000-000000000054'::uuid, 'b0000000-0000-0000-0000-000000000042'::uuid, 'Investment Operation Page Access')
) AS expected(membership_id, group_id, group_name)
JOIN permissions_groups permission_group
  ON permission_group.id = expected.group_id
 AND permission_group.name = expected.group_name
 AND permission_group.is_active = true
 AND permission_group.deleted_at IS NULL
WHERE green_user.id = 'a0000000-0000-0000-0000-000000000011'::uuid
  AND lower(green_user.username) = 'green'
  AND green_user.is_active = true
  AND green_user.deleted_at IS NULL
ON CONFLICT (user_id, group_id) DO NOTHING;

-- Replace the initial single-user wording with neutral role metadata. Rollback
-- retains this harmless descriptive correction.
UPDATE permissions_groups
SET description = 'Read prerequisites for explicitly assigned users of the investment operation directory and OP-01 decision form.',
    updated_by  = 'a0000000-0000-0000-0000-000000000001'::uuid,
    updated_at  = NOW()
WHERE id = 'b0000000-0000-0000-0000-000000000042'::uuid
  AND name = 'Investment Operation Page Access';

-- Fail closed if an upgraded database contains active Green but any intended
-- group membership is absent.
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
        FROM permissions_accounts_groups account_group
        JOIN permissions_groups permission_group
          ON permission_group.id = account_group.group_id
        WHERE account_group.user_id = 'a0000000-0000-0000-0000-000000000011'::uuid
          AND permission_group.id IN (
              'b0000000-0000-0000-0000-000000000011'::uuid,
              'b0000000-0000-0000-0000-000000000040'::uuid,
              'b0000000-0000-0000-0000-000000000041'::uuid,
              'b0000000-0000-0000-0000-000000000042'::uuid
          )
          AND permission_group.is_active = true
          AND permission_group.deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'failed to establish Green function-permission parity with Ben';
    END IF;
END
$$;

COMMIT;
