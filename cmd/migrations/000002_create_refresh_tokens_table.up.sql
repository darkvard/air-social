CREATE TABLE
    refresh_tokens (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        token_hash VARCHAR(512) NOT NULL,
        device_id VARCHAR(255) NOT NULL DEFAULT 'unknown',
        expires_at TIMESTAMPTZ NOT NULL,
        revoked_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );