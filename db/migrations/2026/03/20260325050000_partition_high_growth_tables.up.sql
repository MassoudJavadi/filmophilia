-- ============================================================
-- Partition High-Growth Tables (notifications, activities)
-- Implements monthly range partitioning on created_at
-- ============================================================

-- ============================================================
-- PART 1: Partition notifications table
-- ============================================================

-- Create partitioned notifications table
-- Note: Partition key (created_at) must be part of primary key
CREATE TABLE notifications_partitioned (
    id BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for 2026
CREATE TABLE notifications_2026_01 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE notifications_2026_02 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE notifications_2026_03 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE notifications_2026_04 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE notifications_2026_05 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE notifications_2026_06 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE notifications_2026_07 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE notifications_2026_08 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE notifications_2026_09 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE notifications_2026_10 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE notifications_2026_11 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE notifications_2026_12 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Create monthly partitions for 2027
CREATE TABLE notifications_2027_01 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE notifications_2027_02 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE notifications_2027_03 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE notifications_2027_04 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE notifications_2027_05 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE notifications_2027_06 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE notifications_2027_07 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE notifications_2027_08 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE notifications_2027_09 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE notifications_2027_10 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE notifications_2027_11 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE notifications_2027_12 PARTITION OF notifications_partitioned
    FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');

-- Create default partition for data outside defined ranges
CREATE TABLE notifications_default PARTITION OF notifications_partitioned DEFAULT;

-- Create sequence for notifications
CREATE SEQUENCE notifications_partitioned_id_seq AS BIGINT;
ALTER TABLE notifications_partitioned ALTER COLUMN id SET DEFAULT nextval('notifications_partitioned_id_seq');

-- Migrate data from old table to partitioned table
INSERT INTO notifications_partitioned
SELECT * FROM notifications;

-- Sync sequence to current max ID
SELECT setval('notifications_partitioned_id_seq', COALESCE((SELECT MAX(id) FROM notifications_partitioned), 1));

-- Recreate indexes on partitioned table
CREATE INDEX notifications_partitioned_user_unread_idx
    ON notifications_partitioned (user_id, is_read, created_at DESC);

-- Recreate foreign key
ALTER TABLE notifications_partitioned
    ADD CONSTRAINT notifications_partitioned_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Drop old table and rename new one
DROP TABLE notifications;
ALTER TABLE notifications_partitioned RENAME TO notifications;
ALTER SEQUENCE notifications_partitioned_id_seq RENAME TO notifications_id_seq;
ALTER INDEX notifications_partitioned_user_unread_idx RENAME TO notifications_user_unread_idx;
ALTER TABLE notifications RENAME CONSTRAINT notifications_partitioned_user_id_fkey TO notifications_user_id_fkey;

-- ============================================================
-- PART 2: Partition activities table
-- ============================================================

-- Create partitioned activities table
CREATE TABLE activities_partitioned (
    id BIGINT NOT NULL,
    user_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    entity_type entity_type NOT NULL,
    entity_id BIGINT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for 2026
CREATE TABLE activities_2026_01 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE activities_2026_02 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE activities_2026_03 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE activities_2026_04 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE activities_2026_05 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE activities_2026_06 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE activities_2026_07 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE activities_2026_08 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE activities_2026_09 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE activities_2026_10 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE activities_2026_11 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE activities_2026_12 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Create monthly partitions for 2027
CREATE TABLE activities_2027_01 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');
CREATE TABLE activities_2027_02 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');
CREATE TABLE activities_2027_03 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');
CREATE TABLE activities_2027_04 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-04-01') TO ('2027-05-01');
CREATE TABLE activities_2027_05 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-05-01') TO ('2027-06-01');
CREATE TABLE activities_2027_06 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-06-01') TO ('2027-07-01');
CREATE TABLE activities_2027_07 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-07-01') TO ('2027-08-01');
CREATE TABLE activities_2027_08 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-08-01') TO ('2027-09-01');
CREATE TABLE activities_2027_09 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-09-01') TO ('2027-10-01');
CREATE TABLE activities_2027_10 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-10-01') TO ('2027-11-01');
CREATE TABLE activities_2027_11 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-11-01') TO ('2027-12-01');
CREATE TABLE activities_2027_12 PARTITION OF activities_partitioned
    FOR VALUES FROM ('2027-12-01') TO ('2028-01-01');

-- Create default partition
CREATE TABLE activities_default PARTITION OF activities_partitioned DEFAULT;

-- Create sequence for activities
CREATE SEQUENCE activities_partitioned_id_seq AS BIGINT;
ALTER TABLE activities_partitioned ALTER COLUMN id SET DEFAULT nextval('activities_partitioned_id_seq');

-- Migrate data from old table to partitioned table
INSERT INTO activities_partitioned
SELECT * FROM activities;

-- Sync sequence to current max ID
SELECT setval('activities_partitioned_id_seq', COALESCE((SELECT MAX(id) FROM activities_partitioned), 1));

-- Recreate indexes on partitioned table
CREATE INDEX activities_partitioned_entity_idx
    ON activities_partitioned (entity_type, entity_id);
CREATE INDEX activities_partitioned_user_created_idx
    ON activities_partitioned (user_id, created_at DESC);

-- Recreate foreign key
ALTER TABLE activities_partitioned
    ADD CONSTRAINT activities_partitioned_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Drop old table and rename new one
DROP TABLE activities;
ALTER TABLE activities_partitioned RENAME TO activities;
ALTER SEQUENCE activities_partitioned_id_seq RENAME TO activities_id_seq;
ALTER INDEX activities_partitioned_entity_idx RENAME TO activities_entity_idx;
ALTER INDEX activities_partitioned_user_created_idx RENAME TO activities_user_created_idx;
ALTER TABLE activities RENAME CONSTRAINT activities_partitioned_user_id_fkey TO activities_user_id_fkey;

-- ============================================================
-- PARTITION MAINTENANCE NOTES
-- ============================================================
-- To create new monthly partitions, run:
--
-- CREATE TABLE notifications_YYYY_MM PARTITION OF notifications
--     FOR VALUES FROM ('YYYY-MM-01') TO ('YYYY-MM+1-01');
--
-- CREATE TABLE activities_YYYY_MM PARTITION OF activities
--     FOR VALUES FROM ('YYYY-MM-01') TO ('YYYY-MM+1-01');
--
-- Consider automating this with a cron job or pg_partman extension.
-- To drop old partitions (e.g., data older than 2 years):
--
-- DROP TABLE notifications_YYYY_MM;
-- DROP TABLE activities_YYYY_MM;
-- ============================================================
