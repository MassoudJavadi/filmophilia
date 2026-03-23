UPDATE users
SET password_hash = NULL
WHERE password_hash = '';

ALTER TABLE users
    ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users
    ADD CONSTRAINT users_password_hash_nonempty_chk
    CHECK (password_hash IS NULL OR length(password_hash) > 0);
