--
-- PostgreSQL database dump
--

\restrict 1TcVJ5TngzjjxM1z2t2hqTfxy26yXS1GWnNXsnCe2wes4FJq1NkUJIzyix4S8fl

-- Dumped from database version 17.7
-- Dumped by pg_dump version 17.7

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: department; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.department AS ENUM (
    'DIRECTING',
    'WRITING',
    'ACTING',
    'PRODUCTION',
    'CINEMATOGRAPHY',
    'EDITING',
    'SOUND',
    'ART'
);


--
-- Name: entity_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.entity_type AS ENUM (
    'MOVIE',
    'COMMENT',
    'USER'
);


--
-- Name: notification_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.notification_type AS ENUM (
    'NEW_FOLLOWER',
    'NEW_LIKE',
    'NEW_COMMENT',
    'ACCOUNT_ACTIVATED',
    'ACCOUNT_SUSPENDED',
    'ACCOUNT_BANNED',
    'SYSTEM_ALERT'
);


--
-- Name: reaction_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.reaction_type AS ENUM (
    'LIKE',
    'LOVE',
    'LAUGH',
    'SAD',
    'ANGRY'
);


--
-- Name: role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.role AS ENUM (
    'USER',
    'ADMIN'
);


--
-- Name: user_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.user_status AS ENUM (
    'PENDING',
    'ACTIVE',
    'SUSPENDED',
    'BANNED'
);


