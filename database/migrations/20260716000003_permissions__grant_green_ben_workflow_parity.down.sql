-- Remove only Green permission-group memberships created by the matching up
-- migration or seed 021. Pre-existing memberships with different identifiers,
-- function rights, data scopes, approval groups, stages, and maker-checker
-- rules remain unchanged.

BEGIN;

DELETE FROM permissions_accounts_groups account_group
USING (VALUES
    ('b1000000-0000-0000-0000-000000000051'::uuid, 'b0000000-0000-0000-0000-000000000011'::uuid),
    ('b1000000-0000-0000-0000-000000000052'::uuid, 'b0000000-0000-0000-0000-000000000040'::uuid),
    ('b1000000-0000-0000-0000-000000000053'::uuid, 'b0000000-0000-0000-0000-000000000041'::uuid),
    ('b1000000-0000-0000-0000-000000000054'::uuid, 'b0000000-0000-0000-0000-000000000042'::uuid)
) AS owned(membership_id, group_id)
WHERE account_group.id = owned.membership_id
  AND account_group.user_id = 'a0000000-0000-0000-0000-000000000011'::uuid
  AND account_group.group_id = owned.group_id;

COMMIT;
