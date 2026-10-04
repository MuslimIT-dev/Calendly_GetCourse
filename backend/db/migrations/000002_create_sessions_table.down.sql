DROP TRIGGER IF EXISTS trg_user_sessions_updated_at ON user_sessions;

DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_user_active;

DROP TABLE IF EXISTS user_sessions CASCADE;