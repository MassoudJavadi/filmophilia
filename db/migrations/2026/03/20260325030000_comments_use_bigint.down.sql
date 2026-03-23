ALTER TABLE reactions
    DROP CONSTRAINT IF EXISTS reactions_comment_id_fkey;

ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS comments_parent_id_fkey;

ALTER TABLE reactions
    ALTER COLUMN comment_id TYPE INT USING comment_id::integer;

ALTER TABLE comments
    ALTER COLUMN parent_id TYPE INT USING parent_id::integer,
    ALTER COLUMN id TYPE INT USING id::integer;

ALTER SEQUENCE comments_id_seq AS INTEGER;

ALTER TABLE comments
    ADD CONSTRAINT comments_parent_id_fkey
        FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE SET NULL;

ALTER TABLE reactions
    ADD CONSTRAINT reactions_comment_id_fkey
        FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE;
