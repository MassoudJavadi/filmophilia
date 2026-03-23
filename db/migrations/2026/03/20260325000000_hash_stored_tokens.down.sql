DROP TRIGGER IF EXISTS accounts_hash_tokens_before_write ON accounts;
DROP TRIGGER IF EXISTS sessions_hash_refresh_token_before_write ON sessions;

DROP FUNCTION IF EXISTS hash_accounts_tokens();
DROP FUNCTION IF EXISTS hash_sessions_refresh_token();
DROP FUNCTION IF EXISTS hash_token_value(TEXT);
