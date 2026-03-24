-- ============================================================
-- Rollback: Remove partitioning from notifications and activities
-- ============================================================

-- ============================================================
-- PART 1: Revert notifications partitioning
-- ============================================================

-- Create non-partitioned notifications table
CREATE TABLE notifications_regular (
    id BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

-- Create sequence
CREATE SEQUENCE notifications_regular_id_seq AS BIGINT;
ALTER TABLE notifications_regular ALTER COLUMN id SET DEFAULT nextval('notifications_regular_id_seq');

-- Copy data from partitioned table
INSERT INTO notifications_regular
SELECT * FROM notifications;

-- Sync sequence
SELECT setval('notifications_regular_id_seq', COALESCE((SELECT MAX(id) FROM notifications_regular), 1));

-- Recreate indexes
CREATE INDEX notifications_regular_user_unread_idx
    ON notifications_regular (user_id, is_read, created_at DESC);

-- Recreate foreign key
ALTER TABLE notifications_regular
    ADD CONSTRAINT notifications_regular_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Drop partitioned table (this drops all partitions)
DROP TABLE notifications;
ALTER TABLE notifications_regular RENAME TO notifications;
ALTER SEQUENCE notifications_regular_id_seq RENAME TO notifications_id_seq;
ALTER INDEX notifications_regular_user_unread_idx RENAME TO notifications_user_unread_idx;
ALTER TABLE notifications RENAME CONSTRAINT notifications_regular_user_id_fkey TO notifications_user_id_fkey;

-- ============================================================
-- PART 2: Revert activities partitioning
-- ============================================================

-- Create non-partitioned activities table
CREATE TABLE activities_regular (
    id BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    entity_type entity_type NOT NULL,
    entity_id BIGINT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

-- Create sequence
CREATE SEQUENCE activities_regular_id_seq AS BIGINT;
ALTER TABLE activities_regular ALTER COLUMN id SET DEFAULT nextval('activities_regular_id_seq');

-- Copy data from partitioned table
INSERT INTO activities_regular
SELECT * FROM activities;

-- Sync sequence
SELECT setval('activities_regular_id_seq', COALESCE((SELECT MAX(id) FROM activities_regular), 1));

-- Recreate indexes
CREATE INDEX activities_regular_entity_idx
    ON activities_regular (entity_type, entity_id);
CREATE INDEX activities_regular_user_created_idx
    ON activities_regular (user_id, created_at DESC);

-- Recreate foreign key
ALTER TABLE activities_regular
    ADD CONSTRAINT activities_regular_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Drop partitioned table (this drops all partitions)
DROP TABLE activities;
ALTER TABLE activities_regular RENAME TO activities;
ALTER SEQUENCE activities_regular_id_seq RENAME TO activities_id_seq;
ALTER INDEX activities_regular_entity_idx RENAME TO activities_entity_idx;
ALTER INDEX activities_regular_user_created_idx RENAME TO activities_user_created_idx;
ALTER TABLE activities RENAME CONSTRAINT activities_regular_user_id_fkey TO activities_user_id_fkey;
