DROP TRIGGER IF EXISTS trg_iam_users_updated_at ON iam_users;

DROP FUNCTION IF EXISTS set_updated_at ();

DROP TABLE IF EXISTS iam_users;