CREATE TABLE
    IF NOT EXISTS posts (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL,
        content TEXT NOT NULL DEFAULT '',
        visibility VARCHAR(20) NOT NULL DEFAULT 'public',
        version INT NOT NULL DEFAULT 1,
        likes_count INT NOT NULL DEFAULT 0,
        comments_count INT NOT NULL DEFAULT 0,
        shares_count INT NOT NULL DEFAULT 0,
        original_post_id BIGINT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (original_post_id) REFERENCES posts (id) ON DELETE SET NULL
    );

CREATE INDEX IF NOT EXISTS idx_posts_user_id_id ON posts (user_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_original_post_id ON posts (original_post_id);