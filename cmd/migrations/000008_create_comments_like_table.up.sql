CREATE TABLE
    IF NOT EXISTS comment_likes (
        comment_id BIGINT NOT NULL,
        user_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        PRIMARY KEY (comment_id, user_id),
        FOREIGN KEY (comment_id) REFERENCES comments (id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );