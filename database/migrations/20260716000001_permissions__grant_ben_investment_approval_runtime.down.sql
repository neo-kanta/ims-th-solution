-- Remove the dedicated investment-decision approver role introduced by the up
-- migration. Cascades remove only this migration-owned role's grants and user
-- memberships; operator permissions, approval stages, and maker-checker remain.

BEGIN;

DELETE FROM permissions_groups
WHERE id = 'b0000000-0000-0000-0000-000000000041'
  AND name = 'Investment Decision Approver';

-- Function definitions are canonical application catalog entries and may be
-- shared by other roles, so rollback intentionally retains them.

COMMIT;
