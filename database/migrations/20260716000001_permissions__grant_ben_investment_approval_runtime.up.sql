-- =============================================================================
-- Investment Decision Approver role (production catalog)
-- =============================================================================
-- The approval HTTP routes require both generic approval-engine permissions
-- and the investment decision subject permission. Keep these rights in a
-- dedicated approver role so ordinary Investment Decision Operator members do
-- not inherit them.
--
-- This migration grants only inbox/request visibility plus approve/reject for
-- investment decisions. It does not grant approval configuration, cancellation,
-- revocation, audit administration, stage-2 access, or self-approval.
--
-- This migration creates the role catalog and its rights ONLY. It must never
-- assign the role to any named identity (demo or otherwise) — production
-- migration/bootstrap paths must not assign privileges to named identities.
-- Assigning real users to this role is an operator action performed through
-- the application after deployment; development/test membership for the demo
-- user "ben" is seeded by the demo-only seed at
-- database/seeds/demo/006_ben_investment_decision_approver_seed.sql, gated by
-- APP_ENV so it never reaches production.
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

-- Named-identity assignment intentionally removed: production and upgraded
-- migration paths must never assign this role to a demo or otherwise named
-- identity. See database/seeds/demo/006_ben_investment_decision_approver_seed.sql
-- for the development/test-only membership, gated by APP_ENV.

COMMIT;
