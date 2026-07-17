-- =============================================================================
-- Investment decision/execution/confirmation permission seed (idempotent)
-- =============================================================================
-- Grants the fine-grained INVESTMENT_DECISION_*/INVESTMENT_EXECUTION_*/
-- INVESTMENT_CONFIRMATION_* function permissions (backend/internal/investment/
-- permission/policies.go) needed to use the frontend Operator catalog's
-- three workflows:
--   OP-01  BUY/SELL Single Securities  -> INVESTMENT_DECISION_VIEW/MANAGE/SUBMIT
--   OP-02  Execution & Approvals       -> INVESTMENT_EXECUTION_VIEW/MANAGE,
--                                          INVESTMENT_DECISION_APPROVE
--   OP-03  Review & Print Summary      -> INVESTMENT_DECISION_VIEW
--
-- 007_investment_permission_seed.sql granted Admin the base fund/portfolio/
-- ledger/research permission set but never included these decision/
-- execution/confirmation codes, so neither the seeded admin nor ben account
-- could reach OP-01/OP-02/OP-03 (INVESTMENT_DECISION_CANCEL is included for
-- completeness alongside VIEW/MANAGE/SUBMIT/APPROVE since they're the same
-- decision-lifecycle surface).
--
-- Group: Investment Decision Operator (new) : ben (member)
-- Also grants the same codes directly to the existing Admin group.
--
-- DEV ONLY — references the admin/ben users from 001_initial_seed.sql /
-- 004_investment_process_assignment_seed.sql.
-- =============================================================================

BEGIN;

-- ── Grant directly to Admin ──────────────────────────────────────────────────
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('INVESTMENT_DECISION_VIEW'),
    ('INVESTMENT_DECISION_MANAGE'),
    ('INVESTMENT_DECISION_SUBMIT'),
    ('INVESTMENT_DECISION_CANCEL'),
    ('INVESTMENT_DECISION_APPROVE'),
    ('INVESTMENT_EXECUTION_VIEW'),
    ('INVESTMENT_EXECUTION_MANAGE'),
    ('INVESTMENT_CONFIRMATION_VIEW'),
    ('INVESTMENT_CONFIRMATION_MANAGE')
) AS p(code)
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

-- ── Group for ben (and any future operator granted the same access) ─────────
INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000040',
    'Investment Decision Operator',
    'Operators authorised to create/submit investment decisions, manage executions, and approve decisions via the Operator catalog (OP-01/OP-02/OP-03).',
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
SELECT g.id, p.code, true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
CROSS JOIN (VALUES
    ('INVESTMENT_DECISION_VIEW'),
    ('INVESTMENT_DECISION_MANAGE'),
    ('INVESTMENT_DECISION_SUBMIT'),
    ('INVESTMENT_DECISION_CANCEL'),
    ('INVESTMENT_DECISION_APPROVE'),
    ('INVESTMENT_EXECUTION_VIEW'),
    ('INVESTMENT_EXECUTION_MANAGE'),
    ('INVESTMENT_CONFIRMATION_VIEW'),
    ('INVESTMENT_CONFIRMATION_MANAGE')
) AS p(code)
WHERE g.name = 'Investment Decision Operator'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

-- ── Membership: ben ──────────────────────────────────────────────────────────
INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
SELECT 'a0000000-0000-0000-0000-000000000010'::uuid, g.id, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
WHERE g.name = 'Investment Decision Operator'
ON CONFLICT (user_id, group_id) DO NOTHING;

COMMIT;
