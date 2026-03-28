--
-- PostgreSQL database dump
--


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
    'BANNED',
    'DELETED'
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
-- Name: calculate_landing_sort_score(real, numeric, integer, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_landing_sort_score(p_user_avg_rating real, p_imdb_rating numeric, p_rotten_tomatoes integer, p_metacritic_score integer) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
    community_rating NUMERIC;
    rating_count INT;
BEGIN
    rating_count := (CASE WHEN p_imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN p_rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN p_metacritic_score IS NOT NULL THEN 1 ELSE 0 END);

    IF rating_count > 0 THEN
        community_rating := (
            COALESCE(p_imdb_rating, 0) +
            COALESCE(p_rotten_tomatoes::NUMERIC / 10.0, 0) +
            COALESCE(p_metacritic_score::NUMERIC / 10.0, 0)
        ) / rating_count;
    END IF;

    IF p_user_avg_rating IS NOT NULL AND rating_count > 0 THEN
        RETURN (p_user_avg_rating::NUMERIC * 0.9) + (community_rating * 0.1);
    ELSIF p_user_avg_rating IS NOT NULL THEN
        RETURN p_user_avg_rating::NUMERIC;
    ELSIF rating_count > 0 THEN
        RETURN community_rating;
    ELSE
        RETURN 0;
    END IF;
END;
$$;


--
-- Name: update_movie_landing_sort_score(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_movie_landing_sort_score() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.landing_sort_score := calculate_landing_sort_score(
        NEW.user_avg_rating,
        NEW.imdb_rating,
        NEW.rotten_tomatoes,
        NEW.metacritic_score
    );
    RETURN NEW;
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
                ELSE NULL
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
                    ELSE NULL
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
                    ELSE NULL
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


--
-- Name: update_user_comment_count(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_user_comment_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Only count if not soft-deleted
        IF NEW.deleted_at IS NULL THEN
            UPDATE users SET comment_count = comment_count + 1
            WHERE id = NEW.user_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle soft delete transition
        IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
            -- Comment was soft-deleted
            UPDATE users SET comment_count = GREATEST(0, comment_count - 1)
            WHERE id = NEW.user_id;
        ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
            -- Comment was restored
            UPDATE users SET comment_count = comment_count + 1
            WHERE id = NEW.user_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Only decrement if wasn't already soft-deleted
        IF OLD.deleted_at IS NULL THEN
            UPDATE users SET comment_count = GREATEST(0, comment_count - 1)
            WHERE id = OLD.user_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


--
-- Name: update_user_follow_counts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_user_follow_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment follower_count for the user being followed
        UPDATE users SET follower_count = follower_count + 1
        WHERE id = NEW.following_id;
        -- Increment following_count for the user who follows
        UPDATE users SET following_count = following_count + 1
        WHERE id = NEW.follower_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement follower_count for the user being unfollowed
        UPDATE users SET follower_count = GREATEST(0, follower_count - 1)
        WHERE id = OLD.following_id;
        -- Decrement following_count for the user who unfollows
        UPDATE users SET following_count = GREATEST(0, following_count - 1)
        WHERE id = OLD.follower_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


--
-- Name: update_user_rating_count(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_user_rating_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE users SET rating_count = rating_count + 1
        WHERE id = NEW.user_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE users SET rating_count = GREATEST(0, rating_count - 1)
        WHERE id = OLD.user_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
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
-- Name: activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
)
PARTITION BY RANGE (created_at);


--
-- Name: activities_2026_01; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_01 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_02; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_02 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_03; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_03 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_04; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_04 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_05; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_05 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_06; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_06 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_07; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_07 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_08; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_08 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_09; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_09 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_10; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_10 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_11; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_11 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2026_12; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2026_12 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_01; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_01 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_02; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_02 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_03; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_03 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_04; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_04 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_05; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_05 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_06; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_06 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_07; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_07 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_08; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_08 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_09; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_09 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_10; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_10 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_11; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_11 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_2027_12; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_2027_12 (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: activities_default; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.activities_default (
    id bigint DEFAULT nextval('public.activities_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    action character varying(50) NOT NULL,
    entity_type public.entity_type NOT NULL,
    entity_id bigint NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id bigint NOT NULL,
    user_id integer,
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
    user_avg_rating real,
    user_rating_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    imdb_rating numeric(3,1),
    rotten_tomatoes integer,
    metacritic_score integer,
    letterboxd_rating numeric(3,1),
    rating_sum bigint DEFAULT 0,
    landing_sort_score numeric
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
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
)
PARTITION BY RANGE (created_at);


--
-- Name: notifications_2026_01; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_01 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_02; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_02 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_03; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_03 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_04; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_04 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_05; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_05 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_06; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_06 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_07; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_07 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_08; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_08 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_09; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_09 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_10; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_10 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_11; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_11 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2026_12; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2026_12 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_01; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_01 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_02; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_02 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_03; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_03 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_04; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_04 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_05; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_05 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_06; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_06 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_07; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_07 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_08; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_08 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_09; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_09 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_10; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_10 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_11; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_11 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_2027_12; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_2027_12 (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: notifications_default; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications_default (
    id bigint DEFAULT nextval('public.notifications_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    content text,
    is_read boolean DEFAULT false NOT NULL,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


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
    user_id integer,
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
    deleted_at timestamp with time zone,
    anonymized_at timestamp with time zone,
    follower_count integer DEFAULT 0 NOT NULL,
    following_count integer DEFAULT 0 NOT NULL,
    rating_count integer DEFAULT 0 NOT NULL,
    comment_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT users_anonymization_after_deletion_check CHECK (((anonymized_at IS NULL) OR (deleted_at IS NOT NULL))),
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
-- Name: activities_2026_01; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_01 FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');


--
-- Name: activities_2026_02; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_02 FOR VALUES FROM ('2026-02-01 00:00:00+00') TO ('2026-03-01 00:00:00+00');


--
-- Name: activities_2026_03; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_03 FOR VALUES FROM ('2026-03-01 00:00:00+00') TO ('2026-04-01 00:00:00+00');


--
-- Name: activities_2026_04; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_04 FOR VALUES FROM ('2026-04-01 00:00:00+00') TO ('2026-05-01 00:00:00+00');


--
-- Name: activities_2026_05; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_05 FOR VALUES FROM ('2026-05-01 00:00:00+00') TO ('2026-06-01 00:00:00+00');


--
-- Name: activities_2026_06; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_06 FOR VALUES FROM ('2026-06-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');


--
-- Name: activities_2026_07; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_07 FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00');


--
-- Name: activities_2026_08; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_08 FOR VALUES FROM ('2026-08-01 00:00:00+00') TO ('2026-09-01 00:00:00+00');


--
-- Name: activities_2026_09; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_09 FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');


--
-- Name: activities_2026_10; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_10 FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');


--
-- Name: activities_2026_11; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_11 FOR VALUES FROM ('2026-11-01 00:00:00+00') TO ('2026-12-01 00:00:00+00');


--
-- Name: activities_2026_12; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2026_12 FOR VALUES FROM ('2026-12-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');


--
-- Name: activities_2027_01; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_01 FOR VALUES FROM ('2027-01-01 00:00:00+00') TO ('2027-02-01 00:00:00+00');


--
-- Name: activities_2027_02; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_02 FOR VALUES FROM ('2027-02-01 00:00:00+00') TO ('2027-03-01 00:00:00+00');


--
-- Name: activities_2027_03; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_03 FOR VALUES FROM ('2027-03-01 00:00:00+00') TO ('2027-04-01 00:00:00+00');


--
-- Name: activities_2027_04; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_04 FOR VALUES FROM ('2027-04-01 00:00:00+00') TO ('2027-05-01 00:00:00+00');


--
-- Name: activities_2027_05; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_05 FOR VALUES FROM ('2027-05-01 00:00:00+00') TO ('2027-06-01 00:00:00+00');


--
-- Name: activities_2027_06; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_06 FOR VALUES FROM ('2027-06-01 00:00:00+00') TO ('2027-07-01 00:00:00+00');


--
-- Name: activities_2027_07; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_07 FOR VALUES FROM ('2027-07-01 00:00:00+00') TO ('2027-08-01 00:00:00+00');


--
-- Name: activities_2027_08; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_08 FOR VALUES FROM ('2027-08-01 00:00:00+00') TO ('2027-09-01 00:00:00+00');


--
-- Name: activities_2027_09; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_09 FOR VALUES FROM ('2027-09-01 00:00:00+00') TO ('2027-10-01 00:00:00+00');


--
-- Name: activities_2027_10; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_10 FOR VALUES FROM ('2027-10-01 00:00:00+00') TO ('2027-11-01 00:00:00+00');


--
-- Name: activities_2027_11; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_11 FOR VALUES FROM ('2027-11-01 00:00:00+00') TO ('2027-12-01 00:00:00+00');


--
-- Name: activities_2027_12; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_2027_12 FOR VALUES FROM ('2027-12-01 00:00:00+00') TO ('2028-01-01 00:00:00+00');


--
-- Name: activities_default; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities ATTACH PARTITION public.activities_default DEFAULT;


--
-- Name: notifications_2026_01; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_01 FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');


--
-- Name: notifications_2026_02; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_02 FOR VALUES FROM ('2026-02-01 00:00:00+00') TO ('2026-03-01 00:00:00+00');


--
-- Name: notifications_2026_03; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_03 FOR VALUES FROM ('2026-03-01 00:00:00+00') TO ('2026-04-01 00:00:00+00');


--
-- Name: notifications_2026_04; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_04 FOR VALUES FROM ('2026-04-01 00:00:00+00') TO ('2026-05-01 00:00:00+00');


--
-- Name: notifications_2026_05; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_05 FOR VALUES FROM ('2026-05-01 00:00:00+00') TO ('2026-06-01 00:00:00+00');


--
-- Name: notifications_2026_06; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_06 FOR VALUES FROM ('2026-06-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');


--
-- Name: notifications_2026_07; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_07 FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00');


--
-- Name: notifications_2026_08; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_08 FOR VALUES FROM ('2026-08-01 00:00:00+00') TO ('2026-09-01 00:00:00+00');


--
-- Name: notifications_2026_09; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_09 FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');


--
-- Name: notifications_2026_10; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_10 FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');


--
-- Name: notifications_2026_11; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_11 FOR VALUES FROM ('2026-11-01 00:00:00+00') TO ('2026-12-01 00:00:00+00');


--
-- Name: notifications_2026_12; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2026_12 FOR VALUES FROM ('2026-12-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');


--
-- Name: notifications_2027_01; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_01 FOR VALUES FROM ('2027-01-01 00:00:00+00') TO ('2027-02-01 00:00:00+00');


--
-- Name: notifications_2027_02; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_02 FOR VALUES FROM ('2027-02-01 00:00:00+00') TO ('2027-03-01 00:00:00+00');


--
-- Name: notifications_2027_03; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_03 FOR VALUES FROM ('2027-03-01 00:00:00+00') TO ('2027-04-01 00:00:00+00');


--
-- Name: notifications_2027_04; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_04 FOR VALUES FROM ('2027-04-01 00:00:00+00') TO ('2027-05-01 00:00:00+00');


--
-- Name: notifications_2027_05; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_05 FOR VALUES FROM ('2027-05-01 00:00:00+00') TO ('2027-06-01 00:00:00+00');


--
-- Name: notifications_2027_06; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_06 FOR VALUES FROM ('2027-06-01 00:00:00+00') TO ('2027-07-01 00:00:00+00');


--
-- Name: notifications_2027_07; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_07 FOR VALUES FROM ('2027-07-01 00:00:00+00') TO ('2027-08-01 00:00:00+00');


--
-- Name: notifications_2027_08; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_08 FOR VALUES FROM ('2027-08-01 00:00:00+00') TO ('2027-09-01 00:00:00+00');


--
-- Name: notifications_2027_09; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_09 FOR VALUES FROM ('2027-09-01 00:00:00+00') TO ('2027-10-01 00:00:00+00');


--
-- Name: notifications_2027_10; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_10 FOR VALUES FROM ('2027-10-01 00:00:00+00') TO ('2027-11-01 00:00:00+00');


--
-- Name: notifications_2027_11; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_11 FOR VALUES FROM ('2027-11-01 00:00:00+00') TO ('2027-12-01 00:00:00+00');


--
-- Name: notifications_2027_12; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_2027_12 FOR VALUES FROM ('2027-12-01 00:00:00+00') TO ('2028-01-01 00:00:00+00');


--
-- Name: notifications_default; Type: TABLE ATTACH; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ATTACH PARTITION public.notifications_default DEFAULT;


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


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
    ADD CONSTRAINT activities_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_01 activities_2026_01_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_01
    ADD CONSTRAINT activities_2026_01_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_02 activities_2026_02_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_02
    ADD CONSTRAINT activities_2026_02_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_03 activities_2026_03_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_03
    ADD CONSTRAINT activities_2026_03_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_04 activities_2026_04_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_04
    ADD CONSTRAINT activities_2026_04_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_05 activities_2026_05_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_05
    ADD CONSTRAINT activities_2026_05_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_06 activities_2026_06_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_06
    ADD CONSTRAINT activities_2026_06_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_07 activities_2026_07_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_07
    ADD CONSTRAINT activities_2026_07_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_08 activities_2026_08_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_08
    ADD CONSTRAINT activities_2026_08_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_09 activities_2026_09_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_09
    ADD CONSTRAINT activities_2026_09_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_10 activities_2026_10_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_10
    ADD CONSTRAINT activities_2026_10_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_11 activities_2026_11_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_11
    ADD CONSTRAINT activities_2026_11_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2026_12 activities_2026_12_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2026_12
    ADD CONSTRAINT activities_2026_12_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_01 activities_2027_01_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_01
    ADD CONSTRAINT activities_2027_01_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_02 activities_2027_02_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_02
    ADD CONSTRAINT activities_2027_02_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_03 activities_2027_03_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_03
    ADD CONSTRAINT activities_2027_03_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_04 activities_2027_04_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_04
    ADD CONSTRAINT activities_2027_04_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_05 activities_2027_05_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_05
    ADD CONSTRAINT activities_2027_05_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_06 activities_2027_06_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_06
    ADD CONSTRAINT activities_2027_06_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_07 activities_2027_07_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_07
    ADD CONSTRAINT activities_2027_07_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_08 activities_2027_08_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_08
    ADD CONSTRAINT activities_2027_08_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_09 activities_2027_09_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_09
    ADD CONSTRAINT activities_2027_09_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_10 activities_2027_10_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_10
    ADD CONSTRAINT activities_2027_10_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_11 activities_2027_11_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_11
    ADD CONSTRAINT activities_2027_11_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_2027_12 activities_2027_12_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_2027_12
    ADD CONSTRAINT activities_2027_12_pkey PRIMARY KEY (id, created_at);


--
-- Name: activities_default activities_default_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.activities_default
    ADD CONSTRAINT activities_default_pkey PRIMARY KEY (id, created_at);


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
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_01 notifications_2026_01_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_01
    ADD CONSTRAINT notifications_2026_01_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_02 notifications_2026_02_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_02
    ADD CONSTRAINT notifications_2026_02_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_03 notifications_2026_03_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_03
    ADD CONSTRAINT notifications_2026_03_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_04 notifications_2026_04_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_04
    ADD CONSTRAINT notifications_2026_04_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_05 notifications_2026_05_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_05
    ADD CONSTRAINT notifications_2026_05_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_06 notifications_2026_06_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_06
    ADD CONSTRAINT notifications_2026_06_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_07 notifications_2026_07_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_07
    ADD CONSTRAINT notifications_2026_07_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_08 notifications_2026_08_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_08
    ADD CONSTRAINT notifications_2026_08_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_09 notifications_2026_09_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_09
    ADD CONSTRAINT notifications_2026_09_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_10 notifications_2026_10_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_10
    ADD CONSTRAINT notifications_2026_10_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_11 notifications_2026_11_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_11
    ADD CONSTRAINT notifications_2026_11_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2026_12 notifications_2026_12_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2026_12
    ADD CONSTRAINT notifications_2026_12_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_01 notifications_2027_01_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_01
    ADD CONSTRAINT notifications_2027_01_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_02 notifications_2027_02_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_02
    ADD CONSTRAINT notifications_2027_02_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_03 notifications_2027_03_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_03
    ADD CONSTRAINT notifications_2027_03_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_04 notifications_2027_04_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_04
    ADD CONSTRAINT notifications_2027_04_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_05 notifications_2027_05_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_05
    ADD CONSTRAINT notifications_2027_05_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_06 notifications_2027_06_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_06
    ADD CONSTRAINT notifications_2027_06_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_07 notifications_2027_07_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_07
    ADD CONSTRAINT notifications_2027_07_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_08 notifications_2027_08_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_08
    ADD CONSTRAINT notifications_2027_08_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_09 notifications_2027_09_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_09
    ADD CONSTRAINT notifications_2027_09_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_10 notifications_2027_10_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_10
    ADD CONSTRAINT notifications_2027_10_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_11 notifications_2027_11_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_11
    ADD CONSTRAINT notifications_2027_11_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_2027_12 notifications_2027_12_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_2027_12
    ADD CONSTRAINT notifications_2027_12_pkey PRIMARY KEY (id, created_at);


--
-- Name: notifications_default notifications_default_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications_default
    ADD CONSTRAINT notifications_default_pkey PRIMARY KEY (id, created_at);


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

CREATE INDEX activities_entity_idx ON ONLY public.activities USING btree (entity_type, entity_id);


--
-- Name: activities_2026_01_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_01_entity_type_entity_id_idx ON public.activities_2026_01 USING btree (entity_type, entity_id);


--
-- Name: activities_user_created_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_user_created_idx ON ONLY public.activities USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_01_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_01_user_id_created_at_idx ON public.activities_2026_01 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_02_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_02_entity_type_entity_id_idx ON public.activities_2026_02 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_02_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_02_user_id_created_at_idx ON public.activities_2026_02 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_03_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_03_entity_type_entity_id_idx ON public.activities_2026_03 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_03_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_03_user_id_created_at_idx ON public.activities_2026_03 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_04_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_04_entity_type_entity_id_idx ON public.activities_2026_04 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_04_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_04_user_id_created_at_idx ON public.activities_2026_04 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_05_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_05_entity_type_entity_id_idx ON public.activities_2026_05 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_05_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_05_user_id_created_at_idx ON public.activities_2026_05 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_06_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_06_entity_type_entity_id_idx ON public.activities_2026_06 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_06_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_06_user_id_created_at_idx ON public.activities_2026_06 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_07_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_07_entity_type_entity_id_idx ON public.activities_2026_07 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_07_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_07_user_id_created_at_idx ON public.activities_2026_07 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_08_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_08_entity_type_entity_id_idx ON public.activities_2026_08 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_08_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_08_user_id_created_at_idx ON public.activities_2026_08 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_09_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_09_entity_type_entity_id_idx ON public.activities_2026_09 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_09_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_09_user_id_created_at_idx ON public.activities_2026_09 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_10_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_10_entity_type_entity_id_idx ON public.activities_2026_10 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_10_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_10_user_id_created_at_idx ON public.activities_2026_10 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_11_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_11_entity_type_entity_id_idx ON public.activities_2026_11 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_11_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_11_user_id_created_at_idx ON public.activities_2026_11 USING btree (user_id, created_at DESC);


--
-- Name: activities_2026_12_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_12_entity_type_entity_id_idx ON public.activities_2026_12 USING btree (entity_type, entity_id);


--
-- Name: activities_2026_12_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2026_12_user_id_created_at_idx ON public.activities_2026_12 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_01_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_01_entity_type_entity_id_idx ON public.activities_2027_01 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_01_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_01_user_id_created_at_idx ON public.activities_2027_01 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_02_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_02_entity_type_entity_id_idx ON public.activities_2027_02 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_02_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_02_user_id_created_at_idx ON public.activities_2027_02 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_03_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_03_entity_type_entity_id_idx ON public.activities_2027_03 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_03_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_03_user_id_created_at_idx ON public.activities_2027_03 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_04_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_04_entity_type_entity_id_idx ON public.activities_2027_04 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_04_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_04_user_id_created_at_idx ON public.activities_2027_04 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_05_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_05_entity_type_entity_id_idx ON public.activities_2027_05 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_05_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_05_user_id_created_at_idx ON public.activities_2027_05 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_06_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_06_entity_type_entity_id_idx ON public.activities_2027_06 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_06_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_06_user_id_created_at_idx ON public.activities_2027_06 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_07_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_07_entity_type_entity_id_idx ON public.activities_2027_07 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_07_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_07_user_id_created_at_idx ON public.activities_2027_07 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_08_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_08_entity_type_entity_id_idx ON public.activities_2027_08 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_08_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_08_user_id_created_at_idx ON public.activities_2027_08 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_09_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_09_entity_type_entity_id_idx ON public.activities_2027_09 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_09_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_09_user_id_created_at_idx ON public.activities_2027_09 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_10_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_10_entity_type_entity_id_idx ON public.activities_2027_10 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_10_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_10_user_id_created_at_idx ON public.activities_2027_10 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_11_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_11_entity_type_entity_id_idx ON public.activities_2027_11 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_11_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_11_user_id_created_at_idx ON public.activities_2027_11 USING btree (user_id, created_at DESC);


--
-- Name: activities_2027_12_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_12_entity_type_entity_id_idx ON public.activities_2027_12 USING btree (entity_type, entity_id);


--
-- Name: activities_2027_12_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_2027_12_user_id_created_at_idx ON public.activities_2027_12 USING btree (user_id, created_at DESC);


--
-- Name: activities_default_entity_type_entity_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_default_entity_type_entity_id_idx ON public.activities_default USING btree (entity_type, entity_id);


--
-- Name: activities_default_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX activities_default_user_id_created_at_idx ON public.activities_default USING btree (user_id, created_at DESC);


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
-- Name: comments_user_id_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comments_user_id_active_idx ON public.comments USING btree (user_id) WHERE (user_id IS NOT NULL);


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
-- Name: movies_landing_sort_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_landing_sort_idx ON public.movies USING btree (landing_sort_score DESC NULLS LAST, created_at DESC);


--
-- Name: movies_release_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_release_date_idx ON public.movies USING btree (release_date DESC);


--
-- Name: movies_release_year_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_release_year_idx ON public.movies USING btree (EXTRACT(year FROM release_date));


--
-- Name: movies_title_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movies_title_trgm_idx ON public.movies USING gin (title public.gin_trgm_ops);


--
-- Name: notifications_user_unread_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_user_unread_idx ON ONLY public.notifications USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_01_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_01_user_id_is_read_created_at_idx ON public.notifications_2026_01 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_02_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_02_user_id_is_read_created_at_idx ON public.notifications_2026_02 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_03_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_03_user_id_is_read_created_at_idx ON public.notifications_2026_03 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_04_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_04_user_id_is_read_created_at_idx ON public.notifications_2026_04 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_05_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_05_user_id_is_read_created_at_idx ON public.notifications_2026_05 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_06_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_06_user_id_is_read_created_at_idx ON public.notifications_2026_06 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_07_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_07_user_id_is_read_created_at_idx ON public.notifications_2026_07 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_08_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_08_user_id_is_read_created_at_idx ON public.notifications_2026_08 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_09_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_09_user_id_is_read_created_at_idx ON public.notifications_2026_09 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_10_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_10_user_id_is_read_created_at_idx ON public.notifications_2026_10 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_11_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_11_user_id_is_read_created_at_idx ON public.notifications_2026_11 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2026_12_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2026_12_user_id_is_read_created_at_idx ON public.notifications_2026_12 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_01_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_01_user_id_is_read_created_at_idx ON public.notifications_2027_01 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_02_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_02_user_id_is_read_created_at_idx ON public.notifications_2027_02 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_03_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_03_user_id_is_read_created_at_idx ON public.notifications_2027_03 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_04_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_04_user_id_is_read_created_at_idx ON public.notifications_2027_04 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_05_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_05_user_id_is_read_created_at_idx ON public.notifications_2027_05 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_06_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_06_user_id_is_read_created_at_idx ON public.notifications_2027_06 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_07_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_07_user_id_is_read_created_at_idx ON public.notifications_2027_07 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_08_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_08_user_id_is_read_created_at_idx ON public.notifications_2027_08 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_09_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_09_user_id_is_read_created_at_idx ON public.notifications_2027_09 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_10_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_10_user_id_is_read_created_at_idx ON public.notifications_2027_10 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_11_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_11_user_id_is_read_created_at_idx ON public.notifications_2027_11 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_2027_12_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_2027_12_user_id_is_read_created_at_idx ON public.notifications_2027_12 USING btree (user_id, is_read, created_at DESC);


--
-- Name: notifications_default_user_id_is_read_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_default_user_id_is_read_created_at_idx ON public.notifications_default USING btree (user_id, is_read, created_at DESC);


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
-- Name: ratings_user_movie_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ratings_user_movie_unique_idx ON public.ratings USING btree (user_id, movie_id) WHERE (user_id IS NOT NULL);


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
-- Name: sessions_refresh_token_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_refresh_token_idx ON public.sessions USING btree (refresh_token);


--
-- Name: sessions_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_user_id_idx ON public.sessions USING btree (user_id);


--
-- Name: user_status_logs_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_status_logs_user_idx ON public.user_status_logs USING btree (user_id, created_at DESC);


--
-- Name: users_active_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_active_idx ON public.users USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: users_deleted_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_deleted_at_idx ON public.users USING btree (deleted_at) WHERE (deleted_at IS NOT NULL);


--
-- Name: users_id_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_id_status_idx ON public.users USING btree (id) WHERE (status = 'ACTIVE'::public.user_status);


--
-- Name: users_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_status_idx ON public.users USING btree (status);


--
-- Name: watchlists_user_rank_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX watchlists_user_rank_idx ON public.watchlists USING btree (user_id, rank_position);


--
-- Name: activities_2026_01_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_01_entity_type_entity_id_idx;


--
-- Name: activities_2026_01_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_01_pkey;


--
-- Name: activities_2026_01_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_01_user_id_created_at_idx;


--
-- Name: activities_2026_02_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_02_entity_type_entity_id_idx;


--
-- Name: activities_2026_02_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_02_pkey;


--
-- Name: activities_2026_02_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_02_user_id_created_at_idx;


--
-- Name: activities_2026_03_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_03_entity_type_entity_id_idx;


--
-- Name: activities_2026_03_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_03_pkey;


--
-- Name: activities_2026_03_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_03_user_id_created_at_idx;


--
-- Name: activities_2026_04_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_04_entity_type_entity_id_idx;


--
-- Name: activities_2026_04_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_04_pkey;


--
-- Name: activities_2026_04_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_04_user_id_created_at_idx;


--
-- Name: activities_2026_05_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_05_entity_type_entity_id_idx;


--
-- Name: activities_2026_05_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_05_pkey;


--
-- Name: activities_2026_05_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_05_user_id_created_at_idx;


--
-- Name: activities_2026_06_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_06_entity_type_entity_id_idx;


--
-- Name: activities_2026_06_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_06_pkey;


--
-- Name: activities_2026_06_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_06_user_id_created_at_idx;


--
-- Name: activities_2026_07_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_07_entity_type_entity_id_idx;


--
-- Name: activities_2026_07_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_07_pkey;


--
-- Name: activities_2026_07_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_07_user_id_created_at_idx;


--
-- Name: activities_2026_08_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_08_entity_type_entity_id_idx;


--
-- Name: activities_2026_08_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_08_pkey;


--
-- Name: activities_2026_08_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_08_user_id_created_at_idx;


--
-- Name: activities_2026_09_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_09_entity_type_entity_id_idx;


--
-- Name: activities_2026_09_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_09_pkey;


--
-- Name: activities_2026_09_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_09_user_id_created_at_idx;


--
-- Name: activities_2026_10_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_10_entity_type_entity_id_idx;


--
-- Name: activities_2026_10_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_10_pkey;


--
-- Name: activities_2026_10_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_10_user_id_created_at_idx;


--
-- Name: activities_2026_11_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_11_entity_type_entity_id_idx;


--
-- Name: activities_2026_11_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_11_pkey;


--
-- Name: activities_2026_11_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_11_user_id_created_at_idx;


--
-- Name: activities_2026_12_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2026_12_entity_type_entity_id_idx;


--
-- Name: activities_2026_12_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2026_12_pkey;


--
-- Name: activities_2026_12_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2026_12_user_id_created_at_idx;


--
-- Name: activities_2027_01_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_01_entity_type_entity_id_idx;


--
-- Name: activities_2027_01_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_01_pkey;


--
-- Name: activities_2027_01_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_01_user_id_created_at_idx;


--
-- Name: activities_2027_02_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_02_entity_type_entity_id_idx;


--
-- Name: activities_2027_02_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_02_pkey;


--
-- Name: activities_2027_02_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_02_user_id_created_at_idx;


--
-- Name: activities_2027_03_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_03_entity_type_entity_id_idx;


--
-- Name: activities_2027_03_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_03_pkey;


--
-- Name: activities_2027_03_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_03_user_id_created_at_idx;


--
-- Name: activities_2027_04_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_04_entity_type_entity_id_idx;


--
-- Name: activities_2027_04_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_04_pkey;


--
-- Name: activities_2027_04_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_04_user_id_created_at_idx;


--
-- Name: activities_2027_05_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_05_entity_type_entity_id_idx;


--
-- Name: activities_2027_05_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_05_pkey;


--
-- Name: activities_2027_05_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_05_user_id_created_at_idx;


--
-- Name: activities_2027_06_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_06_entity_type_entity_id_idx;


--
-- Name: activities_2027_06_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_06_pkey;


--
-- Name: activities_2027_06_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_06_user_id_created_at_idx;


--
-- Name: activities_2027_07_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_07_entity_type_entity_id_idx;


--
-- Name: activities_2027_07_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_07_pkey;


--
-- Name: activities_2027_07_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_07_user_id_created_at_idx;


--
-- Name: activities_2027_08_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_08_entity_type_entity_id_idx;


--
-- Name: activities_2027_08_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_08_pkey;


--
-- Name: activities_2027_08_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_08_user_id_created_at_idx;


--
-- Name: activities_2027_09_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_09_entity_type_entity_id_idx;


--
-- Name: activities_2027_09_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_09_pkey;


--
-- Name: activities_2027_09_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_09_user_id_created_at_idx;


--
-- Name: activities_2027_10_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_10_entity_type_entity_id_idx;


--
-- Name: activities_2027_10_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_10_pkey;


--
-- Name: activities_2027_10_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_10_user_id_created_at_idx;


--
-- Name: activities_2027_11_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_11_entity_type_entity_id_idx;


--
-- Name: activities_2027_11_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_11_pkey;


--
-- Name: activities_2027_11_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_11_user_id_created_at_idx;


--
-- Name: activities_2027_12_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_2027_12_entity_type_entity_id_idx;


--
-- Name: activities_2027_12_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_2027_12_pkey;


--
-- Name: activities_2027_12_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_2027_12_user_id_created_at_idx;


--
-- Name: activities_default_entity_type_entity_id_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_entity_idx ATTACH PARTITION public.activities_default_entity_type_entity_id_idx;


--
-- Name: activities_default_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_pkey ATTACH PARTITION public.activities_default_pkey;


--
-- Name: activities_default_user_id_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.activities_user_created_idx ATTACH PARTITION public.activities_default_user_id_created_at_idx;


--
-- Name: notifications_2026_01_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_01_pkey;


--
-- Name: notifications_2026_01_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_01_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_02_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_02_pkey;


--
-- Name: notifications_2026_02_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_02_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_03_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_03_pkey;


--
-- Name: notifications_2026_03_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_03_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_04_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_04_pkey;


--
-- Name: notifications_2026_04_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_04_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_05_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_05_pkey;


--
-- Name: notifications_2026_05_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_05_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_06_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_06_pkey;


--
-- Name: notifications_2026_06_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_06_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_07_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_07_pkey;


--
-- Name: notifications_2026_07_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_07_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_08_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_08_pkey;


--
-- Name: notifications_2026_08_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_08_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_09_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_09_pkey;


--
-- Name: notifications_2026_09_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_09_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_10_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_10_pkey;


--
-- Name: notifications_2026_10_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_10_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_11_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_11_pkey;


--
-- Name: notifications_2026_11_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_11_user_id_is_read_created_at_idx;


--
-- Name: notifications_2026_12_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2026_12_pkey;


--
-- Name: notifications_2026_12_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2026_12_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_01_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_01_pkey;


--
-- Name: notifications_2027_01_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_01_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_02_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_02_pkey;


--
-- Name: notifications_2027_02_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_02_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_03_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_03_pkey;


--
-- Name: notifications_2027_03_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_03_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_04_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_04_pkey;


--
-- Name: notifications_2027_04_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_04_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_05_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_05_pkey;


--
-- Name: notifications_2027_05_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_05_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_06_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_06_pkey;


--
-- Name: notifications_2027_06_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_06_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_07_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_07_pkey;


--
-- Name: notifications_2027_07_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_07_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_08_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_08_pkey;


--
-- Name: notifications_2027_08_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_08_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_09_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_09_pkey;


--
-- Name: notifications_2027_09_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_09_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_10_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_10_pkey;


--
-- Name: notifications_2027_10_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_10_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_11_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_11_pkey;


--
-- Name: notifications_2027_11_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_11_user_id_is_read_created_at_idx;


--
-- Name: notifications_2027_12_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_2027_12_pkey;


--
-- Name: notifications_2027_12_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_2027_12_user_id_is_read_created_at_idx;


--
-- Name: notifications_default_pkey; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_pkey ATTACH PARTITION public.notifications_default_pkey;


--
-- Name: notifications_default_user_id_is_read_created_at_idx; Type: INDEX ATTACH; Schema: public; Owner: -
--

ALTER INDEX public.notifications_user_unread_idx ATTACH PARTITION public.notifications_default_user_id_is_read_created_at_idx;


--
-- Name: accounts accounts_hash_tokens_before_write; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER accounts_hash_tokens_before_write BEFORE INSERT OR UPDATE OF access_token, refresh_token ON public.accounts FOR EACH ROW EXECUTE FUNCTION public.hash_accounts_tokens();


--
-- Name: comments comments_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER comments_updated_at BEFORE UPDATE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: comments comments_user_count_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER comments_user_count_delete AFTER DELETE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_user_comment_count();


--
-- Name: comments comments_user_count_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER comments_user_count_insert AFTER INSERT ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_user_comment_count();


--
-- Name: comments comments_user_count_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER comments_user_count_update AFTER UPDATE OF deleted_at ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_user_comment_count();


--
-- Name: follows follows_user_counts_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER follows_user_counts_delete AFTER DELETE ON public.follows FOR EACH ROW EXECUTE FUNCTION public.update_user_follow_counts();


--
-- Name: follows follows_user_counts_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER follows_user_counts_insert AFTER INSERT ON public.follows FOR EACH ROW EXECUTE FUNCTION public.update_user_follow_counts();


--
-- Name: movies movies_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER movies_updated_at BEFORE UPDATE ON public.movies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: movies trg_movies_landing_sort_score; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_movies_landing_sort_score BEFORE INSERT OR UPDATE OF user_avg_rating, imdb_rating, rotten_tomatoes, metacritic_score ON public.movies FOR EACH ROW EXECUTE FUNCTION public.update_movie_landing_sort_score();


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
-- Name: ratings ratings_user_count_delete; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_user_count_delete AFTER DELETE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_user_rating_count();


--
-- Name: ratings ratings_user_count_insert; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ratings_user_count_insert AFTER INSERT ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_user_rating_count();


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

ALTER TABLE public.activities
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
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


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

ALTER TABLE public.notifications
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
    ADD CONSTRAINT ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


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

