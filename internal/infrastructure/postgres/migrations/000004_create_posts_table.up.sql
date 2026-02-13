CREATE TABLE
    IF NOT EXISTS posts (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL,
        content TEXT NOT NULL DEFAULT '',
        visibility VARCHAR(20) NOT NULL DEFAULT 'public',
        version INT NOT NULL DEFAULT 1,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ,
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_posts_user_id_id ON posts (user_id, id DESC);