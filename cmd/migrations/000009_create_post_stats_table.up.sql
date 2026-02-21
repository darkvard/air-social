CREATE TABLE IF NOT EXISTS post_stats (
    post_id BIGINT PRIMARY KEY,
    likes_count INT NOT NULL DEFAULT 0,
    comments_count INT NOT NULL DEFAULT 0,
    shares_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
