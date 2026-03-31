CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING GIN(username gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_users_full_name_trgm ON users USING GIN(full_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_posts_content_fts ON posts USING GIN(to_tsvector('simple', content)) WHERE deleted_at IS NULL;