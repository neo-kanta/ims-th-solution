-- Remove the dedicated Ben operation-page role introduced by the up
-- migration. Cascades remove only this migration-owned role's grants and user
-- membership; data scopes, operator rights, approval stages, and maker-checker
-- rules remain unchanged.

BEGIN;

DELETE FROM permissions_groups
WHERE id = 'b0000000-0000-0000-0000-000000000042'
  AND name = 'Investment Operation Page Access';

-- Function definitions are canonical or shared catalog entries, so rollback
-- intentionally retains them.

COMMIT;
