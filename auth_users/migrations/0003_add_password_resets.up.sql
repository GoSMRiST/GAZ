CREATE TABLE IF NOT EXISTS password_resets (
                                               id         SERIAL PRIMARY KEY,
                                               user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);