-- =============================================================================
-- Approval module — COMPLIANCE_RELEASE process config seed (idempotent)
-- =============================================================================
-- Seeds the approval group, process configuration, and single-stage approver
-- for COMPLIANCE_RELEASE subjects. Without this seed, any call to
-- SubmitForApproval("COMPLIANCE_RELEASE") fails with "no active process config".
--
-- Group:
--   COMPLIANCE_IRG_REVIEWERS : admin (priority 1)
--
-- Process (COMPANY-global):
--   COMPLIANCE_RELEASE
--     stage 1 GROUP_ANY  COMPLIANCE_IRG_REVIEWERS (final)
--
-- Permission grant:
--   The COMPLIANCE_IRG_REVIEWERS group is granted INVESTMENT_COMPLIANCE_RELEASE_APPROVE
--   so the subject accessor's function-permission check passes for group members.
--
-- DEV ONLY — references the admin user from 001_initial_seed.sql.
-- =============================================================================

BEGIN;

-- ── Group ─────────────────────────────────────────────────────────────────────
INSERT INTO approval__groups (id, group_code, group_name, remarks, is_active, created_by)
VALUES (
    'a9000000-0000-0000-0000-000000000004',
    'COMPLIANCE_IRG_REVIEWERS',
    'Compliance / IRG Reviewers',
    'Dev seed group — compliance officers authorised to release blocked decisions',
    true,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (group_code) DO NOTHING;

-- ── Group members ─────────────────────────────────────────────────────────────
INSERT INTO approval__group_members (id, group_id, user_id, priority_order, member_type, status, is_active, created_by)
VALUES (
    'a9100000-0000-0000-0000-000000000006',
    'a9000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000001',  -- admin
    1, 'SUPERVISOR', 'APPROVED', true,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (group_id, user_id) DO NOTHING;

-- ── Process config: COMPLIANCE_RELEASE ───────────────────────────────────────
INSERT INTO approval__process_configs
    (id, process_code, process_name, process_type, contract_type, contract_id, effective_date,
     is_active, group_approval_enabled, require_team_approval, created_by)
VALUES (
    'a9200000-0000-0000-0000-000000000004',
    'PROC_COMPLIANCE_RELEASE_DEFAULT',
    'Compliance Release Approval (default)',
    'COMPLIANCE_RELEASE', 'COMPANY', NULL, CURRENT_DATE - 1,
    true, true, false,
    'a0000000-0000-0000-0000-000000000001'
)
ON CONFLICT (process_code) DO NOTHING;

-- ── Stage ─────────────────────────────────────────────────────────────────────
INSERT INTO approval__process_stages
    (id, process_config_id, stage_number, stage_name, approver_mode, approval_group_id, required_approval_count, is_final_stage, reject_policy)
VALUES (
    'a9300000-0000-0000-0000-000000000007',
    'a9200000-0000-0000-0000-000000000004',
    1, 'Compliance officer sign-off', 'GROUP_ANY',
    'a9000000-0000-0000-0000-000000000004',
    1, true, 'STOP'
)
ON CONFLICT (process_config_id, stage_number) DO NOTHING;

-- ── Function permission grant ─────────────────────────────────────────────────
-- Ensure a permissions group exists for compliance reviewers and grant the
-- INVESTMENT_COMPLIANCE_RELEASE_APPROVE function permission so the subject
-- accessor's function-permission gate passes for group members.

INSERT INTO permissions_groups (id, name, description, is_active, created_by, updated_by)
VALUES (
    'b0000000-0000-0000-0000-000000000030',
    'Compliance IRG Reviewer',
    'Compliance / IRG officers authorised to release investment decisions blocked by compliance rules.',
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
SELECT g.id, 'INVESTMENT_COMPLIANCE_RELEASE_APPROVE', true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
WHERE g.name = 'Compliance IRG Reviewer'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

-- Also grant to Admin so the demo user can act on compliance release approvals.
INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
SELECT g.id, 'INVESTMENT_COMPLIANCE_RELEASE_APPROVE', true, 'a0000000-0000-0000-0000-000000000001'::uuid
FROM permissions_groups g
WHERE g.name = 'Admin'
ON CONFLICT (group_id, permission_code) DO UPDATE
SET is_granted = EXCLUDED.is_granted;

COMMIT;
