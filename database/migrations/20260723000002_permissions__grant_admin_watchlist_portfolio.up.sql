-- =============================================================================
-- Admin / admin2 — portfolio watchlist access
-- =============================================================================
-- admin and admin2 are both members of the Admin group (see migrations
-- 20260301000005 and 20260722000002) but the watchlist module's function
-- permissions were never granted to that group, so the Portfolio Watchlists
-- tab and its actions were forbidden for both accounts.
--
-- Portfolio-scoped watchlist reads/writes additionally require IAM data-scope
-- (fund access) via HasDataPermission — see
-- backend/internal/watchlist/application/{command/create_item.go,query/list_items.go}.
-- admin already holds the "*" data-scope wildcard from the zz_demo seed;
-- admin2 does not, so it is granted directly here to keep the two accounts at
-- parity, consistent with migration 20260722000002's stated intent to give
-- admin2 the same group-derived, full application-administrator authority as
-- admin.
--
-- This migration does not touch WATCHLIST_EVALUATE (operational-only; normal
-- users, including admins, are not meant to receive it per
-- backend/internal/watchlist/permission/policies.go) or WATCHLIST_ADMIN
-- (a separate, explicit override left for a future dedicated decision).
-- =============================================================================

BEGIN;

-- Migrations run before development seeds on a fresh database. Ensure the
-- canonical watchlist permission definitions exist before inserting rights.
INSERT INTO permissions_function_definitions (code, module, name, description)
VALUES
    ('WATCHLIST_VIEW', 'watchlist', 'Watchlist View', 'List visible watchlist items and alert events.'),
    ('WATCHLIST_MANAGE', 'watchlist', 'Watchlist Manage', 'Create, update, disable, and soft-delete watchlist items and threshold rules.'),
    ('WATCHLIST_ALERT_ACK', 'watchlist', 'Watchlist Alert Acknowledge', 'Acknowledge visible alert events.')
ON CONFLICT (code) DO UPDATE
SET module        = EXCLUDED.module,
    name          = EXCLUDED.name,
    description   = EXCLUDED.description,
    deprecated_at = NULL;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT
    'b0000000-0000-0000-0000-000000000001'::uuid,
    permission_code.code,
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('WATCHLIST_VIEW'),
    ('WATCHLIST_MANAGE'),
    ('WATCHLIST_ALERT_ACK')
) AS permission_code(code)
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = true;

-- Give admin2 the same fund/portfolio data-scope wildcard admin already has
-- (from the zz_demo seed) so portfolio-scoped watchlist calls are not blocked
-- by HasDataPermission.
INSERT INTO permissions_data_rights (user_id, contract_id, is_granted, granted_by)
VALUES (
    'a0000000-0000-0000-0000-000000000002'::uuid,
    '*',
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid
)
ON CONFLICT (user_id, contract_id) DO UPDATE
SET is_granted = true,
    granted_at = NOW(),
    granted_by = EXCLUDED.granted_by;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM iam_users admin_user
        JOIN permissions_accounts_groups account_group
          ON account_group.user_id = admin_user.id
        JOIN permissions_groups permission_group
          ON permission_group.id = account_group.group_id
         AND permission_group.id = 'b0000000-0000-0000-0000-000000000001'::uuid
         AND permission_group.is_active = true
         AND permission_group.deleted_at IS NULL
        WHERE admin_user.id IN (
            'a0000000-0000-0000-0000-000000000001'::uuid,
            'a0000000-0000-0000-0000-000000000002'::uuid
        )
        HAVING COUNT(DISTINCT admin_user.id) = 2
    ) THEN
        RAISE EXCEPTION 'admin and admin2 are not both active members of the Admin group';
    END IF;

    IF 3 <> (
        SELECT COUNT(*)
        FROM permissions_function_rights function_right
        WHERE function_right.group_id = 'b0000000-0000-0000-0000-000000000001'::uuid
          AND function_right.permission_code IN ('WATCHLIST_VIEW', 'WATCHLIST_MANAGE', 'WATCHLIST_ALERT_ACK')
          AND function_right.is_granted = true
    ) THEN
        RAISE EXCEPTION 'failed to grant watchlist function permissions to the Admin group';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM permissions_data_rights data_right
        WHERE data_right.user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND data_right.contract_id = '*'
          AND data_right.is_granted = true
    ) THEN
        RAISE EXCEPTION 'failed to grant admin2 the wildcard data-scope';
    END IF;
END
$$;

COMMIT;
