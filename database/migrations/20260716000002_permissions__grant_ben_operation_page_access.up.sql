-- =============================================================================
-- Investment Operation Page Access role (production catalog)
-- =============================================================================
-- The operation directory and OP-01 decision form need read access to funds,
-- portfolios (including scoped holdings/cash), and instruments in addition to
-- an operator's existing decision operator permissions. Keep these
-- prerequisites in a dedicated assignment role so only explicitly assigned
-- users receive the additional visibility.
--
-- This migration does not add data scopes, transaction-posting permissions,
-- approval-stage permissions, or any maker-checker bypass.
--
-- This migration creates the role catalog and its rights ONLY. It must never
-- assign the role to any named identity (demo or otherwise) — production
-- migration/bootstrap paths must not assign privileges to named identities.
-- Assigning real users to this role is an operator action performed through
-- the application after deployment; development/test membership for the demo
-- user "ben" is seeded by the demo-only seed at
-- database/seeds/demo/007_ben_operation_page_access_seed.sql, gated by
-- APP_ENV so it never reaches production.
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

-- Named-identity assignment intentionally removed: production and upgraded
-- migration paths must never assign this role to a demo or otherwise named
-- identity. See database/seeds/demo/007_ben_operation_page_access_seed.sql
-- for the development/test-only membership, gated by APP_ENV.

COMMIT;
