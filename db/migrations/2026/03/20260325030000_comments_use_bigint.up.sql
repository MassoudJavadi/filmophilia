ALTER TABLE reactions
    DROP CONSTRAINT IF EXISTS reactions_comment_id_fkey;

ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS comments_parent_id_fkey;

ALTER SEQUENCE comments_id_seq AS BIGINT;

ALTER TABLE comments
    ALTER COLUMN id TYPE BIGINT,
    ALTER COLUMN parent_id TYPE BIGINT;

ALTER TABLE reactions
    ALTER COLUMN comment_id TYPE BIGINT;

ALTER TABLE comments
    ADD CONSTRAINT comments_parent_id_fkey
        FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE SET NULL;

ALTER TABLE reactions
    ADD CONSTRAINT reactions_comment_id_fkey
        FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE;
