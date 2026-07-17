-- =============================================================================
-- Ben investment operation-page access
-- =============================================================================
-- The operation directory and OP-01 decision form need read access to funds,
-- portfolios (including scoped holdings/cash), and instruments in addition to
-- Ben's existing decision operator permissions. Keep these prerequisites in a
-- dedicated assignment role so only explicitly assigned users receive the
-- additional visibility.
--
-- This migration does not add data scopes, transaction-posting permissions,
-- approval-stage permissions, or any maker-checker bypass.
-- =============================================================================

BEGIN;

-- Migrations run before development seeds on a fresh database. Ensure the
-- canonical investment permission definitions exist before inserting rights.
INSERT INTO permissions_function_definitions (code, module, name, description)
VALUES
    ('INVESTMENT_FUND_VIEW',       'investment', 'Investment Fund View',       'Read fund master data and fund-level AUM history.'),
    ('INVESTMENT_PORTFOLIO_VIEW',  'investment', 'Investment Portfolio View',  'Read portfolio master data, positions, cash, and valuations.'),
    ('INVESTMENT_INSTRUMENT_VIEW', 'investment', 'Investment Instrument View', 'Read instrument master data and provider mappings.')
ON CONFLICT (code) DO UPDATE
SET module        = EXCLUDED.module,
    name          = EXCLUDED.name,
    description   = EXCLUDED.description,
    deprecated_at = NULL;

-- INVESTMENT_VIEW is the legacy frontend entry-page gate. Preserve any
-- operator-maintained catalog metadata if it already exists.
INSERT INTO permissions_function_definitions (code, module, name, description)
VALUES (
    'INVESTMENT_VIEW',
    'legacy',
    'INVESTMENT_VIEW',
    'Legacy investment module entry-page access.'
)
ON CONFLICT (code) DO NOTHING;

-- This role is owned by this migration. Reject a same-name role with another
-- identifier rather than mutating an operator-managed authorization object.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM permissions_groups
        WHERE name = 'Investment Operation Page Access'
          AND id <> 'b0000000-0000-0000-0000-000000000042'::uuid
    ) THEN
        RAISE EXCEPTION 'Investment Operation Page Access permission group already exists with an unexpected id';
    END IF;
END
$$;

INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000042',
    'Investment Operation Page Access',
    'Read prerequisites for explicitly assigned users of the investment operation directory and OP-01 decision form.',
    true,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO NOTHING;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT
    'b0000000-0000-0000-0000-000000000042'::uuid,
    permission_code.code,
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('INVESTMENT_VIEW'),
    ('INVESTMENT_FUND_VIEW'),
    ('INVESTMENT_PORTFOLIO_VIEW'),
    ('INVESTMENT_INSTRUMENT_VIEW')
) AS permission_code(code)
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = true;

-- Ben is development/demo data and does not exist in a migration-only fresh
-- database. Assign him when present; seed 020 performs the same idempotent
-- assignment after fresh migrations.
INSERT INTO permissions_accounts_groups (id, user_id, group_id, assigned_by)
SELECT
    'b1000000-0000-0000-0000-000000000042',
    ben_user.id,
    'b0000000-0000-0000-0000-000000000042'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM iam_users ben_user
WHERE lower(ben_user.username) = 'ben'
  AND ben_user.is_active = true
  AND ben_user.deleted_at IS NULL
ON CONFLICT (user_id, group_id) DO NOTHING;

-- Fail closed if an upgraded database contains an active Ben account but the
-- effective role assignment or any required grant was not established.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM iam_users
        WHERE lower(username) = 'ben'
          AND is_active = true
          AND deleted_at IS NULL
    ) AND NOT EXISTS (
        SELECT 1
        FROM iam_users ben_user
        JOIN permissions_accounts_groups account_group
          ON account_group.user_id = ben_user.id
        JOIN permissions_groups permission_group
          ON permission_group.id = account_group.group_id
        WHERE lower(ben_user.username) = 'ben'
          AND permission_group.id = 'b0000000-0000-0000-0000-000000000042'::uuid
          AND permission_group.is_active = true
          AND permission_group.deleted_at IS NULL
          AND 4 = (
              SELECT COUNT(*)
              FROM permissions_function_rights function_right
              WHERE function_right.group_id = permission_group.id
                AND function_right.permission_code IN (
                    'INVESTMENT_VIEW',
                    'INVESTMENT_FUND_VIEW',
                    'INVESTMENT_PORTFOLIO_VIEW',
                    'INVESTMENT_INSTRUMENT_VIEW'
                )
                AND function_right.is_granted = true
          )
    ) THEN
        RAISE EXCEPTION 'failed to establish Ben investment operation-page permissions';
    END IF;
END
$$;

COMMIT;
