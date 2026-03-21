-- ============================================================
-- ROLLBACK: Restore original trigger behavior
-- ============================================================

-- Drop triggers
DROP TRIGGER IF EXISTS reactions_like_count_insert ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_update ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_delete ON reactions;

-- Restore original function (counts all reactions, includes review references)
-- Note: This will fail if reviews table doesn't exist
CREATE OR REPLACE FUNCTION update_like_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE comment_id = OLD.comment_id
            ) WHERE id = OLD.comment_id;
        END IF;
        RETURN OLD;
    ELSE
        IF NEW.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions WHERE comment_id = NEW.comment_id
            ) WHERE id = NEW.comment_id;
        END IF;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Recreate original triggers (no UPDATE trigger)
CREATE TRIGGER reactions_like_count_insert
    AFTER INSERT ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_delete
    AFTER DELETE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();
