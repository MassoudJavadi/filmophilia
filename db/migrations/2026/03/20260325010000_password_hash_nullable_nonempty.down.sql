ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_password_hash_nonempty_chk;

UPDATE users
SET password_hash = ''
WHERE password_hash IS NULL;

ALTER TABLE users
    ALTER COLUMN password_hash SET NOT NULL;
