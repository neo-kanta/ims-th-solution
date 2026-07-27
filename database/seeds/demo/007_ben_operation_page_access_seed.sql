-- =============================================================================
-- Ben investment operation-page access seed (development/test only, idempotent)
-- =============================================================================
-- Formerly database/seeds/020_ben_operation_page_access_seed.sql. Migration
-- 20260716000002 creates the dedicated least-privilege "Investment Operation
-- Page Access" role and grants its permissions (reference, always runs); this
-- seed ensures ben's role membership also exists after a fresh migrate + seed
-- or a later idempotent seed rerun. This file is only executed when APP_ENV
-- is development or test (see backend/cmd/seed/sql_seeds.go); it must never
-- reach production.
-- =============================================================================

BEGIN;

INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000042',
    'Investment Operation Page Access',
    'Read prerequisites for explicitly assigned users of the investment operation directory and OP-01 decision form.',
    true,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO UPDATE
SET description = EXCLUDED.description,
    is_active   = EXCLUDED.is_active,
    updated_by  = EXCLUDED.updated_by,
    updated_at  = NOW();

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT permission_group.id, permission_code.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups permission_group
CROSS JOIN (VALUES
    ('INVESTMENT_VIEW'),
    ('INVESTMENT_FUND_VIEW'),
    ('INVESTMENT_PORTFOLIO_VIEW'),
    ('INVESTMENT_INSTRUMENT_VIEW')
) AS permission_code(code)
WHERE permission_group.name = 'Investment Operation Page Access'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
SELECT
    'a0000000-0000-0000-0000-000000000010'::uuid,
    permission_group.id,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups permission_group
WHERE permission_group.name = 'Investment Operation Page Access'
  AND EXISTS (
      SELECT 1
      FROM iam_users ben_user
      WHERE ben_user.id = 'a0000000-0000-0000-0000-000000000010'::uuid
        AND lower(ben_user.username) = 'ben'
        AND ben_user.is_active = true
        AND ben_user.deleted_at IS NULL
  )
ON CONFLICT (user_id, group_id) DO NOTHING;

COMMIT;
