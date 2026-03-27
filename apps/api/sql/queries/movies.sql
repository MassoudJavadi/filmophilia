-- ============================================================
-- PERSONS QUERIES
-- ============================================================

-- name: UpsertPerson :one
-- Insert a person or update their name if the slug already exists
INSERT INTO persons (name, slug)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- ============================================================
-- MOVIES QUERIES
-- ============================================================

-- name: CreateMovie :one
-- Create a new movie record with external ratings
INSERT INTO movies (
    title, slug, overview, poster_url, release_date, runtime, 
    imdb_id, tmdb_id, imdb_rating, rotten_tomatoes, metacritic_score
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: ListMovies :many
-- Get a paginated list of movies for landing-page cards, ordered by weighted rating
WITH scored_movies AS (
    SELECT
        m.id,
        m.title,
        m.slug,
        m.poster_url,
        m.release_date,
        m.user_avg_rating,
        m.user_rating_count,
        m.imdb_rating,
        m.rotten_tomatoes,
        m.metacritic_score,
        m.created_at,
        CASE
            WHEN
                (CASE WHEN m.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN m.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN m.metacritic_score IS NOT NULL THEN 1 ELSE 0 END) > 0
            THEN (
                COALESCE(m.imdb_rating::NUMERIC, 0) +
                COALESCE((m.rotten_tomatoes::NUMERIC / 10.0), 0) +
                COALESCE((m.metacritic_score::NUMERIC / 10.0), 0)
            ) / (
                (CASE WHEN m.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN m.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN m.metacritic_score IS NOT NULL THEN 1 ELSE 0 END)
            )::NUMERIC
            ELSE NULL
        END AS community_rating,
        CASE
            WHEN m.user_rating_count > 0
                 AND (
                    m.imdb_rating IS NOT NULL OR
                    m.rotten_tomatoes IS NOT NULL OR
                    m.metacritic_score IS NOT NULL
                 )
                THEN (m.user_avg_rating::NUMERIC * 0.7) + (
                    (
                        COALESCE(m.imdb_rating::NUMERIC, 0) +
                        COALESCE((m.rotten_tomatoes::NUMERIC / 10.0), 0) +
                        COALESCE((m.metacritic_score::NUMERIC / 10.0), 0)
                    ) / (
                        (CASE WHEN m.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                        (CASE WHEN m.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                        (CASE WHEN m.metacritic_score IS NOT NULL THEN 1 ELSE 0 END)
                    )::NUMERIC
                ) * 0.3
            WHEN m.user_rating_count > 0
                THEN m.user_avg_rating::NUMERIC
            WHEN
                m.imdb_rating IS NOT NULL OR
                m.rotten_tomatoes IS NOT NULL OR
                m.metacritic_score IS NOT NULL
                THEN (
                    COALESCE(m.imdb_rating::NUMERIC, 0) +
                    COALESCE((m.rotten_tomatoes::NUMERIC / 10.0), 0) +
                    COALESCE((m.metacritic_score::NUMERIC / 10.0), 0)
                ) / (
                    (CASE WHEN m.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN m.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN m.metacritic_score IS NOT NULL THEN 1 ELSE 0 END)
                )::NUMERIC
            ELSE 0
        END AS landing_sort_score
    FROM movies m
), paged_movies AS (
    SELECT
        sm.id,
        sm.title,
        sm.slug,
        sm.poster_url,
        sm.release_date,
        sm.user_avg_rating,
        sm.user_rating_count,
        sm.imdb_rating,
        sm.rotten_tomatoes,
        sm.metacritic_score,
        sm.created_at
    FROM scored_movies sm
    ORDER BY sm.landing_sort_score DESC, sm.created_at DESC
    LIMIT $1 OFFSET $2
)
SELECT
    pm.id,
    pm.title,
    pm.slug,
    pm.poster_url,
    pm.release_date,
    pm.user_avg_rating,
    pm.user_rating_count,
    pm.imdb_rating,
    pm.rotten_tomatoes,
    pm.metacritic_score,
    COALESCE(
        (
            ARRAY_AGG(DISTINCT p.name ORDER BY p.name)
            FILTER (WHERE p.name IS NOT NULL)
        )::TEXT[],
        ARRAY[]::TEXT[]
    )::TEXT[] AS directors
FROM paged_movies pm
LEFT JOIN credits c
    ON c.movie_id = pm.id
    AND c.department = 'DIRECTING'
LEFT JOIN persons p ON p.id = c.person_id
GROUP BY
    pm.id,
    pm.title,
    pm.slug,
    pm.poster_url,
    pm.release_date,
    pm.user_avg_rating,
    pm.user_rating_count,
    pm.imdb_rating,
    pm.rotten_tomatoes,
    pm.metacritic_score,
    pm.created_at
ORDER BY
    CASE
        WHEN pm.user_rating_count > 0
             AND (
                pm.imdb_rating IS NOT NULL OR
                pm.rotten_tomatoes IS NOT NULL OR
                pm.metacritic_score IS NOT NULL
             )
            THEN (pm.user_avg_rating::NUMERIC * 0.7) + (
                (
                    COALESCE(pm.imdb_rating::NUMERIC, 0) +
                    COALESCE((pm.rotten_tomatoes::NUMERIC / 10.0), 0) +
                    COALESCE((pm.metacritic_score::NUMERIC / 10.0), 0)
                ) / (
                    (CASE WHEN pm.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN pm.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                    (CASE WHEN pm.metacritic_score IS NOT NULL THEN 1 ELSE 0 END)
                )::NUMERIC
            ) * 0.3
        WHEN pm.user_rating_count > 0
            THEN pm.user_avg_rating::NUMERIC
        WHEN
            pm.imdb_rating IS NOT NULL OR
            pm.rotten_tomatoes IS NOT NULL OR
            pm.metacritic_score IS NOT NULL
            THEN (
                COALESCE(pm.imdb_rating::NUMERIC, 0) +
                COALESCE((pm.rotten_tomatoes::NUMERIC / 10.0), 0) +
                COALESCE((pm.metacritic_score::NUMERIC / 10.0), 0)
            ) / (
                (CASE WHEN pm.imdb_rating IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN pm.rotten_tomatoes IS NOT NULL THEN 1 ELSE 0 END) +
                (CASE WHEN pm.metacritic_score IS NOT NULL THEN 1 ELSE 0 END)
            )::NUMERIC
        ELSE 0
    END DESC,
    pm.created_at DESC;

-- name: GetMovieBySlug :one
-- Get a single movie by its unique slug
SELECT * FROM movies
WHERE slug = $1 LIMIT 1;

-- name: SearchMovies :many
-- Simple search by title or slug
SELECT * FROM movies
WHERE title ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%'
ORDER BY imdb_rating DESC
LIMIT $2 OFFSET $3;

-- name: AdvancedSearchMovies :many
-- Advanced search with filters: query, genres, year range, rating range, sort
SELECT DISTINCT m.*
FROM movies m
LEFT JOIN movie_genres mg ON mg.movie_id = m.id
WHERE
    -- Text search (optional)
    (sqlc.narg(query)::TEXT IS NULL OR m.title ILIKE '%' || sqlc.narg(query)::TEXT || '%')
    -- Genre filter (optional, array of genre IDs)
    AND (sqlc.narg(genre_ids)::INT[] IS NULL OR mg.genre_id = ANY(sqlc.narg(genre_ids)::INT[]))
    -- Year range (optional)
    AND (sqlc.narg(year_from)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) >= sqlc.narg(year_from)::INT)
    AND (sqlc.narg(year_to)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) <= sqlc.narg(year_to)::INT)
    -- IMDB rating range (optional, scale 0-10)
    AND (sqlc.narg(imdb_min)::NUMERIC IS NULL OR m.imdb_rating >= sqlc.narg(imdb_min)::NUMERIC)
    AND (sqlc.narg(imdb_max)::NUMERIC IS NULL OR m.imdb_rating <= sqlc.narg(imdb_max)::NUMERIC)
    -- User rating range (optional, scale 1-10)
    AND (sqlc.narg(user_rating_min)::REAL IS NULL OR m.user_avg_rating >= sqlc.narg(user_rating_min)::REAL)
    AND (sqlc.narg(user_rating_max)::REAL IS NULL OR m.user_avg_rating <= sqlc.narg(user_rating_max)::REAL)
    -- Rotten Tomatoes range (optional, scale 0-100)
    AND (sqlc.narg(rt_min)::INT IS NULL OR m.rotten_tomatoes >= sqlc.narg(rt_min)::INT)
    AND (sqlc.narg(rt_max)::INT IS NULL OR m.rotten_tomatoes <= sqlc.narg(rt_max)::INT)
    -- Metacritic range (optional, scale 0-100)
    AND (sqlc.narg(metacritic_min)::INT IS NULL OR m.metacritic_score >= sqlc.narg(metacritic_min)::INT)
    AND (sqlc.narg(metacritic_max)::INT IS NULL OR m.metacritic_score <= sqlc.narg(metacritic_max)::INT)
ORDER BY
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'title_asc' THEN m.title END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'title_desc' THEN m.title END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'release_date_asc' THEN m.release_date END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'release_date_desc' THEN m.release_date END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'imdb_rating_asc' THEN m.imdb_rating END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'imdb_rating_desc' THEN m.imdb_rating END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'user_rating_asc' THEN m.user_avg_rating END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'user_rating_desc' THEN m.user_avg_rating END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'rotten_tomatoes_asc' THEN m.rotten_tomatoes END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'rotten_tomatoes_desc' THEN m.rotten_tomatoes END DESC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'metacritic_asc' THEN m.metacritic_score END ASC,
    CASE WHEN sqlc.narg(sort_by)::TEXT = 'metacritic_desc' THEN m.metacritic_score END DESC,
    m.imdb_rating DESC NULLS LAST,
    m.release_date DESC NULLS LAST
LIMIT $1 OFFSET $2;

-- name: CountAdvancedSearchMovies :one
-- Count total results for advanced search (for pagination)
SELECT COUNT(DISTINCT m.id)::INT as total
FROM movies m
LEFT JOIN movie_genres mg ON mg.movie_id = m.id
WHERE
    (sqlc.narg(query)::TEXT IS NULL OR m.title ILIKE '%' || sqlc.narg(query)::TEXT || '%')
    AND (sqlc.narg(genre_ids)::INT[] IS NULL OR mg.genre_id = ANY(sqlc.narg(genre_ids)::INT[]))
    AND (sqlc.narg(year_from)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) >= sqlc.narg(year_from)::INT)
    AND (sqlc.narg(year_to)::INT IS NULL OR EXTRACT(YEAR FROM m.release_date) <= sqlc.narg(year_to)::INT)
    AND (sqlc.narg(imdb_min)::NUMERIC IS NULL OR m.imdb_rating >= sqlc.narg(imdb_min)::NUMERIC)
    AND (sqlc.narg(imdb_max)::NUMERIC IS NULL OR m.imdb_rating <= sqlc.narg(imdb_max)::NUMERIC)
    AND (sqlc.narg(user_rating_min)::REAL IS NULL OR m.user_avg_rating >= sqlc.narg(user_rating_min)::REAL)
    AND (sqlc.narg(user_rating_max)::REAL IS NULL OR m.user_avg_rating <= sqlc.narg(user_rating_max)::REAL)
    AND (sqlc.narg(rt_min)::INT IS NULL OR m.rotten_tomatoes >= sqlc.narg(rt_min)::INT)
    AND (sqlc.narg(rt_max)::INT IS NULL OR m.rotten_tomatoes <= sqlc.narg(rt_max)::INT)
    AND (sqlc.narg(metacritic_min)::INT IS NULL OR m.metacritic_score >= sqlc.narg(metacritic_min)::INT)
    AND (sqlc.narg(metacritic_max)::INT IS NULL OR m.metacritic_score <= sqlc.narg(metacritic_max)::INT);

-- ============================================================
-- CREDITS QUERIES
-- ============================================================

-- name: CreateCredit :exec
-- Create a link between a movie and a person (cast/crew)
INSERT INTO credits (movie_id, person_id, department, role, character)
VALUES ($1, $2, $3, $4, $5);

-- name: UpsertCredit :exec
-- Create a credit or ignore if it already exists
INSERT INTO credits (movie_id, person_id, department, role, character, "order")
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;

-- name: GetMovieByTmdbID :one
-- Get a movie by its TMDB ID
SELECT * FROM movies WHERE tmdb_id = $1 LIMIT 1;
