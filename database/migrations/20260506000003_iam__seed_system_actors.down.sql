-- Reverse the system actor seeds. Identified by their stable UUIDs.

DELETE FROM iam_users
WHERE id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002'
);
