CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) UNIQUE NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    is_blocked BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIMEZONE NOT NULL,
    created_at TIMESTAMP WITH TIMEZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIMEZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_refresh_token ON user_sessions (refresh_token);