CREATE INDEX comments_user_created_active_idx
    ON comments (user_id, created_at DESC)
    WHERE deleted_at IS NULL;
