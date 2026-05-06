CREATE TABLE users (
    id SERIAL PRIMARY KEY,

    nickname TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,

    birth_date DATE NOT NULL,
    gender TEXT NOT NULL CHECK (gender IN ('man', 'woman')),

    avatar_url TEXT DEFAULT '/static/avatar/default.png',

    meetings_count INT DEFAULT 0,

    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT NOW()
);