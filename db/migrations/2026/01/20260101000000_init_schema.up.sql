-- ============================================================
-- FILMOPHILIA - Initial Schema
-- Created: 2026-01-01
-- ============================================================

-- ============================================================
-- EXTENSIONS (Must be first for GIN indexes)
-- ============================================================

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE role AS ENUM ('USER', 'ADMIN');
CREATE TYPE user_status AS ENUM ('PENDING', 'ACTIVE', 'SUSPENDED', 'BANNED');
CREATE TYPE reaction_type AS ENUM ('LIKE', 'LOVE', 'LAUGH', 'SAD', 'ANGRY');
CREATE TYPE department AS ENUM (
    'DIRECTING', 'WRITING', 'ACTING', 'PRODUCTION',
    'CINEMATOGRAPHY', 'EDITING', 'SOUND', 'ART'
);
CREATE TYPE entity_type AS ENUM ('MOVIE', 'REVIEW', 'COMMENT', 'USER');
CREATE TYPE notification_type AS ENUM (
    'NEW_FOLLOWER', 'NEW_LIKE', 'NEW_COMMENT', 'NEW_REVIEW',
    'ACCOUNT_ACTIVATED', 'ACCOUNT_SUSPENDED', 'ACCOUNT_BANNED', 'SYSTEM_ALERT'
);

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    username      VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(100),
    avatar_url    TEXT,
    bio           TEXT,
    role          role NOT NULL DEFAULT 'USER',
    status        user_status NOT NULL DEFAULT 'PENDING',
    is_verified   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_email_idx ON users (email);
CREATE INDEX users_username_idx ON users (username);
CREATE INDEX users_status_idx ON users (status);

-- ============================================================
-- USER STATUS LOGS (Audit Trail)
-- ============================================================

