-- Consolidate rating stats maintenance into a single trigger.
-- This keeps the trigger definition aligned with the multi-operation
-- logic already handled inside update_movie_rating_stats().

DROP TRIGGER IF EXISTS ratings_stats_delete ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_insert ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_update ON ratings;
DROP TRIGGER IF EXISTS ratings_stats_after_change ON ratings;

CREATE TRIGGER ratings_stats_after_change
    AFTER INSERT OR UPDATE OR DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();
