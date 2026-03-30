DROP TRIGGER IF EXISTS ratings_stats_after_change ON ratings;

CREATE TRIGGER ratings_stats_delete
    AFTER DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_insert
    AFTER INSERT ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();

CREATE TRIGGER ratings_stats_update
    AFTER UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION update_movie_rating_stats();