--
-- Name: hash_accounts_tokens(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.hash_accounts_tokens() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.access_token = hash_token_value(NEW.access_token);
    NEW.refresh_token = hash_token_value(NEW.refresh_token);
    RETURN NEW;
END;
$$;


--
-- Name: hash_sessions_refresh_token(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.hash_sessions_refresh_token() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.refresh_token = hash_token_value(NEW.refresh_token);
    RETURN NEW;
END;
$$;


--
-- Name: hash_token_value(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.hash_token_value(token_value text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT CASE
        WHEN token_value IS NULL THEN NULL
        WHEN token_value LIKE 'sha256:%' THEN token_value
        ELSE 'sha256:' || encode(digest(token_value, 'sha256'), 'hex')
    END
$$;


--
-- Name: log_user_status_change(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.log_user_status_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO user_status_logs (user_id, old_status, new_status)
        VALUES (NEW.id, OLD.status, NEW.status);
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: update_like_counts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_like_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Decrement like_count if the deleted reaction was a LIKE
        IF OLD.type = 'LIKE' THEN
            UPDATE comments
            SET like_count = GREATEST(like_count - 1, 0)
            WHERE id = OLD.comment_id;
        END IF;
        RETURN OLD;
    ELSIF TG_OP = 'INSERT' THEN
        -- Increment like_count if the new reaction is a LIKE
        IF NEW.type = 'LIKE' THEN
            UPDATE comments
            SET like_count = like_count + 1
            WHERE id = NEW.comment_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle reaction type changes (e.g., LIKE -> LOVE or LOVE -> LIKE)
        IF OLD.type = 'LIKE' AND NEW.type != 'LIKE' THEN
            -- Changed from LIKE to something else: decrement
            UPDATE comments
            SET like_count = GREATEST(like_count - 1, 0)
            WHERE id = OLD.comment_id;
        ELSIF OLD.type != 'LIKE' AND NEW.type = 'LIKE' THEN
            -- Changed from something else to LIKE: increment
            UPDATE comments
            SET like_count = like_count + 1
            WHERE id = NEW.comment_id;
        END IF;

        -- Handle comment_id changes (rare but possible)
        IF OLD.comment_id IS DISTINCT FROM NEW.comment_id THEN
            -- Remove from old comment
            IF OLD.type = 'LIKE' THEN
                UPDATE comments
                SET like_count = GREATEST(like_count - 1, 0)
                WHERE id = OLD.comment_id;
            END IF;
            -- Add to new comment
            IF NEW.type = 'LIKE' THEN
                UPDATE comments
                SET like_count = like_count + 1
                WHERE id = NEW.comment_id;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
END;
$$;


--
-- Name: update_movie_rating_stats(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_movie_rating_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        -- Decrement count and subtract score from sum
        UPDATE movies
        SET
            user_rating_count = GREATEST(user_rating_count - 1, 0),
            rating_sum = GREATEST(rating_sum - OLD.score, 0),
            user_avg_rating = CASE
                WHEN user_rating_count - 1 > 0
                THEN (rating_sum - OLD.score)::REAL / (user_rating_count - 1)
                ELSE 0
            END
        WHERE id = OLD.movie_id;
        RETURN OLD;
    ELSIF TG_OP = 'INSERT' THEN
        -- Increment count and add score to sum
        UPDATE movies
        SET
            user_rating_count = user_rating_count + 1,
            rating_sum = rating_sum + NEW.score,
            user_avg_rating = (rating_sum + NEW.score)::REAL / (user_rating_count + 1)
        WHERE id = NEW.movie_id;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle score changes
        IF OLD.score != NEW.score THEN
            UPDATE movies
            SET
                rating_sum = rating_sum - OLD.score + NEW.score,
                user_avg_rating = CASE
                    WHEN user_rating_count > 0
                    THEN (rating_sum - OLD.score + NEW.score)::REAL / user_rating_count
                    ELSE 0
                END
            WHERE id = NEW.movie_id;
        END IF;

        -- Handle movie_id changes (rare but possible due to unique constraint)
        IF OLD.movie_id != NEW.movie_id THEN
            -- Remove from old movie
            UPDATE movies
            SET
                user_rating_count = GREATEST(user_rating_count - 1, 0),
                rating_sum = GREATEST(rating_sum - OLD.score, 0),
                user_avg_rating = CASE
                    WHEN user_rating_count - 1 > 0
                    THEN (rating_sum - OLD.score)::REAL / (user_rating_count - 1)
                    ELSE 0
                END
            WHERE id = OLD.movie_id;

            -- Add to new movie
            UPDATE movies
            SET
                user_rating_count = user_rating_count + 1,
                rating_sum = rating_sum + NEW.score,
                user_avg_rating = (rating_sum + NEW.score)::REAL / (user_rating_count + 1)
            WHERE id = NEW.movie_id;
        END IF;
        RETURN NEW;
    END IF;
END;
$$;


--
-- Name: update_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.accounts (
    id integer NOT NULL,
    user_id integer NOT NULL,
    provider character varying(50) NOT NULL,
    provider_account_id character varying(255) NOT NULL,
    access_token text,
    refresh_token text,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- Name: activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.activities_id_seq OWNED BY public.activities.id;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    movie_id integer NOT NULL,
    parent_id bigint,
    content text NOT NULL,
    like_count integer DEFAULT 0,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


--
-- Name: credits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.credits (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    person_id integer NOT NULL,
    department public.department NOT NULL,
    role character varying(255) NOT NULL,
    "character" character varying(255),
    "order" integer DEFAULT 0
);


--
-- Name: credits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.credits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: credits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.credits_id_seq OWNED BY public.credits.id;


--
-- Name: follows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.follows (
    id integer NOT NULL,
    follower_id integer NOT NULL,
    following_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT follows_check CHECK ((follower_id <> following_id))
);


--
-- Name: follows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.follows_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: follows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.follows_id_seq OWNED BY public.follows.id;


--
-- Name: genres; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.genres (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    slug character varying(100) NOT NULL,
    tmdb_id integer
);


--
-- Name: genres_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.genres_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: genres_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.genres_id_seq OWNED BY public.genres.id;


--
-- Name: movie_genres; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_genres (
    movie_id integer NOT NULL,
    genre_id integer NOT NULL
);


--
-- Name: movies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movies (
    id integer NOT NULL,
    title character varying(500) NOT NULL,
    slug character varying(500) NOT NULL,
    overview text,
    poster_url text,
    backdrop_url text,
    trailer_url text,
    release_date date,
    runtime integer,
    content_rating character varying(20),
    original_language character varying(10),
    country character varying(100),
    imdb_id character varying(20),
    tmdb_id integer,
    user_avg_rating real DEFAULT 0,
    user_rating_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    imdb_rating numeric(3,1),
    rotten_tomatoes integer,
    metacritic_score integer,
    letterboxd_rating numeric(3,1),
    rating_sum bigint DEFAULT 0
);


--
-- Name: movies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movies_id_seq OWNED BY public.movies.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: persons; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.persons (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    biography text,
    photo_url text,
    birth_date date,
    death_date date,
    birthplace character varying(255),
    tmdb_id integer,
    imdb_id character varying(20),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: persons_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.persons_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: persons_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.persons_id_seq OWNED BY public.persons.id;


--
-- Name: ratings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ratings (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    movie_id integer NOT NULL,
    score integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ratings_score_check CHECK (((score >= 0) AND (score <= 10)))
);


--
-- Name: ratings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ratings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ratings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ratings_id_seq OWNED BY public.ratings.id;


--
-- Name: reactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reactions (
    id bigint NOT NULL,
    user_id integer NOT NULL,
    comment_id bigint NOT NULL,
    type public.reaction_type NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reactions_id_seq OWNED BY public.reactions.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id integer NOT NULL,
    refresh_token text,
    user_agent text,
    ip_address character varying(45),
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_status_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_status_logs (
    id integer NOT NULL,
    user_id integer NOT NULL,
    old_status public.user_status,
    new_status public.user_status NOT NULL,
    reason text,
    changed_by integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_status_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_status_logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_status_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_status_logs_id_seq OWNED BY public.user_status_logs.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    username character varying(50) NOT NULL,
    password_hash character varying(255),
    display_name character varying(100),
    avatar_url text,
    bio text,
    role public.role DEFAULT 'USER'::public.role NOT NULL,
    status public.user_status DEFAULT 'PENDING'::public.user_status NOT NULL,
    is_verified boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_password_hash_nonempty_chk CHECK (((password_hash IS NULL) OR (length((password_hash)::text) > 0)))
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: watchlists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.watchlists (
    id integer NOT NULL,
    user_id integer NOT NULL,
    movie_id integer NOT NULL,
    notes text,
    rank_position real,
    watched_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: watchlists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.watchlists_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: watchlists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.watchlists_id_seq OWNED BY public.watchlists.id;


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- Name: activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ALTER COLUMN id SET DEFAULT nextval('public.activities_id_seq'::regclass);


--
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- Name: credits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credits ALTER COLUMN id SET DEFAULT nextval('public.credits_id_seq'::regclass);


--
-- Name: follows id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows ALTER COLUMN id SET DEFAULT nextval('public.follows_id_seq'::regclass);


--
-- Name: genres id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres ALTER COLUMN id SET DEFAULT nextval('public.genres_id_seq'::regclass);


--
-- Name: movies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movies ALTER COLUMN id SET DEFAULT nextval('public.movies_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: persons id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.persons ALTER COLUMN id SET DEFAULT nextval('public.persons_id_seq'::regclass);


--
-- Name: ratings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ratings ALTER COLUMN id SET DEFAULT nextval('public.ratings_id_seq'::regclass);


--
-- Name: reactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions ALTER COLUMN id SET DEFAULT nextval('public.reactions_id_seq'::regclass);


--
-- Name: user_status_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_status_logs ALTER COLUMN id SET DEFAULT nextval('public.user_status_logs_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: watchlists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlists ALTER COLUMN id SET DEFAULT nextval('public.watchlists_id_seq'::regclass);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: accounts accounts_provider_provider_account_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_provider_provider_account_id_key UNIQUE (provider, provider_account_id);


--
-- Name: activities activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: credits credits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_pkey PRIMARY KEY (id);


--
-- Name: follows follows_follower_id_following_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_follower_id_following_id_key UNIQUE (follower_id, following_id);


--
-- Name: follows follows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_pkey PRIMARY KEY (id);


--
-- Name: genres genres_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_name_key UNIQUE (name);


--
-- Name: genres genres_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_pkey PRIMARY KEY (id);


--
-- Name: genres genres_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_slug_key UNIQUE (slug);


--
-- Name: genres genres_tmdb_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_tmdb_id_key UNIQUE (tmdb_id);


--
-- Name: movie_genres movie_genres_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genres
    ADD CONSTRAINT movie_genres_pkey PRIMARY KEY (movie_id, genre_id);


--
-- Name: movies movies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_pkey PRIMARY KEY (id);


--
-- Name: movies movies_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_slug_key UNIQUE (slug);


--
-- Name: movies movies_tmdb_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_tmdb_id_key UNIQUE (tmdb_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: persons persons_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_pkey PRIMARY KEY (id);


--
-- Name: persons persons_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_slug_key UNIQUE (slug);


--
-- Name: persons persons_tmdb_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_tmdb_id_key UNIQUE (tmdb_id);


--
-- Name: ratings ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_pkey PRIMARY KEY (id);


--
-- Name: ratings ratings_user_id_movie_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_user_id_movie_id_key UNIQUE (user_id, movie_id);


--
-- Name: reactions reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: user_status_logs user_status_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_status_logs
    ADD CONSTRAINT user_status_logs_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: watchlists watchlists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlists
    ADD CONSTRAINT watchlists_pkey PRIMARY KEY (id);


--
-- Name: watchlists watchlists_user_id_movie_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlists
    ADD CONSTRAINT watchlists_user_id_movie_id_key UNIQUE (user_id, movie_id);


--
-- Name: accounts_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX accounts_user_id_idx ON public.accounts USING btree (user_id);


--
-- Name: activities_entity_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_entity_idx ON public.activities USING btree (entity_type, entity_id);


--
-- Name: activities_user_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_user_created_idx ON public.activities USING btree (user_id, created_at DESC);


--
-- Name: comments_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_created_at_idx ON public.comments USING btree (created_at DESC);


--
-- Name: comments_deleted_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_deleted_idx ON public.comments USING btree (deleted_at) WHERE (deleted_at IS NOT NULL);


--
-- Name: comments_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_movie_id_idx ON public.comments USING btree (movie_id);


--
-- Name: comments_parent_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_parent_id_idx ON public.comments USING btree (parent_id);


--
-- Name: comments_user_created_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_user_created_active_idx ON public.comments USING btree (user_id, created_at DESC) WHERE (deleted_at IS NULL);


--
-- Name: credits_department_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX credits_department_idx ON public.credits USING btree (department);


--
-- Name: credits_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX credits_movie_id_idx ON public.credits USING btree (movie_id);


--
-- Name: credits_person_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX credits_person_id_idx ON public.credits USING btree (person_id);


--
-- Name: credits_unique_actor; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX credits_unique_actor ON public.credits USING btree (movie_id, person_id, role, "character") WHERE ("character" IS NOT NULL);


--
-- Name: credits_unique_non_actor; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX credits_unique_non_actor ON public.credits USING btree (movie_id, person_id, role) WHERE ("character" IS NULL);


--
-- Name: follows_follower_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX follows_follower_idx ON public.follows USING btree (follower_id);


--
-- Name: follows_following_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX follows_following_idx ON public.follows USING btree (following_id);


--
-- Name: movie_genres_genre_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movie_genres_genre_idx ON public.movie_genres USING btree (genre_id);


--
-- Name: movies_average_rating_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_average_rating_idx ON public.movies USING btree (user_avg_rating DESC);


--
-- Name: movies_release_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_release_date_idx ON public.movies USING btree (release_date DESC);


--
-- Name: movies_title_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_title_trgm_idx ON public.movies USING gin (title public.gin_trgm_ops);


--
-- Name: notifications_user_unread_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_user_unread_idx ON public.notifications USING btree (user_id, is_read, created_at DESC);


--
-- Name: persons_name_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX persons_name_trgm_idx ON public.persons USING gin (name public.gin_trgm_ops);


--
-- Name: ratings_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ratings_created_at_idx ON public.ratings USING btree (created_at DESC);


--
-- Name: ratings_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ratings_movie_id_idx ON public.ratings USING btree (movie_id);


--
-- Name: reactions_comment_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reactions_comment_id_idx ON public.reactions USING btree (comment_id);


--
-- Name: reactions_user_comment_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX reactions_user_comment_unique ON public.reactions USING btree (user_id, comment_id) WHERE (comment_id IS NOT NULL);


--
-- Name: sessions_expires_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_expires_at_idx ON public.sessions USING btree (expires_at);


--
-- Name: sessions_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_user_id_idx ON public.sessions USING btree (user_id);


--
-- Name: user_status_logs_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_status_logs_user_idx ON public.user_status_logs USING btree (user_id, created_at DESC);


--
-- Name: users_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_status_idx ON public.users USING btree (status);


--
-- Name: watchlists_user_rank_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX watchlists_user_rank_idx ON public.watchlists USING btree (user_id, rank_position);


--
-- Name: accounts accounts_hash_tokens_before_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER accounts_hash_tokens_before_write BEFORE INSERT OR UPDATE OF access_token, refresh_token ON public.accounts FOR EACH ROW EXECUTE FUNCTION public.hash_accounts_tokens();


--
-- Name: comments comments_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER comments_updated_at BEFORE UPDATE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: movies movies_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER movies_updated_at BEFORE UPDATE ON public.movies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: persons persons_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER persons_updated_at BEFORE UPDATE ON public.persons FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: ratings ratings_stats_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_stats_delete AFTER DELETE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_movie_rating_stats();


--
-- Name: ratings ratings_stats_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_stats_insert AFTER INSERT ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_movie_rating_stats();


--
-- Name: ratings ratings_stats_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_stats_update AFTER UPDATE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_movie_rating_stats();


--
-- Name: ratings ratings_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_updated_at BEFORE UPDATE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: reactions reactions_like_count_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER reactions_like_count_delete AFTER DELETE ON public.reactions FOR EACH ROW EXECUTE FUNCTION public.update_like_counts();


--
-- Name: reactions reactions_like_count_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER reactions_like_count_insert AFTER INSERT ON public.reactions FOR EACH ROW EXECUTE FUNCTION public.update_like_counts();


--
-- Name: reactions reactions_like_count_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER reactions_like_count_update AFTER UPDATE ON public.reactions FOR EACH ROW EXECUTE FUNCTION public.update_like_counts();


--
-- Name: sessions sessions_hash_refresh_token_before_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER sessions_hash_refresh_token_before_write BEFORE INSERT OR UPDATE OF refresh_token ON public.sessions FOR EACH ROW EXECUTE FUNCTION public.hash_sessions_refresh_token();


--
-- Name: users users_status_change; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER users_status_change AFTER UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.log_user_status_change();


--
-- Name: users users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: accounts accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: activities activities_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities
    ADD CONSTRAINT activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comments comments_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE;


--
-- Name: comments comments_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.comments(id) ON DELETE SET NULL;


--
-- Name: comments comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: credits credits_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE;


--
-- Name: credits credits_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.persons(id) ON DELETE CASCADE;


--
-- Name: follows follows_follower_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_follower_id_fkey FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: follows follows_following_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_following_id_fkey FOREIGN KEY (following_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: movie_genres movie_genres_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genres
    ADD CONSTRAINT movie_genres_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;


--
-- Name: movie_genres movie_genres_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genres
    ADD CONSTRAINT movie_genres_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: ratings ratings_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE;


--
-- Name: ratings ratings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reactions reactions_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: reactions reactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_status_logs user_status_logs_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_status_logs
    ADD CONSTRAINT user_status_logs_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_status_logs user_status_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_status_logs
    ADD CONSTRAINT user_status_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: watchlists watchlists_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlists
    ADD CONSTRAINT watchlists_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON DELETE CASCADE;


--
-- Name: watchlists watchlists_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlists
    ADD CONSTRAINT watchlists_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 1TcVJ5TngzjjxM1z2t2hqTfxy26yXS1GWnNXsnCe2wes4FJq1NkUJIzyix4S8fl

