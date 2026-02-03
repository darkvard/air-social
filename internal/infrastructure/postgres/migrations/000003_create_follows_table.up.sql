CREATE TABLE
    IF NOT EXISTS follows (
        follower_id BIGINT NOT NULL,
        followee_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        PRIMARY KEY (follower_id, followee_id),
        FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (followee_id) REFERENCES users (id) ON DELETE CASCADE,
        CONSTRAINT no_self_follow CHECK (follower_id <> followee_id)
    );

CREATE INDEX IF NOT EXISTS idx_follows_followee_id ON follows (followee_id);