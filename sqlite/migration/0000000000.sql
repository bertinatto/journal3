CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY,
    api_key TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- ALTER TABLE user
-- ADD COLUMN password TEXT NOT NULL;

-- TODO: posts -> post
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY,
    permalink TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS now (
    id INTEGER PRIMARY KEY,
    content TEXT NOT NULL,
    location TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS about (
    id INTEGER PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
