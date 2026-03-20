-- ============================================================
-- RESTORE REVIEWS (ROLLBACK)
-- ============================================================

-- Restore notification_type enum
ALTER TYPE notification_type RENAME TO notification_type_old;
CREATE TYPE notification_type AS ENUM (
    'NEW_FOLLOWER', 'NEW_LIKE', 'NEW_COMMENT', 'NEW_REVIEW',
    'ACCOUNT_ACTIVATED', 'ACCOUNT_SUSPENDED', 'ACCOUNT_BANNED', 'SYSTEM_ALERT'
);
ALTER TABLE notifications ALTER COLUMN type TYPE notification_type USING type::text::notification_type;
DROP TYPE notification_type_old;

-- Restore entity_type enum
ALTER TYPE entity_type RENAME TO entity_type_old;
CREATE TYPE entity_type AS ENUM ('MOVIE', 'REVIEW', 'COMMENT', 'USER');
ALTER TABLE activities ALTER COLUMN entity_type TYPE entity_type USING entity_type::text::entity_type;
DROP TYPE entity_type_old;

-- Recreate reviews table
CREATE TABLE reviews (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id    INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    title       VARCHAR(255),
    content     TEXT NOT NULL,
    like_count  INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, movie_id)
);

CREATE INDEX reviews_movie_id_idx ON reviews (movie_id);
CREATE INDEX reviews_created_at_idx ON reviews (created_at DESC);
CREATE INDEX reviews_like_count_idx ON reviews (like_count DESC);

CREATE TRIGGER reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Restore reactions table structure
ALTER TABLE reactions ALTER COLUMN comment_id DROP NOT NULL;
ALTER TABLE reactions ADD COLUMN review_id INT REFERENCES reviews(id) ON DELETE CASCADE;
ALTER TABLE reactions ADD CONSTRAINT reactions_check CHECK (
    (review_id IS NOT NULL AND comment_id IS NULL) OR
    (review_id IS NULL AND comment_id IS NOT NULL)
);

CREATE UNIQUE INDEX reactions_user_review_unique
    ON reactions (user_id, review_id) WHERE review_id IS NOT NULL;
CREATE INDEX reactions_review_id_idx ON reactions (review_id);
