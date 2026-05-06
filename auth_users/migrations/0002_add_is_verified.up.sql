CREATE TABLE IF NOT EXISTS email_verifications (
    id SERIAL PRIMARY KEY,

    email TEXT NOT NULL UNIQUE,
    nickname TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    birth_date DATE NOT NULL,
    gender TEXT NOT NULL,

    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    created_at TIMESTAMP DEFAULT now()
);