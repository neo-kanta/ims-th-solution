-- Revoke only the grants added by the matching up migration: the Admin
-- group's watchlist function rights and admin2's wildcard data-scope. The
-- Admin group itself, admin/admin2 accounts, and the watchlist permission
-- catalog definitions are left untouched.

BEGIN;

DELETE FROM permissions_function_rights
WHERE group_id = 'b0000000-0000-0000-0000-000000000001'::uuid
  AND permission_code IN ('WATCHLIST_VIEW', 'WATCHLIST_MANAGE', 'WATCHLIST_ALERT_ACK');

DELETE FROM permissions_data_rights
WHERE user_id = 'a0000000-0000-0000-0000-000000000002'::uuid
  AND contract_id = '*';

-- Function definitions are canonical catalog entries maintained by the
-- watchlist module's Go permission provider, so rollback intentionally
-- retains them.

COMMIT;
