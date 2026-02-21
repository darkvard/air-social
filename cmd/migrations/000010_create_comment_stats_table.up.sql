CREATE TABLE IF NOT EXISTS comment_stats (
    comment_id BIGINT PRIMARY KEY,
    likes_count INT NOT NULL DEFAULT 0,
    replies_count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);
