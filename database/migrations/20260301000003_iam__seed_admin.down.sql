-- Remove seeded function permissions for Admin group
DELETE FROM permissions_function_rights WHERE group_id = 'b0000000-0000-0000-0000-000000000001';

-- Remove admin user from Admin group
DELETE FROM permissions_accounts_groups
WHERE user_id = 'a0000000-0000-0000-0000-000000000001'
  AND group_id = 'b0000000-0000-0000-0000-000000000001';

-- Remove Admin group
DELETE FROM permissions_groups WHERE id = 'b0000000-0000-0000-0000-000000000001';

-- Remove admin user
DELETE FROM iam_users WHERE id = 'a0000000-0000-0000-0000-000000000001';
