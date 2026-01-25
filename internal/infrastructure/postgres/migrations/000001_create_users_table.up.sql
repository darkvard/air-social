CREATE TABLE
    users (
        id BIGSERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        username VARCHAR(50) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        -- Profile  
        full_name VARCHAR(100) NOT NULL DEFAULT '',
        bio VARCHAR(255) NOT NULL DEFAULT '',
        avatar VARCHAR(255) NOT NULL DEFAULT '',
        cover_image VARCHAR(255) NOT NULL DEFAULT '',
        location VARCHAR(100) NOT NULL DEFAULT '',
        website VARCHAR(255) NOT NULL DEFAULT '',
        -- System Info 
        verified BOOLEAN NOT NULL DEFAULT FALSE,
        verified_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        version INT NOT NULL DEFAULT 1
    );