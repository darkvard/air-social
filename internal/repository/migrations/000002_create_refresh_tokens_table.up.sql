CREATE TABLE refresh_tokens (
	id SERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id),
	token_hash TEXT NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	revoked_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT NOW()
);
