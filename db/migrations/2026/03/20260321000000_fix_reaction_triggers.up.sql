-- ============================================================
-- FIX REACTION TRIGGERS AFTER REVIEWS REMOVAL
-- ============================================================

-- Drop existing triggers
DROP TRIGGER IF EXISTS reactions_like_count_insert ON reactions;
DROP TRIGGER IF EXISTS reactions_like_count_delete ON reactions;

-- Replace the function to:
-- 1. Remove all review_id/reviews references (reviews table was dropped)
-- 2. Only count 'LIKE' reactions (column is called like_count, not reaction_count)
-- 3. Handle all operations (INSERT, UPDATE, DELETE)
CREATE OR REPLACE FUNCTION update_like_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = OLD.comment_id AND type = 'LIKE'
        ) WHERE id = OLD.comment_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle reaction type changes (e.g., LIKE -> LOVE)
        -- Update old comment if comment_id changed (unlikely but safe)
        IF OLD.comment_id IS DISTINCT FROM NEW.comment_id THEN
            UPDATE comments SET like_count = (
                SELECT COUNT(*) FROM reactions
                WHERE comment_id = OLD.comment_id AND type = 'LIKE'
            ) WHERE id = OLD.comment_id;
        END IF;
        -- Update new/current comment
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = NEW.comment_id AND type = 'LIKE'
        ) WHERE id = NEW.comment_id;
        RETURN NEW;
    ELSE
        -- INSERT
        UPDATE comments SET like_count = (
            SELECT COUNT(*) FROM reactions
            WHERE comment_id = NEW.comment_id AND type = 'LIKE'
        ) WHERE id = NEW.comment_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for INSERT, UPDATE, and DELETE
CREATE TRIGGER reactions_like_count_insert
    AFTER INSERT ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_update
    AFTER UPDATE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();

CREATE TRIGGER reactions_like_count_delete
    AFTER DELETE ON reactions
    FOR EACH ROW EXECUTE FUNCTION update_like_counts();
