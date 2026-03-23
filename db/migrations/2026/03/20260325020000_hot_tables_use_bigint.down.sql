ALTER TABLE reactions
    ALTER COLUMN id TYPE INT USING id::integer;
ALTER SEQUENCE reactions_id_seq AS INTEGER;

ALTER TABLE ratings
    ALTER COLUMN id TYPE INT USING id::integer;
ALTER SEQUENCE ratings_id_seq AS INTEGER;

ALTER TABLE notifications
    ALTER COLUMN id TYPE INT USING id::integer;
ALTER SEQUENCE notifications_id_seq AS INTEGER;

ALTER TABLE activities
    ALTER COLUMN entity_id TYPE INT USING entity_id::integer,
    ALTER COLUMN id TYPE INT USING id::integer;
ALTER SEQUENCE activities_id_seq AS INTEGER;
