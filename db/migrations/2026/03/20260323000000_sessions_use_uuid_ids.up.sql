CREATE EXTENSION IF NOT EXISTS pgcrypto;

UPDATE sessions
SET id = gen_random_uuid()::text
WHERE id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

ALTER TABLE sessions
ALTER COLUMN id DROP DEFAULT,
ALTER COLUMN id TYPE UUID USING id::uuid,
ALTER COLUMN id SET DEFAULT gen_random_uuid();
