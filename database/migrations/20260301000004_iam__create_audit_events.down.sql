DROP TRIGGER IF EXISTS trg_iam_audit_events_no_delete ON iam_audit_events;
DROP TRIGGER IF EXISTS trg_iam_audit_events_no_update ON iam_audit_events;
DROP FUNCTION IF EXISTS prevent_audit_modification();
DROP TABLE IF EXISTS iam_audit_events;
