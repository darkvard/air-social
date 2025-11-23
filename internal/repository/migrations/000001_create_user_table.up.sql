CREATE TABLE
    users (
        id SERIAL PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        username TEXT NOT NULL,
        password_hash TEXT NOT NULL,
        profile JSONB DEFAULT '{}'::JSONB,
        created_at TIMESTAMPTZ(0) DEFAULT NOW(),
        updated_at TIMESTAMPTZ(0) DEFAULT NOW(),
        version INT DEFAULT 1
    );