-- Add index on sessions.refresh_token for faster token lookups
-- This is critical for performance on refresh token validation
CREATE INDEX CONCURRENTLY IF NOT EXISTS sessions_refresh_token_idx ON sessions (refresh_token);