CREATE TABLE user_status_logs (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    old_status  user_status,
    new_status  user_status NOT NULL,
    reason      TEXT,
    changed_by  INT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX user_status_logs_user_idx ON user_status_logs (user_id, created_at DESC);

-- ============================================================
-- SESSIONS
-- ============================================================

CREATE TABLE sessions (
    id            VARCHAR(255) PRIMARY KEY,
    user_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT,
    user_agent    TEXT,
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- ============================================================
-- ACCOUNTS (OAuth Providers)
-- ============================================================

CREATE TABLE accounts (
    id                  SERIAL PRIMARY KEY,
    user_id             INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider            VARCHAR(50) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    access_token        TEXT,
    refresh_token       TEXT,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_account_id)
);

CREATE INDEX accounts_user_id_idx ON accounts (user_id);

-- ============================================================
-- FOLLOWS
-- ============================================================

CREATE TABLE follows (
    id           SERIAL PRIMARY KEY,
    follower_id  INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (follower_id, following_id),
    CHECK (follower_id != following_id)
);

CREATE INDEX follows_follower_idx ON follows (follower_id);
CREATE INDEX follows_following_idx ON follows (following_id);

-- ============================================================
-- MOVIES
-- ============================================================

CREATE TABLE movies (
    id                SERIAL PRIMARY KEY,
    title             VARCHAR(500) NOT NULL,
    slug              VARCHAR(500) NOT NULL UNIQUE,
    overview          TEXT,
    poster_url        TEXT,
    backdrop_url      TEXT,
    trailer_url       TEXT,
    release_date      DATE,
    runtime           INT,  -- minutes
    content_rating    VARCHAR(20),
    original_language VARCHAR(10),
    country           VARCHAR(100),
    imdb_id           VARCHAR(20),
    tmdb_id           INT UNIQUE,
    average_rating    REAL DEFAULT 0,
    rating_count      INT DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX movies_slug_idx ON movies (slug);
CREATE INDEX movies_release_date_idx ON movies (release_date DESC);
CREATE INDEX movies_average_rating_idx ON movies (average_rating DESC);
CREATE INDEX movies_title_trgm_idx ON movies USING GIN (title gin_trgm_ops);

-- ============================================================
-- PERSONS (Cast & Crew)
-- ============================================================

CREATE TABLE persons (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL UNIQUE,
    biography   TEXT,
    photo_url   TEXT,
    birth_date  DATE,
    death_date  DATE,
    birthplace  VARCHAR(255),
    tmdb_id     INT UNIQUE,
    imdb_id     VARCHAR(20),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX persons_name_trgm_idx ON persons USING GIN (name gin_trgm_ops);

-- ============================================================
-- GENRES
-- ============================================================

CREATE TABLE genres (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL UNIQUE,
    slug    VARCHAR(100) NOT NULL UNIQUE,
    tmdb_id INT UNIQUE
);

-- ============================================================
-- MOVIE_GENRES (Junction)
-- ============================================================

CREATE TABLE movie_genres (
    movie_id INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    genre_id INT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, genre_id)
);

CREATE INDEX movie_genres_genre_idx ON movie_genres (genre_id);

-- ============================================================
-- CREDITS (Cast & Crew per Movie)
-- ============================================================

CREATE TABLE credits (
    id         SERIAL PRIMARY KEY,
    movie_id   INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    person_id  INT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    department department NOT NULL,
    role       VARCHAR(255) NOT NULL,  -- "Director", "Screenplay", "Actor", etc.
    character  VARCHAR(255),            -- For actors only
    "order"    INT DEFAULT 0
);

CREATE INDEX credits_movie_id_idx ON credits (movie_id);
CREATE INDEX credits_person_id_idx ON credits (person_id);
CREATE INDEX credits_department_idx ON credits (department);

-- Partial unique indexes for credit uniqueness
CREATE UNIQUE INDEX credits_unique_non_actor
    ON credits (movie_id, person_id, role)
    WHERE character IS NULL;

CREATE UNIQUE INDEX credits_unique_actor
    ON credits (movie_id, person_id, role, character)
    WHERE character IS NOT NULL;

-- ============================================================
-- RATINGS
-- ============================================================

CREATE TABLE ratings (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id   INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    score      INT NOT NULL CHECK (score >= 0 AND score <= 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, movie_id)
);

CREATE INDEX ratings_movie_id_idx ON ratings (movie_id);
CREATE INDEX ratings_created_at_idx ON ratings (created_at DESC);

-- ============================================================
-- REVIEWS
-- ============================================================

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

-- ============================================================
-- COMMENTS
-- ============================================================

CREATE TABLE comments (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id   INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    parent_id  INT REFERENCES comments(id) ON DELETE SET NULL,
    content    TEXT NOT NULL,
    like_count INT DEFAULT 0,
    deleted_at TIMESTAMPTZ,  -- Soft delete
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX comments_movie_id_idx ON comments (movie_id);
CREATE INDEX comments_parent_id_idx ON comments (parent_id);
CREATE INDEX comments_created_at_idx ON comments (created_at DESC);
CREATE INDEX comments_deleted_idx ON comments (deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- WATCHLISTS
-- ============================================================

CREATE TABLE watchlists (
    id            SERIAL PRIMARY KEY,
    user_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id      INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    notes         TEXT,
    rank_position REAL,  -- Fractional indexing for ordering
    watched_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, movie_id)
);

CREATE INDEX watchlists_user_rank_idx ON watchlists (user_id, rank_position);

-- ============================================================
-- REACTIONS
-- ============================================================

CREATE TABLE reactions (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    review_id  INT REFERENCES reviews(id) ON DELETE CASCADE,
    comment_id INT REFERENCES comments(id) ON DELETE CASCADE,
    type       reaction_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Must target exactly one entity
    CHECK (
        (review_id IS NOT NULL AND comment_id IS NULL) OR
        (review_id IS NULL AND comment_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX reactions_user_review_unique
    ON reactions (user_id, review_id) WHERE review_id IS NOT NULL;
CREATE UNIQUE INDEX reactions_user_comment_unique
    ON reactions (user_id, comment_id) WHERE comment_id IS NOT NULL;
CREATE INDEX reactions_review_id_idx ON reactions (review_id);
CREATE INDEX reactions_comment_id_idx ON reactions (comment_id);

-- ============================================================
-- ACTIVITIES (Feed)
-- ============================================================

CREATE TABLE activities (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      VARCHAR(50) NOT NULL,  -- "rated", "reviewed", "followed", etc.
    entity_type entity_type NOT NULL,
    entity_id   INT NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX activities_user_created_idx ON activities (user_id, created_at DESC);
CREATE INDEX activities_entity_idx ON activities (entity_type, entity_id);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE notifications (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       notification_type NOT NULL,
    title      VARCHAR(255) NOT NULL,
    content    TEXT,
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    metadata   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX notifications_user_unread_idx
    ON notifications (user_id, is_read, created_at DESC);

-- ============================================================
-- TRIGGERS
-- ============================================================

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER movies_updated_at
    BEFORE UPDATE ON movies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER persons_updated_at
    BEFORE UPDATE ON persons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER ratings_updated_at
    BEFORE UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Update movie rating stats on rating changes
CREATE OR REPLACE FUNCTION update_movie_rating_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE movies SET
            average_rating = COALESCE((SELECT AVG(score)::REAL FROM ratings WHERE movie_id = OLD.movie_id), 0),
            rating_count = (SELECT COUNT(*) FROM ratings WHERE movie_id = OLD.movie_id)
        WHERE id = OLD.movie_id;
        RETURN OLD;
    ELSE
        UPDATE movies SET
            average_rating = COALESCE((SELECT AVG(score)::REAL FROM ratings WHERE movie_id = NEW.movie_id), 0),
            rating_count = (SELECT COUNT(*) FROM ratings WHERE movie_id = NEW.movie_id)
        WHERE id = NEW.movie_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ratings_stats_insert
    AFTER INSERT ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_update
    AFTER UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_delete
    AFTER DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

-- Update like counts on reaction changes
CREATE OR REPLACE FUNCTION update_like_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.review_id IS NOT NULL THEN
            UPDATE reviews SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE review_id = OLD.review_id
            ) WHERE id = OLD.review_id;
        END IF;
        IF OLD.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE comment_id = OLD.comment_id
            ) WHERE id = OLD.comment_id;
        END IF;
        RETURN OLD;
    ELSE
        IF NEW.review_id IS NOT NULL THEN
            UPDATE reviews SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE review_id = NEW.review_id
            ) WHERE id = NEW.review_id;
        END IF;
        IF NEW.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE comment_id = NEW.comment_id
            ) WHERE id = NEW.comment_id;
        END IF;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reactions_like_count_insert
    AFTER INSERT ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_delete
    AFTER DELETE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

-- Log user status changes
CREATE OR REPLACE FUNCTION log_user_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO user_status_logs (user_id, old_status, new_status)
        VALUES (NEW.id, OLD.status, NEW.status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_status_change
    AFTER UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION log_user_status_change();
