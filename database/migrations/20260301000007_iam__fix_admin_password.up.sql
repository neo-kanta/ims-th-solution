-- Fix admin password hash to match 'admin123'
-- This is necessary because the original seed had an incorrect hash.
UPDATE iam_users 
SET password_hash = '$2a$12$RlJ58G8tTbtR8.xKogKewOgRjs0RcGzW6S0JTmZbiYUJplGPnZO1C'
WHERE username = 'admin';
