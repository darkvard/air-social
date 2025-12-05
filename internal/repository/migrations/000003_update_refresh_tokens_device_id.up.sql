ALTER TABLE refresh_tokens
ADD device_id TEXT NOT NULL DEFAULT 'unknown';
