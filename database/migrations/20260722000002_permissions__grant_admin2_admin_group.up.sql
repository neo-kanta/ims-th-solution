-- Grant admin2 the same group-derived authority as the bootstrap admin.
-- This intentionally includes permission-request merge/apply authority and
-- broad application-administrator permissions.

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM iam_users
        WHERE id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND username = 'admin2'
          AND is_active = true
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'active canonical admin2 account is missing';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM permissions_groups
        WHERE id = 'b0000000-0000-0000-0000-000000000001'::uuid
          AND name = 'Admin'
          AND is_active = true
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'active canonical Admin group is missing';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM permissions_accounts_groups
        WHERE id = 'b1000000-0000-0000-0000-000000000003'::uuid
          AND (
              user_id <> 'a0000000-0000-0000-0000-000000000002'::uuid
              OR group_id <> 'b0000000-0000-0000-0000-000000000001'::uuid
          )
    ) THEN
        RAISE EXCEPTION 'admin2 Admin-group membership id is already in use';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM audit_logs
        WHERE id = 'd1000000-0000-0000-0000-000000000003'::uuid
          AND (
              action <> 'BOOTSTRAP_ADMIN_GROUP_GRANTED'
              OR entity_id <> 'a0000000-0000-0000-0000-000000000002'
          )
    ) THEN
        RAISE EXCEPTION 'admin2 Admin-group audit id is already in use';
    END IF;
END
$$;

INSERT INTO permissions_accounts_groups (id, user_id, group_id, assigned_by)
VALUES (
    'b1000000-0000-0000-0000-000000000003'::uuid,
    'a0000000-0000-0000-0000-000000000002'::uuid,
    'b0000000-0000-0000-0000-000000000001'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid
)
ON CONFLICT (user_id, group_id) DO NOTHING;

INSERT INTO audit_logs (
    id, actor_user_id, action, module, entity_type, entity_id,
    before_json, after_json
)
VALUES (
    'd1000000-0000-0000-0000-000000000003'::uuid,
    'a0000000-0000-0000-0000-000000000001'::uuid,
    'BOOTSTRAP_ADMIN_GROUP_GRANTED',
    'permissions',
    'iam_user',
    'a0000000-0000-0000-0000-000000000002',
    '{}'::jsonb,
    '{"group":"Admin","authority":"full group-derived administrator permissions","source":"migration 20260722000002"}'::jsonb
)
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM permissions_accounts_groups
        WHERE user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
          AND group_id = 'b0000000-0000-0000-0000-000000000001'::uuid
    ) THEN
        RAISE EXCEPTION 'failed to grant admin2 Admin-group membership';
    END IF;
END
$$;

COMMIT;
