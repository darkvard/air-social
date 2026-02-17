CREATE TABLE
    IF NOT EXISTS comments (
        id BIGSERIAL PRIMARY KEY,
        post_id BIGINT NOT NULL,
        user_id BIGINT NOT NULL,
        parent_id BIGINT,
        content TEXT NOT NULL,
        media JSONB NOT NULL DEFAULT '[]'::jsonb,
        likes_count INT NOT NULL DEFAULT 0,
        version INT NOT NULL DEFAULT 1,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ,
        FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (parent_id) REFERENCES comments (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_comments_post_id_id ON comments (post_id, id DESC);