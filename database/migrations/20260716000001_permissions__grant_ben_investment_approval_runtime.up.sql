-- =============================================================================
-- Ben investment-decision approval runtime permissions
-- =============================================================================
-- Ben is an APPROVED member of FUND_MANAGER_REVIEWERS, but the approval HTTP
-- routes require both generic approval-engine permissions and the investment
-- decision subject permission. Keep these rights in a dedicated approver role
-- so ordinary Investment Decision Operator members do not inherit them.
--
-- This migration grants only inbox/request visibility plus approve/reject for
-- investment decisions. It does not grant approval configuration, cancellation,
-- revocation, audit administration, stage-2 access, or self-approval.
-- =============================================================================

BEGIN;

-- Migrations run before development seeds on a fresh database. Keep these
-- catalog rows available so the permissions_function_rights FK is valid;
-- cmd/seed remains the canonical catalog refresher.
INSERT INTO permissions_function_definitions (code, module, name, description)
VALUES
    ('APPROVAL_VIEW_INBOX',          'approval',   'Approval View Inbox',          'View the personal approval inbox of pending tasks.'),
    ('APPROVAL_VIEW_REQUEST',        'approval',   'Approval View Request',        'View approval requests, timeline and signatures.'),
    ('APPROVAL_APPROVE',             'approval',   'Approval Approve',             'Approve an assigned approval task.'),
    ('APPROVAL_REJECT',              'approval',   'Approval Reject',              'Reject an assigned approval task.'),
    ('INVESTMENT_DECISION_APPROVE',  'investment', 'Investment Decision Approve',  'Approve or reject pending investment decisions via the batch approval screen.')
ON CONFLICT (code) DO UPDATE
SET module        = EXCLUDED.module,
    name          = EXCLUDED.name,
    description   = EXCLUDED.description,
    deprecated_at = NULL;

-- This role is owned by this migration. Reject a same-name role with another
-- identifier rather than mutating an operator-managed authorization object.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM permissions_groups
        WHERE name = 'Investment Decision Approver'
          AND id <> 'b0000000-0000-0000-0000-000000000041'::uuid
    ) THEN
        RAISE EXCEPTION 'Investment Decision Approver permission group already exists with an unexpected id';
    END IF;
END
$$;

INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000041',
    'Investment Decision Approver',
    'Assigned investment-decision reviewers authorised to use the approval inbox and approve or reject tasks assigned to them.',
    true,
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT ON CONSTRAINT uq_permissions_groups_name DO NOTHING;

INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT
    'b0000000-0000-0000-0000-000000000041'::uuid,
    permission_code.code,
    true,
    'a0000000-0000-0000-0000-000000000001'::uuid
FROM (VALUES
    ('APPROVAL_VIEW_INBOX'),
    ('APPROVAL_VIEW_REQUEST'),
    ('APPROVAL_APPROVE'),
    ('APPROVAL_REJECT'),
    ('INVESTMENT_DECISION_APPROVE')
) AS permission_code(code)
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = true;

-- Ben is development/demo data and does not exist in a migration-only fresh
-- database. Assign him when present; seed 019 performs the same idempotent
-- assignment after fresh migrations.
INSERT INTO permissions_accounts_groups (id, user_id, group_id, assigned_by)
SELECT
    'b1000000-0000-0000-0000-000000000041',
    ben_user.id,
    'b0000000-0000-0000-0000-000000000041'::uuid,
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
          AND permission_group.id = 'b0000000-0000-0000-0000-000000000041'::uuid
          AND permission_group.is_active = true
          AND permission_group.deleted_at IS NULL
          AND 5 = (
              SELECT COUNT(*)
              FROM permissions_function_rights function_right
              WHERE function_right.group_id = permission_group.id
                AND function_right.permission_code IN (
                    'APPROVAL_VIEW_INBOX',
                    'APPROVAL_VIEW_REQUEST',
                    'APPROVAL_APPROVE',
                    'APPROVAL_REJECT',
                    'INVESTMENT_DECISION_APPROVE'
                )
                AND function_right.is_granted = true
          )
    ) THEN
        RAISE EXCEPTION 'failed to establish Ben investment-decision approval permissions';
    END IF;
END
$$;

COMMIT;
