CREATE OR REPLACE FUNCTION hash_token_value(token_value TEXT)
RETURNS TEXT
LANGUAGE sql
IMMUTABLE
AS $$
    SELECT CASE
        WHEN token_value IS NULL THEN NULL
        WHEN token_value LIKE 'sha256:%' THEN token_value
        ELSE 'sha256:' || encode(digest(token_value, 'sha256'), 'hex')
    END
$$;

CREATE OR REPLACE FUNCTION hash_sessions_refresh_token()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.refresh_token = hash_token_value(NEW.refresh_token);
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION hash_accounts_tokens()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.access_token = hash_token_value(NEW.access_token);
    NEW.refresh_token = hash_token_value(NEW.refresh_token);
    RETURN NEW;
END;
$$;

UPDATE sessions
SET refresh_token = hash_token_value(refresh_token)
WHERE refresh_token IS NOT NULL
  AND refresh_token NOT LIKE 'sha256:%';

UPDATE accounts
SET
    access_token = hash_token_value(access_token),
    refresh_token = hash_token_value(refresh_token)
WHERE (access_token IS NOT NULL AND access_token NOT LIKE 'sha256:%')
   OR (refresh_token IS NOT NULL AND refresh_token NOT LIKE 'sha256:%');

DROP TRIGGER IF EXISTS sessions_hash_refresh_token_before_write ON sessions;
CREATE TRIGGER sessions_hash_refresh_token_before_write
BEFORE INSERT OR UPDATE OF refresh_token ON sessions
FOR EACH ROW
EXECUTE FUNCTION hash_sessions_refresh_token();

DROP TRIGGER IF EXISTS accounts_hash_tokens_before_write ON accounts;
CREATE TRIGGER accounts_hash_tokens_before_write
BEFORE INSERT OR UPDATE OF access_token, refresh_token ON accounts
FOR EACH ROW
EXECUTE FUNCTION hash_accounts_tokens();
